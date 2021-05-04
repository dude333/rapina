package parsers

/*
	TODO:
	https://query1.finance.yahoo.com/v7/finance/download/RBVA11.SA?period1=1588395063&period2=1619931063&interval=1d&events=history&includeAdjustedClose=true
	https://query1.finance.yahoo.com/v7/finance/download/BBPO11.SA?period1=1619654400&period2=1619740800&interval=1d&events=history&includeAdjustedClose=true
*/

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/dude333/rapina"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type stockQuote struct {
	Stock  string
	Date   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type StockParser struct {
	db   *sql.DB
	stmt *sql.Stmt
	log  rapina.Logger
	mu   sync.Mutex // ensures atomic writes to db
}

func NewStock(db *sql.DB, log rapina.Logger) *StockParser {
	s := &StockParser{db: db, log: log}
	return s
}

//
// Quote returns the quote from DB.
//
func (s *StockParser) Quote(code, date string) (float64, error) {
	query := `SELECT close FROM stock_quotes WHERE stock=$1 AND date=$2;`
	var close float64
	err := s.db.QueryRow(query, code, date).Scan(&close)
	if err == sql.ErrNoRows {
		return 0, errors.New("não encontrado no bd")
	}
	if err != nil {
		return 0, errors.Wrapf(err, "lendo cotação de %s do bd", code)
	}

	return close, nil
}

func (s *StockParser) SaveB3Quotes(filename string) error {
	isNew, err := isNewFile(s.db, filename)
	if !isNew && err == nil { // if error, process file
		s.log.Info("%s já processado anteriormente", filename)
		return errors.New("este arquivo de cotações já foi importado anteriormente")
	}

	if err := s.populateStockQuotes(filename); err != nil {
		return err
	}

	storeFile(s.db, filename)

	return nil
}

func (s *StockParser) populateStockQuotes(filename string) error {
	fh, err := os.Open(filename)
	if err != nil {
		return errors.Wrapf(err, "abrindo arquivo %s", filename)
	}
	defer fh.Close()

	// BEGIN TRANSACTION
	tx, err := s.db.Begin()
	if err != nil {
		return errors.Wrap(err, "Failed to begin transaction")
	}

	dec := transform.NewReader(fh, charmap.ISO8859_1.NewDecoder())
	scanner := bufio.NewScanner(dec)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		q, err := parseB3(line)
		if err != nil {
			continue // ignore line
		}
		fmt.Printf("%+v\n", q)
	}

	// END TRANSACTION
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "Failed to commit transaction")
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrapf(err, "lendo arquivo %s", filename)
	}

	return nil
}

// parseB3 parses the line based on this layout:
// http://www.b3.com.br/data/files/33/67/B9/50/D84057102C784E47AC094EA8/SeriesHistoricas_Layout.pdf
//
//   CAMPO/CONTEÚDO  TIPO E TAMANHO  POS. INIC.	 POS. FINAL
//   TIPREG “01”     N(02)           01          02
//   DATA “AAAAMMDD” N(08)           03          10
//   CODBDI          X(02)           11          12
//   CODNEG          X(12)           13          24
//   TPMERC          N(03)           25          27
//   PREABE          (11)V99         57          69
//   PREMAX          (11)V99         70          82
//   PREMIN          (11)V99         83          95
//   PREULT          (11)V99         109         121
//   QUATOT          N18             153         170
//   VOLTOT          (16)V99         171         188
//
// CODBDI:
//   02 LOTE PADRÃO
//   12 FUNDO IMOBILIÁRIO
//
// TPMERC:
//   010 VISTA
//   020 FRACIONÁRIO
func parseB3(line string) (*stockQuote, error) {
	if len(line) != 245 {
		return nil, errors.New("linha deve conter 245 bytes")
	}

	recType := line[0:2]
	if recType != "01" {
		return nil, fmt.Errorf("registro %s ignorado", recType)
	}

	codBDI := line[10:12]
	if codBDI != "02" && codBDI != "12" {
		return nil, fmt.Errorf("BDI %s ignorado", codBDI)
	}

	tpMerc := line[24:27]
	if tpMerc != "010" && tpMerc != "020" {
		return nil, fmt.Errorf("tipo de mercado %s ignorado", tpMerc)
	}

	date := line[2:6] + "-" + line[6:8] + "-" + line[8:10]
	code := strings.TrimSpace(line[12:24])

	numRanges := [5]struct {
		i, f int
	}{
		{56, 69},   // PREABE = open
		{69, 82},   // PREMAX = high
		{82, 95},   // PREMIN = low
		{108, 121}, // PREULT = close
		{170, 188}, // VOLTOT = volume
	}
	var vals [5]int
	for i, r := range numRanges {
		num, err := strconv.Atoi(line[r.i:r.f])
		if err != nil {
			return nil, err
		}
		vals[i] = num
	}

	return &stockQuote{
		Stock:  code,
		Date:   date,
		Open:   float64(vals[0]) / 100,
		High:   float64(vals[1]) / 100,
		Low:    float64(vals[2]) / 100,
		Close:  float64(vals[3]) / 100,
		Volume: float64(vals[4]) / 100,
	}, nil
}

//
// Save parses the 'stream', get the 'code' stock quotes and
// store it on 'db'. Returns the number of registers saved.
//
func (s *StockParser) Save(stream io.Reader, code string) (int, error) {
	if s.db == nil {
		return 0, errors.New("bd inválido")
	}
	if stream == nil {
		return 0, errors.New("sem dados")
	}

	if err := s.open(); err != nil {
		return 0, err
	}
	defer s.close()

	// Read stream, line by line
	var count, prov int
	isHeader := true
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()

		if isHeader {
			prov = provider(line)
			isHeader = false
			continue
		}

		var q *stockQuote
		var err error
		switch prov {
		case b3:
			q, err = parseB3(line)
		case yahoo:
			q, err = parseYahoo(line, code)
		case alphaVantage:
			q, err = parseAlphaVantage(line, code)
		}
		if err != nil {
			continue // ignore lines with error
		}

		err = s.store(q)
		if err == nil {
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		return count, err
	}

	return count, nil
}

// open prepares the insert statement.
func (s *StockParser) open() error {
	var err error
	insert := `INSERT OR IGNORE INTO stock_quotes 
	(stock, date, open, high, low, close, volume) VALUES (?,?,?,?,?,?,?);`

	s.stmt, err = s.db.Prepare(insert)
	if err != nil || s.stmt == nil {
		return errors.Wrap(err, "insert on stock_quotes")
	}

	return nil
}

// store stores the data using the insert statement.
func (s *StockParser) store(q *stockQuote) error {
	if s.stmt == nil {
		return errors.New("sql statement not initalized")
	}

	s.mu.Lock()

	res, err := s.stmt.Exec(
		q.Stock,
		q.Date,
		q.Open,
		q.High,
		q.Low,
		q.Close,
		q.Volume,
	)

	s.mu.Unlock()

	if err != nil {
		return errors.Wrap(err, "salvando cotação")
	}

	n, err := res.RowsAffected()
	if n == 0 || err != nil {
		return errors.New("registro não salvo (duplicado)")
	}

	return nil
}

// close closes the insert statement.
func (s *StockParser) close() error {
	var err error
	if s.stmt != nil {
		err = s.stmt.Close()
	}
	return err
}

// API providers.
const (
	none int = iota
	alphaVantage
	yahoo
	b3
)

// provider returns stream type based on header
func provider(header string) int {
	if header == "timestamp,open,high,low,close,volume" {
		return alphaVantage
	}
	if header == "Date,Open,High,Low,Close,Adj Close,Volume" {
		return yahoo
	}
	if strings.HasPrefix(header, "00COTAHIST.") {
		return b3
	}
	return none
}

// parseAlphaVantage parses lines downloaded from Alpha Vantage API server
// and returns *stockQuote for 'code'.
func parseAlphaVantage(line, code string) (*stockQuote, error) {
	fields := strings.Split(line, ",")
	if len(fields) != 6 {
		return nil, errors.New("linha inválida") // ignore lines with error
	}

	// Columns: timestamp,open,high,low,close,volume
	var err error
	var floats [5]float64
	for i := 1; i <= 5; i++ {
		floats[i-1], err = strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return nil, errors.Wrap(err, "campo inválido")
		}
	}

	return &stockQuote{
		Stock:  code,
		Date:   fields[0],
		Open:   floats[0],
		High:   floats[1],
		Low:    floats[2],
		Close:  floats[3],
		Volume: floats[4],
	}, nil
}

// parseYahoo parses lines downloaded from Yahoo Finance API server
// and returns *stockQuote for 'code'.
func parseYahoo(line, code string) (*stockQuote, error) {
	fields := strings.Split(line, ",")
	if len(fields) != 7 {
		return nil, errors.New("linha inválida") // ignore lines with error
	}

	// Columns: Date,Open,High,Low,Close,Adj Close,Volume
	var err error
	var floats [6]float64
	for i := 1; i <= 6; i++ {
		floats[i-1], err = strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return nil, errors.Wrap(err, "campo inválido")
		}
	}

	return &stockQuote{
		Stock:  code,
		Date:   fields[0],
		Open:   floats[0],
		High:   floats[1],
		Low:    floats[2],
		Close:  floats[3],
		Volume: floats[5],
	}, nil
}

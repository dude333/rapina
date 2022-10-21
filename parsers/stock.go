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
	"github.com/dude333/rapina/progress"
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

type stockCode struct {
	TckrSymb      string // Code
	SgmtNm        string // value: CASH
	SctyCtgyNm    string // values: SHARES, UNIT, FUNDS
	CrpnNm        string // Company name
	SpcfctnCd     string // values: ON, ON NM, PN N2, etc.
	CorpGovnLvlNm string // values: NOVO MERCADO, NIVEL 2, etc.
}

type StockParser struct {
	db  *sql.DB
	log rapina.Logger
}

//
// NewStock creates the required tables, if necessary, and returns a StockParser instance.
//
func NewStock(db *sql.DB, log rapina.Logger) (*StockParser, error) {
	for _, t := range []string{"status", "stock_quotes", "stock_codes"} {
		if err := createTable(db, t); err != nil {
			return nil, err
		}
	}

	s := &StockParser{db: db, log: log}
	return s, nil
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

//
// Quote returns the company ON stock code, where stockType is:
// ON, PN, UNT, CI [CI = FII]
//
func (s *StockParser) Code(companyName, stockType string) (string, error) {
	query := `SELECT trading_code FROM stock_codes WHERE company_name LIKE ? AND SpcfctnCd LIKE ?;`
	st := strings.ToUpper(stockType + "%")
	var code string
	err := s.db.QueryRow(query, "%"+companyName+"%", st).Scan(&code)
	if err == sql.ErrNoRows {
		return "", errors.New("não encontrado no bd")
	}
	if err != nil {
		return "", errors.Wrapf(err, "lendo código de %s do bd", companyName)
	}

	return code, nil
}

func (s *StockParser) SaveB3Quotes(filename string) error {
	isNew, err := isNewFile(s.db, filename)
	if !isNew && err == nil { // if error, process file
		progress.Warning("%s já processado anteriormente", filename)
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
		q, err := parseB3Quote(line)
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

	scanner := bufio.NewScanner(stream)

	// Read 1st line
	scanner.Scan()
	prov := provider(scanner.Text())

	var r rec
	if err := r.open(s.db, prov); err != nil {
		return 0, err
	}
	defer r.close()

	// Read stream, line by line
	var count int
	for scanner.Scan() {
		line := scanner.Text()

		var q *stockQuote
		var c *stockCode
		var err error
		switch prov {
		case b3Quotes:
			q, err = parseB3Quote(line)
		case yahoo:
			q, err = parseYahoo(line, code)
		case alphaVantage:
			q, err = parseAlphaVantage(line, code)
		case b3Codes:
			c, err = parseB3Code(line)
		}
		if err != nil {
			continue // ignore lines with error
		}

		if q != nil {
			err = r.storeQuote(q)
			if err == nil {
				count++
			}
		}
		if c != nil {
			err = r.storeCode(c)
			if err == nil {
				count++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return count, err
	}

	return count, nil
}

// open prepares the insert statement.
func (s *rec) open(db *sql.DB, provider int) error {
	var err error
	insert := `INSERT OR IGNORE INTO stock_quotes 
	(stock, date, open, high, low, close, volume) VALUES (?,?,?,?,?,?,?);`

	if provider == b3Codes {
		insert = `INSERT OR IGNORE INTO stock_codes 
	(trading_code, company_name, SpcfctnCd, CorpGovnLvlNm) VALUES (?,?,?,?);`
	}

	s.stmt, err = db.Prepare(insert)
	if err != nil || s.stmt == nil {
		return errors.Wrap(err, "insert on db")
	}

	return nil
}

// storeQuote stores the data using the insert statement.
func (s *rec) storeQuote(q *stockQuote) error {
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

// storeQuote stores the data using the insert statement.
func (s *rec) storeCode(c *stockCode) error {
	if s.stmt == nil {
		return errors.New("sql statement not initalized")
	}

	s.mu.Lock()

	res, err := s.stmt.Exec(
		c.TckrSymb, // trading_code
		c.CrpnNm,   // company_name
		c.SpcfctnCd,
		c.CorpGovnLvlNm,
	)

	s.mu.Unlock()

	if err != nil {
		return errors.Wrap(err, "salvando códigos")
	}

	n, err := res.RowsAffected()
	if n == 0 || err != nil {
		return errors.New("registro não salvo (duplicado)")
	}

	return nil
}

// close closes the insert statement.
func (s *rec) close() error {
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
	b3Quotes
	b3Codes
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
		return b3Quotes
	}
	if strings.HasPrefix(header, "RptDt;TckrSymb;Asst;AsstDesc;SgmtNm;MktNm;SctyCtgyNm;XprtnDt;") {
		return b3Codes
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

// parseB3Quote parses the line based on this layout:
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
func parseB3Quote(line string) (*stockQuote, error) {
	if len(line) != 245 {
		return nil, errors.New("linha deve conter 245 bytes")
	}

	recType := line[0:2]
	if recType != "01" {
		return nil, fmt.Errorf("registro %s ignorado", recType)
	}

	codBDI := line[10:12]
	if codBDI != "02" && codBDI != "12" && codBDI != "13" && codBDI != "14" {
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

type rec struct {
	stmt *sql.Stmt
	mu   sync.Mutex // ensures atomic writes to db
}

// parseB3Code parses lines downloaded from B3 server
// and returns *stockCode.
//
func parseB3Code(line string) (*stockCode, error) {
	fields := strings.Split(line, ";")

	// Columns:
	// RptDt;TckrSymb(2);Asst;AsstDesc;SgmtNm(5);MktNm;SctyCtgyNm(7);XprtnDt;XprtnCd;
	// TradgStartDt;TradgEndDt;BaseCd;ConvsCritNm;MtrtyDtTrgtPt;ReqrdConvsInd;
	// ISIN;CFICd;DlvryNtceStartDt;DlvryNtceEndDt;OptnTp;CtrctMltplr;AsstQtnQty;
	// AllcnRndLot;TradgCcy;DlvryTpNm;WdrwlDays;WrkgDays;ClnrDays;RlvrBasePricNm;
	// OpngFutrPosDay;SdTpCd1;UndrlygTckrSymb1;SdTpCd2;UndrlygTckrSymb2;
	// PureGoldWght;ExrcPric;OptnStyle;ValTpNm;PrmUpfrntInd;OpngPosLmtDt;
	// DstrbtnId;PricFctr;DaysToSttlm;SrsTpNm;PrtcnFlg;AutomtcExrcInd;SpcfctnCd(47);
	// CrpnNm(48);CorpActnStartDt;CtdyTrtmntTpNm;MktCptlstn;CorpGovnLvlNm(52)
	if len(fields) != 52 {
		return nil, fmt.Errorf("linha inválida %d", len(fields)) // ignore lines with error
	}

	s := stockCode{
		TckrSymb:      fields[1],
		SgmtNm:        fields[4],
		SctyCtgyNm:    fields[6],
		CrpnNm:        fields[47],
		SpcfctnCd:     fields[46],
		CorpGovnLvlNm: fields[51],
	}

	if s.SgmtNm != "CASH" ||
		(s.SctyCtgyNm != "SHARES" &&
			s.SctyCtgyNm != "FUNDS" &&
			s.SctyCtgyNm != "UNIT") {
		return nil, errors.New("linha ignorada")
	}

	return &s, nil
}

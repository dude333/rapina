package parsers

/*
	TODO:
	https://query1.finance.yahoo.com/v7/finance/download/RBVA11.SA?period1=1588395063&period2=1619931063&interval=1d&events=history&includeAdjustedClose=true
	https://query1.finance.yahoo.com/v7/finance/download/BBPO11.SA?period1=1619654400&period2=1619740800&interval=1d&events=history&includeAdjustedClose=true
*/

import (
	"bufio"
	"database/sql"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/dude333/rapina"
	"github.com/pkg/errors"
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

//
// Save parses the 'stream', get the 'code' stock quotes and
// store it on 'db'. Returns the number of registers saved.
//
func (s *StockParser) Save(stream io.ReadCloser, code string) (int, error) {
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
		case alphaVantage:
			q, err = parseAlphaVantage(line, code)
		case yahoo:
			q, err = parseYahoo(line, code)
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
)

// provider returns stream type based on header
func provider(header string) int {
	if header == "timestamp,open,high,low,close,volume" {
		return alphaVantage
	}
	if header == "Date,Open,High,Low,Close,Adj Close,Volume" {
		return yahoo
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

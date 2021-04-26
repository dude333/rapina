package parsers

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

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

type StockStore struct {
	db   *sql.DB
	stmt *sql.Stmt
}

func NewStockStore(db *sql.DB) (*StockStore, error) {
	s := &StockStore{db: db}
	return s, nil
}

//
// CsvToDB parses the 'stream', get the 'code' stock quotes and
// store it on 'db'.
//
func (s StockStore) CsvToDB(stream io.ReadCloser, code string) error {
	if s.db == nil {
		return errors.New("invalid db")
	}
	if stream == nil {
		return errors.New("empty stream")
	}

	if err := s.open(); err != nil {
		return err
	}
	defer s.close()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")

		var err error
		var floats [5]float64
		for i := 1; i <= 5; i++ {
			floats[i-1], err = strconv.ParseFloat(fields[i], 64)
			if err != nil {
				break
			}
		}
		if err != nil {
			continue
		}

		_ = s.store(&stockQuote{
			Stock:  code,
			Date:   fields[0],
			Open:   floats[0],
			High:   floats[1],
			Low:    floats[2],
			Close:  floats[3],
			Volume: floats[4],
		})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}

// open prepares the insert statement.
func (s *StockStore) open() error {
	if s.db == nil {
		return fmt.Errorf("db not provided")
	}
	if err := createTable(s.db, "stock_quotes"); err != nil {
		return err
	}

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
func (s StockStore) store(q *stockQuote) error {
	if s.stmt == nil {
		return errors.New("sql statement not initalized")
	}

	_, err := s.stmt.Exec(
		q.Stock,
		q.Date,
		q.Open,
		q.High,
		q.Low,
		q.Close,
		q.Volume,
	)
	if err != nil {
		return errors.Wrap(err, "inserting stock quote")
	}

	return nil
}

// close closes the insert statement.
func (s StockStore) close() error {
	var err error
	if s.stmt != nil {
		err = s.stmt.Close()
	}
	return err
}

func (s StockStore) Quote(code, date string) (rapina.Quotation, error) {
	return rapina.Quotation{
		Code: code,
		Date: date,
		Val:  123.45,
	}, nil
}

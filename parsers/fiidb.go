package parsers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/dude333/rapina"
	"github.com/dude333/rapina/progress"
	"github.com/pkg/errors"
)

// Error codes
var (
	ErrDBUnset  = errors.New("database not set")
	ErrNotFound = errors.New("not found")
)

// FIIParser implements sqlite storage for a rapina.FIIParser object.
type FIIParser struct {
	db  *sql.DB
	log rapina.Logger
	mu  sync.Mutex // ensures atomic writes on db
}

// NewFII creates a new instace of FII.
func NewFII(db *sql.DB, log rapina.Logger) (*FIIParser, error) {
	err := createAllTables(db)
	return &FIIParser{
		db:  db,
		log: log,
	}, err
}

// StoreFIIDetails parses the stream data into FIIDetails and returns
// the *FIIDetails.
func (fii *FIIParser) SaveDetails(stream []byte) error {
	fii.mu.Lock()
	defer fii.mu.Unlock()

	if !hasTable(fii.db, "fii_details") {
		if err := createTable(fii.db, "fii_details"); err != nil {
			return err
		}
	}

	var fiiDetails rapina.FIIDetails
	if err := json.Unmarshal(stream, &fiiDetails); err != nil {
		return errors.Wrap(err, "json unmarshal")
	}

	trimFIIDetails(&fiiDetails)

	x := fiiDetails.DetailFund
	if x.CNPJ == "" {
		return errors.New("CNPJ não encontrado")
	}

	insert := `INSERT OR IGNORE INTO fii_details 
		(cnpj, acronym, trading_code, json) 
		VALUES (?,?,?,?);`
	_, err := fii.db.Exec(insert,
		x.CNPJ, x.Acronym, x.TradingCode, stream)

	return err
}

// Details returns the FII Details for the 'code' or
// an empty string if not found in the db.
func (fii *FIIParser) Details(code string) (*rapina.FIIDetails, error) {
	details := rapina.FIIDetails{}

	if fii.db == nil {
		return nil, ErrDBUnset
	}

	var query string
	if len(code) == 4 {
		query = `SELECT json FROM fii_details WHERE acronym=?`
	} else if len(code) == 6 {
		query = `SELECT json FROM fii_details WHERE trading_code=?`
	} else {
		return nil, fmt.Errorf("invalid code '%s'", code)
	}

	var jsonStr []byte
	row := fii.db.QueryRow(query, code)
	err := row.Scan(&jsonStr)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonStr, &details); err != nil {
		progress.ErrorMsg("FII details [%v]: %s\n", err, string(jsonStr))
		return nil, errors.Wrap(err, "json unmarshal")
	}

	return &details, nil
}

// Dividends returns the dividend from the db.
func (fii *FIIParser) Dividends(code, monthYear string) (*[]rapina.Dividend, error) {
	fii.mu.Lock()
	defer fii.mu.Unlock()

	const s = `SELECT trading_code, base_date, value
	FROM fii_dividends 
	WHERE trading_code=$1 
	AND base_date LIKE $2;`
	rows, err := fii.db.Query(s, code, monthYear+"%")
	if err != nil {
		return nil, errors.Wrap(err, "lendo dividendos do bd")
	}
	defer rows.Close()

	dividends := []rapina.Dividend{}
	var (
		tradingCode, baseDate string
		value                 float64
	)
	for rows.Next() {
		err := rows.Scan(&tradingCode, &baseDate, &value)
		if err != nil {
			return nil, err
		}

		// fii.log.Debug("reading: %v %v %v", tradingCode, baseDate, value)

		dividends = append(dividends, rapina.Dividend{
			Code: tradingCode,
			Date: baseDate,
			Val:  value,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(dividends) == 0 {
		return nil, errors.New("dividendos não encontrados")
	}

	return &dividends, nil
}

// SaveDividend parses and stores the map in the db. Returns the parsed stream.
func (fii *FIIParser) SaveDividend(dividend rapina.Dividend) error {
	fii.mu.Lock()
	defer fii.mu.Unlock()

	if err := createTable(fii.db, "fii_dividends"); err != nil {
		return err
	}

	const insert = `INSERT OR IGNORE INTO fii_dividends 
	(trading_code, base_date, payment_date, value) VALUES (?,?,?,?)`
	_, err := fii.db.Exec(insert, dividend.Code, dividend.Date, dividend.PaymentDate, dividend.Val)

	return errors.Wrap(err, "inserting data on fii_dividends")
}

func (fii *FIIParser) SelectFIIDetails(code string) (*rapina.FIIDetails, error) {
	if fii.db == nil {
		return nil, ErrDBUnset
	}

	var query string
	if len(code) == 4 {
		query = `SELECT cnpj, acronym, trading_code FROM fii_details WHERE acronym=?`
	} else if len(code) == 6 {
		query = `SELECT cnpj, acronym, trading_code FROM fii_details WHERE trading_code=?`
	} else {
		return nil, fmt.Errorf("invalid code '%s'", code)
	}

	var cnpj, acronym, tradingCode string
	row := fii.db.QueryRow(query, code)
	err := row.Scan(&cnpj, &acronym, &tradingCode)
	if err != nil {
		return nil, err
	}

	var fiiDetail rapina.FIIDetails
	fiiDetail.DetailFund.CNPJ = cnpj
	fiiDetail.DetailFund.Acronym = acronym
	fiiDetail.DetailFund.TradingCode = tradingCode

	return &fiiDetail, nil
}

/* -------- Utils ----------- */

func trimFIIDetails(f *rapina.FIIDetails) {
	f.DetailFund.CNPJ = strings.TrimSpace(f.DetailFund.CNPJ)
	f.DetailFund.Acronym = strings.TrimSpace(f.DetailFund.Acronym)
	tradingCodes := strings.Split(
		strings.TrimSpace(f.DetailFund.TradingCode), " ")
	f.DetailFund.TradingCode = tradingCodes[0]
}

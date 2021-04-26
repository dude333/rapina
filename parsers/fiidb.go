package parsers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dude333/rapina"
	"github.com/pkg/errors"
)

// Error codes
var (
	ErrDBUnset  = errors.New("database not set")
	ErrNotFound = errors.New("not found")
)

type FIIStore struct {
	db *sql.DB
}

// NewFIIStore creates a new instace of FII.
func NewFIIStore(db *sql.DB) *FIIStore {
	fii := &FIIStore{
		db: db, // will accept null db when caching is no needed
	}
	return fii
}

//
// StoreFIIDetails parses the stream data into FIIDetails and returns
// the *FIIDetails.
//
func (fii FIIStore) StoreFIIDetails(stream []byte) error {
	if fii.db == nil {
		return ErrDBUnset
	}

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
		return fmt.Errorf("wrong CNPJ: %s", x.CNPJ)
	}

	insert := "INSERT OR IGNORE INTO fii_details (cnpj, acronym, trading_code) VALUES (?,?,?)"
	_, err := fii.db.Exec(insert, x.CNPJ, x.Acronym, x.TradingCode)

	return err
}

//
// CNPJ returns the FII CNPJ for the 'code' or
// an empty string if not found in the db.
//
func (fii FIIStore) CNPJ(code string) (string, error) {
	if fii.db == nil {
		return "", ErrDBUnset
	}

	var query string
	if len(code) == 4 {
		query = `SELECT cnpj FROM fii_details WHERE acronym=?`
	} else if len(code) == 6 {
		query = `SELECT cnpj FROM fii_details WHERE trading_code=?`
	} else {
		return "", fmt.Errorf("invalid code '%s'", code)
	}

	var cnpj string
	row := fii.db.QueryRow(query, code)
	err := row.Scan(&cnpj)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	return cnpj, nil
}

//
// StoreFIIDividends stores the map into the db.
//
func (fii FIIStore) StoreFIIDividends(stream map[string]string) error {
	// fmt.Println("----------------------------")
	// fmt.Printf("%+v\n\n", stream)

	if err := createTable(fii.db, "fii_dividends"); err != nil {
		return err
	}

	code := mapFinder("Código de negociação da cota", stream)
	baseDate := mapFinder("Data-base", stream)
	pymtDate := mapFinder("Data do pagamento", stream)
	val := mapFinder("Valor do provento por cota", stream)

	const insert = `INSERT OR IGNORE INTO fii_dividends 
	(trading_code, base_date, payment_date, value) VALUES (?,?,?,?)`
	_, err := fii.db.Exec(insert, code, baseDate, pymtDate, comma2dot(val))

	// fmt.Println(insert, code, baseDate, pymtDate, comma2dot(val))

	return errors.Wrap(err, "inserting data on fii_dividends")
}

func (fii FIIStore) SelectFIIDetails(code string) (*rapina.FIIDetails, error) {
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
	f.DetailFund.TradingCode = strings.TrimSpace(f.DetailFund.TradingCode)
}

func mapFinder(key string, m map[string]string) string {
	for k := range m {
		if strings.Contains(k, key) {
			return m[k]
		}
	}
	return ""
}

func comma2dot(val string) float64 {
	a := strings.ReplaceAll(val, ".", "")
	b := strings.ReplaceAll(a, ",", ".")
	n, _ := strconv.ParseFloat(b, 64)
	return n
}

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
// Dividend returns the dividend from the db.
//
func (fii FIIStore) Dividends(code, monthYear string) (*[]rapina.Dividend, error) {
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

		fmt.Println("[d reading]", tradingCode, baseDate, value)

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

//
// SaveDividend parses and stores the map in the db. Returns the parsed stream.
//
func (fii FIIStore) SaveDividend(stream map[string]string) (*rapina.Dividend, error) {
	// fmt.Println("----------------------------")
	// fmt.Printf("%+v\n\n", stream)

	if err := createTable(fii.db, "fii_dividends"); err != nil {
		return nil, err
	}

	code := mapFinder("Código de negociação da cota", stream)
	baseDate := fixDate(mapFinder("Data-base", stream))
	pymtDate := fixDate(mapFinder("Data do pagamento", stream))
	val := mapFinder("Valor do provento por cota", stream)
	fVal := comma2dot(val)

	const insert = `INSERT OR IGNORE INTO fii_dividends 
	(trading_code, base_date, payment_date, value) VALUES (?,?,?,?)`
	_, err := fii.db.Exec(insert, code, baseDate, pymtDate, fVal)

	// fmt.Println("[d saving]", code, baseDate, fVal)

	d := rapina.Dividend{
		Code: code,
		Date: baseDate,
		Val:  fVal,
	}

	return &d, errors.Wrap(err, "inserting data on fii_dividends")
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

// fixDate converts dates from DD/MM/YYYY to YYYY-MM-DD.
func fixDate(date string) string {
	if len(date) != len("26/04/2021") || strings.Count(date, "/") != 2 {
		return date
	}

	return date[6:10] + "-" + date[3:5] + "-" + date[0:2]
}

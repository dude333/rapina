package parsers

import (
	"fmt"

	"github.com/pkg/errors"
)

// Error codes
var (
	ErrDBUnset = errors.New("database not set")
)

func (fii FII) StoreFIIDetails(fiiDetails *FIIDetails) error {
	if fii.db == nil {
		return ErrDBUnset
	}

	if !hasTable(fii.db, "fii_details") {
		if err := createTable(fii.db, "fii_details"); err != nil {
			return err
		}
	}

	x := fiiDetails.DetailFund
	if x.CNPJ == "" {
		return fmt.Errorf("wrong CNPJ: %s", x.CNPJ)
	}

	insert := "INSERT OR IGNORE INTO fii_details (cnpj, acronym, trading_code) VALUES (?,?,?)"
	_, err := fii.db.Exec(insert, x.CNPJ, x.Acronym, x.TradingCode)

	return err
}

func (fii FII) SelectFIIDetails(code string) (*FIIDetails, error) {
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

	var fiiDetail FIIDetails
	fiiDetail.DetailFund.CNPJ = cnpj
	fiiDetail.DetailFund.Acronym = acronym
	fiiDetail.DetailFund.TradingCode = tradingCode

	return &fiiDetail, nil
}

package parsers

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

func StoreFIIDetail(db *sql.DB, fii *FII) error {
	if !hasTable(db, "fii_detail") {
		if err := createTable(db, "fii_detail"); err != nil {
			return err
		}
	}

	insert := "INSERT INTO fii_detail (cnpj, acronym, trading_code) VALUES (?,?,?)"
	x := fii.DetailFund
	_, err := db.Exec(insert, x.CNPJ, x.Acronym, x.TradingName)

	return err
}

func SelectFIIDetail(db *sql.DB, code string) (*FII, error) {
	var query string
	if len(code) == 4 {
		query = `SELECT cnpj, code, trading_code FROM fii_detail WHERE code=?`
	} else if len(code) == 6 {
		query = `SELECT cnpj, code, trading_code FROM fii_detail WHERE trading_code=?`
	} else {
		return nil, fmt.Errorf("invalid code '%s'", code)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "reading DB")
	}
	defer rows.Close()

	var cnpj, acronym, tradingCode string
	for rows.Next() {
		err = rows.Scan(&cnpj, &acronym, &tradingCode)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
	}

	var fii FII
	fii.DetailFund.CNPJ = cnpj
	fii.DetailFund.Acronym = acronym
	fii.DetailFund.TradingCode = tradingCode

	return &fii, err
}

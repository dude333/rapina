package rapina

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dude333/rapina/parsers"
)

//
// FIIDividends prints the dividends from 'code' fund for 'n' months,
// starting from latest.
//
func FIIDividends(code string, n int) error {
	db, err := openDatabase()
	if err != nil {
		return err
	}
	code = strings.ToUpper(code)

	fii, _ := parsers.NewFII(db)

	cnpj, err := cnpj(fii, code)
	if err != nil {
		return err
	}

	err = fii.FetchFIIDividends(cnpj, n)

	return err
}

//
// cnpj returns the CNPJ from FII code. It first checks the DB and, if not
// found, fetches from B3.
//
func cnpj(fii *parsers.FII, code string) (string, error) {
	fiiDetails, err := fii.SelectFIIDetails(code)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("[x] error", err)
	}
	if err == nil && fiiDetails.DetailFund.CNPJ != "" {
		fmt.Println("DB", code, fiiDetails.DetailFund.CNPJ)
		return fiiDetails.DetailFund.CNPJ, nil
	}
	//
	// Fetch online if DB fails
	fiiDetails, err = fii.FetchFIIDetails(code)
	if err != nil {
		return "", err
	}

	fmt.Println("online", code, fiiDetails.DetailFund.CNPJ)

	return fiiDetails.DetailFund.CNPJ, nil
}

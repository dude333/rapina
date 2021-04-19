package rapina

import (
	"database/sql"
	"fmt"

	"github.com/dude333/rapina/parsers"
)

func FIIDividends(code string) error {
	db, err := openDatabase()
	if err != nil {
		return err
	}

	fii, err := parsers.SelectFIIDetail(db, code)
	if err == nil && err != sql.ErrNoRows {
		fmt.Println("DB", code, fii.DetailFund.CNPJ)
		return nil
	}

	// Fetch online if DB fails
	u := "https://sistemaswebb3-listados.b3.com.br"
	fii, err = parsers.FetchFIIDetails(u, code)
	if err != nil {
		return err
	}

	fmt.Println("online", code, fii.DetailFund.CNPJ)

	return nil
}

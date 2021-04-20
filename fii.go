package rapina

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dude333/rapina/parsers"
)

func FIIDividends(code string) error {
	db, err := openDatabase()
	if err != nil {
		return err
	}
	u := "https://sistemaswebb3-listados.b3.com.br"
	code = strings.ToUpper(code)

	fii, _ := parsers.NewFII(db, u)

	fiiDetails, err := fii.SelectFIIDetails(code)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("[x] error", err)
	}
	if err == nil && fiiDetails.DetailFund.CNPJ != "" {
		fmt.Println("DB", code, fiiDetails.DetailFund.CNPJ)
		return nil
	}

	// Fetch online if DB fails
	fiiDetails, err = fii.FetchFIIDetails(code)
	if err != nil {
		return err
	}

	fmt.Println("online", code, fiiDetails.DetailFund.CNPJ)

	return nil
}

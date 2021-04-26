package reports

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dude333/rapina"
	"github.com/dude333/rapina/fetch"
)

type FIITerminalReport struct {
	db       *sql.DB
	store    rapina.StockStore
	fetchFII *fetch.FII
}

func NewFIITerminalReport(
	db *sql.DB,
	store rapina.StockStore,
	parser rapina.FIIParser) (*FIITerminalReport, error) {
	if db == nil {
		return nil, errors.New("invalid db")
	}
	if store == nil {
		return nil, errors.New("invalid store")
	}
	if parser == nil {
		return nil, errors.New("invalid parser")
	}

	return &FIITerminalReport{
		db:       db,
		store:    store,
		fetchFII: fetch.NewFII(parser),
	}, nil
}

func (t FIITerminalReport) Dividends(code string, n int) error {

	dividends, err := t.fetchFII.Dividends(code, n)
	if err != nil {
		return err
	}

	for _, d := range *dividends {
		fmt.Println("[>]", d)
	}

	// Quotation

	return nil
}

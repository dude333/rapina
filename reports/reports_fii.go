package reports

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/dude333/rapina/fetch"
	"github.com/dude333/rapina/parsers"
)

type FIITerminalReport struct {
	fetchFII   *fetch.FII
	fetchStock *fetch.StockFetch
}

func NewFIITerminalReport(db *sql.DB, apiKey string) (*FIITerminalReport, error) {
	store, err := parsers.NewStockStore(db)
	if err != nil {
		return nil, err
	}
	parser := parsers.NewFIIStore(db)
	if parser == nil {
		return nil, errors.New("invalid parser")
	}
	fetchStock, err := fetch.NewStockFetch(store, apiKey)
	if err != nil {
		return nil, errors.Wrap(err, "new StockFetch instance")
	}
	fetchFII := fetch.NewFII(parser)
	if fetchFII == nil {
		return nil, errors.New("invalid FII fetcher")
	}

	return &FIITerminalReport{
		fetchFII:   fetchFII,
		fetchStock: fetchStock,
	}, nil
}

func (t FIITerminalReport) Dividends(code string, n int) error {

	dividends, err := t.fetchFII.Dividends(code, n)
	if err != nil {
		return err
	}

	line := strings.Repeat("-", 55)
	fmt.Println(line)
	fmt.Println(code)
	fmt.Println(line)
	fmt.Println("  DATA           RENDIMENTO     COTAÇÃO       YELD  ")
	fmt.Println("  ----------     ----------     ----------    ------")

	for _, d := range *dividends {
		q, _ := t.fetchStock.Quote(code, d.Date)
		fmt.Printf("  %s %14.6f %14.6f %8.2f%%\n", d.Date, d.Val, q, 100*d.Val/q)
	}

	fmt.Println(line)
	return nil
}

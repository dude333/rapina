package reports

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/dude333/rapina/fetch"
	"github.com/dude333/rapina/parsers"
)

type FIITerminalReport struct {
	fetchFII   *fetch.FII
	fetchStock *fetch.StockFetch
}

func NewFIITerminalReport(db *sql.DB, apiKey string) (*FIITerminalReport, error) {
	log := NewLogger(os.Stderr)
	store := parsers.NewStockStore(db, log)
	parser := parsers.NewFIIStore(db, log)
	fetchStock := fetch.NewStockFetch(store, log, apiKey)
	fetchFII := fetch.NewFII(parser, log)

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
		fmt.Printf("  %s %14.6f %14.6f ", d.Date, d.Val, q)
		if q > 0 {
			fmt.Printf("%8.2f%%\n", 100*d.Val/q)
		}
	}

	fmt.Println(line)
	return nil
}

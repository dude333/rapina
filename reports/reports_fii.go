package reports

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/dude333/rapina/fetch"
	"github.com/dude333/rapina/parsers"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var line = strings.Repeat("-", 67)

type FIITerminalReport struct {
	fetchFII   *fetch.FII
	fetchStock *fetch.StockFetch
	log        *Logger
}

func NewFIITerminalReport(db *sql.DB, apiKey string) (*FIITerminalReport, error) {
	log := NewLogger(nil)
	store := parsers.NewStockStore(db, log)
	parser := parsers.NewFIIStore(db, log)
	fetchStock := fetch.NewStockFetch(store, log, apiKey)
	fetchFII := fetch.NewFII(parser, log)

	return &FIITerminalReport{
		fetchFII:   fetchFII,
		fetchStock: fetchStock,
		log:        log,
	}, nil
}

func (t FIITerminalReport) Dividends(codes []string, n int) error {

	for _, code := range codes {
		if len(code) != 6 {
			t.log.Error("Código inválido: %s", code)
			t.log.Info("Padrão experado: ABCD11")
			continue
		}

		err := t.PrintDividends(code, n)
		if err != nil {
			t.log.Error("%s", err)
		}
	}
	fmt.Println(line)

	return nil
}

func (t *FIITerminalReport) SetParms(parms map[string]string) {
	if _, ok := parms["verbose"]; ok {
		t.log.SetOut(os.Stderr)
	}
}

func (t FIITerminalReport) PrintDividends(code string, n int) error {
	dividends, err := t.fetchFII.Dividends(code, n)
	if err != nil {
		return err
	}

	fmt.Println(line)
	fmt.Println(code)
	fmt.Println(line)
	fmt.Println("  DATA COM       RENDIMENTO     COTAÇÃO       YELD      YELD a.a.")
	fmt.Println("  ----------     ----------     ----------    ------    ---------")

	p := message.NewPrinter(language.BrazilianPortuguese)

	for _, d := range *dividends {
		q, _ := t.fetchStock.Quote(code, d.Date)
		p.Printf("  %s     R$%8.2f     R$%8.2f ", d.Date, d.Val, q)
		if q > 0 {
			i := d.Val / q
			p.Printf("%8.2f%%    %8.2f%%\n", 100*i, 100*(math.Pow(1+i, 12)-1))
		}
	}

	return nil
}

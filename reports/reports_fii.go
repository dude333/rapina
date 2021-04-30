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

const (
	Rtable = iota + 1
	Rcsv
)

type FIITerminalReport struct {
	fetchFII   *fetch.FII
	fetchStock *fetch.StockFetch
	log        *Logger
	reportType int
}

func NewFIITerminalReport(db *sql.DB, apiKey string) (*FIITerminalReport, error) {
	log := NewLogger(nil)
	store := parsers.NewStockStore(db, log)
	parser, err := parsers.NewFIIStore(db, log)
	if err != nil {
		return nil, err
	}
	fetchStock := fetch.NewStockFetch(store, log, apiKey)
	fetchFII := fetch.NewFII(parser, log)

	return &FIITerminalReport{
		fetchFII:   fetchFII,
		fetchStock: fetchStock,
		log:        log,
		reportType: Rtable,
	}, nil
}

func (t *FIITerminalReport) SetParms(parms map[string]string) {
	if _, ok := parms["verbose"]; ok {
		t.log.SetOut(os.Stderr)
	}
	if r, ok := parms["type"]; ok {
		switch r {
		case "table", "tabela", "tab":
			t.reportType = Rtable
		case "csv":
			t.reportType = Rcsv
		}
	}
}

func (t FIITerminalReport) Dividends(codes []string, n int) error {

	// Header
	if t.reportType == Rcsv {
		fmt.Println("Código,Data Com,Rendimento,Cotação,Yeld,Yeld a.a.")
	}

	for _, code := range codes {
		if len(code) != 6 {
			t.log.Error("Código inválido: %s", code)
			t.log.Info("Padrão experado: ABCD11")
			continue
		}

		var err error
		switch t.reportType {
		case Rcsv:
			err = t.CsvDividends(code, n)
		default:
			err = t.PrintDividends(code, n)
		}
		if err != nil {
			t.log.Error("%s", err)
		}
	}

	// Footer
	if t.reportType == Rtable {
		fmt.Println(line)
	}

	return nil
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
		q, err := t.fetchStock.Quote(code, d.Date)
		if err != nil {
			return err
		}
		p.Printf("  %s     R$%8.2f     R$%8.2f ", d.Date, d.Val, q)
		if q > 0 {
			i := d.Val / q
			p.Printf("%8.2f%%    %8.2f%%", 100*i, 100*(math.Pow(1+i, 12)-1))
		}
		p.Println()
	}

	return nil
}

func (t FIITerminalReport) CsvDividends(code string, n int) error {
	dividends, err := t.fetchFII.Dividends(code, n)
	if err != nil {
		return err
	}

	p := message.NewPrinter(language.BrazilianPortuguese)
	for _, d := range *dividends {
		q, err := t.fetchStock.Quote(code, d.Date)
		if err != nil {
			return err
		}
		var i float64
		if q > 0 {
			i = d.Val / q
		}
		p.Printf(`%s,%s,"%f","%f",`, code, d.Date, d.Val, q)
		p.Printf(`"%f%%","%f%%"`, 100*i, 100*(math.Pow(1+i, 12)-1))
		p.Println()
	}

	return nil
}

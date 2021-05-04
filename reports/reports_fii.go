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

func NewFIITerminalReport(db *sql.DB, apiKey, dataDir string) (*FIITerminalReport, error) {
	log := NewLogger(nil)
	stockParser := parsers.NewStock(db, log)
	fiiParser, err := parsers.NewFIIStore(db, log)
	if err != nil {
		return nil, err
	}
	fetchStock := fetch.NewStockFetch(stockParser, log, apiKey, dataDir)
	fetchFII := fetch.NewFII(fiiParser, log)

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

	// var wg sync.WaitGroup

	for _, code := range codes {
		if len(code) != 6 {
			t.log.Error("Código inválido: %s", code)
			t.log.Info("Padrão experado: ABCD11")
			continue
		}

		// wg.Add(1)
		// go func(code string, n int) {
		var buf *strings.Builder
		var err error
		switch t.reportType {
		case Rcsv:
			buf, err = t.CsvDividends(code, n)
		default:
			buf, err = t.PrintDividends(code, n)
		}
		if err != nil {
			t.log.Error("%s", err)
		} else {
			fmt.Print(buf.String())
		}
		// wg.Done()
		// }(code, n)

	}

	// wg.Wait()

	// Footer
	if t.reportType == Rtable {
		fmt.Println(line)
	}

	return nil
}

func (t FIITerminalReport) PrintDividends(code string, n int) (*strings.Builder, error) {
	dividends, err := t.fetchFII.Dividends(code, n)
	if err != nil {
		return nil, err
	}

	buf := &strings.Builder{}
	p := message.NewPrinter(language.BrazilianPortuguese)

	p.Fprintln(buf, line)
	p.Fprintln(buf, code)
	p.Fprintln(buf, line)
	p.Fprintln(buf, "  DATA COM       RENDIMENTO     COTAÇÃO       YELD      YELD a.a.")
	p.Fprintln(buf, "  ----------     ----------     ----------    ------    ---------")

	for _, d := range *dividends {
		p.Fprintf(buf, "  %s     R$%8.2f     ", d.Date, d.Val)

		q, err := t.fetchStock.Quote(code, d.Date)
		if err != nil {
			t.log.Error("Cotação de %s (%s): %v", code, d.Date, err)
		}
		if q > 0 && err == nil {
			i := d.Val / q
			p.Fprintf(buf, "R$%8.2f %8.2f%%    %8.2f%%", q, 100*i, 100*(math.Pow(1+i, 12)-1))
		}
		p.Fprintf(buf, "\n")
	}

	return buf, nil
}

func (t FIITerminalReport) CsvDividends(code string, n int) (*strings.Builder, error) {
	dividends, err := t.fetchFII.Dividends(code, n)
	if err != nil {
		return nil, err
	}

	buf := &strings.Builder{}
	p := message.NewPrinter(language.BrazilianPortuguese)
	for _, d := range *dividends {
		p.Fprintf(buf, `%s,%s,"%f",`, code, d.Date, d.Val)

		q, err := t.fetchStock.Quote(code, d.Date)
		if err != nil {
			t.log.Error("Cotação de %s (%s): %v", code, d.Date, err)
		}
		if q > 0 && err == nil {
			i := d.Val / q
			p.Fprintf(buf, `"%f","%f%%","%f%%"\n`, q, 100*i, 100*(math.Pow(1+i, 12)-1))
		} else {
			p.Fprintf(buf, `"","",""\n`)
		}
	}

	return buf, nil
}

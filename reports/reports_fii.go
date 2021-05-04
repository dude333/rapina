package reports

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/dude333/rapina"
	"github.com/dude333/rapina/fetch"
	"github.com/dude333/rapina/parsers"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var line = strings.Repeat("-", 67)

// Type of report output
const (
	Rtable = iota + 1
	Rcsv
)

// FIITerminal implements reports related to FII funds on the terminal.
type FIITerminal struct {
	fetchFII   *fetch.FII
	fetchStock *fetch.StockFetch
	log        *Logger
	reportType int
}

// NewFIITerminal creates a new instace of a FIITerminal
func NewFIITerminal(db *sql.DB, apiKey, dataDir string) (*FIITerminal, error) {
	db.SetMaxOpenConns(1)
	log := NewLogger(nil)
	stockParser := parsers.NewStock(db, log)
	fiiParser, err := parsers.NewFII(db, log)
	if err != nil {
		return nil, err
	}
	fetchStock := fetch.NewStock(stockParser, log, apiKey, dataDir)
	fetchFII := fetch.NewFII(fiiParser, log)

	return &FIITerminal{
		fetchFII:   fetchFII,
		fetchStock: fetchStock,
		log:        log,
		reportType: Rtable,
	}, nil
}

// SetParms set the terminal reports parameters.
func (t *FIITerminal) SetParms(parms map[string]string) {
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

// Dividends prints the dividends report on terminal.
func (t FIITerminal) Dividends(codes []string, n int) error {

	// Header
	if t.reportType == Rcsv {
		fmt.Println("Código,Data Com,Rendimento,Cotação,Yeld,Yeld a.a.")
	}

	// Remove codes
	c := make([]string, 0, len(codes))
	for _, code := range codes {
		if len(code) == 6 {
			c = append(c, code)
		} else {
			t.log.Error("Código inválido: %s", code)
			t.log.Info("Padrão experado: ABCD11")
		}
	}
	codes = c

	dividends := make(map[string]*[]rapina.Dividend, len(codes))

	var wg sync.WaitGroup
	for _, code := range codes {
		wg.Add(1)
		go func(code string, n int) {
			div, err := t.fetchFII.Dividends(code, n)
			if err != nil {
				t.log.Error("%s dividendos: %v", code, err)
			}
			dividends[code] = div
			wg.Done()
		}(code, n)
	}
	wg.Wait()

	for _, code := range codes {
		if _, ok := dividends[code]; !ok {
			continue
		}
		var buf *strings.Builder
		var err error
		switch t.reportType {
		case Rcsv:
			buf, err = t.csvDividends(code, dividends[code])
		default:
			buf, err = t.printDividends(code, dividends[code])
		}
		if err != nil {
			t.log.Error("%s", err)
		} else {
			fmt.Print(buf)
		}
	}

	// Footer
	if t.reportType == Rtable {
		fmt.Println(line)
	}

	return nil
}

func (t FIITerminal) printDividends(code string, dividends *[]rapina.Dividend) (*strings.Builder, error) {
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
		buf.WriteByte('\n')
	}

	return buf, nil
}

func (t FIITerminal) csvDividends(code string, dividends *[]rapina.Dividend) (*strings.Builder, error) {
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
			p.Fprintf(buf, `"%f","%f%%","%f%%"`, q, 100*i, 100*(math.Pow(1+i, 12)-1))
		} else {
			buf.WriteString(`"","",""`)
		}
		buf.WriteByte('\n')
	}

	return buf, nil
}

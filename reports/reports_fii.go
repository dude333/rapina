package reports

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/dude333/rapina"
	"github.com/dude333/rapina/fetch"
	"github.com/dude333/rapina/progress"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var line = strings.Repeat("-", 67)

// Type of report output
const (
	Rtable = iota + 1
	Rcsv
	Rcsvrend
)

// FIITerminal implements reports related to FII funds on the terminal.
type FIITerminal struct {
	fetchFII     *fetch.FII
	fetchStock   *fetch.Stock
	reportFormat int
}

type FIITerminalOptions struct {
	APIKey, DataDir string
}

// NewFIITerminal creates a new instace of a FIITerminal
func NewFIITerminal(db *sql.DB, opts FIITerminalOptions) (*FIITerminal, error) {
	var log rapina.Logger

	fetchStock, err := fetch.NewStock(db, log, opts.APIKey, opts.DataDir)
	if err != nil {
		return nil, err
	}

	fetchFII, err := fetch.NewFII(db, log)
	if err != nil {
		return nil, err
	}

	return &FIITerminal{
		fetchFII:     fetchFII,
		fetchStock:   fetchStock,
		reportFormat: Rtable,
	}, nil
}

// SetParms set the terminal reports parameters.
func (t *FIITerminal) SetParms(parms map[string]string) {
	if _, ok := parms["verbose"]; ok {
		progress.SetDebug(true)
	}
	if r, ok := parms["format"]; ok {
		switch r {
		case "table", "tabela", "tab":
			t.reportFormat = Rtable
		case "csv":
			t.reportFormat = Rcsv
		case "csvrend":
			t.reportFormat = Rcsvrend
		}
	}
}

// Dividends prints the dividends report on terminal.
func (t FIITerminal) Dividends(codes []string, n int) error {
	// Header
	if t.reportFormat == Rcsv {
		fmt.Println("Código,Data Com,Rendimento,Cotação,Yeld,Yeld a.a.")
	}
	if t.reportFormat == Rcsvrend {
		fmt.Print(`Código/Data-Com`)
		for _, date := range revMonthsFromToday(n) {
			fmt.Printf(",%s", date)
		}
		fmt.Println()
	}

	// Remove codes
	c := make([]string, 0, len(codes))
	for _, code := range codes {
		if len(code) == 6 {
			c = append(c, code)
		} else {
			progress.ErrorMsg("Código inválido: %s. Padrão esperado: ABCD11.", code)
		}
	}
	codes = c

	dividends := sync.Map{}
	var wg sync.WaitGroup
	for i, code := range codes {
		wg.Add(1)
		i := i
		go func(code string, n int) {
			defer wg.Done()
			div, err := t.fetchFII.Dividends(code, n)
			progress.Debug("[go routine %d] dividends (%d): %v", i, len(*div), div)
			if err != nil {
				progress.ErrorMsg("%s: %v", code, err)
				return
			}
			dividends.Store(code, div)
		}(code, n)
	}
	wg.Wait()

	for _, code := range codes {
		div, ok := dividends.Load(code)
		if !ok {
			continue
		}
		dividendsForCode := div.(*[]rapina.Dividend)
		var buf *strings.Builder
		var err error
		switch t.reportFormat {
		case Rcsv:
			buf, err = t.csvDividends(code, dividendsForCode)
		case Rcsvrend:
			buf, err = t.csvDividendsOnly(code, n, dividendsForCode)
		default:
			buf, err = t.printDividends(code, dividendsForCode)
		}
		if err != nil {
			progress.Error(err)
		} else {
			fmt.Print(buf)
		}
	}

	// Footer
	// if t.reportFormat == Rtable {
	// 	fmt.Println(line)
	// }

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
			progress.ErrorMsg("Cotação de %s (%s): %v", code, d.Date, err)
		}
		if q > 0 && err == nil {
			i := d.Val / q
			p.Fprintf(buf, "R$%8.2f %8.2f%%    %8.2f%%", q, 100*i, 100*(math.Pow(1+i, 12)-1))
		}
		buf.WriteByte('\n')
	}
	buf.WriteByte('\n')

	return buf, nil
}

func (t FIITerminal) csvDividends(code string, dividends *[]rapina.Dividend) (*strings.Builder, error) {
	buf := &strings.Builder{}
	p := message.NewPrinter(language.BrazilianPortuguese)
	for _, d := range *dividends {
		p.Fprintf(buf, `%s,%s,"%f",`, code, d.Date, d.Val)

		q, err := t.fetchStock.Quote(code, d.Date)
		if err != nil {
			progress.ErrorMsg("Cotação de %s (%s): %v", code, d.Date, err)
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

func (t FIITerminal) csvDividendsOnly(code string, n int, dividends *[]rapina.Dividend) (*strings.Builder, error) {
	buf := &strings.Builder{}
	p := message.NewPrinter(language.BrazilianPortuguese)
	buf.WriteString(code)

	for _, month := range revMonthsFromToday(n) {
		found := false
		for _, div := range *dividends {
			if div.Date[0:len("YYYY-MM")] == month {
				p.Fprintf(buf, `,"%f"`, div.Val)
				found = true
				break
			}
		}
		if !found {
			buf.WriteString(`,""`)
		}

	}
	buf.WriteByte('\n')

	return buf, nil
}

func revMonthsFromToday(n int) []string {
	rev := make([]string, 0, n)
	dates := rapina.MonthsFromToday(n)
	for i := len(dates) - 1; i >= 0; i-- {
		rev = append(rev, dates[i][0:len("YYYY-MM")])
	}
	return rev
}

/* ------- MONTHLY REPORTS -------- */

func (t FIITerminal) Monthly(codes []string, n int) error {

	for _, c := range codes {
		ii, err := t.fetchFII.MonthlyReportIDs(c, n)
		progress.Status("indexes: %v (err: %v)", ii, err)
	}

	return nil
}

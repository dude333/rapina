package fetch

/*
	URL List:

	Fundos.NET: where the report IDs are obtained.
	=> https://fnet.bmfbovespa.com.br/fnet/publico/pesquisarGerenciadorDocumentosCVM?paginaCertificados=false&tipoFundo=1
	=> GET
	https://fnet.bmfbovespa.com.br/fnet/publico/pesquisarGerenciadorDocumentosDados?d=3&s=0&l=10&o[0][dataEntrega]=desc&tipoFundo=1&idCategoriaDocumento=14&idTipoDocumento=41&idEspecieDocumento=0&situacao=A&cnpj=28737771000185&dataInicial=01/02/2021&dataFinal=28/02/2021&_=1619467786288
*/

import (
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dude333/rapina"
	"github.com/dude333/rapina/parsers"
	"github.com/dude333/rapina/progress"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
)

const MAX_N = 200

// FII holds the infrastructure data.
type FII struct {
	storage rapina.FIIStorage
}

// NewFII creates a new instace of FII.
func NewFII(db *sql.DB, log rapina.Logger) (*FII, error) {
	storage, err := parsers.NewFII(db, log)
	if err != nil {
		return nil, err
	}

	fii := &FII{
		storage: storage,
	}
	return fii, nil
}

type id int

// Report holds the result of all documents filtered by a criteria defined by a
// http.Get on the B3 server.
type Report struct {
	Data []docID `json:"data"`
}
type docID struct {
	ID          id     `json:"id"`
	Description string `json:"descricaoFundo"`
	DocType     string `json:"tipoDocumento"`
	Status      string `json:"situacaoDocumento"`
}

//
// Dividends gets the report IDs for one company ('cnpj') and then the
// yeld montlhy report for 'n' months, starting from the latest released.
//
func (fii FII) Dividends(code string, n int) (*[]rapina.Dividend, error) {
	dividends, months, err := fii.dividendsFromDB(code, n)
	if err == nil {
		if months >= n {
			return dividends, err
		}
	}

	if _, err := fii.dividendsFromServer(code, n); err != nil {
		return nil, err
	}

	// Return from DB to get data sorted by date
	dividends, _, err = fii.dividendsFromDB(code, n)

	return dividends, err
}

func (fii FII) dividendsFromDB(code string, n int) (*[]rapina.Dividend, int, error) {
	var dividends []rapina.Dividend
	var months int
	for _, monthYear := range rapina.MonthsFromToday(n + 2) {
		d, err := fii.storage.Dividends(code, monthYear)
		if err == nil { // ignore errors
			dividends = append(dividends, *d...)
			months++
		}
		if months == n {
			break
		}
	}

	if len(dividends) == 0 {
		return nil, 0, errors.New("dividendos não encontrados")
	}

	return &dividends, months, nil
}

//
// Dividends gets the report IDs for one company ('cnpj') and then the
// yeld montlhy report for 'n' months, starting from the latest released.
//
// If the number of reports does not match n, it'll retry with a bigger n as
// sometimes reports from follow-on offerings (FPO).
//
func (fii *FII) dividendsFromServer(code string, n int) (*[]rapina.Dividend, error) {
	if n > MAX_N {
		n = MAX_N
	}
	nn := n
	lastLen := -1
	dividends := &[]rapina.Dividend{}

	for len(*dividends) < n && nn <= 200 {
		progress.Status("Relatórios de dividendos: %s", code)

		ids, err := fii.reportIDs(repDividends, code, nn)
		if err != nil {
			return nil, err
		}
		progress.Debug("Report IDs: %v", ids)

		dividends, err = fii.dividendReport(code, ids)
		if err != nil {
			return nil, err
		}
		progress.Debug("Dividends (%d): %v", len(*dividends), *dividends)

		if lastLen == len(*dividends) {
			break
		}
		lastLen = len(*dividends)
		nn += 2 * (n - len(*dividends))
	}

	return dividends, nil
}

//
// dividendReport parses the dividend reports and return their dividends.
//
func (fii *FII) dividendReport(code string, ids []id) (*[]rapina.Dividend, error) {
	yeld := make(map[string]string, len(ids))

	c := colly.NewCollector()
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html")
	})

	c.OnError(func(r *colly.Response, err error) {
		progress.ErrorMsg("Request URL: %v failed with response: %v\nError: %v", r.Request.URL, string(r.Body), err)
	})

	// Handles the html report
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		var fieldName string
		e.ForEach("td", func(_ int, el *colly.HTMLElement) {
			v := strings.Trim(el.Text, " \r\n")
			if v != "" {
				if fieldName == "" {
					fieldName = v
				} else {
					if strings.Contains(fieldName, "Código de negociação da cota") ||
						strings.Contains(fieldName, "Data da informação") {
						progress.Debug("[%s] %-30s => %s\n", code, fieldName, v)
					}
					yeld[fieldName] = v
					fieldName = ""
				}
			}
		})
	})

	// Get the yeld monthly report given the list of 'report IDs' -- returns HTML
	dividends := make([]rapina.Dividend, 0, len(ids))
	for _, id := range ids {
		u := fmt.Sprintf("https://fnet.bmfbovespa.com.br/fnet/publico/exibirDocumento?id=%d&cvm=true", id)
		progress.Debug("Relatórios de dividendos: %s", u)
		if err := c.Visit(u); err != nil {
			return nil, err
		}
		d, err := fii.storage.SaveDividend(yeld)
		if err != nil {
			progress.Error(err)
			continue
		}
		// fmt.Println("from server", d.Code, d.Date, d.Val)
		if d.Code == code {
			dividends = append(dividends, *d)
		}
	}

	return &dividends, nil
}

func (fii *FII) MonthlyReportIDs(code string, n int) ([]id, error) {
	ids, err := fii.reportIDs(repMonthly, code, n)
	if err != nil {
		return []id{}, err
	}
	_, err = fii.monthlyReport(code, ids)
	if err != nil {
		return []id{}, err
	}

	return ids, nil
}

//
// monthlyReport parses the FII monthly reports.
//
func (fii *FII) monthlyReport(code string, ids []id) (*[]rapina.Monthly, error) {
	yeld := make(map[string]string, len(ids))

	c := colly.NewCollector()
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html")
	})

	c.OnError(func(r *colly.Response, err error) {
		progress.ErrorMsg("Request URL: %v failed with response: %v\nError: %v", r.Request.URL, string(r.Body), err)
	})

	// Handles the html report
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		var fieldName string
		e.ForEach("td", func(_ int, el *colly.HTMLElement) {
			v := strings.Trim(el.Text, " \r\n")
			progress.Debug("%q", v)
			if v != "" {
				if fieldName == "" {
					if v[0] < '0' || v[0] > '9' { // Ignore fields starting with number
						fieldName = v
					}
				} else {
					fmt.Printf("%-30s => %s\n", fieldName, v)
					yeld[fieldName] = v
					fieldName = ""
				}
			}
		})
		progress.Status("----------------------")
	})

	// Get the yeld monthly report given the list of 'report IDs' -- returns HTML
	monthly := make([]rapina.Monthly, 0, len(ids))
	for _, id := range ids {
		u := fmt.Sprintf("https://fnet.bmfbovespa.com.br/fnet/publico/exibirDocumento?id=%d&cvm=true", id)
		progress.Debug(u)
		if err := c.Visit(u); err != nil {
			return nil, err
		}
		// d, err := fii.storage.SaveDividend(yeld)
		// if err != nil {
		// 	fii.log.Error("%v", err)
		// 	continue
		// }
		// // fmt.Println("from server", d.Code, d.Date, d.Val)
		// if d.Code == code {
		// 	monthly = append(monthly, *d)
		// }
	}

	return &monthly, nil
}

//
// Details returns the FII Details from DB. If not found:
// fetches from server, stores it in the DB and returns the Details.
//
func (fii *FII) Details(fiiCode string) (*rapina.FIIDetails, error) {
	if len(fiiCode) != 4 && len(fiiCode) != 6 {
		return nil, fmt.Errorf("wrong code '%s'", fiiCode)
	}

	details, err := fii.storage.Details(fiiCode)
	if err == nil && details.DetailFund.CNPJ != "" {
		return details, nil
	}

	progress.Warning("Detalhes do %s não encontrado no bd. Consultando web...", fiiCode)

	// Fetch from server if not found in the database
	data := fmt.Sprintf(`{"typeFund":7,"cnpj":"0","identifierFund":"%s"}`, fiiCode[0:4])
	enc := base64.URLEncoding.EncodeToString([]byte(data))
	fundDetailURL := rapina.JoinURL(
		`https://sistemaswebb3-listados.b3.com.br/fundsProxy/fundsCall/GetDetailFundSIG/`,
		enc,
	)

	tr := &http.Transport{
		DisableCompression: true,
		IdleConnTimeout:    _http_timeout,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(fundDetailURL)
	if err != nil {
		return details, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return details, fmt.Errorf("%s: %s", resp.Status, fundDetailURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "FII Details(%s): reading body", fiiCode)
	}

	err = fii.storage.SaveDetails(body)
	if err != nil {
		return details, errors.Wrap(err, "armazenando detalhes do FII")
	}

	return fii.storage.Details(fiiCode)
}

// Report type
type repType int

const (
	repMonthly repType = iota + 1
	repDividends
)

func (fii *FII) reportIDs(rt repType, code string, n int) ([]id, error) {
	n = minmax(n, 1, MAX_N)

	// Parameters to list the report IDs for the last 'n' dividend reports
	timestamp := strconv.FormatInt(int64(time.Now().UnixNano()/1e6), 10)
	nMonthAgo := time.Now()
	nMonthAgo = nMonthAgo.AddDate(0, -n, -nMonthAgo.Day()+1)
	det, err := fii.Details(code)
	if err != nil {
		return nil, err
	}
	cnpj := det.DetailFund.CNPJ

	var idTipoDocumento, idCategoriaDocumento, d string
	if rt == repMonthly {
		idTipoDocumento = "40"
		idCategoriaDocumento = "6"
		d = "0"
	} else if rt == repDividends {
		idTipoDocumento = "41"
		idCategoriaDocumento = "14"
		d = "2"
	} else {
		return []id{}, errors.New("invalid report type")
	}

	v := url.Values{
		"tipoFundo":            []string{"1"},
		"cnpjFundo":            []string{cnpj},
		"idTipoDocumento":      []string{idTipoDocumento},
		"idCategoriaDocumento": []string{idCategoriaDocumento},
		"d":                    []string{d},
		"idEspecieDocumento":   []string{"0"},
		"situacao":             []string{"A"},
		"s":                    []string{"0"},
		"l":                    []string{"200"}, // 'n*2' latest reports as other codes may appear (e.g.:ABCD11, ABCD12, ABCD13...)
		"dataFinal":            []string{time.Now().Format("02/01/2006")},
		"dataInicial":          []string{nMonthAgo.Format("02/01/2006")},
        "o[0][dataReferencia]":	[]string{"asc"},
		"_":                    []string{timestamp},
	}

	// Get the 'report IDs' for a given company (CNPJ) -- returns JSON
	var report Report
	u := "https://fnet.bmfbovespa.com.br/fnet/publico/pesquisarGerenciadorDocumentosDados?" +
		v.Encode()
	progress.Debug("* Report IDs: %s", u)
	if err := getJSON(u, &report); err != nil {
		return nil, err
	}

	var ids []id
	for _, d := range report.Data {
		if d.Status == "A" {
			ids = append(ids, d.ID)
		}
	}

	return ids, nil
}

// minmax returns n limited to [min, max]
func minmax(n, min, max int) int {
	if n < min {
		n = min
	}
	if n > max {
		n = MAX_N
	}
	return n
}

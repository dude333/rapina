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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dude333/rapina"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
)

// FII holds the infrastructure data.
type FII struct {
	parser rapina.FIIParser
	log    rapina.Logger
}

// NewFII creates a new instace of FII.
func NewFII(parser rapina.FIIParser, log rapina.Logger) *FII {
	fii := &FII{parser: parser, log: log}
	return fii
}

type id int
type Report struct {
	Data []docID `json:"data"`
}
type docID struct {
	ID          id     `json:"id"`
	Description string `json:"descricaoFundo"`
	DocType     string `json:"tipoDocumento"`
	Status      string `json:"situacaoDocumento"`
}

/*
type money float32
type fiiYeld struct {
	FundName  string `json:"Nome do Fundo:"`
	FundCNPJ  string `json:"CNPJ do Fundo:"`
	Admin     string `json:"Nome do Administrador:"`
	AdminCNPJ string `json:"CNPJ do Administrador:"`
	ISIN      string `json:"Código ISIN da cota:"`
	Cod       string `json:"Código de negociação da cota:"`

	ReleaseDate *time.Time `json:"Data da informação"`
	BaseDate    *time.Time `json:"Data-base (último dia de negociação “com” direito ao provento)"`
	PymtDate    *time.Time `json:"Data do pagamento"`
	Value       money      `json:"Valor do provento por cota (R$)"`
	Month       int        `json:"Período de referência"`
	Year        int        `json:"Ano"`
}
*/

//
// Dividends gets the report IDs for one company ('cnpj') and then the
// yeld montlhy report for 'n' months, starting from the latest released.
//
func (fii FII) Dividends(code string, n int) (*[]rapina.Dividend, error) {
	dividends, months, err := fii.dividendsFromDB(code, n)
	if err == nil {
		fii.log.Debug("FROM DB (n=%v months=%v)", n, months)
		if months >= n {
			return dividends, err
		}
	}

	fii.log.Debug("FROM SERVER (%v)", err)
	dividends, err = fii.dividendsFromServer(code, n)

	return dividends, err
}

func (fii FII) dividendsFromDB(code string, n int) (*[]rapina.Dividend, int, error) {
	var dividends []rapina.Dividend
	var months int
	for _, monthYear := range rapina.MonthsFromToday(n + 2) {
		d, err := fii.parser.Dividends(code, monthYear)
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
func (fii FII) dividendsFromServer(code string, n int) (*[]rapina.Dividend, error) {
	var ids []id
	yeld := make(map[string]string, n)
	if n <= 0 {
		n = 1
	}

	c := colly.NewCollector()
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html, application/json")
	})

	c.OnError(func(r *colly.Response, err error) {
		fii.log.Error("Request URL: %v failed with response: %v\nError: %v", r.Request.URL, string(r.Body), err)
	})

	// Handles the json output with the report IDs
	c.OnResponse(func(r *colly.Response) {
		if !strings.Contains(r.Headers.Get("content-type"), "application/json") {
			return
		}
		var report Report
		err := json.Unmarshal(r.Body, &report)
		if err != nil {
			fii.log.Error("json error: %v", err)
			return
		}
		for _, d := range report.Data {
			if d.Status == "A" {
				ids = append(ids, d.ID)
			}
		}
	})

	// Parameters to list the report IDs for the last 'n' dividend reports
	timestamp := strconv.FormatInt(int64(time.Now().UnixNano()/1e6), 10)
	cnpj, err := fii.CNPJ(code)
	if err != nil {
		return nil, err
	}
	v := url.Values{
		"d":                    []string{"2"},
		"s":                    []string{"0"},
		"l":                    []string{strconv.Itoa(n)}, // months
		"o[0][dataEntrega]":    []string{"desc"},
		"tipoFundo":            []string{"1"},
		"cnpjFundo":            []string{cnpj},
		"idCategoriaDocumento": []string{"14"},
		"idTipoDocumento":      []string{"41"},
		"idEspecieDocumento":   []string{"0"},
		"situacao":             []string{"A"},
		"_":                    []string{timestamp},
	}

	// Get the 'report IDs' for a given company (CNPJ) -- returns JSON
	u := "https://fnet.bmfbovespa.com.br/fnet/publico/pesquisarGerenciadorDocumentosDados" +
		"?" + v.Encode()
	if err := c.Visit(u); err != nil {
		return nil, err
	}

	// Handles the html report
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		var fieldName string
		e.ForEach("td", func(_ int, el *colly.HTMLElement) {
			v := strings.Trim(el.Text, " \r\n")
			if v != "" {
				if fieldName == "" {
					fieldName = v
				} else {
					// fmt.Printf("%-30s => %s\n", fieldName, v)
					yeld[fieldName] = v
					fieldName = ""
				}
			}
		})
	})

	// Get the yeld monthly report given the list of 'report IDs' -- returns HTML
	dividends := make([]rapina.Dividend, 0, n)
	for _, i := range ids {
		u = fmt.Sprintf("https://fnet.bmfbovespa.com.br/fnet/publico/exibirDocumento?id=%d&cvm=true", i)
		if err := c.Visit(u); err != nil {
			return nil, err
		}
		d, err := fii.parser.SaveDividend(yeld)
		if err != nil {
			log.Println("[x]", err)
			continue
		}
		// fmt.Println("from server", d.Code, d.Date, d.Val)
		if d.Code == code {
			dividends = append(dividends, *d)
		}
	}

	return &dividends, nil
}

//
// FetchFIIDetails returns the FII CNPJ from DB. If not found:
// fetches from server, stores it in the DB and returns the CNPJ.
//
func (fii FII) CNPJ(fiiCode string) (string, error) {
	if len(fiiCode) != 4 && len(fiiCode) != 6 {
		return "", fmt.Errorf("wrong code '%s'", fiiCode)
	}

	cnpj, err := fii.parser.CNPJ(fiiCode)
	if err != nil {
		return "", err
	}
	if cnpj != "" {
		return cnpj, nil
	}

	// Fetch from server if not found in the database
	data := fmt.Sprintf(`{"typeFund":7,"cnpj":"0","identifierFund":"%s"}`, fiiCode[0:4])
	enc := base64.URLEncoding.EncodeToString([]byte(data))
	fundDetailURL := rapina.JoinURL(
		`https://sistemaswebb3-listados.b3.com.br/fundsProxy/fundsCall/GetDetailFundSIG/`,
		enc,
	)

	tr := &http.Transport{
		DisableCompression: true,
		IdleConnTimeout:    30 * time.Second,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(fundDetailURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%s: %s", resp.Status, fundDetailURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read body")
	}

	_ = fii.parser.StoreFIIDetails(body)

	return fii.parser.CNPJ(fiiCode)
}

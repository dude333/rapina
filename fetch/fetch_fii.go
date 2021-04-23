package fetch

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/pkg/errors"
)

type FIIStore interface {
	CNPJ(code string) (string, error)
	StoreFIIDetails(stream []byte) error
}

// FII holds the infrastructure data.
type FII struct {
	store FIIStore
}

// NewFII creates a new instace of FII.
func NewFII(store FIIStore) *FII {
	fii := &FII{store: store}
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
// FetchFIIDividends gets the report IDs for one company ('cnpj') and then the
// yeld montlhy report for 'n' months, starting from the latest released.
//
func (fii FII) FetchFIIDividends(code string, n int) error {
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
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", string(r.Body), "\nError:", err)
	})

	// Handles the json output with the report IDs
	c.OnResponse(func(r *colly.Response) {
		if !strings.Contains(r.Headers.Get("content-type"), "application/json") {
			return
		}
		var report Report
		err := json.Unmarshal(r.Body, &report)
		if err != nil {
			fmt.Println("json error:", err)
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
		return err
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
		return err
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
					fmt.Printf("%-30s => %s\n", fieldName, v)
					yeld[fieldName] = v
					fieldName = ""
				}
			}
		})
	})

	// Get the yeld monthly report given the list of 'report IDs' -- returns HTML
	for _, i := range ids {
		u = fmt.Sprintf("https://fnet.bmfbovespa.com.br/fnet/publico/exibirDocumento?id=%d&cvm=true", i)
		if err := c.Visit(u); err != nil {
			return err
		}
		fmt.Println("----------------------------")

		// fmt.Printf("%+v\n", yeld)
	}

	return nil
}

//
// FetchFIIDetails returns the FII CNPJ from DB. If not found:
// fetches from server, stores it in the DB and returns the CNPJ.
//
func (fii FII) CNPJ(fiiCode string) (string, error) {
	if len(fiiCode) != 4 && len(fiiCode) != 6 {
		return "", fmt.Errorf("wrong code '%s'", fiiCode)
	}

	cnpj, err := fii.store.CNPJ(fiiCode)
	if err != nil {
		return "", err
	}
	if cnpj != "" {
		return cnpj, nil
	}

	// Fetch from server if not found in the database
	data := fmt.Sprintf(`{"typeFund":7,"cnpj":"0","identifierFund":"%s"}`, fiiCode[0:4])
	enc := base64.URLEncoding.EncodeToString([]byte(data))
	fundDetailURL := JoinURL(
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

	_ = fii.store.StoreFIIDetails(body)

	return fii.store.CNPJ(fiiCode)
}

/* ------- Utils ------- */

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// JoinURL joins strings as URL paths
func JoinURL(base string, paths ...string) string {
	p := path.Join(paths...)
	return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(p, "/"))
}

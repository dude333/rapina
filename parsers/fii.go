package parsers

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/pkg/errors"
)

type id int
type data struct {
	Data []docID `json:"data"`
}
type docID struct {
	ID      id     `json:"id"`
	Descr   string `json:"descricaoFundo"`
	TipoDoc string `json:"tipoDocumento"`
	Sit     string `json:"situacaoDocumento"`
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
// FetchFII gets the report IDs for one company and then the
// yeld montlhy reports.
//
func FetchFII(baseURL string) error {
	var ids []id

	c := colly.NewCollector()
	yeld := make(map[string]string, 20)

	// Handles the html report
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		var fieldName string

		e.ForEach("td", func(_ int, el *colly.HTMLElement) {
			v := strings.Trim(el.Text, " \r\n")
			if v != "" {
				if fieldName == "" {
					fieldName = v
				} else {
					fmt.Printf("%s => %s\n", fieldName, v)
					yeld[fieldName] = v
					fieldName = ""
				}
			}
		})

	})

	// Handles the json output with the report IDs
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html, application/json")
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", string(r.Body), "\nError:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		if !strings.Contains(r.Headers.Get("content-type"), "application/json") {
			return
		}
		var d data
		err := json.Unmarshal(r.Body, &d)
		if err != nil {
			fmt.Println("json error:", err)
		} else {
			for _, x := range d.Data {
				if x.Sit == "A" {
					ids = append(ids, x.ID)
				}
			}
		}
	})

	v := url.Values{
		"d":                    []string{"2"},
		"s":                    []string{"0"},
		"l":                    []string{"2"}, // months
		"o[0][dataEntrega]":    []string{"desc"},
		"tipoFundo":            []string{"1"},
		"cnpjFundo":            []string{"14410722000129"},
		"idCategoriaDocumento": []string{"14"},
		"idTipoDocumento":      []string{"41"},
		"idEspecieDocumento":   []string{"0"},
		"situacao":             []string{"A"},
		"_":                    []string{"1609254186709"},
	}

	// Get the 'report IDs' for a given company (CNPJ)
	u := JoinURL(baseURL, "/pesquisarGerenciadorDocumentosDados?", v.Encode())
	if err := c.Visit(u); err != nil {
		return err
	}

	// Get the yeld monthly report given the list of 'report IDs'
	for _, i := range ids {
		u = JoinURL(baseURL, fmt.Sprintf("/exibirDocumento?id=%d&cvm=true", i))
		if err := c.Visit(u); err != nil {
			return err
		}
		fmt.Printf("%+v\n", yeld)
	}

	return nil
}

//
// FetchFIIs downloads the list of FIIs to get their code (e.g. 'HGLG'),
// then it uses this code to retrieve its details to get the CNPJ.
// Original baseURL: https://sistemaswebb3-listados.b3.com.br.
//
func FetchFIIList(baseURL string) ([]string, error) {
	listFundsURL := JoinURL(baseURL, `/fundsProxy/fundsCall/GetListFundDownload/eyJ0eXBlRnVuZCI6NywicGFnZU51bWJlciI6MSwicGFnZVNpemUiOjIwfQ==`)
	// fundsDetailsURL := `https://sistemaswebb3-listados.b3.com.br/fundsProxy/fundsCall/GetDetailFundSIG`

	tr := &http.Transport{
		DisableCompression: true,
		IdleConnTimeout:    30 * time.Second,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(listFundsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	unq, err := strconv.Unquote(string(body))
	if err != nil {
		return nil, err
	}
	txt, err := base64.StdEncoding.DecodeString(unq)
	if err != nil {
		return nil, err
	}

	var codes []string

	for _, line := range strings.Split(string(txt), "\n") {
		p := strings.Split(line, ";")
		if len(p) > 3 && len(p[3]) == 4 {
			codes = append(codes, p[3])
		}
	}

	return codes, nil
}

// FII details
type FII struct {
	DetailFund struct {
		Acronym               string      `json:"acronym"`
		TradingName           string      `json:"tradingName"`
		TradingCode           string      `json:"tradingCode"`
		TradingCodeOthers     string      `json:"tradingCodeOthers"`
		Cnpj                  string      `json:"cnpj"`
		Classification        string      `json:"classification"`
		WebSite               string      `json:"webSite"`
		FundAddress           string      `json:"fundAddress"`
		FundPhoneNumberDDD    string      `json:"fundPhoneNumberDDD"`
		FundPhoneNumber       string      `json:"fundPhoneNumber"`
		FundPhoneNumberFax    string      `json:"fundPhoneNumberFax"`
		PositionManager       string      `json:"positionManager"`
		ManagerName           string      `json:"managerName"`
		CompanyAddress        string      `json:"companyAddress"`
		CompanyPhoneNumberDDD string      `json:"companyPhoneNumberDDD"`
		CompanyPhoneNumber    string      `json:"companyPhoneNumber"`
		CompanyPhoneNumberFax string      `json:"companyPhoneNumberFax"`
		CompanyEmail          string      `json:"companyEmail"`
		CompanyName           string      `json:"companyName"`
		QuotaCount            string      `json:"quotaCount"`
		QuotaDateApproved     string      `json:"quotaDateApproved"`
		Codes                 []string    `json:"codes"`
		CodesOther            interface{} `json:"codesOther"`
		Segment               interface{} `json:"segment"`
	} `json:"detailFund"`
	ShareHolder struct {
		ShareHolderName           string `json:"shareHolderName"`
		ShareHolderAddress        string `json:"shareHolderAddress"`
		ShareHolderPhoneNumberDDD string `json:"shareHolderPhoneNumberDDD"`
		ShareHolderPhoneNumber    string `json:"shareHolderPhoneNumber"`
		ShareHolderFaxNumber      string `json:"shareHolderFaxNumber"`
		ShareHolderEmail          string `json:"shareHolderEmail"`
	} `json:"shareHolder"`
}

//
// FetchFIIDetails gets all the details from a fund and returns a *FII struct.
//
func FetchFIIDetails(baseURL string, fiiCode string) (*FII, error) {
	data := fmt.Sprintf(`{"typeFund":7,"cnpj":"0","identifierFund":"%s"}`, fiiCode)
	enc := base64.URLEncoding.EncodeToString([]byte(data))
	fundDetailURL := JoinURL(baseURL, `/fundsProxy/fundsCall/GetDetailFundSIG/`, enc)

	tr := &http.Transport{
		DisableCompression: true,
		IdleConnTimeout:    30 * time.Second,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(fundDetailURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %s", resp.Status, fundDetailURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read body")
	}

	var fii FII

	if err := json.Unmarshal(body, &fii); err != nil {
		return nil, errors.Wrap(err, "json unmarshal")
	}

	return &fii, nil
}

// JoinURL joins strings as URL paths
func JoinURL(base string, paths ...string) string {
	p := path.Join(paths...)
	return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(p, "/"))
}

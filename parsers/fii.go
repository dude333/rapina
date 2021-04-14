package parsers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly"
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

func FetchFII(baseURL string) error {
	var ids []id
	n := 1

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
		// fmt.Println("Visiting", r.URL)
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
					// fmt.Printf("%d (%s - %s)\n", x.ID, x.Descr, x.TipoDoc)
					ids = append(ids, x.ID)
				}
			}
		}
		// fmt.Println(n, string(r.Body))
		n++
	})

	v := url.Values{
		"d":                    []string{"2"},
		"s":                    []string{"0"},
		"l":                    []string{"2"},
		"o[0][dataEntrega]":    []string{"desc"},
		"tipoFundo":            []string{"1"},
		"cnpjFundo":            []string{"14410722000129"},
		"idCategoriaDocumento": []string{"14"},
		"idTipoDocumento":      []string{"41"},
		"idEspecieDocumento":   []string{"0"},
		"situacao":             []string{"A"},
		"_":                    []string{"1609254186709"},
	}

	u := baseURL + "/pesquisarGerenciadorDocumentosDados?" + v.Encode()
	if err := c.Visit(u); err != nil {
		return err
	}

	for _, i := range ids {
		u = fmt.Sprintf("%s/exibirDocumento?id=%d&cvm=true", baseURL, i)
		if err := c.Visit(u); err != nil {
			return err
		}
		fmt.Printf("%+v\n", yeld)
	}

	return nil
}

package parsers

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func companies(url string) {
	c := colly.NewCollector(
		// Restrict crawling to specific domains
		// colly.AllowedDomains("bvmf.bmfbovespa.com.br"),
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),
		// Allow crawling to be done in parallel / async
		colly.Async(false),
		colly.CacheDir(".data/cache"),
	)

	// Find and visit all links
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		// if e.Attr("class") != "GridRow_SiteBmfBovespa GridBovespaItemStyle" {
		// 	return
		// }

		e.ForEachWithBreak("a", func(_ int, elem *colly.HTMLElement) bool {
			if strings.Contains(elem.Attr("href"), "ResumoEmpresaPrincipal.aspx") {
				fmt.Println("   >", elem.Text)
			}
			return false // get only the 1st elem
		})

	})

	// c.OnResponse(func(r *colly.Response) {
	// 	fmt.Println(string(r.Body))
	// })

	// c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting", r.URL)
	// })

	c.Visit(url)
	// c.Visit("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/BuscaEmpresaListada.aspx?segmento=Bancos")
	// c.Visit("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/BuscaEmpresaListada.aspx?segmento=Explora%c3%a7%c3%a3o+de+Im%c3%b3veis")
}

func Sectors() {
	c := colly.NewCollector(
		// Restrict crawling to specific domains
		// colly.AllowedDomains("bvmf.bmfbovespa.com.br"),
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),
		// Allow crawling to be done in parallel / async
		colly.Async(false),
		colly.CacheDir(".data/cache"),
	)

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// e.Request.Visit(e.Attr("href"))
		// fmt.Printf("> %v\n", e.Attr("href"))
	})

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		// if e.Attr("class") != "GridRow_SiteBmfBovespa GridBovespaItemStyle" {
		// 	return
		// }

		// h, _ := e.DOM.Find("td").Html()
		// fmt.Printf("=> %#v \n", h)

		var sector string
		var subsectors []string
		c := 0
		e.ForEach("td", func(_ int, elem *colly.HTMLElement) {
			elem.DOM.Each(func(_ int, s *goquery.Selection) {
				h, _ := s.Html()
				if c == 0 {
					sector = h
				} else if c == 1 {
					subsectors = strings.Split(h, "<br/>")
					last := subsectors[0]
					for i := range subsectors {
						if subsectors[i] == "" {
							subsectors[i] = last
						}
						last = subsectors[i]
					}
				}
				c++
			})

			elem.ForEach("a[href]", func(i int, elem *colly.HTMLElement) {
				if strings.Contains(elem.Attr("href"), "BuscaEmpresaListada.aspx") {
					fmt.Printf("\n=> %s > %s > %s:\n", sector, subsectors[i], elem.Text) //, elem.Attr("href"))
					companies("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/" + elem.Attr("href"))
				}
			})
		})

		// s := col[2]
		// if len(col[2]) > 3 {
		// 	s = s[2:]
		// }
		// fmt.Printf("|=> %s \n |==> %s \n  |====> %s\n", col[0], col[1], s)

	})

	c.OnResponse(func(r *colly.Response) {
		// fmt.Println(string(r.Body))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/BuscaEmpresaListada.aspx?opcao=1&indiceAba=1&Idioma=pt-br")
}

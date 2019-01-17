package parsers

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func companies() {
	c := colly.NewCollector(
		// Restrict crawling to specific domains
		colly.AllowedDomains("bvmf.bmfbovespa.com.br"),
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),
		// Allow crawling to be done in parallel / async
		colly.Async(false),
	)

	// Find and visit all links
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		// if e.Attr("class") != "GridRow_SiteBmfBovespa GridBovespaItemStyle" {
		// 	return
		// }

		e.ForEachWithBreak("a", func(_ int, elem *colly.HTMLElement) bool {
			if strings.Contains(elem.Attr("href"), "ResumoEmpresaPrincipal.aspx") {
				fmt.Println("a> ", elem.Text)
			}
			return false // get only the 1st elem
		})

	})

	c.Visit("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/BuscaEmpresaListada.aspx?segmento=Bancos")
	// c.Visit("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/BuscaEmpresaListada.aspx?segmento=Explora%c3%a7%c3%a3o+de+Im%c3%b3veis")
}

func sectors() {
	c := colly.NewCollector(
		// Restrict crawling to specific domains
		// colly.AllowedDomains("bvmf.bmfbovespa.com.br"),
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),
		// Allow crawling to be done in parallel / async
		colly.Async(false),
		colly.CacheDir("./cache"),
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

		e.ForEach("td", func(_ int, elem *colly.HTMLElement) {
			// 	// if strings.Contains(elem.Attr("href"), "ResumoEmpresaPrincipal.aspx") {
			// 	fmt.Printf("=> %#v \n", elem.DOM.Find("td").Text())
			// 	// }
			// h, _ := elem.DOM.Html()
			// fmt.Printf("=> %#v \n", h)

			elem.DOM.Each(func(_ int, s *goquery.Selection) {
				a := s.Find("a[href]").Text()
				if len(a) == 0 {
					h, _ := s.Html()
					h = strings.Replace(h, "<br/>", ";", -1)
					fmt.Printf("|=> %#v\n", h)
				}
			})

			elem.ForEach("a[href]", func(_ int, elem *colly.HTMLElement) {
				if strings.Contains(elem.Attr("href"), "BuscaEmpresaListada.aspx") {
					fmt.Printf("==> %-40s \t [%s]\n", elem.Text, elem.Attr("href"))
				}
			})
		})

	})

	c.OnResponse(func(r *colly.Response) {
		// fmt.Println(string(r.Body))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/BuscaEmpresaListada.aspx?opcao=1&indiceAba=1&Idioma=pt-br")
}

package parsers

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

//
// Sectors grab data from B3 website and prints out to a yaml file
// with all companies grouped by sector, subsector, segment
//
func Sectors() {
	c := colly.NewCollector(
		// Restrict crawling to specific domains
		// colly.AllowedDomains("bvmf.bmfbovespa.com.br"),
		colly.AllowURLRevisit(),
		colly.Async(false),
		colly.CacheDir(".data/cache"),
	)

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		var sector string
		var subsectors []string
		c := 0
		e.ForEach("td", func(_ int, elem *colly.HTMLElement) {
			elem.DOM.Each(func(_ int, s *goquery.Selection) {
				h, _ := s.Html()
				if c == 0 {
					sector = h
					fmt.Println("  - Setor:", sector)
					fmt.Println("    Subsetores:")
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

			lastSub := ""
			elem.ForEach("a[href]", func(i int, elem *colly.HTMLElement) {
				if strings.Contains(elem.Attr("href"), "BuscaEmpresaListada.aspx") {
					// fmt.Printf("\n=> %s > %s > %s:\n", sector, subsectors[i], elem.Text) //, elem.Attr("href"))
					if subsectors[i] != lastSub {
						fmt.Println("      - Subsetor:", subsectors[i])
						fmt.Println("        Segmentos:")
					}
					lastSub = subsectors[i]
					fmt.Println("          - Segmento:", elem.Text)
					fmt.Println("            Empresas:")
					companies("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/" + elem.Attr("href"))
				}
			})
		})
	})

	fmt.Println("Setores:")
	c.Visit("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/BuscaEmpresaListada.aspx?opcao=1&indiceAba=1&Idioma=pt-br")
}

//
// companies lists all companies in the same sector/subsector/segment
//
func companies(url string) {
	c := colly.NewCollector(
		// Restrict crawling to specific domains
		// colly.AllowedDomains("bvmf.bmfbovespa.com.br"),
		colly.AllowURLRevisit(),
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
				fmt.Println("              -", elem.Text)
			}
			return false // get only the 1st elem
		})

	})

	c.Visit(url)
}

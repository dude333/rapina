package parsers

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

//
// SectorsToYaml grab data from B3 website and prints out to a yaml file
// with all companies grouped by sector, subsector, segment
//
func SectorsToYaml(yamlFile string) (err error) {
	progress := []string{"/", "-", "\\", "|", "-", "\\"}
	var p int32

	if !overwritePrompt(yamlFile) {
		return fmt.Errorf("arquivo %s não foi alterado", yamlFile)
	}
	f, err := os.Create(yamlFile)
	if err != nil {
		return errors.Wrapf(err, "falha ao criar arquivo %s", yamlFile)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

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
					fmt.Fprintln(w, "  - Setor:", sector)
					fmt.Fprintln(w, "    Subsetores:")
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
						fmt.Fprintln(w, "      - Subsetor:", subsectors[i])
						fmt.Fprintln(w, "        Segmentos:")
					}
					lastSub = subsectors[i]
					fmt.Fprintln(w, "          - Segmento:", elem.Text)
					fmt.Fprintln(w, "            Empresas:")
					companies(w, "http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/"+elem.Attr("href"))
				}

				fmt.Printf("\r[%s]", progress[p%6])
				p++
			})
		})
	})

	fmt.Print("[ ] Lendo informações do site da B3")
	fmt.Fprintln(w, "Setores:")

	c.Visit("http://bvmf.bmfbovespa.com.br/cias-listadas/empresas-listadas/BuscaEmpresaListada.aspx?opcao=1&indiceAba=1&Idioma=pt-br")

	fmt.Println()
	w.Flush()

	return
}

//
// companies lists all companies in the same sector/subsector/segment
//
func companies(w *bufio.Writer, url string) {
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
				fmt.Fprintln(w, "              -", elem.Text)
			}
			return false // get only the 1st elem
		})

	})

	c.Visit(url)
}

//
// overwritePrompt prompts to overwrite file if it exists
func overwritePrompt(filename string) bool {
	if _, err := os.Stat(filename); err == nil { // check if file exists
		fmt.Printf("\n[?] Deseja sobrescrever o arquivo \"%s\"? (s/N) ", filename)
		reader := bufio.NewReader(os.Stdin)
		prompt, _ := reader.ReadString('\n')
		if !strings.EqualFold(prompt, "s\n") && !strings.EqualFold(prompt, "sim\n") {
			return false
		}
	}

	return true
}

// S contains the sectors
type S struct {
	Sectors []Sector `yaml:"Setores"`
}

// Sector is divided into subsectors
type Sector struct {
	Name       string      `yaml:"Setor"`
	Subsectors []Subsector `yaml:"Subsetores"`
}

// Subsector is divided into segments
type Subsector struct {
	Name     string    `yaml:"Subsetor"`
	Segments []Segment `yaml:"Segmentos"`
}

// Segment contains companies from the same sector/subsector/segment
type Segment struct {
	Name      string   `yaml:"Segmento"`
	Companies []string `yaml:"Empresas"`
}

//
// FromSector returns all companies from the same sector as the 'company'
//
func FromSector(company, yamlFile string) (companies []string, sectorName string, err error) {

	y, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		err = errors.Wrapf(err, "ReadFile: %v", err)
		return
	}

	s := S{}
	yaml.Unmarshal(y, &s)

	for _, sector := range s.Sectors {
		for _, subsector := range sector.Subsectors {
			for _, segment := range subsector.Segments {
				if FuzzyMatch(company, segment.Companies, 2) {
					companies = segment.Companies
					sectorName = strings.Join([]string{sector.Name, subsector.Name, segment.Name}, " > ")
					return
				}
			}
		}
	}

	return
}

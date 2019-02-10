package reports

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dude333/rapina/parsers"
	"github.com/pkg/errors"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//
// ListCompanies shows all available companies
//
func ListCompanies(db *sql.DB) (com []string, err error) {
	com, err = companies(db)

	if err != nil {
		fmt.Println("[x] Falha:", err)
		return
	}

	// Sort accents correctly
	cl := collate.New(language.BrazilianPortuguese, collate.Loose)
	cl.SortStrings(com)

	return
}

//
// ListSector shows all companies from the same sector as 'company'
//
func ListSector(db *sql.DB, company, yamlFile string) (err error) {
	// Companies from the same sector
	secCo, secName, err := parsers.FromSector(company, yamlFile)
	if len(secCo) <= 1 || err != nil {
		err = errors.Wrap(err, "erro ao ler arquivo dos setores "+yamlFile)
		return
	}

	// All companies stored on db
	list, err := ListCompanies(db)
	if err != nil {
		err = errors.Wrap(err, "erro ao listar empresas")
		return
	}

	// Translate company names to match the name stored on db
	fmt.Printf("%-40s %s\n", "ARQUIVO YAML", "BANCO DE DADOS")
	fmt.Printf("%-40s %s\n", strings.Repeat("-", 40), strings.Repeat("-", 40))
	for _, s := range secCo {
		z := parsers.FuzzyFind(s, list, 3)
		if len(z) > 0 {
			fmt.Printf("%-40s %s\n", s, z)
		} else {
			fmt.Printf("%-40s %s\n", s, "Nao encontrado")
		}
	}
	fmt.Printf("\nSETOR: %s\n", secName)

	return
}

//
// ListCompaniesProfits lists companies by net profit: more sustainable growth
// listed first
//
func ListCompaniesProfits(db *sql.DB) error {

	list, err := companies(db)
	if err != nil {
		return fmt.Errorf("falha ao obter a lista de empresas (%v)", err)
	}

	yi, yf, err := timeRange(db)
	yi-- // Considering penultimate period
	if err != nil {
		return fmt.Errorf("falha ao obter a faixa de datas (%v)", err)
	}

	var sep string
	fmt.Printf("%20s ", " ") // Space to match company name
	for y := yi; y <= yf; y++ {
		fmt.Printf("%10d ", y)
		sep += fmt.Sprintf("%s ", strings.Repeat("-", 10))
	}
	fmt.Println()
	fmt.Printf("%20s %s\n", " ", sep)

	pt := message.NewPrinter(language.Portuguese)

	for _, c := range list {
		profits, err := companyProfits(db, c)
		if err != nil {
			return fmt.Errorf("falha ao obter lucros de %s (%v)", c, err)
		}
		if len(profits) < 4 {
			continue
		}
		profitable := true
		for i := 1; i < len(profits); i++ {
			if profits[i].profit < 0 || profits[i].profit < profits[i-1].profit {
				profitable = false
				break
			}
		}
		if !profitable {
			continue
		}

		fmt.Printf("%-20.20s %s", c, strings.Repeat(" ", (profits[0].year-yi)*11))
		for _, p := range profits {
			pt.Printf("%10.0f ", p.profit)
		}
		fmt.Println()
	}

	return nil
}

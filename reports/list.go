package reports

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dude333/rapina/parsers"
	"github.com/pkg/errors"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
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

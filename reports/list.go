package reports

import (
	"database/sql"
	"fmt"
	"math"
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
func ListCompanies(db *sql.DB) (names []string, err error) {
	info, err := companies(db)

	if err != nil {
		fmt.Println("[x] Falha:", err)
		return
	}

	if len(info) == 0 {
		err = fmt.Errorf("lista vazia")
		return
	}

	// Extract companies names
	names = make([]string, len(info))
	for i, co := range info {
		names[i] = co.name
	}

	// Sort accents correctly
	cl := collate.New(language.BrazilianPortuguese, collate.Loose)
	cl.SortStrings(names)

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
func ListCompaniesProfits(db *sql.DB, rate float32) error {

	info, err := companies(db)
	if err != nil {
		return fmt.Errorf("falha ao obter a lista de empresas (%v)", err)
	}

	yi, yf, err := timeRange(db)
	if err != nil {
		return fmt.Errorf("falha ao obter a faixa de datas (%v)", err)
	}

	// Header
	var sep string
	fmt.Printf("%20s ", " ") // Space to match company name
	for y := yi; y <= yf; y++ {
		fmt.Printf("%10d ", y)
		sep += fmt.Sprintf("%s ", strings.Repeat("-", 10))
	}
	fmt.Printf("%10s\n", "CAGR")
	sep += fmt.Sprintf("%s ", strings.Repeat("-", 10))
	fmt.Printf("%20s %s\n", " ", sep)

	// Profits
	pt := message.NewPrinter(language.Portuguese)
	for _, co := range info {
		profits, err := companyProfits(db, co.id)
		if err != nil {
			return fmt.Errorf("falha ao obter lucros de %s (%v)", co.name, err)
		}

		// FILTERS ----------------------------
		// At least 4 years
		if len(profits) < 4 {
			continue
		}
		// Ignore if there is no recent data
		if profits[len(profits)-1].year < yf-1 {
			continue
		}

		pi := profits[0].profit
		pf := profits[len(profits)-1].profit

		if pf < pi {
			continue
		}

		// Ignore if next profix < 'rate' * current profit
		profitable := true
		for i := 1; i < len(profits); i++ {
			if profits[i].profit < 0 || profits[i].profit < (1+rate)*profits[i-1].profit {
				profitable = false
				break
			}
		}
		if !profitable {
			continue
		}

		// COMPANY NAME -----------------------
		fmt.Printf("%-20.20s ", co.name)

		// PROFIT VALUES ----------------------
		i := 0
		for y := yi; y <= yf; y++ {
			if i < len(profits) && profits[i].year == y {
				pt.Printf("%10.0f ", profits[i].profit)
				i++
			} else {
				fmt.Printf("%10s ", " ")
			}
		}

		// CAGR -------------------------------
		if pi != 0 && pf != 0 && pf*pi >= 0 {
			cagr := math.Pow(float64(pf/pi), 1/float64(yf-yi-1)) - 1
			pt.Printf("%10.1f%%", cagr*100)
		}
		fmt.Println()

	} // next c

	fmt.Printf("\nEmpresas com lucros crescentes e variação mínima de %0.0f%% de um ano para o outro.\n\n", rate*100)

	return nil
}

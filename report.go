package rapina

import (
	"fmt"
	"sort"

	"github.com/dude333/rapina/parsers"
	"github.com/dude333/rapina/reports"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

const outputDir = "reports"

//
// Report a company from DB to Excel
//
func Report(company string, path, yamlFile string) (err error) {

	db, err := openDatabase()
	if err != nil {
		return errors.Wrap(err, "fail to open db")
	}

	if path == "" {
		path = outputDir
	}

	return reports.Report(db, company, path, yamlFile)
}

//
// ListCompanies a company from DB to Excel
//
func ListCompanies() (err error) {
	db, err := openDatabase()
	if err != nil {
		return errors.Wrap(err, "fail to open db")
	}

	com, err := reports.ListCompanies(db)
	if err != nil {
		return errors.Wrap(err, "erro ao listar empresas")
	}
	for _, c := range com {
		fmt.Println(c)
	}

	return
}

//
// ListSector shows all companies from the same sector as 'company'
//
func ListSector(company, yamlFile string) (err error) {
	db, err := openDatabase()
	if err != nil {
		return errors.Wrap(err, "fail to open db")
	}

	err = reports.ListSector(db, company, yamlFile)
	if err != nil {
		return errors.Wrap(err, "erro ao listar empresas")
	}

	return
}

//
// SelectCompany returns the company name compared to the names
// stored in the DB
//
func SelectCompany(company string, scriptMode bool) string {
	db, err := openDatabase()
	if err != nil {
		fmt.Println("[x]", err)
		return ""
	}

	com, err := reports.ListCompanies(db)
	if err != nil {
		fmt.Println("[x]", err)
		return ""
	}

	company = parsers.RemoveDiacritics(company)

	// Do a fuzzy match on the company name against
	// all companies listed on the DB
	matches := make([]string, 0, 10)
	for _, c := range com {
		if fuzzy.MatchFold(company, parsers.RemoveDiacritics(c)) {
			matches = append(matches, c)
		}
	}

	// Script mode
	if len(matches) >= 1 && scriptMode {
		rank := fuzzy.RankFindFold(company, matches)
		if len(rank) <= 0 {
			return ""
		}
		sort.Sort(rank)
		return rank[0].Target
	}

	// Interactive menu
	if len(matches) >= 1 {
		result := promptUser(matches)
		return result
	}

	return ""
}

//
// promptUser presents a navigable list to be selected on CLI
//
func promptUser(list []string) (result string) {
	templates := &promptui.SelectTemplates{
		Help: `{{ "Use estas teclas para navegar:" | faint }} {{ .NextKey | faint }} ` +
			`{{ .PrevKey | faint }} {{ .PageDownKey | faint }} {{ .PageUpKey | faint }} ` +
			`{{ if .Search }} {{ "and" | faint }} {{ .SearchKey | faint }} {{ "toggles search" | faint }}{{ end }}`,
	}

	prompt := promptui.Select{
		Label:     "Selecione a Empresa",
		Items:     list,
		Templates: templates,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	return
}

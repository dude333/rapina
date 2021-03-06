package rapina

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dude333/rapina/reports"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

const outputDir = "reports"

//
// Report a company from DB to Excel
//
func Report(p Parms) (err error) {

	db, err := openDatabase()
	if err != nil {
		return errors.Wrap(err, "fail to open db")
	}

	if p.OutputDir == "" {
		p.OutputDir = outputDir
	}

	file, err := filename(p.OutputDir, p.Company)
	if err != nil {
		return err
	}

	parms := reports.Parms{
		DB:       db,
		Company:  p.Company,
		Filename: file,
		YamlFile: p.YamlFile,
		Reports:  p.Reports,
	}
	return reports.Report(parms)
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
// ListCompaniesProfits lists companies profits
//
func ListCompaniesProfits(rate float32) (err error) {
	db, err := openDatabase()
	if err != nil {
		return errors.Wrap(err, "fail to open db")
	}

	err = reports.ListCompaniesProfits(db, rate)
	if err != nil {
		return errors.Wrap(err, "erro ao listar lucros")
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

	companies, err := reports.ListCompanies(db)
	if err != nil {
		fmt.Println("[x]", err)
		return ""
	}

	// Do a fuzzy match on the company name against
	// all companies listed on the DB
	matches := make([]string, 0, 10)
	for _, c := range companies {
		if fuzzy.MatchNormalizedFold(company, c) {
			matches = append(matches, c)
		}
	}

	// Script mode
	if len(matches) >= 1 && scriptMode {
		rank := fuzzy.RankFindNormalizedFold(company, matches)
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

//
// filename cleans up the filename and returns the path/filename
func filename(path, name string) (filepath string, err error) {
	clean := func(r rune) rune {
		switch r {
		case ' ', ',', '/', '\\':
			return '_'
		}
		return r
	}
	path = strings.TrimSuffix(path, "/")
	name = strings.TrimSuffix(name, ".")
	name = strings.Map(clean, name)
	filepath = path + "/" + name + ".xlsx"

	const max = 50
	var x int
	for x = 1; x <= max; x++ {
		_, err = os.Stat(filepath)
		if err == nil {
			// File exists, try again with another name
			filepath = fmt.Sprintf("%s/%s(%d).xlsx", path, name, x)
		} else if os.IsNotExist(err) {
			err = nil // reset error
			break
		} else {
			err = fmt.Errorf("file %s stat error: %v", filepath, err)
			return
		}
	}

	if x > max {
		err = fmt.Errorf("remova o arquivo %s/%s.xlsx antes de continuar", path, name)
		return
	}

	// Create directory
	_ = os.Mkdir(path, os.ModePerm)

	// Check if the directory was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.Wrap(err, "diretório não pode ser criado")
	}

	return
}

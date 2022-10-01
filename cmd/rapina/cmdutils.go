package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

// Directory where the DB and downloaded files are stored
const dataDir = ".data"
const yamlFile = "./setores.yml"

// Parms holds the input parameters
type Parms struct {
	// Company name to be processed
	Company string
	// SpcfctnCd to indentify the ticker
	SpcfctnCd string
	// Report format (xlsx/stdout)
	Format string
	// OutputDir: path of the output xlsx
	OutputDir string
	// YamlFile: file with the companies' sectors
	YamlFile string
	// Reports is a map with the reports and reports items to be printed
	Reports map[string]bool
}

//
// openDatabase to be used by parsers and reporting
//
func openDatabase() (db *sql.DB, err error) {
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return nil, err
	}
	connStr := "file:" + dataDir + "/rapina.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
	db, err = sql.Open("sqlite3", connStr)
	if err != nil {
		return db, errors.Wrap(err, "database open failed")
	}
	db.SetMaxOpenConns(1)

	return
}

//
// promptUser presents a navigable list to be selected on CLI
//
func promptUser(list []string, label string) (result string) {
	if label == "" {
		label = "Selecione a Empresa"
	}
	templates := &promptui.SelectTemplates{
		Help: `{{ "Use estas teclas para navegar:" | faint }} {{ .NextKey | faint }} ` +
			`{{ .PrevKey | faint }} {{ .PageDownKey | faint }} {{ .PageUpKey | faint }} ` +
			`{{ if .Search }} {{ "and" | faint }} {{ .SearchKey | faint }} {{ "toggles search" | faint }}{{ end }}`,
	}

	prompt := promptui.Select{
		Label:     label,
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
func filename(path, name string) (fpath string, err error) {
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
	fpath = filepath.FromSlash(path + "/" + name + ".xlsx")

	const max = 50
	var x int
	for x = 1; x <= max; x++ {
		_, err = os.Stat(fpath)
		if err == nil {
			// File exists, try again with another name
			fpath = fmt.Sprintf("%s/%s(%d).xlsx", path, name, x)
		} else if os.IsNotExist(err) {
			err = nil // reset error
			break
		} else {
			err = fmt.Errorf("file %s stat error: %v", fpath, err)
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

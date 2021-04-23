// Copyright © 2018 Adriano P
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package fetch

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dude333/rapina/parsers"
	"github.com/dustin/go-humanize"
	_ "github.com/mattn/go-sqlite3" // requires CGO_ENABLED=1 and gcc
	"github.com/pkg/errors"
)

var (
	// ErrFileNotFound error
	ErrFileNotFound = errors.New("file not found")
	// ErrItemNotFound for string not found on []string
	ErrItemNotFound = errors.New("item not found")
)

//
// CVM fetches all statements from a range
// of years
//
func CVM(db *sql.DB, dataDir string) error {
	now := time.Now().Year()
	try(processQuarterlyReport, db, dataDir, "Arquivo ITR não encontrado", now, now-1, 2)
	try(processAnnualReport, db, dataDir, "Arquivo DFP não encontrado", now-1, 2009, 2)
	try(processFREReport, db, dataDir, "Arquivo FRE não encontrado", now-1, 2009, 2)

	return nil
}

type fn func(*sql.DB, string, int) error

// try to run the function 'f' 'n' times, in case there are network errors.
func try(f fn, db *sql.DB, dataDir, errMsg string, now, limit, n int) {
	tries := n
	var err error

	for year := now; tries > 0 && year >= limit; year-- {
		fmt.Printf("[>] %d ---------------------\n", year)
		err = f(db, dataDir, year)
		if err == ErrFileNotFound {
			fmt.Printf("[x] %s\n", errMsg)
			tries--
			continue
		} else if err != nil {
			fmt.Printf("[x] Erro ao processar arquivo de %d: %v\n", year, err)
			tries--
		} else {
			tries = n
		}
	}
}

// processAnnualReport will get data from .zip files downloaded
// directly from CVM and insert its data into the DB
func processAnnualReport(db *sql.DB, dataDir string, year int) error {

	url := fmt.Sprintf("http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/DADOS/dfp_cia_aberta_%d.zip", year)
	zipfile := fmt.Sprintf("%s/dfp_%d.zip", dataDir, year)

	// Download files from CVM server
	fmt.Print("[          ] Download do arquivo DFP")
	files, err := fetchFiles(url, dataDir, zipfile)
	if err != nil {
		return err
	}

	dataTypes := []string{"BPA", "BPP", "DRE", "DFC_MD", "DFC_MI", "DVA"}

	for _, dt := range dataTypes {
		pattern := fmt.Sprintf("dfp_cia_aberta_%s_con_%d.csv", dt, year)
		reqFile, err := findFile(files, pattern)
		if err == ErrItemNotFound {
			filesCleanup(files)
			return fmt.Errorf("arquivo %s não encontrado", reqFile)
		}
		files, _ = removeItem(files, reqFile)

		// Import file into DB
		if err = parsers.ImportCsv(db, dt, reqFile); err != nil {
			return err
		}
	}

	filesCleanup(files) // remove remaining (unused) files

	return nil
}

//
// processQuarterlyReport download quarter files from CVM and store them on DB
//
func processQuarterlyReport(db *sql.DB, dataDir string, year int) error {

	url := fmt.Sprintf("http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/ITR/DADOS/ITR_CIA_ABERTA_%d.zip", year)
	zipfile := fmt.Sprintf("%s/itr_%d.zip", dataDir, year)

	// Download files from CVM server
	fmt.Print("[          ] Download do arquivo ITR")
	files, err := fetchFiles(url, dataDir, zipfile)
	if err != nil {
		return err
	}

	dataTypes := []string{"BPA", "BPP", "DRE", "DFC_MD", "DFC_MI", "DVA"}

	for _, dt := range dataTypes {
		pattern := fmt.Sprintf("ITR_CIA_ABERTA_%s_con_%d.csv", dt, year)
		reqFile, err := findFile(files, pattern)
		if err == ErrItemNotFound {
			filesCleanup(files)
			return fmt.Errorf("arquivo %s não encontrado", reqFile)
		}
		files, _ = removeItem(files, reqFile)

		// Import file into DB (the trick is to add ITR to the data type so the
		// ImportCSV loads that into the ITR table)
		if err = parsers.ImportCsv(db, dt+"_ITR", reqFile); err != nil {
			return err
		}
	}

	filesCleanup(files) // remove remaining (unused) files

	return nil
}

//
// processFREReport download FRE (Reference Form) files from CVM and store
// them on DB.
//
func processFREReport(db *sql.DB, dataDir string, year int) error {
	url := fmt.Sprintf("http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/FRE/DADOS/fre_cia_aberta_%d.zip", year)
	zipfile := fmt.Sprintf("%s/fre_%d.zip", dataDir, year)

	// Download files from CVM server
	fmt.Print("[          ] Download do arquivo FRE")
	files, err := fetchFiles(url, dataDir, zipfile)
	if err != nil {
		return err
	}

	patterns := []string{"fre_cia_aberta_distribuicao_capital_%d.csv"}

	for _, p := range patterns {
		pattern := fmt.Sprintf(p, year)
		reqFile, err := findFile(files, pattern)
		if err == ErrItemNotFound {
			filesCleanup(files)
			return fmt.Errorf("arquivo %s não encontrado", reqFile)
		}
		files, _ = removeItem(files, reqFile)

		if err = parsers.ImportCsv(db, "FRE", reqFile); err != nil {
			return err
		}

	}

	filesCleanup(files) // remove remaining (unused) files

	return nil
}

//
// fetchFiles on CVM server
//
func fetchFiles(url, dataDir string, zipfile string) ([]string, error) {

	// Download file from CVM server
	err := downloadFile(url, zipfile)
	fmt.Println()
	if err != nil {
		return nil, ErrFileNotFound
	}

	// Unzip and list files
	files, err := Unzip(zipfile, dataDir)
	os.Remove(zipfile)
	if err != nil {
		return nil, errors.Wrap(err, "could not unzip file")
	}

	return files, nil
}

// WriteCounter counts the number of bytes written the io.Writer.
// source: https://golangcode.com/download-a-file-with-progress/
type WriteCounter struct {
	Total uint64
}

// Write implements the io.Writer interface and will be passed to io.TeeReader().
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.printProgress()
	return n, nil
}

func (wc WriteCounter) printProgress() {
	fmt.Printf("\r[  %7s", humanize.Bytes(wc.Total))
}

//
// downloadFile source: https://stackoverflow.com/a/33853856/276311
//
func downloadFile(url, filepath string) (err error) {
	// Create dir if necessary
	basepath := path.Dir(filepath)
	if err = os.MkdirAll(basepath, os.ModePerm); err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}

	// https://www.joeshaw.org/dont-defer-close-on-writable-files/
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	return
}

//
// Sectors checks if the configuration file is already populated.
// If 'force' is set or if the config is empty, it retrieves data from B3,
// unzip and extract a spreadsheet containing a list of companies divided by
// sector, subsector, and segment; then this info is set into the config file.
//
func Sectors(yamlFile string) (err error) {
	err = parsers.SectorsToYaml(yamlFile)

	return
}

//
// filesCleanup
//
func filesCleanup(files []string) {
	// Clean up
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			fmt.Println("could not delete file", f)
		}
	}
}

//
// find returns the smallest index i at which x == a[i],
// or -1 if there is no such index.
//
func find(a []string, x string) int {
	for i, n := range a {
		n = strings.Replace(n, "\\", "/", -1)
		if x == n {
			return i
		}
	}
	return -1
}

//
// removeItem removes 'item' from 'list' (changes the list order)
//
func removeItem(list []string, item string) ([]string, error) {
	if len(list) == 0 {
		return list, nil
	}
	idx := find(list, item)
	if idx == -1 {
		return list, ErrItemNotFound
	}
	list[idx] = list[len(list)-1]  // Replace it with the last one.
	return list[:len(list)-1], nil // Chop off the last one.
}

//
// findFile finds an item on list that matches pattern (case insensitive)
//
func findFile(list []string, pattern string) (string, error) {

	for i := range list {
		f := filepath.Base(list[i])
		if strings.EqualFold(f, pattern) {
			return list[i], nil
		}
	}

	return "", ErrItemNotFound
}

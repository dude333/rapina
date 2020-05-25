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

package rapina

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dude333/rapina/parsers"
	_ "github.com/mattn/go-sqlite3" // requires CGO_ENABLED=1 and gcc
	"github.com/pkg/errors"
)

// Directory where the DB and downloaded files are stored
const dataDir = ".data"

var (
	// ErrFileNotFound error
	ErrFileNotFound = errors.New("file not found")
	// ErrItemNotFound for string not found on []string
	ErrItemNotFound = errors.New("item not found")
)

//
// FetchCVM fetches all statements from a range
// of years
//
func FetchCVM() (err error) {
	db, err := openDatabase()
	if err != nil {
		return err
	}

	now := time.Now().Year()
	tries := 2
OUTER_QTR:
	for year := now; tries > 0 && year >= (now-1); year-- {
		fmt.Printf("[>] %d ---------------------\n", year)
		err := processQuarterlyReport(db, year)
		if err == ErrFileNotFound {
			fmt.Println("[x] Arquivo ITR não encontrado")
			tries--
			continue OUTER_QTR
		} else if err != nil {
			fmt.Printf("[x] Erro ao processar arquivo: %v\n", err)
			tries--
		} else {
			tries = 2
		}
	}

	tries = 2
OUTER:
	for year := now - 1; tries > 0 && year >= 2009; year-- {
		fmt.Printf("[>] %d ---------------------\n", year)
		for _, report := range []string{"BPA", "BPP", "DRE", "DFC_MD", "DFC_MI", "DVA"} {
			err := processAnnualReport(db, report, year)
			if err == ErrFileNotFound {
				fmt.Printf("[x] ---- Arquivo %s não encontrado", report)
				tries--
				continue OUTER
			} else if err != nil {
				fmt.Printf("[x] Erro ao processar %s de %d: %v\n", report, year, err)
				tries--
			} else {
				tries = 2
			}
		}
	}

	return
}

// processAnnualReport will get data from .zip files downloaded
// directly from CVM and insert its data into the DB
func processAnnualReport(db *sql.DB, dataType string, year int) error {

	dt := strings.ToLower(dataType)
	url := fmt.Sprintf("http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/%s/DADOS/%s_cia_aberta_%d.zip", dataType, dt, year)
	zipfile := fmt.Sprintf("%s/%s_%d.zip", dataDir, dt, year)
	reqFile := fmt.Sprintf("%s/%s_cia_aberta_con_%d.csv", dataDir, dt, year)

	// Download files from CVM server
	fmt.Print("[ ] Download do arquivo ", dataType)
	files, err := fetchFiles(url, zipfile)
	if err != nil {
		return err
	}

	// Keep 'reqFile' and remove all other files
	idx := find(files, reqFile)
	if idx == -1 {
		filesCleanup(files)
		return ErrFileNotFound
	}
	files[idx] = files[len(files)-1] // Replace it with the last one.
	files = files[:len(files)-1]     // Chop off the last one.
	filesCleanup(files)

	// Import file into DB
	if err = parsers.ImportCsv(db, dataType, reqFile); err != nil {
		return err
	}

	return nil
}

//
// processQuarterlyReport download quarter files from CVM and store them on DB
//
func processQuarterlyReport(db *sql.DB, year int) error {

	url := fmt.Sprintf("http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/ITR/DADOS/ITR_CIA_ABERTA_%d.zip", year)
	zipfile := fmt.Sprintf("%s/itr_%d.zip", dataDir, year)

	// Download files from CVM server
	fmt.Print("[ ] Download do arquivo ITR")
	files, err := fetchFiles(url, zipfile)
	if err != nil {
		return err
	}

	dataTypes := []string{"BPA", "BPP", "DRE", "DFC_MD", "DFC_MI", "DVA"}

	for _, dt := range dataTypes {
		pattern := fmt.Sprintf("%s/ITR_CIA_ABERTA_%s_con_%d.csv", dataDir, dt, year)
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
// fetchFiles on CVM server
//
func fetchFiles(url, zipfile string) ([]string, error) {

	// Download file from CVM server
	err := downloadFile(url, zipfile)
	if err != nil {
		fmt.Println("\r[x")
		return nil, ErrFileNotFound
	}
	fmt.Println("\r[√")

	// Unzip and list files
	files, err := Unzip(zipfile, dataDir)
	os.Remove(zipfile)
	if err != nil {
		return nil, errors.Wrap(err, "could not unzip file")
	}

	return files, nil
}

//
// downloadFile source: https://stackoverflow.com/a/33853856/276311
//
func downloadFile(url, filepath string) (err error) {
	// Create dir if necessary
	basepath := path.Dir(filepath)
	os.MkdirAll(basepath, os.ModePerm)
	//
	err = nil
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

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
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

//
// FetchSectors checks if the configuration file is already populated.
// If 'force' is set or if the config is empty, it retrieves data from B3,
// unzip and extract a spreadsheet containing a list of companies divided by
// sector, subsector, and segment; then this info is set into the config file.
//
func FetchSectors(yamlFile string) (err error) {
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
		if strings.EqualFold(list[i], pattern) {
			return list[i], nil
		}
	}

	return "", ErrItemNotFound
}

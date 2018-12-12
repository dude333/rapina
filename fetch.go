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

	"github.com/dude333/rapina/parsers"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

const dataDir = ".data"

//
// FetchCVM fetches all statements from a range
// of years
//
func FetchCVM(begin, end int) (err error) {
	// Check year
	if begin < 1900 || begin > 2100 || end < 1900 || end > 2100 {
		return errors.Wrap(err, "ano inválido")
	}
	if begin > end {
		aux := end
		end = begin
		begin = aux
	}

	// fetchB3()

	db, err := openDatabase()
	if err != nil {
		return err
	}

	for year := begin; year <= end; year++ {
		fmt.Printf("[✓] %d ---------------------\n", year)
		for _, report := range []string{"BPA", "BPP", "DRE", "DFC_MD", "DFC_MI", "DVA"} {
			if err = processReport(db, report, year); err != nil {
				fmt.Printf("[x] Erro ao processar %s de %d: %v\n", report, year, err)
			}
		}
	}

	fmt.Print("[ ] Inserindo código das contas")
	err = parsers.CodeAccounts(db)
	if err == nil {
		fmt.Print("\r[✓")
	} else {
		fmt.Print("\r[x")
	}
	fmt.Println()

	return
}

// processReport will get data from .zip files downloaded
// directly from CVM and insert its data into the DB
func processReport(db *sql.DB, dataType string, year int) (err error) {
	var file string

	if file, err = fetchFile(dataType, year); err != nil {
		return err
	}
	if err = parsers.Exec(db, dataType, file); err != nil {
		return err
	}

	return
}

//
// downloadFile source: https://stackoverflow.com/a/33853856/276311
//
func downloadFile(filepath string, url string) (err error) {
	// Create dir if necessary
	basepath := path.Dir(filepath)
	os.MkdirAll(basepath, os.ModePerm)

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
		return errors.Wrapf(err, "bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

//
// fetchFile on CVM server
//
func fetchFile(dataType string, year int) (reqFile string, err error) {
	dt := strings.ToLower(dataType)
	url := fmt.Sprintf("http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/%s/DADOS/%s_cia_aberta_%d.zip", dataType, dt, year)
	zipfile := fmt.Sprintf("%s/%s_%d.zip", dataDir, dt, year)
	reqFile = fmt.Sprintf("%s/%s_cia_aberta_con_%d.csv", dataDir, dt, year)

	// Check if files already exists
	if _, err := os.Stat(reqFile); !os.IsNotExist(err) {
		return reqFile, nil
	}

	// Download file from CVM server
	fmt.Printf("[ ] Baixando %s %d\r", dataType, year)
	err = downloadFile(zipfile, url)
	if err != nil {
		fmt.Println("[x")
		return "", errors.Wrap(err, "could not download file")
	}
	fmt.Println("[✓")

	// Unzip and list files
	var files []string
	files, err = Unzip(zipfile, dataDir)
	if err != nil {
		os.Remove(zipfile)
		return "", errors.Wrap(err, "could not unzip file")
	}
	files = append(files, zipfile)

	// File pattern:
	// xxx_cia_aberta_con_yyy.csv
	idx := find(files, reqFile)
	if idx == -1 {
		filesCleanup(files)
		return "", errors.Errorf("file %s not found", reqFile)
	}

	files[idx] = files[len(files)-1] // Replace it with the last one.
	files = files[:len(files)-1]     // Chop off the last one.
	filesCleanup(files)

	return
}

//
// fetchB3 downloads the sectoral classification file from B3
//
func fetchB3() (xlsxfile string, err error) {
	xlsxfile = dataDir + "/setorial.xlsx"
	fmt.Print("[ ] Baixando arquivo de classificação setorial da B3\r")
	zipfile := dataDir + "/sectorial.zip"

	// Check if files already exists
	if _, err := os.Stat(zipfile); !os.IsNotExist(err) {
		return xlsxfile, nil
	}

	// Download file from B3 server
	// TODO: check file url as it can be updated
	err = downloadFile(zipfile, "http://www.b3.com.br/lumis/portal/file/fileDownload.jsp?fileId=8AA8D0975A2D7918015A3C81693D4CA4")
	if err != nil {
		fmt.Println("[x")
		return
	}
	fmt.Println("[✓")

	// Unzip and list files
	var files []string
	files, err = Unzip(zipfile, dataDir)
	if err != nil {
		return "", errors.Wrap(err, "could not unzip file")
	}
	if len(files) <= 0 {
		return "", errors.Wrap(err, "zip file is empty")
	}
	files = append(files, zipfile)

	// Considering  there is only one file
	os.Remove(xlsxfile)
	os.Rename(files[0], xlsxfile)

	files[0] = files[len(files)-1] // Replace it with the last one.
	files = files[:len(files)-1]   // Chop off the last one.
	filesCleanup(files)

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

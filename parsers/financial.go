// financial.go
// Parses data from csv files containing financial statements

package parsers

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var (
	// ErrAccumITR error for accumulatd quarterly results
	ErrAccumITR = fmt.Errorf("accumulated quarterly results")
)

//
// ImportCsv start the data import process, including the database creation
// if necessary
//
func ImportCsv(db *sql.DB, dataType string, file string) (err error) {

	// Create status table
	if err = createTable(db, "STATUS"); err != nil {
		return err
	}

	// Create companies table
	if err = createTable(db, "COMPANIES"); err != nil {
		return err
	}

	// Check table version, wipe it if version differs from current version, and
	// (re)create the table
	for _, t := range []string{dataType, "MD5"} {
		if v, table := dbVersion(db, t); v != currentDbVersion {
			if v > 0 {
				fmt.Printf("[i] Apagando tabela %s versão %d (versão atual: %d)\n", table, v, currentDbVersion)
			}
			wipeDB(db, t)
		}
		if err = createTable(db, t); err != nil {
			return err
		}

	}

	isNew, err := isNewFile(db, file)
	if !isNew && err == nil { // if error, process file
		fmt.Printf("[ ] %s já processado anteriormente\n", dataType)
		return
	}

	if dataType == "FRE" {
		err = populateFRE(db, file)
	} else {
		err = populateTable(db, dataType, file)
	}
	if err == nil {
		fmt.Print("\r[√")
		storeFile(db, file)
	} else {
		fmt.Print("\r[x")
	}
	fmt.Println()

	return err
}

//
// populateTable loop thru file and insert its lines into DB
//
func populateTable(db *sql.DB, dataType, file string) (err error) {
	progress := []string{"/", "-", "\\", "|", "-", "\\"}
	p := 0

	table, err := whatTable(dataType)
	if err != nil {
		return err
	}

	companies, _ := loadCompanies(db)

	fh, err := os.Open(file)
	if err != nil {
		return errors.Wrapf(err, "erro ao abrir arquivo %s", file)
	}
	defer fh.Close()

	dec := transform.NewReader(fh, charmap.ISO8859_1.NewDecoder())

	// BEGIN TRANSACTION
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "Failed to begin transaction")
	}

	// Data used inside loop
	sep := func(r rune) bool {
		return r == ';'
	}
	header := make(map[string]int) // stores the header item position (e.g., DT_FIM_EXERC:9)
	scanner := bufio.NewScanner(dec)
	count := 0
	insert := ""
	var stmt *sql.Stmt

	// Loop thru file, line by line
	fmt.Print("[ ] Processando arquivo ", dataType)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		fields := strings.FieldsFunc(line, sep)

		if len(header) == 0 { // HEADER
			// Get header positioning
			for i, h := range fields {
				header[h] = i
			}
			// Prepare insert statement
			insert = fmt.Sprintf(`INSERT OR IGNORE INTO %s (
				ID, ID_CIA, CODE, YEAR, DATA_TYPE,
				VERSAO,
				MOEDA, ESCALA_MOEDA, 
				DT_FIM_EXERC,
				CD_CONTA, DS_CONTA, VL_CONTA
			) VALUES (
				?, ?, ?, ?, "%s",
				?,
				?, ?,
				?,
				?, ?, ?
				);`, table, dataType)
			stmt, err = tx.Prepare(insert)
			if err != nil {
				err = errors.Wrapf(err, "erro ao preparar insert (verificar cabeçalho do arquivo %s)", file)
				return
			}
			defer stmt.Close()

		} else { // VALUES

			if len(fields) <= 14 {
				continue
			}

			// UPDATE COMPANIES
			n1, ok1 := header["CNPJ_CIA"]
			n2, ok2 := header["DENOM_CIA"]
			if ok1 && ok2 && n1 >= 0 && n1 < len(fields) && n2 >= 0 && n2 < len(fields) {
				updateCompanies(companies, fields[header["CNPJ_CIA"]], fields[header["DENOM_CIA"]])
			}

			// INSERT
			f, err := prepareFields(dataType, header, fields, companies)
			if err == ErrAccumITR {
				continue // ignore accumulated ITR data
			}
			if err != nil {
				return errors.Wrap(err, "falha ao preparar registro")
			}
			_, err = stmt.Exec(f...)
			if err != nil {
				return errors.Wrap(err, "falha ao inserir registro")
			}
		}

		// fmt.Println("-------------------------------")
		if count++; count%1000 == 0 {
			fmt.Printf("\r[%s", progress[p%6])
			p++
		}
	}

	fmt.Print("\r[*")

	// END TRANSACTION
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Failed to commit transaction")
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrapf(err, "erro ao ler arquivo %s", file)
	}

	saveCompanies(db, companies)

	return
}

//
// prepareFields prepares all fields (columns) to be inserted on the DB.
//
// Returns:
// ID, ID_CIA, CODE, YEAR,
// VERSAO,
// MOEDA, ESCALA_MOEDA,
// DT_FIM_EXERC,
// CD_CONTA, DS_CONTA, VL_CONTA
//
// Tip: to convert Unix timestamp to date on sqlite: strftime('%Y-%m-%d', DT_REFER, 'unixepoch')
//
func prepareFields(dataType string, header map[string]int, fields []string, companies map[string]company) ([]interface{}, error) {
	// AUX FUNCTIONS
	val := func(key string) string {
		v, ok := header[key]
		if !ok {
			return ""
		}
		return fields[v]
	}

	// Convert date string (YYYY-MM-DD) into Unix timestamp
	tim := func(key string) int64 {
		v, ok := header[key]
		if !ok {
			return 0
		}
		t, err := time.Parse("2006-01-02", fields[v])
		if err != nil {
			return 0
		}
		return t.Unix()
	}

	// REFERENCE DATE
	v, ok := header["DT_FIM_EXERC"]
	if !ok {
		return nil, fmt.Errorf("DT_FIM_EXERC não encontrado")
	}
	if len(fields[v]) < 4 || tim("DT_FIM_EXERC") == 0 {
		return nil, fmt.Errorf("DT_FIM_EXERC incorreto: %v", fields[v])
	}
	// Check if quarterly data contains data from 90 days, except for "BPA_ITR" and "BPP_ITR"
	if dataType != "BPA_ITR" && dataType != "BPP_ITR" && strings.HasSuffix(dataType, "_ITR") {
		t1 := tim("DT_INI_EXERC")
		t2 := tim("DT_FIM_EXERC")
		days := (t2 - t1) / 60 / 60 / 24
		if days < 80 || days > 100 {
			return nil, ErrAccumITR
		}
	}
	year := fields[v][:4]

	// CNPJ_CIA and DENOM_CIA are replaced by company id
	cnpj := val("CNPJ_CIA")
	c, ok := companies[cnpj]
	if !ok {
		return nil, fmt.Errorf("CNPJ %s não encontrado", cnpj)
	}
	companyID := c.id

	// Unique value to be used as PRIMARY KEY
	hash := Hash(cnpj + val("GRUPO_DFP") + val("DT_FIM_EXERC") + val("VERSAO") + val("CD_CONTA") + val("VL_CONTA"))

	// Output -- need to follow INSERT sequence
	var f []interface{}
	f = append(f, hash)                                                             // ID
	f = append(f, companyID)                                                        // ID_CIA
	f = append(f, acctCode(fields[header["CD_CONTA"]], fields[header["DS_CONTA"]])) // CODE
	f = append(f, year)                                                             // YEAR

	f = append(f, val("VERSAO"))
	f = append(f, val("MOEDA"))
	f = append(f, val("ESCALA_MOEDA"))
	f = append(f, tim("DT_FIM_EXERC"))
	f = append(f, val("CD_CONTA"))
	f = append(f, val("DS_CONTA"))
	f = append(f, val("VL_CONTA"))

	return f, nil
}

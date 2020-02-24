// financial.go
// Parses data from csv files containing financial statements

package parsers

import (
	"bufio"
	"database/sql"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var fnvHash = fnv.New32a()

//
// GetHash returns the FNV-1 non-cryptographic hash
//
func GetHash(s string) uint32 {
	fnvHash.Write([]byte(s))
	defer fnvHash.Reset()

	return fnvHash.Sum32()
}

//
// ImportCsv start the data import process, including the database creation
// if necessary
//
func ImportCsv(db *sql.DB, dataType string, file string) (err error) {

	// Create status table
	if err = createTable(db, "STATUS"); err != nil {
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

	// Remove indexes to speed up insertion
	// fmt.Println("[i] Apagando índices do bd para acelerarar processamento")
	// dropIndexes(db)

	fmt.Print("[ ] Processando arquivo ", dataType)
	err = populateTable(db, dataType, file)
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
	progress2 := []string{".", "!", "|", "\\", "\"", "'", " "}
	p := 0
	p2 := 0

	table, err := whatTable(dataType)
	if err != nil {
		return err
	}

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
	deleteTable := make(map[string]bool)
	scanner := bufio.NewScanner(dec)
	count := 0
	insert := ""
	var stmt, delStmt *sql.Stmt

	// Loop thru file, line by line
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
			insert = fmt.Sprintf(`INSERT OR IGNORE INTO %s (ID,CODE,DATA_TYPE,YEAR,%s) VALUES (?,?,"%s",?%s);`,
				table, strings.Join(fields, ","), dataType, strings.Repeat(",?", len(fields)))
			stmt, err = tx.Prepare(insert)
			if err != nil {
				err = errors.Wrapf(err, "erro ao preparar insert (verificar cabeçalho do arquivo %s)", file)
				return
			}
			defer stmt.Close()

			// Prepare delete statement (to avoid duplicated data in case of updated data from CVM)
			delete := fmt.Sprintf(`DELETE FROM %s WHERE CNPJ_CIA = ? AND DATA_TYPE = "%s" AND YEAR = ?;`,
				table, dataType)
			delStmt, err = tx.Prepare(delete)
			if err != nil {
				err = errors.Wrapf(err, "erro ao preparar delete")
				return
			}
			defer delStmt.Close()
		} else { // VALUES

			if len(header) != len(fields) {
				fmt.Fprintf(os.Stderr, "[x] Linha com %d campos ao invés de %d\n", len(fields), len(header))
			} else {
				// DELETE
				cnpj := fields[header["CNPJ_CIA"]]
				year := fields[header["DT_REFER"]][:4]
				if _, ok := deleteTable[cnpj+dataType+year]; !ok {
					deleteTable[cnpj+dataType+year] = true
					y, err := strconv.Atoi(year)
					if err == nil {
						fmt.Printf("\r[%s", progress2[p2%7])
						p2++
						delStmt.Exec(cnpj, y)
					}
				}

				// INSERT
				hash := GetHash(line)
				f, err := prepareFields(hash, header, fields)
				_, err = stmt.Exec(f...)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		// fmt.Println("-------------------------------")
		if count++; count%1000 == 0 {
			fmt.Printf("\r[%s", progress[p%6])
			p++
		}
	}

	// END TRANSACTION
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Failed to commit transaction")
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrapf(err, "erro ao ler arquivo %s", file)
	}

	return
}

//
// prepareFields changes dates from 'YYYY-MM-DD' to Unix epoch, set the code
// based on CD_CONTA and DS_CONTA, and set the DATA_TYPE with current time.
// Returns: ID, CODE, DATA_TYPE, <fields from file following header order>.
//
// To convert date on sqlite: strftime('%Y-%m-%d', DT_REFER, 'unixepoch')
//
func prepareFields(hash uint32, header map[string]int, fields []string) (f []interface{}, err error) {
	list := []string{"DT_REFER", "DT_INI_EXERC", "DT_FIM_EXERC"}
	layout := "2006-01-02"

	for _, dt := range list {
		if i, ok := header[dt]; ok {
			var t time.Time
			t, err = time.Parse(layout, fields[i])
			if err != nil {
				err = errors.Wrap(err, "data invalida "+fields[i])
				return
			}
			fields[i] = fmt.Sprintf("%v", t.Unix())
		}
	}

	year, err := strconv.Atoi(fields[header["DT_REFER"]][:4])
	if err != nil {
		errors.Wrapf(err, "registro com ano errado (%s)", fields[header["DT_REFER"]][:4])
		return
	}

	f = make([]interface{}, len(fields)+3)
	f[0] = hash                                                             // ID
	f[1] = acctCode(fields[header["CD_CONTA"]], fields[header["DS_CONTA"]]) // CODE
	f[2] = year
	for i, v := range fields {
		f[i+3] = v
	}

	return
}

//
// RemoveDiacritics transforms, for example, "žůžo" into "zuzo"
//
func RemoveDiacritics(original string) (result string) {
	isMn := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
	}

	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ = transform.String(t, original)

	return
}

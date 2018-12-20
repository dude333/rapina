package parsers

import (
	"bufio"
	"database/sql"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
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
// Exec start the data import process, including the database creation
// if necessary
//
func Exec(db *sql.DB, dataType string, file string) (err error) {
	err = createTable(db, dataType)
	if err != nil {
		return err
	}

	createTable(db, "MD5")

	isNew, err := isNewFile(db, file)
	if !isNew && err == nil { // if error then process file
		fmt.Printf("[ ] %s já processado anteriormente\n", dataType)
		return
	}

	fmt.Print("[ ] Processando arquivo ", dataType)
	err = populateTable(db, dataType, file)
	if err == nil {
		fmt.Print("\r[✓")
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
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.FieldsFunc(line, sep)

		if len(header) == 0 { // HEADER
			// Get header positioning
			for i, h := range fields {
				header[h] = i
			}
			// Prepare insert statement
			insert = fmt.Sprintf(`INSERT OR IGNORE INTO %s (ID,CODE,%s) VALUES (?,?%s);`,
				table, strings.Join(fields, ","), strings.Repeat(",?", len(fields)))
			stmt, err = tx.Prepare(insert)
			if err != nil {
				err = errors.Wrap(err, "Erro ao preparar insert")
				return
			}
			defer stmt.Close()

		} else { // VALUES

			if len(header) != len(fields) {
				fmt.Fprintf(os.Stderr, "[x] Linha com %d campos ao invés de %d\n", len(fields), len(header))
			} else {
				hash := GetHash(line)
				code := GetHash(fields[header["CD_CONTA"]] + fields[header["DS_CONTA"]])
				f, err := prepareFields(header, hash, code, fields)
				_, err = stmt.Exec(f...)
				if err != nil {
					log.Fatal(err)
				}
				// if err = insertLine(tx, dataType, &header, fields); err != nil {
				// fmt.Printf("[x] %s: %v\n", dataType, err)
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
// prepareFields changes date from 'YYYY-MM-DD' to Unix epoch
// To convert back on sqlite: strftime('%Y-%m-%d', DT_REFER, 'unixepoch')
//
func prepareFields(header map[string]int, hash, code uint32, fields []string) (f []interface{}, err error) {
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

	f = make([]interface{}, len(fields)+2)
	f[0] = hash
	f[1] = code
	for i, v := range fields {
		f[i+2] = v
	}

	return
}

//
// toUtf8 convert iso8859-1 to utf8
// https://stackoverflow.com/a/13511463/276311
//
func toUtf8(iso8859_1_buf []byte) string {
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

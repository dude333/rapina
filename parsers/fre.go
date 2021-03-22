package parsers

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var (
	// ErrCNPJNotFound error
	ErrCNPJNotFound = fmt.Errorf("CNPJ not found")
)

func populateFRE(db *sql.DB, file string) (int, error) {
	progress := []string{"/", "-", "\\", "|", "-", "\\"}
	p := 0
	var err error

	table, err := whatTable("FRE")
	if err != nil {
		return 0, err
	}

	companies, _ := loadCompanies(db)

	fh, err := os.Open(file)
	if err != nil {
		return 0, errors.Wrapf(err, "erro ao abrir arquivo %s", file)
	}
	defer fh.Close()

	dec := transform.NewReader(fh, charmap.ISO8859_1.NewDecoder())

	// BEGIN TRANSACTION
	tx, err := db.Begin()
	if err != nil {
		return 0, errors.Wrap(err, "Failed to begin transaction")
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
	fmt.Print("[ ] Processando arquivo FRE")
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
				ID, ID_CIA, YEAR, 
				Versao,
				Quantidade_Total_Acoes_Circulacao, 
				Percentual_Total_Acoes_Circulacao 
			) VALUES (
				?, ?, ?,
				?,
				?,
				?
				);`, table)
			stmt, err = tx.Prepare(insert)
			if err != nil {
				err = errors.Wrapf(err, "erro ao preparar insert (verificar cabeçalho do arquivo %s)", file)
				return count, err
			}
			defer stmt.Close()

		} else { // VALUES

			if len(fields) <= 12 {
				continue
			}
			// INSERT
			f, err := prepareFREFields(header, fields, companies)
			if err == ErrCNPJNotFound {
				continue
			}
			if err != nil {
				fmt.Println(line)
				fmt.Printf("\r[x] Falha ao preparar registro: %v\n", err)
				fmt.Print("[ ] Processando arquivo FRE")
				continue
			}
			_, err = stmt.Exec(f...)
			if err != nil {
				return count, errors.Wrap(err, "falha ao inserir registro")
			}
		}

		if count++; count%60 == 0 {
			fmt.Printf("\r[%s", progress[p%6])
			p++
		}
	}

	fmt.Print("\r[*")

	// END TRANSACTION
	err = tx.Commit()
	if err != nil {
		return count, errors.Wrap(err, "Failed to commit transaction")
	}

	if err := scanner.Err(); err != nil {
		return count, errors.Wrapf(err, "erro ao ler arquivo %s", file)
	}

	return count, nil
}

func prepareFREFields(header map[string]int, fields []string, companies map[string]company) ([]interface{}, error) {
	if len(fields) < len(header)-1 {
		return nil, fmt.Errorf("len(fields)=%d != len(header)=%d", len(fields), len(header))
	}

	// val checks and gets the value from a map
	val := func(key string) string {
		v, ok := header[key]
		if !ok {
			return ""
		}
		return fields[v]
	}

	// YEAR
	v, ok := header["Data_Referencia"]
	if !ok {
		return nil, fmt.Errorf("Data_Referencia não encontrado")
	}
	if len(fields[v]) != 10 {
		return nil, fmt.Errorf("DT_FIM_EXERC incorreto: %v", fields[v])
	}
	year := fields[v][:4]

	// CNPJ_Companhia is replaced by company id
	cnpj := val("CNPJ_Companhia")
	c, ok := companies[cnpj]
	if !ok {
		return nil, ErrCNPJNotFound
	}
	companyID := c.id

	// Free float
	ff := val("Percentual_Total_Acoes_Circulacao")
	var freeFloat float32
	if ff != "" {
		if f, err := strconv.ParseFloat(ff, 32); err == nil {
			freeFloat = float32(f / 100)
		}
	}

	// Total shares considering the free float
	shares := val("Quantidade_Total_Acoes_Circulacao")
	var totalShares float32
	if shares != "" {
		if f, err := strconv.ParseFloat(shares, 32); err == nil {
			if freeFloat > 0 {
				totalShares = float32(f) / freeFloat
			}
		}
	}

	// Unique value to be used as PRIMARY KEY
	hash := Hash(cnpj + val("Data_Referencia") + val("Versao") + val("ID_Documento") + val("Quantidade_Total_Acoes_Circulacao"))

	// Output -- need to match INSERT sequence
	var f []interface{}
	f = append(f, hash)      // ID
	f = append(f, companyID) // ID_CIA
	f = append(f, year)      // YEAR

	f = append(f, val("Versao"))
	f = append(f, totalShares)
	f = append(f, freeFloat)

	return f, nil
}

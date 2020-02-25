package parsers

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

const currentDbVersion = 20022423

var createTableMap = map[string]string{
	"dfp": `CREATE TABLE IF NOT EXISTS dfp
	(
		"ID" PRIMARY KEY,
		"CODE" integer,
		"YEAR" string,
		"DATA_TYPE" string,

		"CNPJ_CIA" varchar(20),
		"DT_REFER" integer,
		"VERSAO" integer,
		"DENOM_CIA" varchar(100),
		"CD_CVM" integer,
		"GRUPO_DFP" varchar(206),
		"MOEDA" varchar(4),
		"ESCALA_MOEDA" varchar(7),
		"ESCALA_DRE" varchar(7),
		"ORDEM_EXERC" varchar(9),
		"DT_INI_EXERC" integer,
		"DT_FIM_EXERC" integer,
		"CD_CONTA" varchar(18),
		"DS_CONTA" varchar(100),
		"VL_CONTA" real
	);`,

	"codes": `CREATE TABLE IF NOT EXISTS codes
	(
		"CODE" INTEGER NOT NULL PRIMARY KEY,
		"NAME" varchar(100)
	);`,

	"md5": `CREATE TABLE IF NOT EXISTS md5
	(
		md5 NOT NULL PRIMARY KEY
	);`,

	"status": `CREATE TABLE IF NOT EXISTS status
	(
		table_name TEXT NOT NULL PRIMARY KEY,
		version integer
	);`,
}

//
// whatTable for the data type
//
func whatTable(dataType string) (table string, err error) {
	switch dataType {
	case "BPA", "BPP", "DRE", "DFC_MD", "DFC_MI", "DVA":
		table = "dfp"
	case "CODES":
		table = "codes"
	case "MD5":
		table = "md5"
	case "STATUS":
		table = "status"
	default:
		return "", errors.Wrapf(err, "tipo de informação inexistente: %s", dataType)
	}

	return
}

//
// createTable creates the table if not created yet
//
func createTable(db *sql.DB, dataType string) (err error) {

	table, err := whatTable(dataType)
	if err != nil {
		return err
	}

	_, err = db.Exec(createTableMap[table])
	if err != nil {
		return errors.Wrap(err, "erro ao criar tabela "+table)
	}

	err = createIndex(db, table)
	if err != nil {
		return errors.Wrap(err, "erro ao criar índice para table "+table)
	}

	if dataType != "STATUS" {
		_, err = db.Exec(fmt.Sprintf(`INSERT OR REPLACE INTO status (table_name, version) VALUES ("%s",%d)`, table, currentDbVersion))
		if err != nil {
			return errors.Wrap(err, "erro ao atualizar tabela "+table)
		}
	}

	return nil
}

//
// dbVersion returns the version stored in DB
//
func dbVersion(db *sql.DB, dataType string) (v int, table string) {
	table, err := whatTable(dataType)
	if err != nil {
		return
	}

	sqlStmt := `SELECT version FROM status WHERE table_name = ?`
	err = db.QueryRow(sqlStmt, table).Scan(&v)
	if err != nil {
		return
	}

	return
}

//
// wipeDB drops the table! Use with care
//
func wipeDB(db *sql.DB, dataType string) (err error) {
	table, err := whatTable(dataType)
	if err != nil {
		return
	}

	_, err = db.Exec("DROP TABLE IF EXISTS " + table)
	if err != nil {
		return errors.Wrap(err, "erro ao apagar tabela")
	}

	return
}

//
// CreateIndexes to optimize queries
// group 1: used for parsers.CodeAccounts
// group 2: used for reports
// group 3: all
//
func CreateIndexes(db *sql.DB, group int) (err error) {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS dfp_code_accounts ON dfp (CD_CONTA, DS_CONTA);",
		"CREATE INDEX IF NOT EXISTS dfp_code_accounts_ds ON dfp (DS_CONTA);",
		"CREATE INDEX IF NOT EXISTS dfp_metrics ON dfp (CODE, DENOM_CIA, ORDEM_EXERC, DT_REFER, VL_CONTA);",
		"CREATE INDEX IF NOT EXISTS dfp_date ON dfp (DT_REFER, DENOM_CIA);",
		"CREATE INDEX IF NOT EXISTS dfp_cnpj ON dfp (CNPJ_CIA, DENOM_CIA);",
	}

	groups := [3][2]int{
		{0, 1},
		{2, 4},
		{0, 4},
	}

	if group < 1 || group > len(groups) {
		group = 2
	} else {
		group--
	}

	for i := groups[group][0]; i <= groups[group][1]; i++ {
		_, err = db.Exec(indexes[i])
		if err != nil {
			return errors.Wrap(err, "erro ao criar índice")
		}
	}

	return nil
}

//
// DropIndexes created by CreateIndexes
//
func DropIndexes(db *sql.DB) (err error) {
	indexes := []string{
		"DROP INDEX IF EXISTS dfp_code_accounts;",
		"DROP INDEX IF EXISTS dfp_code_accounts_ds;",
		"DROP INDEX IF EXISTS dfp_metrics;",
		"DROP INDEX IF EXISTS dfp_date;",
		"DROP INDEX IF EXISTS dfp_cnpj;",
	}

	for _, idx := range indexes {
		_, err = db.Exec(idx)
		if err != nil {
			return errors.Wrap(err, "erro ao apagar índice")
		}
	}

	return nil
}

//
// OptimizeReport creates an index for optimize Report.
//
func OptimizeReport(db *sql.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS dfp_metrics ON dfp (CODE, DENOM_CIA, ORDEM_EXERC, DT_REFER, VL_CONTA);",
		"CREATE INDEX IF NOT EXISTS dfp_cnpj ON dfp (CNPJ_CIA, DENOM_CIA);",
	}

	for _, idx := range indexes {
		_, err := db.Exec(idx)
		if err != nil {
			return errors.Wrap(err, "erro ao criar índice")
		}
	}

	return nil
}

func createIndex(db *sql.DB, table string) error {
	if table == "dfp" {
		idx := "CREATE INDEX IF NOT EXISTS dfp_year ON dfp (YEAR, DATA_TYPE);"
		_, err := db.Exec(idx)
		if err != nil {
			return errors.Wrap(err, "erro ao criar índice")
		}
	}
	return nil
}

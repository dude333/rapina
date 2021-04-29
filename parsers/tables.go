package parsers

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

const currentDbVersion = 210305
const currentFIIDbVersion = 210426
const currentStockQuotesVersion = 210305

var createTableMap = map[string]string{
	"dfp": `CREATE TABLE IF NOT EXISTS dfp
	(
		"ID" PRIMARY KEY,
		"ID_CIA" integer,
		"CODE" integer,
		"YEAR" string,
		"DATA_TYPE" string,

		"VERSAO" integer,
		"MOEDA" varchar(4),
		"ESCALA_MOEDA" varchar(7),
		"DT_FIM_EXERC" integer,
		"CD_CONTA" varchar(18),
		"DS_CONTA" varchar(100),
		"VL_CONTA" real
	);`,

	"itr": `CREATE TABLE IF NOT EXISTS itr
	(
		"ID" PRIMARY KEY,
		"ID_CIA" integer,
		"CODE" integer,
		"YEAR" string,
		"DATA_TYPE" string,

		"VERSAO" integer,
		"MOEDA" varchar(4),
		"ESCALA_MOEDA" varchar(7),
		"DT_FIM_EXERC" integer,
		"CD_CONTA" varchar(18),
		"DS_CONTA" varchar(100),
		"VL_CONTA" real
	);`,

	"fre": `CREATE TABLE IF NOT EXISTS fre
	(
		"ID" PRIMARY KEY,
		"ID_CIA" integer,
		"YEAR" string,

		"Versao" integer,
		"Quantidade_Total_Acoes_Circulacao" integer,
		"Percentual_Total_Acoes_Circulacao" real
	);`,

	"codes": `CREATE TABLE IF NOT EXISTS codes
	(
		"CODE" INTEGER NOT NULL PRIMARY KEY,
		"NAME" varchar(100)
	);`,

	"companies": `CREATE TABLE IF NOT EXISTS companies
	(
		"ID" INTEGER NOT NULL PRIMARY KEY,
		"CNPJ" varchar(20),
		"NAME" varchar(100)
	);`,

	"md5": `CREATE TABLE IF NOT EXISTS md5
	(
		md5 NOT NULL PRIMARY KEY
	);`,

	"fii_details": `CREATE TABLE IF NOT EXISTS fii_details
	(
		cnpj TEXT NOT NULL PRIMARY KEY,
		acronym varchar(4),
		trading_code varchar(6)
	);`,

	"stock_quotes": `CREATE TABLE IF NOT EXISTS stock_quotes
	(
		stock varchar(12) NOT NULL,
		date   varchar(10) NOT NULL,
		open   real,
		high   real,
		low    real,
		close  real,
		volume real
	);`,

	"fii_dividends": `CREATE TABLE IF NOT EXISTS fii_dividends
	(
		trading_code varchar(12) NOT NULL, 
		base_date varchar(10) NOT NULL, 
		payment_date varchar(10), 
		value real
	);`,

	"status": `CREATE TABLE IF NOT EXISTS status
	(
		table_name TEXT NOT NULL PRIMARY KEY,
		version integer
	);`,
}

func allTables() []string {
	keys := make([]string, len(createTableMap))
	i := 0
	for k := range createTableMap {
		keys[i] = k
		i++
	}
	return keys
}

//
// whatTable for the data type
//
func whatTable(dataType string) (table string, err error) {
	switch dataType {
	case "dfp", "BPA", "BPP", "DRE", "DFC_MD", "DFC_MI", "DVA":
		table = "dfp"
	case "itr", "BPA_ITR", "BPP_ITR", "DRE_ITR", "DFC_MD_ITR", "DFC_MI_ITR", "DVA_ITR":
		table = "itr"
	case "fre", "FRE":
		table = "fre"
	case "codes", "CODES":
		table = "codes"
	case "md5", "MD5":
		table = "md5"
	case "status", "STATUS":
		table = "status"
	case "companies", "COMPANIES":
		table = "companies"
	case "fii_details":
		table = dataType
	case "fii_dividends":
		table = dataType
	case "stock_quotes":
		table = dataType
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

	err = createIndexes(db, table)
	if err != nil {
		return errors.Wrap(err, "erro ao criar índice para table "+table)
	}

	if dataType == "STATUS" {
		return nil
	}

	version := currentDbVersion
	switch dataType {
	case "fii_details":
		version = currentFIIDbVersion
	case "fii_dividends":
		version = currentFIIDbVersion
	case "stock_quotes":
		version = currentStockQuotesVersion
	}

	_, err = db.Exec(fmt.Sprintf(`INSERT OR REPLACE INTO status (table_name, version) VALUES ("%s",%d)`, table, version))
	if err != nil {
		return errors.Wrap(err, "erro ao atualizar tabela "+table)
	}

	return nil
}

func createAllTables(db *sql.DB) (err error) {
	if err := createTable(db, "status"); err != nil {
		return err
	}
	for _, t := range allTables() {
		if t == "status" {
			continue
		}
		if err := createTable(db, t); err != nil {
			return err
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
// createIndexes create indexes based on table name
//
func createIndexes(db *sql.DB, table string) error {
	indexes := []string{}

	switch table {
	case "dfp":
		indexes = []string{
			"CREATE INDEX IF NOT EXISTS dfp_metrics ON dfp (CODE, ID_CIA, YEAR, VL_CONTA);",
			"CREATE INDEX IF NOT EXISTS dfp_year_ver ON dfp (ID_CIA, YEAR, VERSAO);",
		}
	case "itr":
		indexes = []string{
			"CREATE INDEX IF NOT EXISTS itr_metrics ON itr (CODE, ID_CIA, YEAR, VL_CONTA);",
			"CREATE INDEX IF NOT EXISTS itr_quarter_ver ON itr (ID_CIA, DT_FIM_EXERC, VERSAO);",
		}
	case "stock_quotes":
		indexes = []string{
			"CREATE UNIQUE INDEX IF NOT EXISTS stock_quotes_stockdate ON stock_quotes (stock, date);",
		}
	case "fii_dividends":
		indexes = []string{
			"CREATE UNIQUE INDEX IF NOT EXISTS fii_dividends_pk ON fii_dividends (trading_code, base_date);",
		}
	}

	for _, idx := range indexes {
		_, err := db.Exec(idx)
		if err != nil {
			return errors.Wrap(err, "erro ao criar índice")
		}
	}

	return nil
}

//
// hasTable checks if the table exists
//
func hasTable(db *sql.DB, tableName string) bool {
	sqlStmt := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	var n string
	err := db.QueryRow(sqlStmt, tableName).Scan(&n)
	if err != nil {
		return false
	}

	return n == tableName
}

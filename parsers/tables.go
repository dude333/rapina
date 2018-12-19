package parsers

import (
	"database/sql"

	"github.com/pkg/errors"
)

var createTableMap = map[string]string{
	"dfp": `CREATE TABLE IF NOT EXISTS dfp
	(
		"ID" PRIMARY KEY,
		"CODE" integer,

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
		"CODE" PRIMARY KEY,
		"NAME" varchar(100)
	);`,

	"md5": `CREATE TABLE IF NOT EXISTS md5
	(
		"md5" PRIMARY KEY
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
		return errors.Wrap(err, "erro ao criar tabela")
	}

	return nil
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
		"CREATE INDEX IF NOT EXISTS dfp_dt_refer ON dfp (DT_REFER);",
	}

	groups := [3][2]int{
		{0, 1},
		{2, 2},
		{0, 2},
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
// DropIndexes created by createIndexes
//
func DropIndexes(db *sql.DB) (err error) {
	indexes := []string{
		"DROP INDEX IF EXISTS dfp_code_accounts;",
		"DROP INDEX IF EXISTS dfp_code_accounts_ds;",
		"DROP INDEX IF EXISTS dfp_dt_refer;",
	}

	for _, idx := range indexes {
		_, err = db.Exec(idx)
		if err != nil {
			return errors.Wrap(err, "erro ao apagar índice")
		}
	}

	return nil
}

package cmd

import (
	"database/sql"

	"github.com/pkg/errors"
)

// Directory where the DB and downloaded files are stored
const dataDir = ".data"

//
// openDatabase to be used by parsers and reporting
//
func openDatabase() (db *sql.DB, err error) {
	connStr := "file:" + dataDir + "/rapina.db?cache=shared&mode=rwc&_journal_mode=WAL"
	db, err = sql.Open("sqlite3", connStr)
	if err != nil {
		return db, errors.Wrap(err, "database open failed")
	}

	return
}

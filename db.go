package rapina

import (
	"database/sql"

	"github.com/pkg/errors"
)

//
// openDatabase to be used by parsers and reporting
//
func openDatabase() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", dataDir+"/rapina.db")
	if err != nil {
		return db, errors.Wrap(err, "database open failed")
	}

	return
}

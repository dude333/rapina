package parsers

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func TestIsNewFile(t *testing.T) {
	db, err := openDatabase()
	if err != nil {
		t.Errorf("cannot open db: %v", err)
		return
	}

	createTable(db, "MD5")

	isNew, err := isNewFile(db, "/home/adr/go/src/github.com/dude333/rapina/cli/.data/bpa_cia_aberta_con_2013.csv")

	t.Errorf("[t] %v %v", isNew, err)
}

func openDatabase() (db *sql.DB, err error) {

	db, err = sql.Open("sqlite3", "/home/adr/go/src/github.com/dude333/rapina/cli/.data/rapina.db")
	if err != nil {
		return db, errors.Wrap(err, "database open failed")
	}

	return
}

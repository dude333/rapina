package parsers

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func TestIsNewFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode") // used in CI
	}

	db, err := openDatabase()
	if err != nil {
		t.Errorf("cannot open db: %v", err)
		return
	}

	if err := createTable(db, "MD5"); err != nil {
		t.Errorf("could not create table: %v", err)
	}

	file := "../cli/.data/bpa_cia_aberta_con_2017.csv"
	isNew, err := isNewFile(db, file)
	expected := false
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		expected = true
	}

	if isNew == expected {
		t.Errorf("isNewFile returned %v. If 'rapina get' has run before it should've returned false.\nError: [%v]", expected, err)
	}
}

func openDatabase() (db *sql.DB, err error) {

	db, err = sql.Open("sqlite3", "../bin/.data/rapina.db")
	if err != nil {
		return db, errors.Wrap(err, "database open failed")
	}

	return
}

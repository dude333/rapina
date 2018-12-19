package parsers

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"os"
)

//
// isNewFile checks the database to see if this file has been
// processed already
//
func isNewFile(db *sql.DB, filename string) (isNew bool, err error) {
	isNew = true

	md5, err := md5FromFile(filename)
	if err != nil {
		return
	}

	sqlStmt := `SELECT md5 FROM md5 WHERE md5 = ?`
	err = db.QueryRow(sqlStmt, md5).Scan(&md5)
	if err != nil {
		return
	}

	isNew = false
	return
}

//
// storeFile into md5 table (only successfully processed files)
//
func storeFile(db *sql.DB, filename string) (md5 string) {
	md5, err := md5FromFile(filename)
	if err != nil {
		return ""
	}
	insert := fmt.Sprintf(`INSERT OR IGNORE INTO md5 (md5) VALUES ("%s")`, md5)
	_, err = db.Exec(insert)
	if err != nil {
		return ""
	}
	return md5
}

//
// md5FromFile
//
func md5FromFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

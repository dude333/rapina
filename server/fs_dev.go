// +build dev

package server

import (
	"io/fs"
	"log"
	"os"
)

var _fs = os.DirFS(".")
var _contentFS fs.FS

func init() {
	var err error
	_contentFS, err = fs.Sub(_fs, "templates")
	if err != nil {
		log.Fatal(err)
	}
}

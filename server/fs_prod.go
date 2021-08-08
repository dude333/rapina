// +build !dev

package server

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed templates
var _fs embed.FS
var _contentFS fs.FS

func init() {
	var err error
	_contentFS, err = fs.Sub(_fs, "templates")
	if err != nil {
		log.Fatal(err)
	}
}

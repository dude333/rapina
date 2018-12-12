package reports

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

//
// filename cleans up the filename and returns the path/filename
func filename(path, name string) (filepath string, err error) {
	clean := func(r rune) rune {
		switch r {
		case ' ', ',', '/', '\\':
			return '_'
		}
		return r
	}
	name = strings.TrimSuffix(name, ".")
	name = strings.Map(clean, name)
	filepath = path + "/" + name + ".xlsx"

	// Create directory
	_ = os.Mkdir(path, os.ModePerm)

	// Check if the directory was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.Wrap(err, "diretório não pode ser criado")
	}

	return
}

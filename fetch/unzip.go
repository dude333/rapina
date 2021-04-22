package fetch

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
// Source: https://golangcode.com/unzip-files-in-go/
//
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		if !valid(f.Name) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			if err = os.MkdirAll(fpath, os.ModePerm); err != nil {
				return nil, err
			}

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			fmt.Printf("[          ] Unziping %s", fpath)
			counter := &WriteCounter{}
			_, err = io.Copy(outFile, io.TeeReader(rc, counter))
			fmt.Println()

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}

func valid(filename string) bool {
	n := strings.ToLower(filename)

	if strings.Contains(n, "_ind_") {
		return false
	}

	list := []string{"_bpa_", "_bpp_", "_dfc_", "_dre_", "_dva_", "fre_"}

	for _, item := range list {
		if strings.Contains(n, item) {
			return true
		}
	}

	return false
}

package rapina

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestFilename(t *testing.T) {
	tempDir, _ := ioutil.TempDir("", "rapina-test")

	table := []struct {
		path     string
		name     string
		expected string
	}{
		{tempDir + "/test", "sample", tempDir + "/test/sample.xlsx"},
		{tempDir, "File 100", tempDir + "/File_100.xlsx"},
		{tempDir, "An,odd/file\\name", tempDir + "/An_odd_file_name.xlsx"},
	}

	for _, x := range table {
		returned, err := filename(x.path, x.name)
		expected := filepath.FromSlash(x.expected)
		if err != nil {
			t.Errorf("filename returned an error %v.", err)
		} else if returned != expected {
			t.Errorf("filename got: %s, want: %s.", returned, expected)
		}
	}

}

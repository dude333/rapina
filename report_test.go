package rapina

import "testing"

func TestFilename(t *testing.T) {
	table := []struct {
		path     string
		name     string
		expected string
	}{
		{"/tmp/test", "sample", "/tmp/test/sample.xlsx"},
		{"/tmp", "File 100", "/tmp/File_100.xlsx"},
		{"/tmp", "An,odd/file\\name", "/tmp/An_odd_file_name.xlsx"},
	}

	for _, x := range table {
		returned, err := filename(x.path, x.name)
		if err != nil {
			t.Errorf("filename returned an error %v.", err)
		} else if returned != x.expected {
			t.Errorf("filename got: %s, want: %s.", returned, x.expected)
		}
	}

	_, err := filename("/tmp2", "test")
	if err == nil {
		t.Errorf("filename should have returned an error.")
	}

}

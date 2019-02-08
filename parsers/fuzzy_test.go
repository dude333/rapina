package parsers

import "testing"

func TestFuzzyFind(t *testing.T) {
	list := []struct {
		src      string
		trg      []string
		maxDist  int
		expected string
	}{
		{"ABCD", []string{"ABC", "ACD"}, 2, "ABC"},
		{"ABCD", []string{"XYZ", "ACD"}, 1, "ACD"},
		{"ABCDÉ", []string{"XYZ", "ACD", "ABCDE"}, 0, "ABCDE"},
		{"ABCDÉ FGH", []string{"XYZ", "ACD", "FGH"}, 6, "FGH"},
		{"BCO ABC", []string{"XYZ", "BANCO ABC", "FGH"}, 0, "BANCO ABC"},
	}

	for _, l := range list {
		r := FuzzyFind(l.src, l.trg, l.maxDist)
		if r != l.expected {
			t.Errorf("Expected: %s, got: %s", l.expected, r)
		}
	}
}

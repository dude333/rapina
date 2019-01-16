package parsers

import (
	"testing"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func TestSectorsToYaml(t *testing.T) {
	err := SectorsToYaml("../cli/.data/B3sectors.xlsx", "../cli/.data/sectors.yaml")
	if err != nil {
		t.Errorf("Error: %v\n", err)
	}
}

func TestFromSector(t *testing.T) {
	s, _ := FromSector("GRENDENE S.A.", "../cli/setores.yaml")
	expected := [...]string{"ALPARGATAS", "CAMBUCI", "GRENDENE", "NIKE", "VULCABRAS"}
	if len(s) != 5 {
		t.Errorf("\n- Expected:  %v\n- Got:       %v", expected, s)
	}

	var arr [5]string
	copy(arr[:], s)
	if arr != expected {
		t.Errorf("\n- Expected:  %v\n- Got:       %v", expected, s)
	}
}

func TestMatch(t *testing.T) {
	testList := []string{
		"ALGAR TELEC",
		"ATT INC",
		"OI",
		"SPRINT",
		"TELEBRAS",
		"TELEF BRASIL",
		"TIM PART",
		"VERIZON",
		"ITAUUNIBANCO",
		"ITAU UNIBANCO",
		"ITAU UNIBANCO HOLDING",
		"ITAU UNIBANCO HOLDING S.A.",
	}

	company := RemoveDiacritics("ITAÚ UNIBANCO HOLDING S.A.")

	for _, l := range testList {
		l = RemoveDiacritics(l)
		r := fuzzy.RankMatchFold(l, company)
		if r != 100000 {
			t.Errorf("Expecting x from %s ~= %s, received %d", company, l, r)
		}
	}
}

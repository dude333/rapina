package parsers

import "testing"

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

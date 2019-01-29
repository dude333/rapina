package parsers

import (
	"testing"
)

func TestFromSector(t *testing.T) {
	s, _ := FromSector("GRENDENE S.A.", "../cli/setores.yml")
	expected := [...]string{"ALPARGATAS S.A.", "CAMBUCI S.A.", "GRENDENE S.A.", "VULCABRAS/AZALEIA S.A."}
	if len(s) != 4 {
		t.Errorf("\n- Expected:  %v\n- Got:       %v", expected, s)
	}

	var arr [4]string
	copy(arr[:], s)
	if arr != expected {
		t.Errorf("\n- Expected:  %v\n- Got:       %v", expected, s)
	}
}

package reports

import "testing"

func TestRemoveDiacritics(t *testing.T) {
	list := []struct {
		str string
		exp string
	}{
		{"ITAÚ", "ITAU"},
		{"SÃO", "SAO"},
		{"São Paulo", "Sao Paulo"},
		{"ÁÉÍÓÚáéíóúÀàÃÕãõÇç", "AEIOUaeiouAaAOaoCc"},
	}

	for _, l := range list {
		if removeDiacritics(l.str) != l.exp {
			t.Errorf("Expecting %s, received %s", l.exp, removeDiacritics(l.str))
		}
	}
}

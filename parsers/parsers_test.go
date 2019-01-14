package parsers

import "testing"

func TestGetHash(t *testing.T) {
	table := []struct {
		s string
		h uint32
	}{
		{"test1", 2569220284},
		{"random data", 1626193638},
		{"excel", 1973829744},
		{"One More...12345!", 2258028052},
	}
	for _, x := range table {
		h := GetHash(x.s)
		if h != x.h {
			t.Errorf("Hash was incorrect, got: %d, want: %d.", h, x.h)
		}
	}
}

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
		if RemoveDiacritics(l.str) != l.exp {
			t.Errorf("Expecting %s, received %s", l.exp, RemoveDiacritics(l.str))
		}
	}
}

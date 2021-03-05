package reports

import (
	"testing"

	p "github.com/dude333/rapina/parsers"
)

// AssertEqual checks if values are equal
func AssertEqual(t *testing.T, msg string, a interface{}, b interface{}) {
	if a == b {
		return
	}
	// debug.PrintStack()
	t.Errorf("%s was incorrect, received %v, expected %v.", msg, a, b)
}

func TestIdent(t *testing.T) {
	table := []struct {
		in         string
		expected   string
		isBaseItem bool
	}{
		{"1", "", true},
		{"1.1", "  ", true},
		{"1.1.2", "    ", false},
		{"2", "", true},
		{"2.3.4.5", "      ", false},
		{"2.10", "  ", true},
		{"3", "", true},
		{"3.1", "", true},
		{"3.1.2", "  ", false},
	}

	for _, x := range table {
		spaces, baseItem := ident(x.in)
		if spaces != x.expected {
			t.Errorf("ident was incorrect for %s, got: '%s', want: '%s'.", x.in, spaces, x.expected)
		}
		if baseItem != x.isBaseItem {
			t.Errorf("ident was incorrect for %s, got: %v, want: %v.", x.in, baseItem, x.isBaseItem)
		}
	}
}

func TestZeroIfNeg(t *testing.T) {
	for x := float32(10); x >= -10; x -= 0.1 {
		y := zeroIfNeg(x)
		if (x >= 0 && x != y) || (x < 0 && y != 0) {
			t.Errorf("zeroIfNeg was incorrect, got: %f, want: %f", y, x)
		}
	}
}

func TestSafeDiv(t *testing.T) {
	for x := float32(10); x >= -10; x -= 0.1 {
		y := safeDiv(2, x)
		if (x == 0 && y != 0) || (y != 2/x) {
			t.Errorf("safeDiv was incorrect, got: %f, want: %f", y, x)
		}
	}
}

func TestMetricsList(t *testing.T) {
	v := make(map[uint32]float32)

	for x := uint32(p.Caixa); x <= uint32(p.Dividendos); x++ {
		v[x] = float32(x) * 123456
	}
	l := metricsList(v)

	seq := []float32{
		v[p.Equity],
		0,
		v[p.Vendas],
		v[p.EBIT] - v[p.Deprec], // EBITDA
		v[p.EBIT],
		v[p.ResulFinanc],
		v[p.ResulOpDescont],
		v[p.LucLiq],
		0,
	}

	for i, val := range seq {
		AssertEqual(t, "metricsList ["+l[i].descr+"]", l[i].val, val)
	}

}

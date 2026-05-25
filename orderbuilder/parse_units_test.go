package orderbuilder

import "testing"

func TestParseUnits(t *testing.T) {
	cases := []struct {
		amount   float64
		decimals int
		want     string
	}{
		{0, 6, "0"},
		{1, 6, "1000000"},
		{1.5, 6, "1500000"},
		{100, 6, "100000000"},
		{0.5, 6, "500000"},
		{0.001, 6, "1000"},
		{0.000001, 6, "1"},
		{1234.5, 6, "1234500000"},
		{0.1, 6, "100000"},
		// Decimals=0 (no shift):
		{42, 0, "42"},
		{0, 0, "0"},
	}
	for _, c := range cases {
		got := parseUnits(c.amount, c.decimals)
		if got != c.want {
			t.Errorf("parseUnits(%v, %d) = %q, want %q", c.amount, c.decimals, got, c.want)
		}
	}
}

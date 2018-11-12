package numjson

import "testing"

func TestFormatFloat(t *testing.T) {
	for want, f := range map[string]float64{
		"0.2":    0.2,
		"5e-8":   5e-8,
		"0.0005": 5e-4,
	} {
		got := FormatFloat(f, 64)
		if got != want {
			t.Fatalf("Invalid float formating %s != %s", got, want)

		}
	}

}

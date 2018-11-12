package numjson

import (
	"math"
	"strconv"
	"testing"
)

func TestParseFloat(t *testing.T) {
	NaN := math.NaN()

	test := func(s string, want float64) {
		t.Helper()
		f := ParseFloat(s)
		if !math.IsNaN(want) {
			want, _ = strconv.ParseFloat(s, 64)
		}
		if f == want || (math.IsNaN(want) && math.IsNaN(f)) {
			return
		}
		t.Errorf("Invalid parse %q ->\n%.20f !=\n%.20f", s, f, want)
	}
	for s, f := range map[string]float64{
		"":        NaN,
		" ":       NaN,
		"-a7.2 ":  NaN,
		"-7a.2 ":  NaN,
		"10.2E-5": 10.2E-5,
		"10.2E+4": 10.2E+4,
		"10.2E5":  10.2E+5,
		"1":       1,
		"-1":      -1,
		"-1.2":    -1.2,
		"1.3":     1.3,
	} {
		test(s, f)
	}
	// Tests from github.com/valyala/fastjson/fastfloat
	// Invalid first char
	test("", NaN)
	test("  ", NaN)
	test("foo", NaN)
	test(" bar ", NaN)
	test("-", NaN)
	test("--", NaN)
	test("-.", NaN)
	test("-.e", NaN)
	test("+112", NaN)
	test("++", NaN)
	test("e123", NaN)
	test("E123", NaN)
	test("-e12", NaN)
	test(".", NaN)
	test("..34", NaN)
	test("-.32", NaN)
	test("-.e3", NaN)
	test(".e+3", NaN)

	// Invalid suffix
	test("1foo", NaN)
	test("1  foo", NaN)
	test("12.34.56", NaN)
	test("13e34.56", NaN)
	test("12.34e56e4", NaN)
	test("12.", NaN)
	test("123..45", NaN)
	test("123ee34", NaN)
	test("123e", NaN)
	test("123e+", NaN)
	test("123E-", NaN)
	test("123E+.", NaN)
	test("-123e-23foo", NaN)

	// Integer
	test("0", 0)
	test("-0", 0)
	test("0123", 123)
	test("-00123", -123)
	test("1", 1)
	test("-1", -1)
	test("1234567890123456", 1234567890123456)
	test("12345678901234567", 12345678901234567)
	test("123456789012345678", 123456789012345678)
	test("1234567890123456789", 1234567890123456789)
	test("12345678901234567890", 12345678901234567890)
	test("-12345678901234567890", -12345678901234567890)

	// Fractional part
	test("0.1", 0.1)
	test("-0.1", -0.1)
	test("-0.123", -0.123)
	test("12345.1234567890123456", 12345.1234567890123456)
	test("12345.12345678901234567", 12345.12345678901234567)
	test("12345.123456789012345678", 12345.123456789012345678)
	test("12345.1234567890123456789", 12345.1234567890123456789)
	test("12345.12345678901234567890", 12345.12345678901234567890)
	test("-12345.12345678901234567890", -12345.12345678901234567890)

	// Exponent part
	test("0e0", 0)
	test("123e+001", 123e1)
	test("0e12", 0)
	test("-0E123", 0)
	test("-0E-123", 0)
	test("-0E+123", 0)
	test("123e12", 123e12)
	test("-123E-12", -123E-12)
	test("-123e-400", 0)
	test("123e456", math.Inf(1))   // too big exponent
	test("-123e456", math.Inf(-1)) // too big exponent

	// Fractional + exponent part
	test("0.123e4", 0.123e4)
	test("-123.456E-10", -123.456E-10)

}

func TestParseUint(t *testing.T) {
	n, ok := ParseUint(`42`)
	if !ok {
		t.Fatalf("Failed to parse uint")
	}
	if n != 42 {
		t.Fatalf("Invalid parse result %d", n)

	}
	_, ok = ParseUint(`-42`)
	if ok {
		t.Fatalf("Should fail to parse uint")
	}
	_, ok = ParseUint(`42.01`)
	if ok {
		t.Fatalf("Should fail to parse uint")
	}

}

func TestParseInt(t *testing.T) {
	n, ok := ParseInt(`42`)
	if !ok {
		t.Fatalf("Failed to parse int")
	}
	if n != 42 {
		t.Fatalf("Invalid parse result %d", n)

	}
	n, ok = ParseInt(`-42`)
	if !ok {
		t.Fatalf("Failed to parse int")
	}
	if n != -42 {
		t.Fatalf("Invalid parse result %d", n)
	}
	_, ok = ParseInt(`42.01`)
	if ok {
		t.Fatalf("Should fail to parse int")
	}

}

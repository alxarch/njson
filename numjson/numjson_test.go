package numjson

import (
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestParseUint(t *testing.T) {
	assert := require.New(t)
	{
		n, tail := parseUint("123", 0)
		assert.Equal("", tail)
		assert.Equal(123, int(n))
	}
	{
		n, tail := parseUint("123.25", 0)
		assert.Equal(".25", tail)
		assert.Equal(123, int(n))
	}
}

func TestParse(t *testing.T) {
	type testCase struct {
		Input  string
		Expect Number
	}
	for _, tc := range []testCase{
		{"10.2E-5", Float64(10.2E-5)},
		{"", Number{}},
		{" ", Number{}},
		{"-a7.2 ", Number{}},
		{"-7a.2 ", Number{}},
		{"10.2E+4", Float64(10.2E+4)},
		{"10.2E5", Float64(10.2E+5)},
		{"1", Int64(1)},
		{"-1", Int64(-1)},
		{"-1.2", Float64(-1.2)},
		{"1.3", Float64(1.3)},
		// Tests from github.com/valyala/fastjson/fastfloat
		// Invalid first char
		{"", Number{}},
		{"  ", Number{}},
		{"foo", Number{}},
		{" bar ", Number{}},
		{"-", Number{}},
		{"--", Number{}},
		{"-.", Number{}},
		{"-.e", Number{}},
		{"+112", Number{}},
		{"++", Number{}},
		{"e123", Number{}},
		{"E123", Number{}},
		{"-e12", Number{}},
		{".", Number{}},
		{"..34", Number{}},
		{"-.32", Number{}},
		{"-.e3", Number{}},
		{".e+3", Number{}},

		// Invalid suffix
		{"1foo", Number{}},
		{"1  foo", Number{}},
		{"12.34.56", Number{}},
		{"13e34.56", Number{}},
		{"12.34e56e4", Number{}},
		{"12.", Number{}},
		{"123..45", Number{}},
		{"123ee34", Number{}},
		{"123e", Number{}},
		{"123e+", Number{}},
		{"123E-", Number{}},
		{"123E+.", Number{}},
		{"-123e-23foo", Number{}},

		// Integer
		{"0", Int64(0)},
		{"-0", Int64(0)},
		{"0123", Number{}},
		{"-00123", Number{}},
		{"1", Int64(1)},
		{"-1", Int64(-1)},
		{"1234567890123456", Int64(1234567890123456)},
		{"12345678901234567", Int64(12345678901234567)},
		{"123456789012345678", Int64(123456789012345678)},
		{"1234567890123456789", Int64(1234567890123456789)},
		{"12345678901234567890", Uint64(12345678901234567890)},

		// Fractional part
		{"0.1", Float64(0.1)},
		{"-0.1", Float64(-0.1)},
		{"-0.123", Float64(-0.123)},
		{"12345.1234567890123456", Float64(12345.1234567890123456)},
		{"12345.12345678901234567", Float64(12345.12345678901234567)},
		{"12345.123456789012345678", Float64(12345.123456789012345678)},
		{"12345.1234567890123456789", Float64(12345.1234567890123456789)},
		{"12345.12345678901234567890", Float64(12345.12345678901234567890)},
		{"-12345.12345678901234567890", Float64(-12345.12345678901234567890)},

		// Exponent part
		{"0e0", Float64(0)},
		{"123e+001", Float64(123e1)},
		{"0e12", Float64(0)},
		{"-0E123", Float64(0)},
		{"-0E-123", Float64(0)},
		{"-0E+123", Float64(0)},
		{"123e12", Float64(123e12)},
		{"-123E-12", Float64(-123E-12)},
		{"-123e-400", Float64(0)},
		{"123e456", Float64(math.Inf(1))},   // too big exponent
		{"-123e456", Float64(math.Inf(-1))}, // too big exponent

		// Fractional + exponent part
		{"0.123e4", Float64(0.123e4)},
		{"-123.456E-10", Float64(-123.456E-10)},
	} {
		tc := tc
		t.Run("parse_"+tc.Input, func(t *testing.T) {
			num, err := Parse(tc.Input)
			if !tc.Expect.IsValid() && err == nil {
				t.Errorf("expected error while parsing %q", tc.Input)
				return
			}
			if num.Type() != tc.Expect.Type() {
				t.Errorf("invalid result type %q ->\n%s !=\n%s", tc.Input, num.Type(), tc.Expect.Type())

			}
			if num.String() != tc.Expect.String() {
				t.Errorf("invalid result value %q ->\n%s !=\n%s", tc.Input, num, tc.Expect)
			}
		})
	}



}

//func TestParseUint(t *testing.T) {
//	n, ok := ParseUint(`42`)
//	if !ok {
//		t.Fatalf("Failed to parse uint")
//	}
//	if n != 42 {
//		t.Fatalf("Invalid parse result %d", n)
//
//	}
//	_, ok = ParseUint(`-42`)
//	if ok {
//		t.Fatalf("Should fail to parse uint")
//	}
//	_, ok = ParseUint(`42.01`)
//	if ok {
//		t.Fatalf("Should fail to parse uint")
//	}
//
//}
//
//func TestParseInt(t *testing.T) {
//	n, ok := ParseInt(`42`)
//	if !ok {
//		t.Fatalf("Failed to parse int")
//	}
//	if n != 42 {
//		t.Fatalf("Invalid parse result %d", n)
//
//	}
//	n, ok = ParseInt(`-42`)
//	if !ok {
//		t.Fatalf("Failed to parse int")
//	}
//	if n != -42 {
//		t.Fatalf("Invalid parse result %d", n)
//	}
//	_, ok = ParseInt(`42.01`)
//	if ok {
//		t.Fatalf("Should fail to parse int")
//	}
//}

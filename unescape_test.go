package njson_test

import (
	"testing"
	"unicode/utf8"

	"github.com/alxarch/njson"
)

func TestUnescape(t *testing.T) {
	b := make([]byte, 64)
	test := func(u, s string) {
		if b = b[:njson.Unescape(b[:cap(b)], u)]; string(b) != s {
			t.Errorf("Invalid unescape:\n%q %d\n%q %d", s, utf8.RuneCountInString(s), b, utf8.RuneCount((b)))
		}
	}
	test("goo", "goo")
	test("goo\\n", "goo\n")
	test("goo\\u0002!", "goo\x02!")
	test("\\uD834\\uDD1E", "𝄞")
	test("\\r", "\r")
	test("\\t", "\t")
	test("\\f", "\f")
	test("\\b", "\b")
	test("\\\\", "\\")
	test("\\\"", "\"")
	test("\\/", "/")
}

func TestEscapeString(t *testing.T) {
	b := make([]byte, 64)
	test := func(u, s string) {
		if b = njson.AppendEscaped(b[:0], s); string(b) != u {
			t.Errorf("Invalid escape:\n%q %d\n%q %d", u, utf8.RuneCountInString(s), b, utf8.RuneCount((b)))
		}
		if b = njson.EscapeBytes(b[:0], []byte(s)); string(b) != u {
			t.Errorf("Invalid escape:\n%q %d\n%q %d", u, utf8.RuneCountInString(s), b, utf8.RuneCount((b)))
		}
	}
	test("goo", "goo")
	test("goo\\n", "goo\n")
	test("goo\\u0002!", "goo\x02!")
	test("𝄞", "𝄞")
	test("\\r", "\r")
	test("\\t", "\t")
	test("\\f", "\f")
	test("\\b", "\b")
	test("\\\\", "\\")
	test("\\\"", "\"")
	test("\\/", "/")
}

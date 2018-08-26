package strjson_test

import (
	"testing"
	"unicode/utf8"

	"github.com/alxarch/njson/strjson"
)

func TestUnescape(t *testing.T) {
	b := make([]byte, 64)
	test := func(u, s string) {
		if b = b[:strjson.UnescapeTo(b[:cap(b)], u)]; string(b) != s {
			t.Errorf("Invalid unescape:\nexpect: %q %d\nactual: %q %d", s, utf8.RuneCountInString(s), b, utf8.RuneCount((b)))
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
		if b = strjson.Escape(b[:0], s); string(b) != u {
			t.Errorf("Invalid escape:\n%q %d\n%q %d", u, utf8.RuneCountInString(s), b, utf8.RuneCount((b)))
		}
		if b = strjson.EscapeBytes(b[:0], []byte(s)); string(b) != u {
			t.Errorf("Invalid escape bytes:\nexpect: %q %d\nactual: %q %d", u, utf8.RuneCountInString(s), b, utf8.RuneCount((b)))
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

func TestEscapeRune(t *testing.T) {
	testRune := func(s, esc string) {
		r, n := utf8.DecodeRuneInString(s)
		_ = n
		// buf := [utf8.UTFMax]byte{}
		// enc := buf[:utf8.EncodeRune(buf[:], r)]
		// qr := strconv.AppendQuoteRuneToASCII(nil, r)
		// t.Logf("%X %d %x %s", r, n, enc, qr)
		t.Run(s, func(t *testing.T) {
			if b := strjson.EscapeRune(nil, r); string(b) != esc {
				t.Errorf("Invalid escape %s %s", b, esc)
			}
		})
	}
	testRune("\f", "\\u000C")
	// U+1D11E MUSICAL SYMBOL G CLEF
	testRune("𝄞", `\uD834\uDD1E`)
	testRune("\uFFFD", `\uFFFD`)
	// Zero width space
	testRune(string([]byte{0xe2, 0x80, 0x8b}), `\u200B`)

}

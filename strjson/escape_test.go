package strjson

import (
	"testing"
	"unicode/utf8"
)

func Test_Escape3(t *testing.T) {
	b := make([]byte, 64)
	test := func(u, s string) {
		t.Helper()
		if b = Escape(b[:0], s, true); string(b) != u {
			t.Errorf("Invalid escape:\n%s %d\n%s %d", u, utf8.RuneCountInString(u), b, utf8.RuneCount((b)))
		}
	}
	test("foo𝄞bar", "foo𝄞bar")
	test(`\"Hello\nThis should be\u0002escaped𝄞\"foo`, "\"Hello\nThis should be\x02escaped𝄞\"foo")
	test("goo\\u0002!", "goo\x02!")
	test("goo", "goo")
	test("goo\\n", "goo\n")
	test("\\r", "\r")
	test("\\t", "\t")
	test("\\f", "\f")
	test("\\b", "\b")
	test("\\\\", "\\")
	test("\\\"", "\"")
	test("\\/", "/")
}

// func BenchmarkEscape2(b *testing.B) {
// 	buf := make([]byte, 64)
// 	s := "\"Hello\nThis should be\x02escaped𝄞\""
// 	e := `\"Hello\nThis should be\u0002escaped𝄞\"`
// 	buf = Escape2(buf[:0], s)
// 	if string(buf) != e {
// 		b.Errorf("Invalid escape %s", string(buf))
// 	}
// 	b.ReportAllocs()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		Escape2(buf[:0], s)
// 	}
// }
func BenchmarkEscape(b *testing.B) {
	buf := make([]byte, 64)
	s := "\"Hello\nThis should be\x02escaped𝄞\""
	e := `\"Hello\nThis should be\u0002escaped𝄞\"`
	buf = Escape(buf[:0], s, false)
	if string(buf) != e {
		b.Errorf("Invalid escape %s", string(buf))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Escape(buf[:0], s, false)
	}
}

// func TestEscapeRune(t *testing.T) {
// 	testRune := func(s, esc string) {
// 		r, n := utf8.DecodeRuneInString(s)
// 		_ = n
// 		// buf := [utf8.UTFMax]byte{}
// 		// enc := buf[:utf8.EncodeRune(buf[:], r)]
// 		// qr := strconv.AppendQuoteRuneToASCII(nil, r)
// 		// t.Logf("%X %d %x %s", r, n, enc, qr)
// 		t.Run(s, func(t *testing.T) {
// 			if b := strjson.EscapeRune(nil, r); string(b) != esc {
// 				t.Errorf("Invalid escape %s %s", b, esc)
// 			}
// 		})
// 	}
// 	testRune("\f", "\\u000C")
// 	// U+1D11E MUSICAL SYMBOL G CLEF
// 	testRune("𝄞", `\uD834\uDD1E`)
// 	testRune("\uFFFD", `\uFFFD`)
// 	// Zero width space
// 	testRune(string([]byte{0xe2, 0x80, 0x8b}), `\u200B`)

// }

package strjson_test

import (
	"testing"
	"unicode/utf8"

	"github.com/alxarch/njson/strjson"
)

func TestUnescape(t *testing.T) {
	b := make([]byte, 64)
	test := func(u, s string) {
		if b = strjson.Unescape(b[:0], u); string(b) != s {
			t.Errorf("Invalid unescape:\nexpect: %q %d\nactual: %q %d", s, utf8.RuneCountInString(s), b, utf8.RuneCount((b)))
		}
	}
	test("\\uD834\\uDD1E", "𝄞")
	test("goo\\u0002!", "goo\x02!")
	test("goo\\n", "goo\n")
	test("goo", "goo")
	test("\\r", "\r")
	test("\\t", "\t")
	test("\\f", "\f")
	test("\\b", "\b")
	test("\\\\", "\\")
	test("\\\"", "\"")
	test("\\/", "/")
}

func BenchmarkUnescapeRune(b *testing.B) {
	b.Run("ascii", BenchmarkUnescapeRuneASCII)
	b.Run("utf8", BenchmarkUnescapeRuneUTF8)
	b.Run("utf16", BenchmarkUnescapeRuneUTF16)
}

func BenchmarkUnescapeRuneASCII(b *testing.B) {
	testRune("\x02", `\u0002`, 6)(b)
}
func BenchmarkUnescapeRuneUTF8(b *testing.B) {
	testRune("ᾊ", `\u1F8A`, 6)(b)
}
func BenchmarkUnescapeRuneUTF16(b *testing.B) {
	testRune("𝄞", `\uD834\uDD1E`, 12)(b)
}

func testRune(want, got string, i int) func(b *testing.B) {
	buf := make([]byte, 64)
	j := 0
	return func(b *testing.B) {
		buf, j = strjson.UnescapeRune(buf[:0], got, 0)
		if string(buf) != want {
			b.Errorf("Invalid rune: %s != %s", buf, want)
			return
		}
		if i != j {
			b.Errorf("Invalid size: %d != %d", j, i)
			return
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = strjson.UnescapeRune(buf[:0], got, 0)
		}

	}
}
func BenchmarkUnescape(b *testing.B) {
	s := `\"Hello\nThis should be\u0002escaped𝄞\"foo`
	unescaped := "\"Hello\nThis should be\x02escaped𝄞\"foo"
	// s := `Lorem ipsum \n dolor.\uD834\uDD1E\uD834\uDD1E`
	// unescaped := []byte("Lorem ipsum \n dolor.𝄞𝄞")
	b.ReportAllocs()

	b.Run("unescape", func(b *testing.B) {
		buf := make([]byte, 0, 512)
		buf = strjson.Unescape(buf[:0], s)
		if string(buf) != unescaped {
			b.Errorf("Invalid unescape: %s", buf)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = strjson.Unescape(buf[:0], s)
		}
	})
}

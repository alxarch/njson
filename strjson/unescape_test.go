package strjson

import (
	"testing"
	"unicode/utf8"
)

func TestUnescaped(t *testing.T) {
	test := func(u, s string) {
		if e := Unescaped(u); e != s {
			t.Errorf("Invalid unescape:\nexpect: %q %d\nactual: %q %d", s, utf8.RuneCountInString(s), e, utf8.RuneCountInString(e))
		}
		if e := AppendUnescaped(nil, u); string(e) != s {
			t.Errorf("Invalid unescape:\nexpect: %q %d\nactual: %q %d", s, utf8.RuneCountInString(s), e, utf8.RuneCount(e))
		}
	}
	test("", "")
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

func TestUnescape(t *testing.T) {
	b := make([]byte, 64)
	n := 0
	test := func(u, s string) {
		if n = Unescape(b[:cap(b)], u); string(b[:n]) != s {
			t.Errorf(`Invalid unescape:
input: %s
expect: %q %d
actual: %q %d`, u,
				s, utf8.RuneCountInString(s),
				b[:n], utf8.RuneCount(b[:n]))
		}
		if b = AppendUnescaped(b[:0], s); string(b) != s {
			t.Errorf(`Invalid append unescape:
input: %s
expect: %q %d
actual: %q %d`, u,
				s, utf8.RuneCountInString(s),
				b[:n], utf8.RuneCount(b[:n]))

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

func BenchmarkUnescape(b *testing.B) {
	s := `\"Hello\nThis should be\u0002escaped𝄞\"foo`
	unescaped := "\"Hello\nThis should be\x02escaped𝄞\"foo"
	// s := `Lorem ipsum \n dolor.\uD834\uDD1E\uD834\uDD1E`
	// unescaped := []byte("Lorem ipsum \n dolor.𝄞𝄞")
	buf := make([]byte, 64)
	n := Unescape(buf, s)
	if n == -1 {
		b.Errorf("Invalid unescape: %s", buf)
	}
	if string(buf[:n]) != unescaped {
		b.Errorf("Invalid unescape\nactual: %s\nexpect: %s", buf[:n], unescaped)
	}
	b.ReportAllocs()

	b.Run("Unescape", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Unescape(buf, s)
		}
	})
	b.Run("UnescapeString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Unescaped(s)
		}
	})
}

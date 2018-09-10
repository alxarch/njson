package strjson

import (
	"testing"
	"unicode/utf8"
)

func Test_Escape(t *testing.T) {
	b := make([]byte, 64)
	test := func(u, s string) {
		t.Helper()
		if b = AppendEscaped(b[:0], s, true); string(b) != u {
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

func BenchmarkEscape(b *testing.B) {
	buf := make([]byte, 64)
	s := "\"Hello\nThis should be\x02escaped𝄞\""
	e := `\"Hello\nThis should be\u0002escaped𝄞\"`
	buf = AppendEscaped(buf[:0], s, false)
	if string(buf) != e {
		b.Errorf("Invalid escape %s", string(buf))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AppendEscaped(buf[:0], s, false)
	}
}

func Test_escapeUTF8(t *testing.T) {
	b := make([]byte, 64)
	for r, e := range map[rune]string{
		0:              "\\u0000",
		utf8.RuneError: "\\uFFFD",
		'\uACAB':       "\\uACAB",
	} {
		if b = escapeUTF8(b[:0], r); string(b) != e {
			t.Errorf("Invalid escapeUTF8 result %s != %s", b, e)
		}
	}
}

func Test_EscapeString(t *testing.T) {
	test := func(u, s string) {
		if e := Escaped(s, true, false); e != u {
			t.Errorf("Invalid escape:\n%s %d\n%s %d", u, utf8.RuneCountInString(u), e, utf8.RuneCountInString(e))
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

func BenchmarkEscapeString(b *testing.B) {
	s := "\"Hello\nThis should be\x02escaped𝄞\""
	e := `\"Hello\nThis should be\u0002escaped𝄞\"`
	got := Escaped(s, false, false)
	if got != e {
		b.Errorf("Invalid escape %s", got)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Escaped(s, false, false)
	}
}

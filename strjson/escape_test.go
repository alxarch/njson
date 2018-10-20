package strjson

import (
	"fmt"
	"testing"
	"unicode/utf8"
)

func Test_Escape(t *testing.T) {
	test := func(u, s string) {
		t.Helper()
		if b := AppendEscaped(nil, s, true); string(b) != u {
			t.Errorf("Invalid escape:\n%s %d\n%s %d", u, utf8.RuneCountInString(u), b, utf8.RuneCount((b)))
		}
		if s := Escaped(s, true, false); s != u {
			t.Errorf("Invalid escape:\n%s %d\n%s %d", u, utf8.RuneCountInString(u), s, utf8.RuneCountInString(s))
		}
		u = fmt.Sprintf(`"%s"`, u)
		if s := Escaped(s, true, true); s != u {
			t.Errorf("Invalid escape:\n%s %d\n%s %d", u, utf8.RuneCountInString(u), s, utf8.RuneCountInString(s))
		}
	}
	test("", "")
	test("foo𝄞bar", "foo𝄞bar")
	test(`\"Hello\nThis should be\u0002escaped𝄞\"foo`, "\"Hello\nThis should be\x02escaped𝄞\"foo")
	test("goo\\u0002!", "goo\x02!")
	test("goo\\u2028!", "goo\u2028!")
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
		utf8.RuneError: "\\ufffd",
		'\uACAB':       "\\uacab",
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

func Test_EscapeHTML(t *testing.T) {
	html := "<p>Foo</p>"
	expect := `\u003cp\u003eFoo\u003c\/p\u003e`
	if actual := Escaped(html, true, false); actual != expect {
		t.Errorf("Invalid HTML escape: %s != %s", actual, expect)
	}
	if actual := Escaped(html, false, false); actual != `<p>Foo<\/p>` {
		t.Errorf("Invalid HTML escape: %s != %s", actual, expect)
	}
	if actual := AppendEscaped(nil, html, true); string(actual) != expect {
		t.Errorf("Invalid HTML escape: %s != %s", actual, expect)
	}

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

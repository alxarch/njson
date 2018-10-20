package njson

import (
	"strings"
	"testing"
)

func Test_parser(t *testing.T) {
	d := Document{}
	n, tail, err := d.Parse(smallJSON)
	if err != nil {
		t.Error(err)
		return
	}
	if n.get() == nil {
		t.Errorf("Nil root")
		return
	}
	if strings.TrimSpace(tail) != "" {
		t.Errorf("Non empty tail: %s", tail)
	}

}

func Benchmark_parser(b *testing.B) {
	d := Document{}
	b.SetBytes(int64(len(mediumJSON)))
	for i := 0; i < b.N; i++ {
		d.Reset()
		d.Parse(mediumJSON)
	}
}
func BenchmarkParseD(b *testing.B) {
	b.Run("small.json", benchmarkD(smallJSON))
	b.Run("medium.min.json", benchmarkD(mediumJSON))
	b.Run("medium.json", benchmarkD(mediumJSONFormatted))
	b.Run("large.json", benchmarkD(largeJSON))
	b.Run("twitter.json", benchmarkD(twitterJSON))
	b.Run("canada.json", benchmarkD(canadaJSON))

}

func benchmarkD(src string) func(b *testing.B) {
	d := Document{}

	return func(b *testing.B) {
		d.Reset()
		n, tail, err := d.Parse(src)
		if err != nil {
			b.Errorf("Parse error: %s", err)
			return
		}
		if strings.TrimSpace(tail) != "" {
			b.Errorf("Non empty tail: %d", len(tail))
			return
		}
		if n.get() == nil {
			b.Errorf("Nil root")
			return
		}
		b.ReportAllocs()
		b.SetBytes(int64(len(src)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			d.Reset()
			n, tail, err = d.Parse(src)
		}
		_ = n
	}
}

func TestDocument_Object(t *testing.T) {
	d := Blank()
	defer d.Close()
	n := d.Object()
	n.Set("foo", d.Text("bar"))
	if data, err := n.AppendJSON(nil); err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	} else if s := `{"foo":"bar"}`; string(data) != s {
		t.Errorf("Invalid json: %s != %s", data, s)
	}
	n.Set("bar", n)
	if data, err := n.AppendJSON(nil); err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	} else if s := `{"foo":"bar","bar":{"foo":"bar"}}`; string(data) != s {
		t.Errorf("Invalid json: %s != %s", data, s)
	}

}

func TestDocument_Create(t *testing.T) {
	d := Document{}
	n := d.Null()
	assertEqual(t, n.get(), &node{
		info:   vNull | infRoot,
		raw:    strNull,
		values: nil,
	})
	n = d.Text("foo")
	assertEqual(t, n.get(), &node{
		info:   vString | infRoot,
		raw:    "foo",
		values: nil,
	})
	n = d.TextHTML("<p>Foo</p>")
	assertEqual(t, n.get(), &node{
		info:   vString | infRoot,
		raw:    `\u003cp\u003eFoo\u003c\/p\u003e`,
		values: nil,
	})
	n = d.Number(42)
	assertEqual(t, n.get(), &node{
		info:   vNumber | infRoot,
		raw:    `42`,
		values: nil,
	})
	n = d.Array()
	assertEqual(t, n.get(), &node{
		info:   vArray | infRoot,
		raw:    ``,
		values: nil,
	})
	n = d.Object()
	assertEqual(t, n.get(), &node{
		info:   vObject | infRoot,
		raw:    ``,
		values: nil,
	})
	n = d.True()
	assertEqual(t, n.get(), &node{
		info:   vBoolean | infRoot,
		raw:    `true`,
		values: nil,
	})
	n = d.False()
	assertEqual(t, n.get(), &node{
		info:   vBoolean | infRoot,
		raw:    `false`,
		values: nil,
	})

}

func Test_DocumentReset(t *testing.T) {
	d := new(Document)
	n := d.Text("foo")
	assertEqual(t, n.Document(), d)
	d.Reset()
	if n.Document() != nil {
		t.Errorf("Node document not nil after reset")
	}

}

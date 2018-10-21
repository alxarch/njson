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

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}
func TestDocument_ParseUnsafe(t *testing.T) {
	d := Document{}
	data := []byte(`["foo"]`)
	n, _, err := d.ParseUnsafe(data)
	assertNoError(t, err)
	raw := n.Index(0).Raw()
	assertEqual(t, raw, "foo")
	d2 := Document{}
	obj := d2.Object()
	obj.Set("foo", n)
	b, err := obj.AppendJSON(nil)
	assertNoError(t, err)
	assertEqual(t, string(b), `{"foo":["foo"]}`)
	data[2] = 'b'
	data[3] = 'a'
	data[4] = 'z'
	assertEqual(t, raw, "baz")
	b, err = obj.AppendJSON(nil)
	assertNoError(t, err)
	assertEqual(t, string(b), `{"foo":["foo"]}`)

}

func TestDocument_ncopy(t *testing.T) {
	d := new(Document)
	n := d.Object()
	n.Set("foo", d.Text("bar"))
	assertEqual(t, len(d.nodes), 2)
	id := d.ncopy(d, d.get(0))
	assertEqual(t, id, uint(2))
	assertEqual(t, len(d.nodes), 4)
	other := new(Document)
	other.Object()
	id = d.ncopy(other, other.get(0))
	assertEqual(t, id, uint(4))
	assertEqual(t, len(d.nodes), 5)
	n.Set("bar", other.Root())
	data, err := n.AppendJSON(nil)
	assertNoError(t, err)
	assertEqual(t, string(data), `{"foo":"bar","bar":{}}`)

}

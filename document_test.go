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
	if n.value() == nil {
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

func TestDocument_Object(t *testing.T) {
	d := Blank()
	defer d.Close()
	n := d.Object().Node()
	n.Object().Set("foo", d.Text("bar"))
	if data, err := n.AppendJSON(nil); err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	} else if s := `{"foo":"bar"}`; string(data) != s {
		t.Errorf("Invalid json: %s != %s", data, s)
	}
	n.Object().Set("bar", n)
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
	assertEqual(t, n.value(), &value{
		typ:      TypeNull,
		flags:     flagRoot,
		raw:      strNull,
		children: nil,
	})
	n = d.Text("foo")
	assertEqual(t, n.value(), &value{
		typ:      TypeString,
		flags:     flagRoot,
		raw:      "foo",
		children: nil,
	})
	n = d.TextHTML("<p>Foo</p>")
	assertEqual(t, n.value(), &value{
		typ:      TypeString,
		flags:     flagRoot,
		raw:      `\u003cp\u003eFoo\u003c\/p\u003e`,
		children: nil,
	})
	n = d.Number(42)
	assertEqual(t, n.value(), &value{
		typ:      TypeNumber,
		flags:     flagRoot,
		raw:      `42`,
		children: nil,
	})
	n = d.Array().Node()
	assertEqual(t, n.value(), &value{
		typ:      TypeArray,
		flags:     flagRoot,
		raw:      ``,
		children: nil,
	})
	n = d.Object().Node()
	assertEqual(t, n.value(), &value{
		typ:      TypeObject,
		flags:     flagRoot,
		raw:      ``,
		children: nil,
	})
	n = d.True()
	assertEqual(t, n.value(), &value{
		typ:      TypeBoolean,
		flags:     flagRoot,
		raw:      `true`,
		children: nil,
	})
	n = d.False()
	assertEqual(t, n.value(), &value{
		typ:      TypeBoolean,
		flags:     flagRoot,
		raw:      `false`,
		children: nil,
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

func TestDocument_ncopy(t *testing.T) {
	d := new(Document)
	n := d.Object().Node()
	n.Object().Set("foo", d.Text("bar"))
	assertEqual(t, len(d.values), 2)
	id := d.copyValue(d, d.get(0))
	assertEqual(t, id, uint(2))
	assertEqual(t, len(d.values), 4)
	other := new(Document)
	other.Object()
	id = d.copyValue(other, other.get(0))
	assertEqual(t, id, uint(4))
	assertEqual(t, len(d.values), 5)
	n.Object().Set("bar", other.Root())
	data, err := n.AppendJSON(nil)
	assertNoError(t, err)
	assertEqual(t, string(data), `{"foo":"bar","bar":{}}`)
}

package njson_test

import (
	"io/ioutil"
	"testing"

	"github.com/alxarch/njson"
)

func TestDocumentParse(t *testing.T) {
	doc := njson.Document{}
	p := njson.Parser{}
	for _, src := range []string{
		// `[]`,
		// `{}`,
		// `[{"foo":"bar"},2,3]`,
		// `{"answer":42}`,
		// `{"answer":"42"}`,
		// `{"answer":true}`,
		// `{"answer":null}`,
		// `{"answer":false}`,
		// `{"results":[]}`,
		// `{"results":[42],"error":null}`,

		`{"baz":{"foo":"bar"}}`,
		// `{"foo":"bar","bar":23,"baz":{"foo":21.2}}`,
		// `{"results":[{"id":42,"name":"answer"},{"id":43,"name":"answerplusone"}],"error":null}`,
		smallJSON,
		mediumJSON,
		largeJSON,
	} {
		root, err := p.Parse(src, &doc)
		if err != nil {
			t.Errorf("Parse error: %s", err)
		} else if root == nil {
			t.Errorf("Nil root")
		} else if out := root.AppendTo(nil); string(out) != src {
			t.Errorf("Invalid root:\nexpect: %s\nactual: %s", src, out)
		}
	}

}

func benchmark(src string) func(b *testing.B) {
	doc := njson.Document{}
	out := []byte{}
	p := njson.Parser{}

	return func(b *testing.B) {
		b.SetBytes(int64(len(src)))
		if root, err := p.Parse(src, &doc); err != nil {
			b.Errorf("Parse error: %s", err)
		} else if out := root.AppendTo(out[:0]); string(out) != src {
			b.Errorf("Invalid parse")
		}
	}
}
func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.Run("small.json", benchmark(smallJSON))
		b.Run("medium.json", benchmark(mediumJSON))
		b.Run("large.json", benchmark(largeJSON))
	}

}

var (
	largeJSON  string
	mediumJSON string
	smallJSON  string
)

func init() {
	if data, err := ioutil.ReadFile("testdata/large.json"); err != nil {
		panic(err)
	} else {
		largeJSON = string(data)
	}
	if data, err := ioutil.ReadFile("testdata/medium.min.json"); err != nil {
		panic(err)
	} else {
		mediumJSON = string(data)
	}
	if data, err := ioutil.ReadFile("testdata/small.json"); err != nil {
		panic(err)
	} else {
		smallJSON = string(data)
	}
}

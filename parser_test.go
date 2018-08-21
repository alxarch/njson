package njson_test

import (
	"io/ioutil"
	"testing"

	"github.com/alxarch/njson"
)

func TestDocumentParse(t *testing.T) {
	doc := njson.Document{}
	for _, src := range []string{
		`[]`,
		`{}`,
		`[{"foo":"bar"},2,3]`,
		`{"answer":42}`,
		`{"answer":"42"}`,
		`{"answer":true}`,
		`{"answer":null}`,
		`{"answer":false}`,
		`{"results":[]}`,
		`{"results":[42],"error":null}`,

		`{"baz":{"foo":"bar"}}`,
		`{"foo":"bar","bar":23,"baz":{"foo":21.2}}`,
		`{"results":[{"id":42,"name":"answer"},{"id":43,"name":"answerplusone"}],"error":null}`,
		smallJSON,
		mediumJSON,
		largeJSON,
	} {
		doc.Reset()
		root, err := doc.Parse(src)
		if err != nil {
			t.Errorf("Parse error: %s", err)
		} else if root == nil {
			t.Errorf("Nil root")
		} else if out, _ := root.AppendJSON(nil); string(out) != src {
			t.Errorf("Invalid root:\nexpect: %s\nactual: %s", src, out)
		}
	}

}

func benchmark(src string) func(b *testing.B) {
	doc := njson.BlankDocument()
	defer doc.Close()
	out := []byte{}

	return func(b *testing.B) {
		// b.SetBytes(int64(len(src)))
		if root, err := doc.Parse(src); err != nil {
			b.Errorf("Parse error: %s", err)
		} else if out, _ := root.AppendJSON(out[:0]); string(out) != src {
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

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		src     string
		wantErr bool
	}{
		{`{}`, false},
		{`"foobarbaz"`, false},
		{`1.2`, false},
		{`0`, false},
		{`-1`, false},
		{`-1.2E-3`, false},
		{`NaN`, false},
		{`true`, false},
		{`false`, false},
		{`null`, false},
		{`{"foo\n":"bar"}`, false},
		{`-a4`, true},
		{`[{"foo":"bar"},2,3]`, false},
		{`{"answer":42}`, false},
		{`{"answer":"42"}`, false},
		{`{"answer":true}`, false},
		{`{"answer":null}`, false},
		{`{"answer":false}`, false},
		{`{"results":[]}`, false},
		{`{"results":[42],"error":null}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			d := njson.Document{}
			root, err := d.Parse(tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Parse(%q) Unexpected error: %v", tt.src, err)
				return
			}
			if tt.wantErr {
				return
			}
			if root == nil {
				t.Errorf("Parser.Parse() nil root")
				return
			}
			out, _ := root.AppendJSON(nil)
			if string(out) != tt.src {
				t.Errorf("Parser.Parse() invalid node:\nactual: %s\nexpect: %s", out, tt.src)
			}
		})
	}
}

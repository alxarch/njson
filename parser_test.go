package njson

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestParseQuick(t *testing.T) {
	s := largeJSON
	testParse(t, s, s)
}
func TestParse(t *testing.T) {
	for in, out := range map[string]string{
		mediumJSONFormatted: mediumJSON,
	} {
		testParse(t, in, out)
	}
	for _, src := range []string{
		largeJSON,
		`{"results":[]}`,
		`{"answer":42}`,
		`{"foo":"bar"}`,
		`{"empty":""}`,
		`{"foo":1,"bar":2,"baz":3}`,
		`[]`,
		`["foo","bar"]`,
		`[42,42]`,
		`{}`,
		`42`,
		`-1.0`,
		`{"answer":42.0}`,
		`{"answer":"42"}`,
		`{"answer":42,"notanswer":"41"}`,
		`{"answer":true}`,
		`{"answer":null}`,
		`{"answer":false}`,
		`{"results":[42,1],"error":null}`,
		`[{"foo":"bar"},2,3]`,

		`{"baz":{"foo":"bar\"baz"}}`,
		`{"foo":"bar","bar":23,"baz":{"foo":21.2}}`,
		`{"results":[{"id":42,"name":"answer"},{"id":43,"name":"answerplusone"}],"error":null}`,
		smallJSON,
		mediumJSON,
	} {
		testParse(t, src, src)
	}

}

func testParse(t *testing.T, input, output string) {
	t.Helper()
	d := Blank()
	defer d.Close()
	p, tail, err := d.Parse(input)
	if err != nil {
		t.Error(input, err)
	} else if out, _ := d.AppendJSON(nil, p.id); string(out) != output {
		t.Errorf("Invalid root:\nexpect: %s\nactual: %s", output, out)
	} else if strings.TrimSpace(tail) != "" {
		t.Errorf("Tail not empty: %q", tail)
	}
}

func TestParser_Parse(t *testing.T) {
	p := Blank()
	defer p.Close()
	tests := []struct {
		args    string
		wantN   Node
		wantS   string
		wantErr bool
	}{
		// {`-a7`, nil, `-a7`, true},
	}
	for _, tt := range tests {
		t.Run(strconv.Quote(tt.args), func(t *testing.T) {
			gotN, gotS, err := p.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotS != tt.wantS {
				t.Errorf("Parser.Parse() tail = %q, want %q", gotS, tt.wantS)
			}
			if !reflect.DeepEqual(gotN, tt.wantN) {
				t.Errorf("Parser.Parse() = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}

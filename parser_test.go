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
	} else if out, _ := p.AppendJSON(nil); string(out) != output {
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

func Test_parseInvalidInput(t *testing.T) {
	type testCase struct {
		input string
		err   error
		tail  string
	}
	for _, tt := range []testCase{
		{`{foo:1}`, abort(1, TypeObject, 'f', []rune{'}', '"'}), `foo:1}`},
		{`fals`, UnexpectedEOF(TypeBoolean), `fals`},
		{`falso`, abort(0, TypeBoolean, `falso`, strFalse), `falso`},
		{`a`, abort(0, TypeAnyValue, 'a', "any value"), `a`},
		{`{"foo":b}`, abort(7, TypeAnyValue, 'b', "any value"), `b}`},
		{`{"foo"`, UnexpectedEOF(TypeObject), ``},
		{`{"foo`, UnexpectedEOF(TypeObject), ``},
		{`{"foo"}`, abort(6, TypeObject, '}', delimNameSeparator), `}`},
		{`{"foo":1`, UnexpectedEOF(TypeObject), ``},
		{`{`, UnexpectedEOF(TypeObject), ``},
		{`[`, UnexpectedEOF(TypeArray), ``},
		{`[1}`, abort(2, TypeArray, '}', []rune{delimValueSeparator, delimEndArray}), `}`},
		{`[1 `, UnexpectedEOF(TypeArray), ``},
		{`["foo`, UnexpectedEOF(TypeString), `foo`},
		{`tru`, UnexpectedEOF(TypeBoolean), `tru`},
		{`truz`, abort(0, TypeBoolean, `truz`, strTrue), `truz`},
		{`nul`, UnexpectedEOF(TypeNull), `nul`},
		{`nulz`, abort(0, TypeNull, `nulz`, strNull), `nulz`},
		{`{"foo":"bar","bar":baz}`, abort(19, TypeAnyValue, 'b', "any value"), `baz}`},
		{`{"foo":"bar",}`, abort(13, TypeObject, '}', delimString), `}`},
		{`{"foo":"bar", `, UnexpectedEOF(TypeObject), ``},
	} {
		d := Document{}
		n, tail, err := d.Parse(tt.input)
		// e, _ := err.(*ParseError)
		assertEqual(t, err.Error(), tt.err.Error())
		assertEqual(t, n, Node{0, 0, nil})
		if e, ok := err.(*ParseError); ok {
			assertEqual(t, tail, "")
			assertEqual(t, tt.input[e.Pos():], tt.tail)
		}
	}

}
func TestDocument_ParseResetParse(t *testing.T) {
	d := Document{}
	d.Parse(`{"foo":"bar","bar":"baz"}`)
	assertEqual(t, len(d.nodes), 3)
	d.Reset()
	d.Parse(`"foo"`)
	assertEqual(t, len(d.nodes), 1)
	assertEqual(t, d.nodes[0].raw, "foo")
}

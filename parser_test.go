package njson

import (
	"reflect"
	"strconv"
	"testing"
)

func TestParseQuick(t *testing.T) {
	s := `{"answer":42}`
	testParse(t, s, s)
}
func TestParse(t *testing.T) {
	for in, out := range map[string]string{
		mediumJSONFormatted: mediumJSON,
	} {
		testParse(t, in, out)
	}
	for _, src := range []string{
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
		largeJSON,
	} {
		testParse(t, src, src)
	}

}

func testParse(t *testing.T, input, output string) {
	t.Helper()
	p := Parser{}
	root, tail, err := p.Parse(input)
	if err != nil {
		t.Error(input, err)
	} else if root == nil {
		t.Errorf("Nil root")
	} else if out, _ := root.AppendJSON(nil); string(out) != output {
		t.Errorf("Invalid root:\nexpect: %s\nactual: %s", output, out)
	} else if tail != "" {
		t.Errorf("Tail not empty: %q", tail)
	}
}

// func Test_scanNumberAt(t *testing.T) {
// 	type args struct {
// 		c   byte
// 		s   string
// 		pos int
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    string
// 		wantEnd int
// 		wantInf Info
// 	}{
// 		{`42`, args{'4', `42`, 0}, `42`, 2, vNumberUint},
// 		{`-42`, args{'-', `-42`, 0}, `-42`, 3, vNumberInt},
// 		{`-42.0`, args{'-', `-42.0`, 0}, `-42.0`, 5, vNumberFloat},
// 		{`-a42.0`, args{'-', `-a42.0`, 0}, `-a`, 1, HasError},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, gotEnd, gotInf := scanNumberAt(tt.args.c, tt.args.s, tt.args.pos)
// 			if got != tt.want {
// 				t.Errorf("scanNumberAt() got = %v, want %v", got, tt.want)
// 			}
// 			if gotEnd != tt.wantEnd {
// 				t.Errorf("scanNumberAt() gotEnd = %v, want %v", gotEnd, tt.wantEnd)
// 			}
// 			if gotInf != tt.wantInf {
// 				t.Errorf("scanNumberAt() gotInf = %v, want %v", gotInf, tt.wantInf)
// 			}
// 		})
// 	}
// }

func TestParser_Parse(t *testing.T) {
	p := Get()
	defer p.Close()
	tests := []struct {
		args    string
		wantN   *Node
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

package generator_test

import (
	"testing"

	"github.com/alxarch/njson"
)

//go:generate go run ../cmd/njson/njson.go -all -w generator_test

type Foo struct {
	Bar string
}

func TestFooUnmarshal(t *testing.T) {
	foo := Foo{}
	d := njson.BlankDocument()
	defer d.Close()
	if n, err := d.Parse(`{"Bar": "baz"}`); err != nil {
		t.Errorf("Unexpected parse error: %s", err)
	} else if err := foo.UnmarshalNodeJSON(n); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if foo.Bar != "baz" {
		t.Errorf("Invalid unmarshal: %v", foo)
	}

}

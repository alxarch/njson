package generator_test

import (
	"net/url"
	"testing"

	"github.com/alxarch/njson/njsonutil"
)

//go:generate go run ../cmd/njson/njson.go -w generator_test "_$"

type Foo_ struct {
	Bar string `json:"bar,omitempty"`
}

func TestFooUnmarshal(t *testing.T) {
	test := njsonutil.UnmarshalTest
	t.Run("Bar Foo", test(Foo_{}))
	t.Run("Blank Foo", test(Foo_{"Bar"}))
	t.Run("Foo null input", test(Foo_{}, njsonutil.Input([]byte(`null`))))

}

type Coords struct {
	Lat float64
	Lon float64
}

type NamedCoords_ struct {
	Coords
	Name string `json:"name,omitempty"`
}

func TestNamedCoords(t *testing.T) {
	test := njsonutil.UnmarshalTest
	t.Run("Empty", test(NamedCoords_{}))
	t.Run("All fields", test(NamedCoords_{Coords{1.2, 1.3}, "Foo"}))

}

type Params_ struct {
	Values url.Values
	OK     bool
}

func TestParams_(t *testing.T) {
	test := njsonutil.UnmarshalTest
	t.Run("Empty", test(Params_{}))
	t.Run("All fields", test(Params_{Values: url.Values{
		"foo": []string{"bar"},
	},
		OK: false,
	}))

}

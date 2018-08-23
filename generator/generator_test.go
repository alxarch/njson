package generator_test

import (
	"testing"

	"github.com/alxarch/njson/njsonutil"
)

//go:generate go run ../cmd/njson/njson.go -w generator_test ^Test

type TestFoo struct {
	Bar string
}

func TestFooUnmarshal(t *testing.T) {
	test := njsonutil.UnmarshalTest
	t.Run("Bar Foo", test(TestFoo{}))
	t.Run("Blank Foo", test(TestFoo{"Bar"}))
	t.Run("Foo null input", test(TestFoo{}, njsonutil.Input([]byte(`null`))))

}

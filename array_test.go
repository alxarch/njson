package njson_test

import (
	"fmt"
	"github.com/alxarch/njson"
	"reflect"
	"testing"
)

func TestArrayIterator(t *testing.T) {
	doc := njson.Document{}
	n, _, err := doc.Parse(`["foo","bar","baz"]`)
	assertNoError(t, err)
	arr := njson.Array(n)
	iter := arr.Iter()
	defer iter.Close()
	expect := []string{"foo", "bar", "baz"}
	for i := 0; iter.Next(); i++ {
		s, typ := iter.Node().ToString()
		assertEqual(t, njson.TypeString, typ)
		assertEqual(t, expect[i], s)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func assert(t *testing.T, ok bool, msg string, a ...interface{}) {
	t.Helper()
	if !ok {
		t.Fatalf("Assertion failed: %s", fmt.Sprintf(msg, a...))
	}
}
func assertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Assertion failed: %v != %v", a, b)
	}
}

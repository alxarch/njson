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
	iter := arr.Iterate()
	defer iter.Close()
	expect := []string{"foo", "bar", "baz"}
	assertEqual(t, 3, iter.Len())
	for i := 0; iter.Next(); i++ {
		assertEqual(t, 3, iter.Len())
		assertEqual(t, false, arr.IsMutable())
		s, typ := iter.Node().ToString()
		assertEqual(t, njson.TypeString, typ)
		assertEqual(t, expect[i], s)
	}
	iter.Close()
	assertEqual(t, true, arr.IsMutable())
}

func TestArray_Len(t *testing.T) {
	for input, expect := range map[string]int{
		`["foo", "bar","baz"]`: 3,
		`[]`:                   0,
		`{}`:                   -1,
		`"foo"`:                -1,
		`true`:                 -1,
		`0`:                    -1,
	} {
		doc := njson.Document{}
		n, tail, err := doc.Parse(input)
		assertNoError(t, err)
		assertEqual(t, "", tail)
		assertEqual(t, expect, n.Array().Len())
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

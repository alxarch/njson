package njson

import (
	"fmt"
	"reflect"
	"testing"
)

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

package njson

import (
	"testing"
)

func TestTypeError(t *testing.T) {
	n := &node{}
	err := n.TypeError(TypeAnyValue)
	if err.Error() != "Invalid type InvalidToken not in [String Object Array Number Boolean Null]" {
		t.Errorf("Invalid error message %s", err)
	}
}

func Test_parseError(t *testing.T) {
	var err error
	err = parseError{}.Err()
	assertNoError(t, err)
	err = abort(2, TypeInvalid, "foo", "bar")
	assertEqual(t, err.Error(), "Invalid parser state at position 2 foo bar")
	err = parseError{'?', []rune{'"', '}'}, 2, TypeString}.Err()
	assertEqual(t, err.Error(), "Invalid token '?' != ['\"' '}'] at position 2 while scanning String")
	err = parseError{'?', nil, 2, TypeString}.Err()
	assertEqual(t, err.Error(), "Invalid token '?' at position 2 while scanning String")

}

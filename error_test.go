package njson

import (
	"testing"
)

func TestTypeError(t *testing.T) {
	n := &Node{}
	err := n.TypeError(TypeAnyValue)
	if err.Error() != "Invalid type InvalidToken not in [String Object Array Number Boolean Null]" {
		t.Errorf("Invalid error message %s", err)
	}
}

func Test_ParseError(t *testing.T) {
	var err error
	err = (*ParseError)(nil)
	assertEqual(t, err.Error(), "<nil>")
	err = UnexpectedEOF(TypeString)
	assertEqual(t, err.Error(), "Unexpected end of input while scanning String")
	err = &ParseError{'?', []rune{'"', '}'}, 2, TypeString}
	assertEqual(t, err.Error(), "Invalid token '?' != ['\"' '}'] at position 2 while scanning String")

}

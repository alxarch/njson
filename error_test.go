package njson

import (
	"testing"
)

func TestTypeError(t *testing.T) {
	n := &Node{}
	err := n.TypeError(TypeAnyValue)
	if err.Error() != "value type is not valid" {
		t.Errorf("invalid error message: %s", err)
	}
}

func Test_ParseError(t *testing.T) {
	var err error
	err = (*ParseError)(nil)
	assertEqual(t, err.Error(), "<nil>")
	err = UnexpectedEOF(TypeString)
	assertEqual(t, err.Error(), "unexpected end of input while scanning String")
	err = &ParseError{'?', []rune{'"', '}'}, 2, TypeString}
	assertEqual(t, err.Error(), "invalid token '?' != ['\"' '}'] at position 2 while scanning String")

}

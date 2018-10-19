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

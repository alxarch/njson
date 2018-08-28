package njson_test

import (
	"testing"

	"github.com/alxarch/njson"
)

func TestTypeError(t *testing.T) {
	n := &njson.Node{}
	err := n.TypeError(njson.TypeAnyValue)
	if err.Error() != "Invalid type InvalidToken not in [String Number Boolean Null Array Object]" {
		t.Errorf("Invalid error message %s", err)
	}
}

package njson_test

import (
	"testing"

	"github.com/alxarch/njson"
)

func TestToHexDigit(t *testing.T) {
	d := njson.ToHexDigit('d')
	if d != 13 {
		t.Errorf("Invalid hex digit: %d", d)
	}
}

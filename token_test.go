package njson_test

import (
	"testing"

	"github.com/alxarch/njson"
)

func TestType_Types(t *testing.T) {
	typ := njson.TypeNumber
	ts := typ.Types()
	if len(ts) != 1 {
		t.Errorf("Invalid types: %s", ts)
		return
	}
	if ts[0] != njson.TypeNumber {
		t.Errorf("Invalid types: %s", ts)
	}
	typ |= njson.TypeObject
	ts = typ.Types()
	if len(ts) != 2 {
		t.Errorf("Invalid types: %s", ts)
		return
	}
	if ts[0] != njson.TypeNumber {
		t.Errorf("Invalid types: %s", ts[0])
	}
	if ts[1] != njson.TypeObject {
		t.Errorf("Invalid types: %s", ts[1])
	}
	// t.Error(njson.TypeError{njson.TypeString, typ})
}

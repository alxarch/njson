package njson_test

import (
	"testing"

	"github.com/alxarch/njson"
)

func TestType_String(t *testing.T) {
	for expect, typ := range map[string]njson.Type{
		"InvalidToken":  njson.TypeInvalid,
		"Number":        njson.TypeNumber,
		"Array":         njson.TypeArray,
		"Boolean":       njson.TypeBoolean,
		"Object":        njson.TypeObject,
		"Null":          njson.TypeNull,
		"String":        njson.TypeString,
		"Key":           njson.TypeKey,
		"AnyValue":      njson.TypeAnyValue,
		"[Number Null]": njson.TypeNumber | njson.TypeNull,
	} {
		if actual := typ.String(); actual != expect {
			t.Errorf("Invalid string %s != %s", actual, expect)
		}
	}

}
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

package njson

import (
	"testing"
)

func TestType_String(t *testing.T) {
	for expect, typ := range map[string]Type{
		"InvalidToken":  TypeInvalid,
		"Number":        TypeNumber,
		"Array":         TypeArray,
		"Boolean":       TypeBoolean,
		"Object":        TypeObject,
		"Null":          TypeNull,
		"String":        TypeString,
		"AnyValue":      TypeAnyValue,
		"[Number Null]": TypeNumber | TypeNull,
	} {
		if actual := typ.String(); actual != expect {
			t.Errorf("Invalid string %s != %s", actual, expect)
		}
	}

}
func TestType_Types(t *testing.T) {
	typ := TypeNumber
	ts := typ.Types()
	if len(ts) != 1 {
		t.Errorf("Invalid types: %s", ts)
		return
	}
	if ts[0] != TypeNumber {
		t.Errorf("Invalid types: %s", ts)
	}
	typ |= TypeObject
	ts = typ.Types()
	if len(ts) != 2 {
		t.Errorf("Invalid types: %s", ts)
		return
	}
	if ts[1] != TypeNumber {
		t.Errorf("Invalid types: %s", ts[0])
	}
	if ts[0] != TypeObject {
		t.Errorf("Invalid types: %s", ts[1])
	}
	// t.Error(TypeError{TypeString, typ})
}

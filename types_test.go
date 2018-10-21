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
	assertEqual(t, Type(0).Types(), []Type{})
}

func TestType_IsValue(t *testing.T) {
	assertEqual(t, TypeNumber.IsValue(), true)
	assertEqual(t, TypeObject.IsValue(), true)
	assertEqual(t, TypeString.IsValue(), true)
	assertEqual(t, TypeArray.IsValue(), true)
	assertEqual(t, TypeBoolean.IsValue(), true)
	assertEqual(t, TypeNull.IsValue(), true)
	assertEqual(t, TypeAnyValue.IsValue(), true)
	assertEqual(t, TypeInvalid.IsValue(), false)
	assertEqual(t, Type(1<<7).IsValue(), false)
}

func TestInfo_IsValue(t *testing.T) {
	assertEqual(t, vNumber.IsValue(), true)
	assertEqual(t, vObject.IsValue(), true)
	assertEqual(t, vString.IsValue(), true)
	assertEqual(t, vArray.IsValue(), true)
	assertEqual(t, vBoolean.IsValue(), true)
	assertEqual(t, vNull.IsValue(), true)
	assertEqual(t, infAnyValue.IsValue(), true)
	assertEqual(t, info(0).IsValue(), false)
}

func TestInfo_IsNull(t *testing.T) {
	assertEqual(t, vNumber.IsNull(), false)
	assertEqual(t, vObject.IsNull(), false)
	assertEqual(t, vString.IsNull(), false)
	assertEqual(t, vArray.IsNull(), false)
	assertEqual(t, vBoolean.IsNull(), false)
	assertEqual(t, vNull.IsNull(), true)
	assertEqual(t, infAnyValue.IsNull(), false)
	assertEqual(t, info(0).IsNull(), false)
}

func TestInfo_IsNumber(t *testing.T) {
	assertEqual(t, vNumber.IsNumber(), true)
	assertEqual(t, vObject.IsNumber(), false)
	assertEqual(t, vString.IsNumber(), false)
	assertEqual(t, vArray.IsNumber(), false)
	assertEqual(t, vBoolean.IsNumber(), false)
	assertEqual(t, vNull.IsNumber(), false)
	assertEqual(t, infAnyValue.IsNumber(), false)
	assertEqual(t, info(0).IsNumber(), false)
}

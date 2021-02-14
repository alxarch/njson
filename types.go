package njson

import (
	"fmt"
	"github.com/alxarch/njson/strjson"
	"math/bits"
)

// Type is the type of a node.
type Type uint8

// Token types and type masks
const (
	TypeInvalid Type = iota
	TypeString  Type = 1 << iota
	TypeObject
	TypeArray
	TypeNumber
	TypeBoolean
	TypeNull
	TypeAnyValue = TypeString | TypeNumber | TypeBoolean | TypeObject | TypeArray | TypeNull
	TypeScalar = TypeString | TypeNumber | TypeBoolean
	TypeImmutable = TypeString | TypeNumber | TypeBoolean | TypeNull
	TypeComposite = TypeArray|TypeObject
)

// Types returns all types of a typemask
func (t Type) Types() (types []Type) {
	if t == 0 {
		return []Type{}
	}
	if bits.OnesCount(uint(t)) == 1 {
		return []Type{t}
	}
	for i := Type(0); i < 8; i++ {
		tt := Type(1 << i)
		if t&tt != 0 {
			types = append(types, tt)
		}
	}
	return
}

func (t Type) IsScalar() bool {
	return t&TypeScalar != 0
}
func (t Type) IsComposite() bool {
	return t&TypeComposite != 0
}
func (t Type) IsNull() bool {
	return t == TypeNull
}
func (t Type) IsValid() bool {
	return t&TypeAnyValue != 0
}
func (t Type) IsImmutable() bool {
	return t&TypeImmutable != 0
}

const (
	strFalse = "false"
	strTrue  = "true"
	strNull  = "null"
	strNaN   = "NaN"
)

func (t Type) String() string {
	switch t {
	case TypeInvalid:
		return "Invalid"
	case TypeString:
		return "String"
	case TypeArray:
		return "Array"
	case TypeObject:
		return "Object"
	case TypeNumber:
		return "Number"
	case TypeNull:
		return "Null"
	case TypeBoolean:
		return "Boolean"
	case TypeAnyValue:
		return "AnyValue"
	default:
		if bits.OnesCount(uint(t)) > 1 {
			return fmt.Sprint(t.Types())
		}
		return "Invalid"
	}
}

type flags strjson.Flags

const (
	flagRoot flags = 1 << 7
	flagNew        = flagRoot | flags(strjson.FlagJSON)
)

func (f flags) IsRoot() bool {
	return f&flagRoot == flagRoot
}

package njson

import (
	"fmt"
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

const (
	strFalse = "false"
	strTrue  = "true"
	strNull  = "null"
	strNaN   = "NaN"
)

func (t Type) String() string {
	switch t {
	case TypeInvalid:
		return "InvalidToken"
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
		return "InvalidToken"
	}
}

type flags uint8

const (
	flagRoot flags = 1 << iota
	flagEscapedString
	flagUnescapedString
)

func (f flags) IsRoot() bool {
	return f&flagRoot == flagRoot
}
func (f flags) IsSimpleString() bool {
	return f&flagEscapedString == 0
}
func (f flags) IsUnescaped() bool {
	return f&flagUnescapedString == flagUnescapedString
}

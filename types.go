package njson

import (
	"fmt"
	"math/bits"
)

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

type Info uint16

const (
	vString     = Info(TypeString)
	vNumber     = Info(TypeNumber)
	vNull       = Info(TypeNull)
	vBoolean    = Info(TypeBoolean)
	vArray      = Info(TypeArray)
	vObject     = Info(TypeObject)
	vFalse      = vBoolean
	vTrue       = vBoolean | IsTrue
	vNumberUint = vNumber | NumberZeroDecimal | NumberParsed
	vNumberInt  = vNumber | NumberZeroDecimal | NumberSigned | NumberParsed
)
const (
	NumberSigned Info = 1 << (iota + 8)
	NumberZeroDecimal
	NumberParsed
	Unescaped
	IsTrue
	HasError Info = 1 << 15
)

func (i Info) Unescaped() bool {
	return i == Unescaped|vString
}

func (i Info) NumberParsed() bool {
	const parsed = NumberParsed | vNumber
	return i&parsed > NumberParsed
}

func (i Info) Type() Type {
	return Type(i)
}

func (i Info) HasLen() bool {
	return i&(vObject|vArray) != 0
}
func (i Info) HasRaw() bool {
	return i&(vObject|vArray) == 0
}

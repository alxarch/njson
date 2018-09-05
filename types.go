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
	TypeKey      // = TypeString | TypeObject
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
	case TypeKey:
		return "Key"
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
	vString       = Info(TypeString)
	vNumber       = Info(TypeNumber)
	vNull         = Info(TypeNull)
	vBoolean      = Info(TypeBoolean)
	vArray        = Info(TypeArray)
	vObject       = Info(TypeObject)
	vKey          = Info(TypeKey)
	vFalse        = vBoolean
	vTrue         = vBoolean | IsTrue
	vNumberUint   = vNumber
	vNumberInt    = vNumber | NumberSigned
	vNumberFloat  = vNumber | NumberFloat
	vNumberFloatZ = vNumber | NumberFloat | NumberZeroDecimal
)
const (
	NumberSigned Info = 1 << (iota + 8)
	NumberFloat
	NumberZeroDecimal
	NumberParsed
)
const (
	HasError  Info = 1 << 15
	Unescaped Info = 1 << (iota + 8)
)
const (
	IsTrue Info = 1 << (iota + 8)
)

func (i Info) Unescaped() bool {
	const unescaped = Unescaped | vQuoted
	return i&unescaped > Unescaped
}

func (i Info) NumberParsed() bool {
	const parsed = NumberParsed | vNumber
	return i&parsed > NumberParsed
}

const vQuoted = vString | vKey

func (i Info) Quoted() bool {
	return i&vQuoted != 0
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

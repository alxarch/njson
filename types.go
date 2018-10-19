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

// Info is a bitmask with type info for a node.
type Info uint16

const (
	vString  = Info(TypeString)
	vNumber  = Info(TypeNumber)
	vNull    = Info(TypeNull)
	vBoolean = Info(TypeBoolean)
	vArray   = Info(TypeArray)
	vObject  = Info(TypeObject)
)

// Type flags
const (
	_ Info = 1 << (iota + 8)
	Unsafe
	Root
)

// IsRoot checks if IsRoot flag is set.
func (i Info) IsRoot() bool {
	return i&Root == Root
}

// IsSafe checks if Unsafe flag is set.
func (i Info) IsSafe() bool {
	return i&Unsafe == 0
}

// Type retutns the Type part of Info.
func (i Info) Type() Type {
	return Type(i)
}

// HasLen returns if an Info's type has length. (ie is Object or Array)
func (i Info) HasLen() bool {
	return i&(vObject|vArray) != 0
}

// IsNull checks if t is TypeNull
func (t Type) IsNull() bool {
	return t == TypeNull
}

// IsNull checks if i is TypeNull
func (i Info) IsNull() bool {
	return i.Type() == TypeNull
}

// IsArray checks if i is TypeArray
func (i Info) IsArray() bool {
	return i.Type() == TypeArray
}

// IsArray checks if t is TypeArray
func (t Type) IsArray() bool {
	return t == TypeArray
}

// IsValue checks if t matches TypeAnyValue
func (t Type) IsValue() bool {
	return t&TypeAnyValue != 0
}

// IsValue checks if i matches TypeAnyValue
func (i Info) IsValue() bool {
	const vAnyValue = Info(TypeAnyValue)
	return i&vAnyValue != 0
}

// IsString checks if t is TypeString
func (t Type) IsString() bool {
	return t == TypeString
}

// IsString checks if i is TypeString
func (i Info) IsString() bool {
	return i.Type() == TypeString
}

// IsNumber checks if i is TypeNumber
func (i Info) IsNumber() bool {
	return i.Type() == TypeNumber
}

// IsObject checks if i is TypeObject
func (i Info) IsObject() bool {
	return i.Type() == TypeObject
}

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

// info is a bitmask with type info for a node.
type info uint16

const (
	vString  = info(TypeString)
	vNumber  = info(TypeNumber)
	vNull    = info(TypeNull)
	vBoolean = info(TypeBoolean)
	vArray   = info(TypeArray)
	vObject  = info(TypeObject)
)

// Type flags
const (
	_ info = 1 << (iota + 8)
	infUnsafe
	infRoot
)

// IsRoot checks if IsRoot flag is set.
func (i info) IsRoot() bool {
	return i&infRoot == infRoot
}

// IsSafe checks if Unsafe flag is set.
func (i info) IsSafe() bool {
	return i&infUnsafe == 0
}
func (i info) Flags() info {
	return i & 0xFF00
}

// Type retutns the Type part of Info.
func (i info) Type() Type {
	return Type(i)
}

// IsNull checks if i is TypeNull
func (i info) IsNull() bool {
	return i.Type() == TypeNull
}

// IsBoolean checks if i is TypeBoolean
func (i info) IsBoolean() bool {
	return i.Type() == TypeBoolean
}

// IsArray checks if i is TypeArray
func (i info) IsArray() bool {
	return i.Type() == TypeArray
}

// IsValue checks if t matches TypeAnyValue
func (t Type) IsValue() bool {
	return t&TypeAnyValue != 0
}

const infAnyValue = info(TypeAnyValue)

// IsValue checks if i matches TypeAnyValue
func (i info) IsValue() bool {
	return i&infAnyValue != 0
}

// IsString checks if i is TypeString
func (i info) IsString() bool {
	return i.Type() == TypeString
}

// IsNumber checks if i is TypeNumber
func (i info) IsNumber() bool {
	return i.Type() == TypeNumber
}

// IsObject checks if i is TypeObject
func (i info) IsObject() bool {
	return i.Type() == TypeObject
}

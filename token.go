package njson

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
)

// Token is a JSON token
type Token struct {
	info  ValueInfo
	extra uint16 // Used for unescaped token index in keys/strings
	src   string
	num   uint64
}

// Type is the token type.
type Type byte

// Token types and type masks
const (
	TypeInvalid Type = iota
	TypeString  Type = 1 << iota
	TypeNumber
	TypeBoolean
	TypeNull
	TypeArray
	TypeObject
	TypeKey
	TypeAnyValue = TypeString | TypeNumber | TypeBoolean | TypeObject | TypeArray | TypeNull
	TypeSized    = TypeObject | TypeArray
	// typeSourceOK = TypeString | TypeNumber | TypeBoolean | TypeNull
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

func (t *Token) parseFloat() (float64, bool) {
	f, err := strconv.ParseFloat(t.src, 64)
	if err != nil {
		t.info |= ValueError
		return 0, false
	}
	t.num = math.Float64bits(f)
	t.info |= ValueReady
	if math.Trunc(f) == f {
		t.info |= ValueZeroDecimal
	}
	return f, true
}

// ToBool returns the boolean value of a token and whether the conversion is lossless
func (i ValueInfo) ToBool() (bool, bool) {
	return i == ValueTrue, i&ValueInfo(TypeBoolean) != 0
}
func negative(u uint64) uint64 {
	return ^(u - 1)
}

// ToUint returns the int value of a token and whether the conversion is lossless
func (t Token) ToUint() (uint64, bool) {
	switch t.info {
	case ValueNegativeInteger:
		return t.num, false
	case ValuePositiveInteger:
		return t.num, true
	case ValuePositiveFloatZeroDecimal:
		f := math.Float64frombits(t.num)
		return uint64(f), true
	case ValueNegativeFloatZeroDecimal:
		return 0, false
	default:
		return 0, false
	}
}

// ToInt returns the int value of a token and whether the conversion is lossless
func (t Token) ToInt() (int64, bool) {
	switch t.info {
	case ValueNegativeInteger:
		return int64(t.num), negative(t.num) <= math.MaxInt64
	case ValuePositiveInteger:
		return int64(t.num), t.num <= math.MaxInt64
	case ValuePositiveFloatZeroDecimal:
		f := math.Float64frombits(t.num)
		return int64(f), f <= math.MaxInt64
	case ValueNegativeFloatZeroDecimal:
		f := math.Float64frombits(t.num)
		return int64(f), f >= math.MinInt64
	default:
		return 0, false
	}
}

// ToFloat returns the float value of a token and whether the conversion is lossless
func (t Token) ToFloat() (float64, bool) {
	switch t.info {
	case ValueNegativeInteger:
		return -float64(negative(t.num)), true
	case ValuePositiveInteger:
		return float64(t.num), true
	default:
		if t.info.IsFloat() {
			return math.Float64frombits(t.num), true
		}
		return 0, false
	}
}

var (
	uNaN = math.Float64bits(math.NaN())
	fNaN = math.NaN()
)

var (
	errInvalidJSONString = errors.New("Invalid JSON string")
)

// Escaped return the JSON escaped string form.
func (t Token) Escaped() string {
	return t.src
}

// Bytes returns the raw bytes of the JSON values.
func (t Token) Bytes() []byte {
	return []byte(t.src)
}

// UnsafeBytes returns the bytes of the JSON value without copying.
func (t Token) UnsafeBytes() []byte {
	return s2b(t.src)
}

// Type returns the token type.
func (t Token) Type() Type {
	return t.info.Type()
}

// ValueInfo holds type and value information for a Token
type ValueInfo uint16

// Number flags
const (
	ValueFloat ValueInfo = 1 << (iota + 8)
	ValueReady
	ValueNegative
	ValueZeroDecimal
	ValueError
	ValueNumberFloatReady         = ValueInfo(TypeNumber) | ValueFloat | ValueReady
	ValuePositiveFloatZeroDecimal = ValueNumberFloatReady | ValueZeroDecimal
	ValueNegativeFloatZeroDecimal = ValueNumberFloatReady | ValueZeroDecimal | ValueNegative
	ValueNegativeInteger          = ValueInfo(TypeNumber) | ValueNegative
	ValuePositiveInteger          = ValueInfo(TypeNumber)
)

// String flags
const (
	ValueUnescaped ValueInfo = 1 << (iota + 8)
)

// Boolean flags
const (
	ValueTrue  ValueInfo = ValueInfo(TypeBoolean) | 1<<(iota+8)
	ValueFalse ValueInfo = ValueInfo(TypeBoolean)
)

// HasError reports if there was a number parse error.
func (i ValueInfo) HasError() bool {
	return i&ValueError == ValueError
}

// HasZeroDecimal reports if a number value has zero decimal part
func (i ValueInfo) HasZeroDecimal() bool {
	return i&ValueZeroDecimal == ValueZeroDecimal
}

// Type returns the token Type part of the info.
func (i ValueInfo) Type() Type {
	return Type(i)
}

const needsEscape = ValueUnescaped | ValueInfo(TypeString) | ValueInfo(TypeKey)

// NeedsEscape checks if value needs escaping.
func (i ValueInfo) NeedsEscape() bool {
	// This works because type bits are on the right side :)
	return (i & needsEscape) > ValueUnescaped
}

// IsFloat checks if the value is a number that needs float to represent.
func (i ValueInfo) IsFloat() bool {
	return i&ValueNumberFloatReady == ValueNumberFloatReady
}

const unparsedFloat = ValueInfo(TypeNumber) | ValueFloat

// IsUnparsedFloat checks if the value is a number that needs float to represent and it needs parsing.
func (i ValueInfo) IsUnparsedFloat() bool {
	return i&ValueNumberFloatReady == unparsedFloat
}

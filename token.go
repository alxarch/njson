package njson

import (
	"errors"
	"math"
	"math/bits"
	"strconv"
)

// Token is an intermediate JSON representation that allows DOM traversal of JSON documents
type Token struct {
	info ValueInfo
	size uint16
	src  string
	num  uint64
}

type Type byte

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
	typeSourceOK = TypeString | TypeNumber | TypeBoolean | TypeNull
)

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

func (t Type) hasSource() bool {
	return t&typeSourceOK != 0
}

func (t Type) v() ValueInfo {
	return ValueInfo(t)
}

const (
	strFalse = "false"
	strTrue  = "true"
	strNull  = "null"
	strNaN   = "NaN"
)

func (t Type) String() string {
	switch t {
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
	default:
		return "InvalidToken"
	}
}

func (t *Token) unquote() (s string) {
	return t.src[1 : len(t.src)-1]
}

func (t *Token) Len() int {
	switch t.Type() {
	case TypeArray, TypeObject:
		return int(t.size)
	default:
		return 0
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

func (t *Token) ToUint() (uint64, bool) {
	switch t.info {
	case ValuePositiveInteger:
		return t.num, true
	case ValueIntegerFloat:
		return uint64(math.Float64frombits(t.num)), true
	case ValueNumberFloatUnparsed:
		if f, ok := t.parseFloat(); ok && f >= 0 && t.info.HasZeroDecimal() {
			return uint64(f), true
		}
		fallthrough
	default:
		return 0, false
	}
}
func (t *Token) ToBool() (bool, bool) {
	switch t.info {
	case ValueTrue:
		return true, true
	case ValueFalse:
		return false, true
	default:
		return false, false
	}
}
func negative(u uint64) uint64 {
	return ^(u - 1)
}

func (t *Token) ToInt() (int64, bool) {
	switch t.info {
	case ValueNegativeInteger:
		return int64(t.num), negative(t.num) <= math.MaxInt64
	case ValuePositiveInteger:
		return int64(t.num), t.num <= math.MaxInt64
	case ValueIntegerFloat:
		return int64(math.Float64frombits(t.num)), true
	case ValueNumberFloatUnparsed:
		if f, ok := t.parseFloat(); ok && t.info.HasZeroDecimal() {
			return int64(f), true
		}
		fallthrough
	default:
		return 0, false
	}
}

func (t *Token) ToFloat() (float64, bool) {
	switch t.info {
	case ValueNumberFloat:
		return math.Float64frombits(t.num), true
	case ValuePositiveInteger:
		return float64(t.num), true
	case ValueNegativeInteger:
		return float64(int64(t.num)), true
	case ValueNumberFloatUnparsed:
		return t.parseFloat()
	default:
		return 0, false
	}
}

// SizeHint returns a size hint for array and object elements.
// The maximum value for a hint is math.MaxUint16 - 1
// It's best to use the Len() method that handles this corner case
// by traversing the list of tokens.
func (t *Token) SizeHint() uint16 {
	return t.size
}

var (
	uNaN = math.Float64bits(math.NaN())
	fNaN = math.NaN()
)

var (
	errInvalidJSONString = errors.New("Invalid JSON string")
)

func (t *Token) ToJSON() string {
	return t.src
}
func (t *Token) Bytes() []byte {
	return s2b(t.src)
}

func hexByte(b []byte, pos int) (c byte) {
	return ToHexDigit(b[pos])<<4 | ToHexDigit(b[pos])
}

func hexDigit(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return (c - '0'), true
	case 'a' <= c && c <= 'f':
		return (c - 'a'), true
	case 'A' <= c && c <= 'F':
		return (c - 'A'), true
	default:
		return c, false
	}
}

func equalStrBytes(s string, b []byte) bool {
	if len(s) == len(b) {
		for i := 0; i < len(s); i++ {
			if s[i] != b[i] {
				return false
			}
		}
		return true
	}
	return false
}

func (t *Token) Type() Type {
	if t == nil {
		return 0
	}
	return t.info.Type()
}

func (t *Token) Unescaped() string {
	if t.info.IsQuoted() {
		if t.info&ValueUnescaped == ValueUnescaped {
			buf := make([]byte, len(t.src))
			buf = buf[:Unescape(buf, t.unquote())]
			return b2s(buf)
		}
		return t.unquote()
	}
	return t.src
}

func (t *Token) UnescapedBytes() (buf []byte) {
	if t.info.IsQuoted() {
		if t.info&ValueUnescaped == ValueUnescaped {
			buf = make([]byte, len(t.src))
			buf = buf[:Unescape(buf, t.unquote())]
		} else {
			buf = s2b(t.unquote())
		}
		return
	}
	return s2b(t.src)
}

type ValueInfo uint16

// Number flags
const (
	ValueFloat ValueInfo = 1 << (iota + 8)
	ValueReady
	ValueNegative
	ValueZeroDecimal
	ValueError
	ValueNumberFloatUnparsed = ValueInfo(TypeNumber) | ValueFloat
	ValueNumberFloat         = ValueInfo(TypeNumber) | ValueFloat | ValueReady
	ValueIntegerFloat        = ValueNumberFloat | ValueZeroDecimal
	ValueNegativeInteger     = ValueInfo(TypeNumber) | ValueNegative
	ValuePositiveInteger     = ValueInfo(TypeNumber)
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

func (i ValueInfo) HasError() bool {
	return i&ValueError == ValueError
}
func (i ValueInfo) HasZeroDecimal() bool {
	return i&ValueZeroDecimal == ValueZeroDecimal
}

func (i ValueInfo) Type() Type {
	return Type(i)
}
func (i ValueInfo) IsQuoted() bool {
	return i&ValueInfo(TypeString|TypeKey) != 0
}

func (i ValueInfo) NeedsEscape() bool {
	return (i&ValueUnescaped == ValueUnescaped) && i.IsQuoted()
}

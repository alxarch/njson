package numjson

import (
	"math"
	"strconv"
)

type Type uint8

const (
	_ Type = iota
	Int
	Uint
	Float
)

func (t Type) String() string {
	switch t {
	case Int:
		return "int"
	case Uint:
		return "uint"
	case Float:
		return "float"
	default:
		return "invalid"
	}
}


type Number struct {
	typ   Type
	value uint64
}

func Int64(i int64) Number {
	return Number{
		value: uint64(i),
		typ:   Int,
	}
}

func Uint64(u uint64) Number {
	return Number{
		value: u,
		typ:   Uint,
	}
}

func Float64(f float64) Number {
	return Number{
		value: math.Float64bits(f),
		typ:   Float,
	}
}
func FromParts(typ Type, data uint64) Number {
	return Number{
		typ:   typ,
		value: data,
	}
}

func (n Number) Type() Type {
	return n.typ
}

func (n Number) IsValid() bool {
	return n.typ != 0
}
func (n Number) IsZero() bool {
	return n.IsValid() && n.value == 0
}

func (n Number) String() string {
	switch n.typ {
	case Int:
		return strconv.FormatInt(int64(n.value), 10)
	case Uint:
		return strconv.FormatUint(n.value, 10)
	case Float:
		return FormatFloat(math.Float64frombits(n.value), 32)
	default:
		return ""
	}
}

func (n Number) Value() interface{} {
	switch n.typ {
	case Float:
		return math.Float64frombits(n.value)
	case Int:
		return int64(n.value)
	case Uint:
		return n.value
	default:
		return nil
	}
}

// Float64 converts n to float64 without considering precision.
func (n Number) Float64() float64 {
	switch n.typ {
	case Float:
		return math.Float64frombits(n.value)
	case Int:
		return float64(int64(n.value))
	case Uint:
		return float64(n.value)
	default:
		return math.NaN()
	}
}

// Int64 converts n to int64 without considering precision.
func (n Number) Int64() int64 {
	switch n.typ {
	case Int:
		return int64(n.value)
	case Float:
		return int64(math.Float64frombits(n.value))
	case Uint:
		return int64(n.value)
	default:
		return 0
	}
}

// Uint64 converts n to uint64 without considering precision.
func (n Number) Uint64() uint64 {
	switch n.typ {
	case Uint:
		return n.value
	case Int:
		return uint64(int64(n.value))
	case Float:
		return uint64(math.Float64frombits(n.value))
	default:
		return 0
	}
}

// Neg returns the negative of n
func (n Number) Neg() Number {
	switch n.typ {
	case Int:
		return Int64(-int64(n.value))
	case Uint:
		return Uint64(-n.value)
	case Float:
		return Float64(-math.Float64frombits(n.value))
	default:
		return Number{}
	}
}

func Compare(a, b Number) (int, bool) {
	switch a.typ {
	case Int:
		valA := int64(a.value)
		switch b.typ {
		case Uint:
			if valA < 0 {
				return 1, true
			}
			return cmpUint64(uint64(valA), b.value), true
		case Float:
			return cmpFloat64(float64(valA), math.Float64frombits(b.value)), true
		case Int:
			return cmpInt64(valA, int64(b.value)), true
		default:
			return 0, false
		}
	case Float:
		valA := math.Float64frombits(a.value)
		switch b.typ {
		case Float:
			return cmpFloat64(valA, math.Float64frombits(b.value)), true
		case Int:
			return cmpFloat64(valA, float64(int64(b.value))), true
		case Uint:
			if valA < 0 {
				return 1, true
			}
			return cmpFloat64(valA, float64(b.value)), true
		default:
			return 0, false
		}
	case Uint:
		valA := a.value
		switch b.typ {
		case Int:
			valB := int64(b.value)
			if valB < 0 {
				return -1, true
			}
			return cmpUint64(valA, uint64(valB)), true
		case Float:
			return cmpFloat64(float64(valA), math.Float64frombits(b.value)), true
		case Uint:
			return cmpUint64(valA, b.value), true
		default:
			return 0, false
		}
	default:
		return 0, false
	}
}

func cmpFloat64(a, b float64) int {
	if a < b {
		return 1
	}
	if a > b {
		return -1
	}
	return 0
}

func cmpUint64(a, b uint64) int {
	if a > b {
		return -1
	}
	if a < b {
		return 1
	}
	return 0
}

func cmpInt64(a, b int64) int {
	if a > b {
		return -1
	}
	if a < b {
		return 1
	}
	return 0
}

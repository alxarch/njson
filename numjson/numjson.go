package numjson

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"unsafe"
)

type Type uint8

const (
	_ Type = iota
	Int
	Uint
	Float
	BigInt
	BigFloat
)

func (t Type) String() string {
	switch t {
	case Int:
		return "int"
	case Uint:
		return "uint"
	case Float:
		return "float"
	case BigInt:
		return "bigint"
	case BigFloat:
		return "bigfloat"
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
func NewBigInt(b *big.Int) Number {
	v := big.NewInt(0).Set(b)
	return Number{
		value: uint64(uintptr(unsafe.Pointer(v))),
		typ:   BigInt,
	}
}

func NewBigFloat(b *big.Float) Number {
	v := big.NewFloat(0).Set(b)
	return Number{
		value: uint64(uintptr(unsafe.Pointer(v))),
		typ:   BigFloat,
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
		typ: Float,
	}
}

func (n Number) Type() Type {
	return n.typ
}

func (n Number) IsValid() bool {
	return n.typ != 0
}

func (n Number) String() string {
	switch n.typ {
	case Int:
		return strconv.FormatInt(int64(n.value), 10)
	case Uint:
		return strconv.FormatUint(n.value, 10)
	case Float:
		return strconv.FormatFloat(math.Float64frombits(n.value), 'f', -1, 64)
	case BigInt:
		return n.BigInt().String()
	case BigFloat:
		return n.BigFloat().String()
	default:
		return ""
	}
}
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

func (n Number) BigInt() *big.Int {
	if n.typ == BigInt {
		return (*big.Int)(unsafe.Pointer(uintptr(n.value)))
	}
	return nil
}
func (n Number) BigFloat() *big.Float {
	if n.typ == BigFloat {
		return (*big.Float)(unsafe.Pointer(uintptr(n.value)))
	}
	return nil
}

func ParseBigFloat(s string) (Number, error) {
	if b, ok := big.NewFloat(0).SetString(s); ok {
		return Number{
			typ:   BigFloat,
			value: uint64(uintptr(unsafe.Pointer(b))),
		}, nil
	}
	return Number{}, &strconv.NumError{
		Func: "ParseBigFloat",
		Num:  s,
	}
}

func ParseBigInt(s string) (Number, error) {
	if b, ok := big.NewInt(0).SetString(s, 10); ok {
		return Number{
			typ:   BigInt,
			value: uint64(uintptr(unsafe.Pointer(b))),
		}, nil
	}
	return Number{}, &strconv.NumError{
		Func: "ParseBigInt",
		Num:  s,
	}
}

func Parse(s string) (Number, error) {
	const funcName = "ParseNumber"
	var (
		tail   = s
		signed bool
		n      uint64
		c      byte
		f float64
	)
	// fast path for single digit integers
	if len(tail) == 1 {
		c = tail[0]
		if isDigit(c) {
			return Number{
				value: uint64(int64(c - '0')),
				typ:   Int,
			}, nil
		}
		return Number{}, &strconv.NumError{
			Func: funcName,
			Num:  s,
			Err:  fmt.Errorf("expecting a digit, found %q", c),
		}
	}

	// Read sign
	tail, signed = readNegative(tail)
	if len(tail) == 0 {
		return Number{}, &strconv.NumError{
			Func: funcName,
			Num:  s,
			Err:  errors.New("empty number string"),
		}
	}
	// Read integer part
	if len(tail) > 0 {
		c, tail = tail[0], tail[1:]
		if c == '0' {
			// continue to decimal
		} else if '1' <= c && c <= '9' {
			// read integer part
			n, tail = parseUint(tail, uint64(c-'0'))
			if n >= cutoff {
				return parseBig(s)
			}
			// continue to decimal
		} else {
			return Number{}, &strconv.NumError{
				Func: funcName,
				Num:  s,
				Err:  fmt.Errorf("expecting a digit, found %q", c),
			}
		}
	} else {
		return Number{}, &strconv.NumError{
			Func: funcName,
			Num:  s,
			Err:  errors.New("empty number string"),
		}
	}
	// Read decimal/scientific part
	if len(tail) > 0 {
		c, tail = tail[0], tail[1:]
	} else {
		goto makeInt
	}

	f = float64(n)

	// Read decimal part
	if c == '.' {
		numDigits := len(tail)
		n, tail = parseUint(tail, 0)
		f += float64(n) * math.Pow10(len(tail)-numDigits)
		if n >= cutoff {
			return ParseBigFloat(s)
		}
		if len(tail) > 0 {
			c, tail = tail[0], tail[1:]
		} else {
			goto makeFloat
		}
	}
	// Read scientific notation
	if c == 'e' || c == 'E' {
		exp, err := strconv.ParseInt(tail, 10, strconv.IntSize)
		if err != nil {
			return Number{}, &strconv.NumError{
				Func: funcName,
				Num:  s,
				Err:  err,
			}
		}
		f *= math.Pow10(int(exp))
		goto makeFloat
	}
	return Number{}, &strconv.NumError{
		Func: funcName,
		Num:  s,
		Err:  errors.New("unexpected characters after integral part"),
	}
makeFloat:
	if signed {
		return Number{
			value: math.Float64bits(-f),
			typ:   Float,
		}, nil
	}
	return Number{
		value: math.Float64bits(f),
		typ:   Float,
	}, nil
makeInt:
	if signed {
		const cutoffInt64 = uint64(1 << 63)
		if n <= cutoffInt64 {
			return Number{
				typ:   Int,
				value: -n,
			}, nil
		}
	} else if n <= math.MaxInt64 {
		return Number{
			typ:   Int,
			value: uint64(int64(n)),
		}, nil
	} else if n <= math.MaxUint64 {
		return Number{
			typ:   Uint,
			value: n,
		}, nil
	}
	b := big.NewInt(0).SetUint64(n)
	if signed {
		b.Neg(b)
	}
	return Number{
		typ:   BigInt,
		value: uint64(uintptr(unsafe.Pointer(b))),
	}, nil
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}


func readNegative(s string) (string, bool) {
	if len(s) > 0 && s[0] == '-' {
		return s[1:], true
	}
	return s, false
}


func parseUint(s string, n uint64) (uint64, string) {
	var c byte
	for i := 0; 0 <= i && i < len(s); i++ {
		c = s[i]
		if isDigit(c) && n < cutoff {
			n = n*10 + uint64(c-'0')
		} else {
			return n, s[i:]
		}
	}
	return n, ""
}

const cutoff = math.MaxUint64/10 + 1

func parseBig(s string) (Number, error) {
	if isFloat(s) {
		return ParseBigFloat(s)
	}
	return ParseBigInt(s)
}

func isFloat(s string) bool {
	for i := 0; 0 <= i && i < len(s);i++ {
		switch s[i] {
		case '.', 'e', 'E':
			return true
		}
	}
	return false
}
//
//func parseNumber(s string) (uint64, Type) {
//	tail, neg := readNegative(s)
//	num, dec, exp, tail := readNumberParts(tail)
//	if tail != "" {
//		return 0, 0
//	}
//	n, err := strconv.ParseUint(num, 10, 64)
//	if err != nil {
//		return 0, 0
//	}
//	if dec == "" && exp == "" {
//		if neg {
//			return uint64(-int64(n)), Int64
//		}
//		return uint64(int64(n)), Int64
//	}
//	f := float64(n)
//	if dec != "" {
//		n, err := strconv.ParseUint(dec, 10, 64)
//		if err != nil {
//			return 0, 0
//		}
//		f += float64(n) * math.Pow10(-len(dec))
//	}
//	if exp != "" {
//		n, err := strconv.ParseInt(exp, 10, strconv.IntSize)
//		if err != nil {
//			return 0, 0
//		}
//		f *= math.Pow10(int(n))
//	}
//	return math.Float64bits(f), Float
//}
//
//func readNumberParts(s string) (num, dec, exp, tail string) {
//	num, tail = readIntegral(s)
//	var c byte
//	if len(tail) > 0 {
//		c, tail = tail[0], tail[1:]
//	} else {
//		return
//	}
//	if c == '.' {
//		dec, tail = readDecimal(tail)
//		if len(tail) > 0 {
//			c, tail = tail[0], tail[1:]
//		} else {
//			return
//		}
//	}
//	if c == 'e' || c == 'E' {
//		exp, tail = readExponent(tail)
//	}
//	return
//}
//
//func readIntegral(s string) (string, string) {
//	if len(s) > 0 && s[0] == '0' {
//		return s[:1], s[1:]
//	}
//	for i := 0; 0 <= i && i < len(s); i++ {
//		if !isDigit(s[i]) {
//			return s[:i], s[i:]
//		}
//	}
//	return s, ""
//}
//
//func readDecimal(s string) (string, string) {
//	const minPow10 = 323
//	if len(s) > minPow10 {
//		s = s[:minPow10]
//	}
//	for i := 0; 0 <= i && i < len(s); i++ {
//		if !isDigit(s[i]) {
//			return s[:i], s[i:]
//		}
//	}
//	return s, ""
//}
//
//func readExponent(s string) (string, string) {
//	i := 0
//	if len(s) > 0 && s[0] == '+' || s[0] == '-' {
//		i++
//	}
//	for ; 0 <= i && i < len(s); i++ {
//		if !isDigit(s[i]) {
//			return s[:i], s[i:]
//		}
//	}
//	return s, ""
//}
//var fNaN = math.NaN()
//
//// ParseFloat parses a float number from a string.
//func ParseFloat(s string) float64 {
//	if len(s) == 1 && isDigit(s[0]) {
//		return float64(s[0] - '0')
//	}
//	var (
//		i      uint
//		j      int
//		c      byte
//		signed bool
//		num    uint64
//		dec    uint64
//		f      float64
//	)
//	if len(s) > 0 {
//		c = s[0]
//		if c == '-' {
//			signed = true
//			i = 1
//		}
//		if c == '0' {
//			if len(s) > 1 {
//				c, i = s[1], 1
//				goto decimal
//			}
//		}
//		if len(s) == 1 {
//			if isDigit(c) {
//				return float64(c - '0')
//			}
//			return fNaN
//		}
//	}
//	const cutoff = math.MaxUint64/10 + 1
//	for ; i < uint(len(s)); i++ {
//		c = s[i]
//		if isDigit(c) {
//			if num >= cutoff {
//				return fNaN
//			}
//			num = num*10 + uint64(c-'0')
//			j++
//			continue
//		}
//		if j == 0 {
//			return fNaN
//		}
//		goto decimal
//	}
//	if 0 < j && j <= 20 {
//		if signed {
//			return -float64(num)
//		}
//		return float64(num)
//	}
//	if j == 0 {
//		return fNaN
//	}
//	goto fallback
//decimal:
//	if c == '.' {
//		j = 0
//		for i++; i < uint(len(s)); i++ {
//			c = s[i]
//			if '0' <= c && c <= '9' {
//				dec = 10*dec + uint64(c-'0')
//				j++
//				continue
//			}
//			if j > 0 {
//				goto scientific
//			}
//			return fNaN
//		}
//		if 0 < j && j <= 19 {
//			f = float64(num) + float64(dec)*math.Pow10(-j)
//			if signed {
//				return -f
//			}
//			return f
//		}
//		if j == 0 {
//			return fNaN
//		}
//		goto fallback
//	}
//scientific:
//	if c == 'e' || c == 'E' {
//		signed := false
//		exp := 0
//		jj := 0
//		for i++; i < uint(len(s)); i++ {
//			c = s[i]
//			if '0' <= c && c <= '9' {
//				jj++
//				exp = 10*exp + int(c-'0')
//				continue
//			}
//			if jj == 0 {
//				switch c {
//				case '-':
//					signed = true
//					continue
//				case '+':
//					continue
//				}
//			}
//			return fNaN
//		}
//		if jj == 0 {
//			return fNaN
//		}
//		if exp > 300 {
//			goto fallback
//		}
//		if signed {
//			f = (float64(num) + float64(dec)*math.Pow10(-j)) * math.Pow10(-exp)
//		} else {
//			f = float64(num)*math.Pow10(exp) + float64(dec)*math.Pow10(exp-j)
//		}
//		goto done
//	}
//	return fNaN
//done:
//	if signed {
//		return -f
//	}
//	return f
//fallback:
//	return fNaN
//	f, err := strconv.ParseFloat(s, 64)
//	if err != nil {
//		return fNaN
//	}
//	return f
//}
//
//const (
//	maxSafeIntegerFloat64 = 9007199254740991
//	minSafeIntegerFloat64 = -9007199254740991
//)
//
//// ParseInt parses an int from string
//func ParseInt(s string) (int64, bool) {
//	f := ParseFloat(s)
//	if minSafeIntegerFloat64 <= f && f <= maxSafeIntegerFloat64 && math.Trunc(f) == f {
//		return int64(f), true
//	}
//	n, err := strconv.ParseInt(s, 10, 64)
//	return n, err == nil
//}
//
//// ParseUint parses a uint from string
//func ParseUint(s string) (uint64, bool) {
//	f := ParseFloat(s)
//	if 0 <= f && f <= maxSafeIntegerFloat64 && math.Trunc(f) == f {
//		return uint64(f), true
//	}
//	n, err := strconv.ParseUint(s, 10, 64)
//	return n, err == nil
//}

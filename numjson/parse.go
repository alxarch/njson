package numjson

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// Parse parses a JSON number literal.
//
// It returns a *strconv.NumError if it fails to parse s.
// The inner error is either a *TooBigError or an *InvalidSyntaxError.
// By detecting the *TooBigError the caller can fall back to special handling for values
// that are too big to represent as one of the types supported by Number (int64, uint64, float64).
func Parse(s string) (Number, error) {
	var (
		tail   = s
		signed bool
		n      uint64
		c      byte
		f      float64
	)
	// fast path for single digit integers
	if len(tail) == 1 {
		c = tail[0]
		if isDigit(c) {
			return FromParts(Int, uint64(int64(c-'0'))), nil
		}
		return invalidSyntax(s, tail, "expecting a digit, found %q", c)
	}

	// Read sign
	tail, signed = readSign(tail)
	// Read integer part
	if len(tail) > 0 {
		c, tail = tail[0], tail[1:]
		if c == '0' {
			// continue to decimal
		} else if '1' <= c && c <= '9' {
			// read integer part
			n, tail = parseUint(tail, uint64(c-'0'))
			if n >= cutoff && len(tail) > 0 && isDigit(tail[0]) {
				return tooBig(s, "integer part too big")
			}
			// continue to decimal
		} else {
			return invalidSyntax(s, tail, "expecting a digit, found %q", c)
		}
	} else {
		return invalidSyntax(s, tail, "empty integer part")
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
		tailSize := len(tail)
		n, tail = parseUint(tail, 0)
		if n >= cutoff && len(tail) > 0 && isDigit(tail[0]) {
			return tooBig(s, "decimal part too big to parse as uint64")
		}
		if exp := tailSize - len(tail); exp > 0 {
			f += float64(n) * math.Pow10(-exp)
			if len(tail) > 0 {
				c, tail = tail[0], tail[1:]
			} else {
				goto makeFloat
			}
		} else {
			return invalidSyntax(s, tail, "empty fractional part after '.'")
		}
	}
	// Read scientific notation
	if c == 'e' || c == 'E' {
		exp, err := strconv.ParseInt(tail, 10, strconv.IntSize)
		if err != nil {
			return Number{}, numError(s, err)
		}
		f *= math.Pow10(int(exp))
		goto makeFloat
	}
	return invalidSyntax(s, tail, "unexpected character %q after integer part", c)
makeFloat:
	if signed {
		return FromParts(Float, math.Float64bits(-f)), nil
	}
	return FromParts(Float, math.Float64bits(f)), nil
makeInt:
	if signed {
		const cutoffInt64 = uint64(1 << 63)
		if n <= cutoffInt64 {
			return FromParts(Int, -n), nil
		}
	} else if n <= math.MaxInt64 {
		return FromParts(Int, uint64(int64(n))), nil
	} else if n <= math.MaxUint64 {
		return FromParts(Uint, n), nil
	}
	b := big.NewInt(0).SetUint64(n)
	if signed {
		b.Neg(b)
		if b.IsInt64() {
			return FromParts(Int, uint64(b.Int64())), nil
		}
		return tooBig(s, "value too big to represent as int64")
	}
	if b.IsUint64() {
		return FromParts(Uint, n), nil
	}
	return tooBig(s, "value too big to represent as uint64")
}

func numError(num string, err error) error {
	const (
		funcName = "ParseNumber"
	)
	return &strconv.NumError{
		Func: funcName,
		Num:  num,
		Err:  err,
	}
}

func invalidSyntax(num, tail, msg string, args ...interface{}) (Number, error) {
	msg = fmt.Sprintf(msg, args...)
	pos := len(num) - len(tail)
	return Number{}, numError(num, &InvalidSyntaxError{msg: msg, pos: pos})
}
func tooBig(num, msg string) (Number, error) {
	return Number{}, numError(num, &TooBigError{msg: msg})
}

type TooBigError struct {
	msg string
}

func (e *TooBigError) Error() string {
	return e.msg
}

type InvalidSyntaxError struct {
	msg string
	pos int
}

func (e *InvalidSyntaxError) Error() string {
	return fmt.Sprintf("invalid JSON number syntax at position %d: %s", e.pos, e.msg)
}
func (e *InvalidSyntaxError) Position() int {
	return e.pos
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func readSign(s string) (string, bool) {
	if len(s) > 0 && s[0] == '-' {
		return s[1:], true
	}
	return s, false
}

const cutoff = math.MaxUint64/10 + 1

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

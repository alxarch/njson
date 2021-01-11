package starlarknjson

import (
	"errors"
	"go.starlark.net/starlark"
	"math"
	"math/big"
	"strconv"
)

func readNumber(s string) (starlark.Value, error) {
	var c byte
	var dec string
	var exp string
	neg, tail := readSign(s)
	num, tail := readDigits(tail)
	if num == "" {
		return nil, errors.New("invalid JSON number empty integer")
	}
	if len(num) > 1 && num[0] == '0' {
		return nil, errors.New("invalid JSON number leading zeros")
	}
	if len(tail) > 0 {
		c, tail = tail[0], tail[1:]
	}
	if c == '.' {
		dec, tail = readDigits(tail)
		if len(tail) > 0 {
			c, tail = tail[0], tail[1:]
		}
	}
	if c == 'e' || c == 'E' {
		exp, tail = readExp(tail)
	}
	if tail != "" {
		return nil, errors.New("invalid JSON number tail")
	}
	if dec == "" && exp == "" {
		return intFromParts(neg, num)
	}
	return floatFromParts(neg, num, dec, exp)
}


func floatFromParts(neg bool, num, dec, exp string) (starlark.Value, error) {
	n, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		return nil, err
	}
	f := float64(n)
	if neg {
		f = -f
	}
	if dec != "" {
		d, err := strconv.ParseUint(dec, 10, 64)
		if err != nil {
			return nil, err
		}
		f += float64(d)*math.Pow10(-len(dec))
	}
	if exp != "" {
		e, err := strconv.ParseInt(exp, 10, strconv.IntSize)
		if err != nil {
			return nil, err
		}
		f *= math.Pow10(int(e))
	}
	return starlark.Float(f), nil
}

func intFromParts(neg bool, num string) (starlark.Value, error) {
	n, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		if b, ok := big.NewInt(0).SetString(num, 10); ok {
			if neg {
				b.Neg(b)
			}
			return starlark.MakeBigInt(b), nil
		}
		return nil, err
	}
	v := starlark.MakeUint64(n)
	if neg {
		v = v.Mul(starlark.MakeInt(-1))
	}
	return v, nil
}

func readSign(s string) (bool, string) {
	if len(s) > 0 && s[0] == '-' {
		return true, s[1:]
	}
	return false, s
}
func readExp(s string) (string, string) {
	var sign string
	if len(s) > 0 && s[0] == '+' || s[0] == '-' {
		sign, s = s[:1], s[1:]
	}
	digits, tail := readDigits(s)
	return sign + digits, tail
}

func readDigits(s string) (string, string) {
	for i := 0; 0 <= i && i < len(s); i++ {
		if !isDigit(s[i]) {
			return s[:i], s[i:]
		}
	}
	return s, ""
}
func isDigit(c byte) bool {
	return '0' <= c && c < '9'
}

package numjson

import (
	"math"
	"strconv"
)

var fNaN = math.NaN()

func ParseFloat(s string) float64 {
	var (
		i      uint
		j      int
		c      byte
		num    uint64
		dec    uint64
		f      float64
		signed bool
	)
	if len(s) > 0 {
		if c = s[0]; c == '-' {
			signed = true
			i = 1
		}
		if len(s) == 1 {
			if '0' <= c && c <= '9' {
				return float64(c - '0')
			}
			return fNaN
		}
	}

	for ; i < uint(len(s)); i++ {
		c = s[i]
		if '0' <= c && c <= '9' {
			num = num*10 + uint64(c-'0')
			j++
			continue
		}
		if j == 0 {
			return fNaN
		}
		// c = s[i]
		goto decimal
	}
	if 0 < j && j <= 18 {
		if signed {
			return -float64(num)
		}
		return float64(num)
	}
	if j == 0 {
		return fNaN
	}
	goto fallback
decimal:
	if c == '.' {
		j = 0
		for i++; i < uint(len(s)); i++ {
			c = s[i]
			if '0' <= c && c <= '9' {
				dec = 10*dec + uint64(c-'0')
				j++
				continue
			}
			if j > 0 {
				goto scientific
			}
			return fNaN
		}
		if 0 < j && j <= 18 {
			f = float64(num) + float64(dec)*math.Pow10(-j)
			if signed {
				return -f
			}
			return f
		}
		if j == 0 {
			return fNaN
		}
		goto fallback
	}
scientific:
	if c == 'e' || c == 'E' {
		signed := false
		exp := 0
		jj := 0
		for i++; i < uint(len(s)); i++ {
			c = s[i]
			if '0' <= c && c <= '9' {
				jj++
				exp = 10*exp + int(c-'0')
				continue
			}
			if jj == 0 {
				switch c {
				case '-':
					signed = true
					continue
				case '+':
					continue
				}
			}
			return fNaN
		}
		if jj == 0 {
			return fNaN
		}
		if exp > 300 {
			goto fallback
		}
		if signed {
			f = (float64(num) + float64(dec)*math.Pow10(-j)) * math.Pow10(-exp)
		} else {
			f = float64(num)*math.Pow10(exp) + float64(dec)*math.Pow10(exp-j)
		}
		goto done
	}
	return fNaN
done:
	if signed {
		return -f
	}
	return f
fallback:
	f, _ = strconv.ParseFloat(s, 64)
	return f

}

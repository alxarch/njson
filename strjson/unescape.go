package strjson

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Unescape unescapes the JSON string into a buffer.
// The buffer must have sufficient size.
// A buffer with the same size as the input string
// is big enough for all cases.
func Unescape(dst []byte, s string) int {
	if len(s) == 0 {
		return 0
	}
	var (
		c    byte
		i, j int
		pos  int
		ss   string
	)
unescape:
	for pos = i; 0 <= i && i < len(s); i++ {
		if c = s[i]; c != delimEscape {
			continue
		}
		// Copy escaped part of the string
		if pos < i && 0 <= pos {
			j += writeStringAt(dst, s[pos:i], j)
		}

		if i++; 0 <= i && i < len(s) {
			switch c = s[i]; c {
			case '"', '/', '\\':
				// keep c
			case 'n':
				c = '\n'
			case 'r':
				c = '\r'
			case 't':
				c = '\t'
			case 'b':
				c = '\b'
			case 'f':
				c = '\f'
			case 'u':
				ss = s[i:]
				goto unescapeRune
			default:
				// Invalid escape, append as is
				goto invalidUnescape
			}
			goto unescapeByte
		}
		// There's an escape slash at the last byte of the string.
		if 0 <= j && j < len(dst) {
			dst[j] = '\\'
			j++
		}
		return j
	}
	if 0 <= pos && pos < len(s) {
		j += writeStringAt(dst, s[pos:], j)
	}
	return j
unescapeRune:
	if len(ss) > 4 {
		r1 := rune(fromHex(ss[1])) << 12
		r1 |= rune(fromHex(ss[2])) << 8
		r1 |= rune(fromHex(ss[3])) << 4
		r1 |= rune(fromHex(ss[4]))
		if r1 < utf8.RuneSelf {
			c = byte(r1)
			i += 4
			goto unescapeByte
		}
		i += 5
		if utf16.IsSurrogate(r1) {
			if len(ss) > 10 && ss[5] == delimEscape && ss[6] == 'u' {
				i += 6
				r2 := rune(fromHex(ss[7])) << 12
				r2 |= rune(fromHex(ss[8])) << 8
				r2 |= rune(fromHex(ss[9])) << 4
				r2 |= rune(fromHex(ss[10]))
				r1 = utf16.DecodeRune(r1, r2)
			} else {
				r1 = utf8.RuneError
			}
		}
		if 0 <= j && j < len(dst) {
			j += utf8.EncodeRune(dst[j:], r1)
		}
		goto unescape
	}
	// Treat is as an invalid escape.
	// This ensures max unescaped length is <= to string length
	c = 'u'
invalidUnescape:
	if 0 <= j && j < len(dst) {
		dst[j] = '\\'
		j++
	}
unescapeByte:
	if 0 <= j && j < len(dst) {
		dst[j] = c
		j++
	}
	i++
	goto unescape
}

const (
	delimEscape = '\\'
	delimString = '"'
)

func writeStringAt(dst []byte, s string, pos int) int {
	if 0 <= pos && pos < len(dst) {
		dst = dst[pos:]
		if len(dst) >= len(s) {
			return copy(dst[:len(s)], s)
		}
	}
	return 0
}

// AppendUnescaped appends the unescaped form of a JSON string to a buffer.
// It's a convenience wrapper around Unescape([]byte, string) int
func AppendUnescaped(dst []byte, s string) []byte {
	if len(s) == 0 {
		return dst
	}
	var (
		n      int
		offset = len(dst)
		buf    []byte
	)
	if n = offset + len(s); cap(dst) < n {
		buf = make([]byte, n)
		copy(buf, dst)
		dst = buf
		buf = buf[offset:]
	} else {
		buf = dst[len(dst):cap(dst)]
	}
	if n = Unescape(buf, s) + offset; 0 <= n && n <= cap(dst) {
		return dst[:n]
	}
	return nil
}

// Unescaped returns the unescaped form of an escaped JSON string.
// If the input string requires unescaping it allocates a new string
// so if possible use Unescape([]byte, string) int for best performance.
func Unescaped(s string) string {
	if len(s) == 0 {
		return ""
	}
	var (
		b   strings.Builder
		c   byte
		i   = strings.IndexByte(s, delimEscape)
		pos int
		ss  string
	)
	if i == -1 {
		return s
	}
	b.Grow(len(s))
unescape:
	for ; 0 <= i && i < len(s); i++ {
		if c = s[i]; c != delimEscape {
			continue
		}
		// Copy escaped part of the string
		if pos < i && 0 <= pos {
			b.WriteString(s[pos:i])
		}

		if i++; 0 <= i && i < len(s) {
			switch c = s[i]; c {
			case '"', '/', '\\':
				// keep c
			case 'n':
				c = '\n'
			case 'r':
				c = '\r'
			case 't':
				c = '\t'
			case 'b':
				c = '\b'
			case 'f':
				c = '\f'
			case 'u':
				ss = s[i:]
				goto unescapeRune
			default:
				// Invalid escape, append as is
				goto invalidUnescape
			}
			goto unescapeByte
		}
		// There's an escape slash at the last byte of the string.
		b.WriteByte('\\')
		return b.String()
	}
	if 0 <= pos && pos < len(s) {
		b.WriteString(s[pos:])
	}
	return b.String()
unescapeRune:
	if len(ss) > 4 {
		r1 := rune(fromHex(ss[1])) << 12
		r1 |= rune(fromHex(ss[2])) << 8
		r1 |= rune(fromHex(ss[3])) << 4
		r1 |= rune(fromHex(ss[4]))
		if r1 < utf8.RuneSelf {
			c = byte(r1)
			i += 4
			goto unescapeByte
		}
		i += 5
		if utf16.IsSurrogate(r1) {
			if len(ss) > 10 && ss[5] == delimEscape && ss[6] == 'u' {
				i += 6
				r2 := rune(fromHex(ss[7])) << 12
				r2 |= rune(fromHex(ss[8])) << 8
				r2 |= rune(fromHex(ss[9])) << 4
				r2 |= rune(fromHex(ss[10]))
				r1 = utf16.DecodeRune(r1, r2)
			} else {
				r1 = utf8.RuneError
			}
		}
		b.WriteRune(r1)
		pos = i
		goto unescape
	}
	// Treat is as an invalid escape.
	// This ensures max unescaped length is <= to string length
	c = 'u'
invalidUnescape:
	b.WriteByte('\\')
unescapeByte:
	b.WriteByte(c)
	i++
	pos = i
	goto unescape

}

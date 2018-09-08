package strjson

import (
	"unicode/utf16"
	"unicode/utf8"
)

// func readHex2(s string, buf []byte, i int) bool {
// 	if 0 <= i && i < len(s) && len(buf) > 2 {
// 		if s = s[i:]; len(s) > 1 {
// 			buf[1] = toHex(s[0])
// 			buf[2] = toHex(s[1])
// 			buf[0] |= buf[1]<<4 | buf[2]
// 			return buf[1] != 0xff && buf[2] != 0xff
// 		}
// 	}
// 	return false
// }

// func readHex(s string, b []byte) byte {
// 	b[2] = FromHex(s[1])
// 	b[1] = FromHex(s[0])
// 	b[0] = b[1]<<4 | b[2]
// 	return b[2] & b[1]
// }

// func readRune(b []byte) rune {
// 	_ = b[1]
// 	return rune(uint16(b[0])<<8 | uint16(b[1]))
// }

// Quoted appends JSON quoted value of s.
func Quoted(b []byte, s string) []byte {
	b = append(b, delimString)
	b = Escape(b, s, false)
	b = append(b, delimString)
	return b
}

const (
	delimEscape = '\\'
	delimString = '"'
)

// MaxUnescapedLen returns a safe size for a buffer to fill with unescaped s bytes.
func MaxUnescapedLen(s string) int {
	// The only cases that need more characters than the input are:
	// - an invalid unicode escape at the end of the string with no hex digits following (ie `foo\u`)
	// - an invalid escape (ie `foo\zbar`)
	// - an escape slash at the end of the string (ie `foo\`)
	// In the worst case senario where the whole string is comprised of wrong escapes
	// we have to allocate 3 bytes for every 2 bytes of the input.
	return 3 * len(s) / 2
}

func writeStringAt(buf []byte, s string, i int) int {
	if len(s) > 0 {
		if 0 <= i && i < len(buf) {
			if buf = buf[i:]; len(buf) >= len(s) {
				buf = buf[:len(s)]
				return copy(buf, s)
			}
		}
	}
	return 0
}

func sliceAt(s string, i int) string {
	if 0 <= i && i <= len(s) {
		return s[i:]
	}
	return ""
}
func writeByteAt(buf []byte, c byte, i int) {
	if 0 <= i && i < len(buf) {
		buf[i] = c
	}
}
func writeAt(w, p []byte, i int) int {
	if 0 <= i && i < len(w) {
		if w = w[i:]; len(w) >= len(p) {
			w = w[:len(p)]
			return copy(w, p)
		}
	}
	return 0
}

func UnescapeRune(dst []byte, s string, i int) ([]byte, int) {
	buf := [utf8.UTFMax]byte{}
	r1 := utf8.RuneError
	r2 := utf8.RuneError
	if len(s) > 5 {
		r1 = rune(FromHex(s[2])) << 12
		r1 |= rune(FromHex(s[3])) << 8
		if r1 == 0 {
			return append(dst, FromHex(s[4])<<4|FromHex(s[5])), i + 6
		}
		r1 |= rune(FromHex(s[4])) << 4
		r1 |= rune(FromHex(s[5]))
		i += 6
		if utf16.IsSurrogate(r1) {
			if len(s) > 11 && s[6] == delimEscape && s[7] == 'u' {
				r2 = rune(FromHex(s[8])) << 12
				r2 |= rune(FromHex(s[9])) << 8
				r2 |= rune(FromHex(s[10])) << 4
				r2 |= rune(FromHex(s[11]))
				i += 6
			}
			r1 = utf16.DecodeRune(r1, r2)
		}
	}

	switch utf8.EncodeRune(buf[:], r1) {
	case 1:
		dst = append(dst, buf[0])
	case 2:
		dst = append(dst, buf[0], buf[1])
	default:
		dst = escapeError(dst)
	case 3:
		dst = append(dst, buf[0], buf[1], buf[2])
	case 4:
		dst = append(dst, buf[0], buf[1], buf[2], buf[3])
	}
	return dst, i
}

// Unescape appends the unescaped form of a string to a buffer.
func Unescape(dst []byte, s string) []byte {
	if len(s) == 0 {
		return dst
	}
	var (
		c    byte
		i, j int
		ss   string
	)
	if j = len(dst) + len(s); cap(dst) < j {
		buf := make([]byte, len(dst), j)
		copy(buf, dst)
	}
unescape:
	for j = i; 0 <= i && i < len(s); i++ {
		if c = s[i]; c != '\\' {
			continue
		}
		if 0 <= j && j < i {
			dst = append(dst, s[j:i]...)
		}

		if j = i + 1; 0 <= j && j < len(s) {
			c = s[j]
		} else {
			// There's an escape slash at the end of the string.
			return append(dst, c)
		}
		switch c {
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
			dst = append(dst, '\\', c)
			i += 2
			goto unescape
		}
		dst = append(dst, c)
		i += 2
		goto unescape
	}
	// if len(s) > 0 {
	// 	dst = append(dst, s...)
	// }
	if 0 <= j && j < len(s) {
		dst = append(dst, s[j:]...)
	}
	return dst
unescapeRune:
	buf := [utf8.UTFMax]byte{}
	r1 := utf8.RuneError
	r2 := utf8.RuneError
	if len(ss) > 5 {
		i += 6
		r1 = rune(FromHex(ss[2])) << 12
		r1 |= rune(FromHex(ss[3])) << 8
		if r1 == 0 {
			dst = append(dst, FromHex(ss[4])<<4|FromHex(ss[5]))
			goto unescape
		}
		r1 |= rune(FromHex(ss[4])) << 4
		r1 |= rune(FromHex(ss[5]))
		if utf16.IsSurrogate(r1) {
			if len(ss) > 11 && ss[6] == delimEscape && ss[7] == 'u' {
				i += 6
				r2 = rune(FromHex(ss[8])) << 12
				r2 |= rune(FromHex(ss[9])) << 8
				r2 |= rune(FromHex(ss[10])) << 4
				r2 |= rune(FromHex(ss[11]))
			}
			r1 = utf16.DecodeRune(r1, r2)
		}
	}

	switch utf8.EncodeRune(buf[:], r1) {
	case 1:
		dst = append(dst, buf[0])
	case 2:
		dst = append(dst, buf[0], buf[1])
	default:
		dst = escapeError(dst)
	case 3:
		dst = append(dst, buf[0], buf[1], buf[2])
	case 4:
		dst = append(dst, buf[0], buf[1], buf[2], buf[3])
	}
	goto unescape
}

// // UnescapeTo unescapes a string inside dst buffer which must have sufficient size (ie 3*len(s)/2).
// func UnescapeTo(dst []byte, s string) (n int) {
// 	if n = strings.IndexByte(s, delimEscape); n == -1 {
// 		n = len(s)
// 	}

// 	var (
// 		c      byte
// 		r1, r2 rune
// 		buf    = [utf8.UTFMax]byte{}
// 		b1, b2 = buf[:], buf[1:]
// 		i      = copy(dst, s[:n])
// 	)
// 	for ; 0 <= i && i < len(s); i++ {
// 		if c = s[i]; c != '\\' {
// 			dst[n] = c
// 			n++
// 			continue
// 		}
// 		if i++; i == len(s) {
// 			n += encodeError(dst)
// 			return
// 		}
// 		switch c = s[i]; c {
// 		case '"', '/', '\\':
// 			dst[n] = c
// 			n++
// 		case 'u':
// 			r1 = utf8.RuneError
// 			if i+4 < len(s) && readHex(s[i+1:], b1)&readHex(s[i+3:], b2) != 0xff {
// 				r1 = readRune(b1)
// 				i += 4
// 			}
// 			switch {
// 			case r1 == utf8.RuneError:
// 			case utf8.ValidRune(r1):
// 			case utf16.IsSurrogate(r1):
// 				r2 = utf8.RuneError
// 				if i+6 < len(s) && s[i+1] == delimEscape && s[i+2] == 'u' {
// 					i += 2
// 					if readHex(s[i+1:], b1)&readHex(s[i+3:], b2) != 0xff {
// 						r2 = readRune(buf[:])
// 						i += 4
// 					}
// 				}
// 				// Will be utf8.RuneError if not a valid surrogate pair
// 				r1 = utf16.DecodeRune(r1, r2)
// 			default:
// 				r1 = utf8.RuneError
// 			}
// 			// Safe to write to dst because if r1 size is 3 if any error occured
// 			n += utf8.EncodeRune(dst[n:], r1)
// 		default:
// 			if c, ok := namedEscapes[c]; ok {
// 				dst[n] = c
// 				n++
// 			} else {
// 				n += encodeError(dst)
// 			}
// 		}
// 	}
// 	return
// }

package strjson

import (
	"strings"
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
// 	b[2] = fromHex(s[1])
// 	b[1] = fromHex(s[0])
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
	b = Escape(b, s)
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

// Unescape appends the unescaped form of a string to a buffer.
func Unescape(dst []byte, s string) []byte {
	var (
		c      byte
		r1, r2 rune
		buf    = [utf8.UTFMax]byte{}
		i      = strings.IndexByte(s, '\\')
		ss     string
	)
	if 0 <= i && i < len(s) {
		if cap(dst)-len(dst) < len(s) {
			// Buffer probably doesn't have enough capacity to store the string
			tmp := make([]byte, len(dst)+i, len(dst)+MaxUnescapedLen(s))
			copy(tmp, dst)
			copy(tmp, s[:i])
			dst = tmp
		} else {
			dst = append(dst, s[:i]...)
		}
	} else {
		dst = append(dst, s...)
		return dst
	}
	for ; 0 <= i && i < len(s); i++ {
		if c = s[i]; c != '\\' {
			dst = append(dst, c)
			continue
		}
		// dst = append(dst, s[:i]...)
		ss = s[i:]
		i++
		// i = -1
		if len(ss) > 1 {
			c = ss[1]
			// s = s[2:]
		} else {
			c = '?'
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
			r1 = utf8.RuneError
			if len(ss) > 5 {
				buf[0] = fromHex(ss[2])
				buf[1] = fromHex(ss[3])
				buf[2] = fromHex(ss[4])
				buf[3] = fromHex(ss[5])
				if buf[0]&buf[1]&buf[2]&buf[3] != 0xff {
					r1 = rune(uint16(buf[0])<<12 | uint16(buf[1])<<8 | uint16(buf[2])<<4 | uint16(buf[3]))
					i += 4
					// s = s[4:]
				}
			}
			switch {
			case r1 < utf8.RuneSelf:
				dst = append(dst, byte(r1))
				continue
			case utf16.IsSurrogate(r1):
				r2 = utf8.RuneError
				if len(ss) > 11 && ss[6] == delimEscape && ss[7] == 'u' {
					i += 2
					buf[0] = fromHex(ss[8])
					buf[1] = fromHex(ss[9])
					buf[2] = fromHex(ss[10])
					buf[3] = fromHex(ss[11])
					if buf[0]&buf[1]&buf[2]&buf[3] != 0xff {
						r2 = rune(uint16(buf[0])<<12 | uint16(buf[1])<<8 | uint16(buf[2])<<4 | uint16(buf[3]))
						i += 4
						// s = s[6:]
					}
				}
				r1 = utf16.DecodeRune(r1, r2)
			}
			switch utf8.EncodeRune(buf[:], r1) {
			case 1:
				dst = append(dst, buf[0])
			case 2:
				dst = append(dst, buf[0], buf[1])
			default:
				fallthrough
			case 3:
				dst = append(dst, buf[0], buf[1], buf[2])
			case 4:
				dst = append(dst, buf[0], buf[1], buf[2], buf[3])
			}
			continue
		default:
			// append rune error
			dst = append(dst, 0xef, 0xbf, 0xbc)
			continue
		}
		// append escaped byte
		dst = append(dst, c)

	}
	// if len(s) > 0 {
	// 	dst = append(dst, s...)
	// }
	return dst
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

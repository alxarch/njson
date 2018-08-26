package strjson

import (
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

func readHex(s string, b []byte) (ok bool) {
	_ = s[1]
	_ = b[2]
	b[1] = fromHex(s[0])
	b[2] = fromHex(s[1])
	b[0] = (b[1] << 4) | b[2]
	ok = b[1] != 0xff && b[2] != 0xff
	return
}

func readRune(b []byte) rune {
	_ = b[1]
	return rune(uint16(b[0])<<8 | uint16(b[1]))
}

// Unescape appends unescaped s to b
func Unescape(b []byte, s string) []byte {
	// Ensure b has enough space to unescape s
	offset := len(b)
	n := offset + len(s)
	if cap(b) < n {
		b = append(b, s...)
	} else {
		b = b[:n]
	}
	n = offset + UnescapeTo(b[:offset], s)
	return b[:n]
}

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

// go:generate go run ../njsonutil/cmd/genmask/genmask.go -w hex.go -pkg strjson ToHex FromHex

func toHex(c byte) byte {
	return maskToHex[c]
	// switch {
	// case 0 <= c && c <= 9:
	// 	return c + '0'
	// case 10 <= c && c <= 15:
	// 	return c + 'A' - 10
	// default:
	// 	return 0xff
	// }
}

func fromHex(c byte) byte {
	return maskFromHex[c]
	// switch {
	// case '0' <= c && c <= '9':
	// 	return c - '0'
	// case 'a' <= c && c <= 'z':
	// 	return 10 + c - 'a'
	// case 'A' <= c && c <= 'Z':
	// 	return 10 + c - 'A'
	// default:
	// 	return 0xff
	// }
}

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

func encodeError(buf []byte) int {
	_ = buf[2]
	buf[0] = 0xef
	buf[1] = 0xbf
	buf[2] = 0xbc
	return 3
}

// UnescapeTo unescapes a string inside dst buffer which must have sufficient size (ie 3*len(s)/2).
func UnescapeTo(dst []byte, s string) (n int) {
	if n = strings.IndexByte(s, delimEscape); n == -1 {
		n = len(s)
	}

	var (
		c      byte
		end    = len(s)
		r1, r2 rune
		buf    = [utf8.UTFMax]byte{}
		i      = copy(dst, s[:n])
	)
	for ; i < end; i++ {
		if c = s[i]; c != '\\' {
			dst[n] = c
			n++
			continue
		}
		if i++; i == end {
			n += encodeError(dst)
			return
		}
		switch c = s[i]; c {
		case '"', '/', '\\':
			dst[n] = c
			n++
		case 'u':
			r1 = utf8.RuneError
			if i+4 < end && readHex(s[i+1:], buf[:]) && readHex(s[i+3:], buf[1:]) {
				r1 = readRune(buf[:])
				i += 4
			}
			switch {
			case r1 == utf8.RuneError:
			case utf8.ValidRune(r1):
			case utf16.IsSurrogate(r1):
				r2 = utf8.RuneError
				if i+2 < end && s[i+1] == delimEscape && s[i+2] == 'u' {
					i += 2
					if i+4 < end && readHex(s[i+1:], buf[:]) && readHex(s[i+3:], buf[1:]) {
						r2 = readRune(buf[:])
						i += 4
					}
				}
				// Will be utf8.RuneError if not a valid surrogate pair
				r1 = utf16.DecodeRune(r1, r2)
			default:
				r1 = utf8.RuneError
			}
			// Safe to write to dst because if r1 size is 3 if any error occured
			n += utf8.EncodeRune(dst[n:], r1)
		default:
			if c, ok := namedEscapes[c]; ok {
				dst[n] = c
				n++
			} else {
				n += encodeError(dst)
			}
		}
	}
	return
}

// Escape appends escaped string to a buffer.
func Escape(dst []byte, s string) []byte {
	for _, r := range s {
		switch {
		case r < utf8.RuneSelf:
			switch r {
			case '"', '\\', '/':
				dst = append(dst, '\\', byte(r))
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			case '\b':
				dst = append(dst, '\\', 'b')
			case '\f':
				dst = append(dst, '\\', 'f')
			default:
				if unicode.IsPrint(r) {
					dst = append(dst, byte(r))
				} else {
					dst = append(dst, '\\', 'u', '0', '0',
						toHex(byte(r)>>4),
						toHex(byte(r)),
					)
				}
			}
		case unicode.IsPrint(r):
			buf := [utf8.UTFMax]byte{}
			dst = append(dst, buf[:utf8.EncodeRune(buf[:], r)]...)
		case r < 0x10000:
			return append(dst, '\\', 'u',
				toHex(byte(r>>12)),
				toHex(byte(r>>8)&0xF),
				toHex(byte(r)>>4),
				toHex(byte(r)&0xF),
			)
		case utf16.IsSurrogate(r):
			dst = escapeUTF16(dst, r)
		default:
			dst = escapeError(dst)
		}
	}
	return dst
}

// EscapeBytes appends escaped bytes.
func EscapeBytes(dst []byte, s []byte) []byte {
	var (
		n    = len(s)
		i, j int
		c    byte
		r    rune
	)
	for i, j = 0, 1; i < n; i, j = i+j, 1 {
		if c = s[i]; c < utf8.RuneSelf {
			switch c {
			case '"', '\\', '/':
				dst = append(dst, '\\', c)
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			case '\b':
				dst = append(dst, '\\', 'b')
			case '\f':
				dst = append(dst, '\\', 'f')
			default:
				if unicode.IsPrint(rune(c)) {
					dst = append(dst, c)
				} else {
					dst = escapeByte(dst, c)
				}
			}
		} else if r, j = utf8.DecodeRune(s[i:]); unicode.IsPrint(r) {
			dst = append(dst, s[i:i+j]...)
		} else {
			dst = EscapeRune(dst, r)
		}
	}
	return dst
}

func escapeByte(dst []byte, c byte) []byte {
	return append(dst, '\\', 'u', '0', '0',
		toHex(c>>4),
		toHex(c),
	)
}

func escapeUTF8(dst []byte, r rune) []byte {
	return append(dst, '\\', 'u',
		toHex(byte(r>>12)),
		toHex(byte(r>>8)&0xF),
		toHex(byte(r)>>4),
		toHex(byte(r)&0xF),
	)
}
func escapeError(dst []byte) []byte {
	return append(dst, `\uFFFD`...)
}
func escapeUTF16(dst []byte, r rune) []byte {
	if r1, r2 := utf16.EncodeRune(r); r1 != utf8.RuneError {
		return append(dst, '\\', 'u',
			toHex(byte(r1>>12)),
			toHex(byte(r1>>8)&0xF),
			toHex(byte(r1)>>4),
			toHex(byte(r1)&0xF),
			'\\', 'u',
			toHex(byte(r2>>12)),
			toHex(byte(r2>>8)&0xF),
			toHex(byte(r2)>>4),
			toHex(byte(r2)&0xF),
		)

	}
	return escapeError(dst)

}

// EscapeRune escapes a rune to JSON unicode escape.
func EscapeRune(dst []byte, r rune) []byte {
	switch {
	case r < utf8.RuneSelf:
		return escapeByte(dst, byte(r))
	case r > utf8.MaxRune:
		r = utf8.RuneError
		fallthrough
	case r < 0x10000:
		return escapeUTF8(dst, r)
	default:
		return escapeUTF16(dst, r)
	}
}

var namedEscapes = map[byte]byte{
	'n': '\n',
	'r': '\r',
	't': '\t',
	'b': '\b',
	'f': '\f',
}

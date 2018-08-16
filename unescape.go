package njson

import (
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

func readHex(b1, b2 byte, b *byte) bool {
	if b1 = ToHexDigit(b1); b1 == 0xff {
		return false
	}
	if b2 = ToHexDigit(b2); b2 == 0xff {
		return false
	}
	*b = b1<<4 | b2
	return true
}

func readHexByte(s string, b []byte, i, j int) bool {
	var d byte
	if d = ToHexDigit(s[i]); d == 0xff {
		return false
	}
	b[j] = d << 4
	if d = ToHexDigit(s[i+1]); d == 0xff {
		return false
	}
	b[j] |= d
	return true
}

func readRune(b0, b1 byte) rune {
	return rune(uint16(b0)<<8 | uint16(b1))
}

func AppendUnescaped(b []byte, src string) []byte {
	offset := len(b)
	n := offset + len(src)
	if cap(b) < n {
		b = append(b, src...)
	} else {
		b = b[:n]
	}
	n = offset + Unescape(b[:offset], src)
	return b[:n]
}

func Unescape(dst []byte, src string) (n int) {
	if n = strings.IndexByte(src, delimEscape); n == -1 {
		n = len(src)
	}

	var (
		c      byte
		end    = len(src)
		r1, r2 rune
		buf    = [utf8.UTFMax]byte{}
		i      = copy(dst, src[:n])
	)
	for ; i < end; i++ {
		if c = src[i]; c != '\\' {
			dst[n] = c
			n++
			continue
		}
		if i++; i == end {
			dst[n] = c
			n++
			return
		}
		switch c = src[i]; c {
		case '"', '/', '\\':
			dst[n] = c
			n++
		case 'u':
			r1 = utf8.RuneError
			if i+4 < end {
				if readHex(src[i+1], src[i+2], &buf[0]) && readHex(src[i+3], src[i+4], &buf[1]) {
					r1 = readRune(buf[0], buf[1])
				}
				i += 4
			}
			switch {
			case r1 == utf8.RuneError:
			case utf8.ValidRune(r1):
			case utf16.IsSurrogate(r1):
				r2 = utf8.RuneError
				if i+6 < end && src[i+1] == delimEscape && src[i+2] == 'u' {
					if readHex(src[i+3], src[i+4], &buf[0]) && readHex(src[i+5], src[i+6], &buf[1]) {
						r2 = readRune(buf[0], buf[1])
					}
					i += 6
				}
				r1 = utf16.DecodeRune(r1, r2)
			default:
				r1 = utf8.RuneError
			}
			// Safe to write to dst because if r1 size is 2 if any error occured
			n += utf8.EncodeRune(dst[n:], r1)
		default:
			if c = ToNamedEscape(c); c != 0 {
				dst[n] = c
				n++
			} else {
				n += utf8.EncodeRune(dst[n:], utf8.RuneError)
			}
		}
	}
	return
}

func QuoteString(s string) string {
	out := make([]byte, 1, ((3*len(s))/2)+2)
	out[0] = '"'
	out = EscapeString(out, s)
	out = append(out, '"')
	return b2s(out)
}

func EscapeString(dst []byte, s string) []byte {
	buf := [3 * utf8.UTFMax]byte{}
	for _, r := range s {
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
				dst = append(dst, buf[:utf8.EncodeRune(buf[:], r)]...)
				continue
			}
			switch utf8.RuneLen(r) {
			case 1:
				dst = append(dst, '\\', 'u', '0', '0',
					ToHex(byte(r)>>4),
					ToHex(byte(r)),
				)
			case 2:
				utf8.EncodeRune(buf[:], r)
				dst = append(dst, '\\', 'u',
					ToHex(buf[0]>>4),
					ToHex(buf[0]),
					ToHex(buf[1]>>4),
					ToHex(buf[1]),
				)
			default:
				r1, r2 := utf16.EncodeRune(r)
				dst = append(dst, '\\', 'u',
					ToHex(byte(uint16(r1)>>12)),
					ToHex(byte(uint16(r1)>>8)),
					ToHex(byte(uint16(r1)>>4)),
					ToHex(byte(uint16(r1))),
					'\\', 'u',
					ToHex(byte(uint16(r2)>>12)),
					ToHex(byte(uint16(r2)>>8)),
					ToHex(byte(uint16(r2)>>4)),
					ToHex(byte(uint16(r2))),
				)
			}
		}
	}
	return dst
}

func EscapeBytes(dst []byte, s []byte) []byte {
	buf := [3 * utf8.UTFMax]byte{}
	n := len(s)
	for i := 0; i < n; i++ {
		switch c := s[i]; c {
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
			r, j := utf8.DecodeRune(s[i:])
			if r == utf8.RuneError {
				dst = append(dst, '\\', 'u', 'F', 'F', 'F', 'D')
				continue
			}
			if unicode.IsPrint(r) {
				switch j {
				case 1:
					dst = append(dst, s[i])
				case 2:
					dst = append(dst, s[i], s[i+1])
					i++
				case 3:
					dst = append(dst, s[i], s[i+1], s[i+2])
					i += 2
				case 4:
					dst = append(dst, s[i], s[i+1], s[i+2], s[i+3])
					i += 3
				}
				continue
			}
			switch j {
			case 1:
				dst = append(dst, '\\', 'u', '0', '0',
					ToHex(byte(r)>>4),
					ToHex(byte(r)),
				)
			case 2:
				utf8.EncodeRune(buf[:], r)
				dst = append(dst, '\\', 'u',
					ToHex(buf[0]>>4),
					ToHex(buf[0]),
					ToHex(buf[1]>>4),
					ToHex(buf[1]),
				)
			default:
				r1, r2 := utf16.EncodeRune(r)
				dst = append(dst, '\\', 'u',
					ToHex(byte(uint16(r1)>>12)),
					ToHex(byte(uint16(r1)>>8)),
					ToHex(byte(uint16(r1)>>4)),
					ToHex(byte(uint16(r1))),
					'\\', 'u',
					ToHex(byte(uint16(r2)>>12)),
					ToHex(byte(uint16(r2)>>8)),
					ToHex(byte(uint16(r2)>>4)),
					ToHex(byte(uint16(r2))),
				)
			}
		}
	}
	return dst
}

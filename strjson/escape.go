package strjson

import (
	"unicode/utf8"
)

// // EscapeRune escapes a rune to JSON unicode escape.
// func EscapeRune(dst []byte, r rune) []byte {
// 	switch {
// 	case r < utf8.RuneSelf:
// 		return escapeByte(dst, byte(r))
// 	case r > utf8.MaxRune:
// 		r = utf8.RuneError
// 		fallthrough
// 	case r < 0x10000:
// 		return escapeUTF8(dst, r)
// 	default:
// 		return escapeUTF16(dst, r)
// 	}
// }

// func appendRuneAt(dst []byte, r rune, i int) []byte {
// retry:
// 	if 0 <= i && i < len(dst) {
// 		if buf := dst[i:]; len(buf) > 3 {
// 			buf[0] = ToHex(byte(r>>12)]
// 			buf[1] = ToHex(byte(r>>8)&0xF]
// 			buf[2] = ToHex(byte(r)>>4]
// 			buf[3] = ToHex(byte(r)&0xF]
// 			return dst
// 		}
// 	}
// 	buf := make([]byte, 2*len(dst)+8)
// 	copy(buf, dst)
// 	goto retry
// }

// func appendAt(dst []byte, b []byte, i int) []byte {
// retry:
// 	if 0 <= i && i < len(dst) {
// 		if buf := dst[i:]; len(buf) > len(b) {
// 			copy(buf, b)
// 			return dst
// 		}
// 	}
// 	buf := make([]byte, 2*len(dst)+len(b))
// 	copy(buf, dst)
// 	goto retry
// }
// func appendStringAt(dst []byte, s string, i int) []byte {
// retry:
// 	if 0 <= i && i < len(dst) {
// 		if buf := dst[i:]; len(buf) >= len(s) {
// 			copy(buf, s)
// 			return dst
// 		}
// 	}
// 	buf := make([]byte, 2*len(dst)+len(s))
// 	copy(buf, dst)
// 	dst = buf
// 	goto retry
// }

// func appendEscapedAt(dst []byte, c byte, i int) []byte {
// 	if 0 <= i && i < len(dst) {
// 		if buf := dst[i:]; len(buf) > 1 {
// 			buf[0] = '\\'
// 			buf[1] = c
// 			return dst
// 		}
// 	}
// 	dst = append(dst, '\\', c)
// 	return dst[:cap(dst)]
// }

// func appendByteAt(dst []byte, c byte, i int) []byte {
// 	if 0 <= i && i < len(dst) {
// 		dst[i] = c
// 		return dst
// 	}
// 	buf := make([]byte, 2*len(dst)+1)
// 	copy(buf, dst)
// 	if 0 <= i && i < len(buf) {
// 		buf[i] = c
// 		return buf
// 	}
// 	return nil
// }
// func hexRune(buf []byte, r rune) []byte {
// 	if len(buf) > 5 {
// 		buf[2] = ToHex(byte(r>>12)]
// 		buf[3] = ToHex(byte(r>>8)&0xF]
// 		buf[4] = ToHex(byte(r)>>4]
// 		buf[5] = ToHex(byte(r)&0xF]
// 	}
// 	return buf
// }

// // Escape appends escaped string to a buffer.
// func Escape(dst []byte, s string) []byte {
// 	for _, r := range s {
// 		switch {
// 		case r < utf8.RuneSelf:
// 			switch r {
// 			case '"', '\\', '/':
// 				dst = append(dst, '\\', byte(r))
// 			case '\n':
// 				dst = append(dst, '\\', 'n')
// 			case '\r':
// 				dst = append(dst, '\\', 'r')
// 			case '\t':
// 				dst = append(dst, '\\', 't')
// 			case '\b':
// 				dst = append(dst, '\\', 'b')
// 			case '\f':
// 				dst = append(dst, '\\', 'f')
// 			default:
// 				if unicode.IsPrint(r) {
// 					dst = append(dst, byte(r))
// 				} else {
// 					dst = append(dst, '\\', 'u', '0', '0',
// 						ToHex(byte(r)>>4],
// 						ToHex(byte(r)],
// 					)
// 				}
// 			}
// 		case unicode.IsPrint(r):
// 			buf := [utf8.UTFMax]byte{}
// 			switch utf8.EncodeRune(buf[:], r) {
// 			case 1:
// 				dst = append(dst, buf[0])
// 			case 2:
// 				dst = append(dst, buf[0], buf[1])
// 			default:
// 				dst = escapeError(dst)
// 			case 3:
// 				dst = append(dst, buf[0], buf[1], buf[2])
// 			case 4:
// 				dst = append(dst, buf[0], buf[1], buf[2], buf[3])
// 			}
// 		case r < 0x10000:
// 			return append(dst, '\\', 'u',
// 				ToHex(byte(r>>12)],
// 				ToHex(byte(r>>8)&0xF],
// 				ToHex(byte(r)>>4],
// 				ToHex(byte(r)&0xF],
// 			)
// 		case utf16.IsSurrogate(r):
// 			dst = escapeUTF16(dst, r)
// 		default:
// 			dst = escapeError(dst)
// 		}
// 	}
// 	return dst
// }

// // EscapeBytes appends escaped bytes.
// func EscapeBytes(dst []byte, s []byte) []byte {
// 	var (
// 		i, j int
// 		c    byte
// 		r    rune
// 	)
// 	for i, j = 0, 1; 0 <= i && i < len(s); i, j = i+j, 1 {
// 		if c = s[i]; c < utf8.RuneSelf {
// 			switch c {
// 			case '"', '\\', '/':
// 				dst = append(dst, '\\', c)
// 			case '\n':
// 				dst = append(dst, '\\', 'n')
// 			case '\r':
// 				dst = append(dst, '\\', 'r')
// 			case '\t':
// 				dst = append(dst, '\\', 't')
// 			case '\b':
// 				dst = append(dst, '\\', 'b')
// 			case '\f':
// 				dst = append(dst, '\\', 'f')
// 			default:
// 				if unicode.IsPrint(rune(c)) {
// 					dst = append(dst, c)
// 				} else {
// 					dst = escapeByte(dst, c)
// 				}
// 			}
// 		} else if r, j = utf8.DecodeRune(s[i:]); unicode.IsPrint(r) {
// 			dst = append(dst, s[i:i+j]...)
// 		} else {
// 			dst = EscapeRune(dst, r)
// 		}
// 	}
// 	return dst
// }

func escapeByte(dst []byte, c byte) []byte {
	return append(dst, '\\', 'u', '0', '0', ToHex(c>>4), ToHex(c))
}

func escapeUTF8(dst []byte, r rune) []byte {
	return append(dst, '\\', 'u',
		ToHex(byte(r>>12)),
		ToHex(byte(r>>8)&0xF),
		ToHex(byte(r)>>4),
		ToHex(byte(r)&0xF),
	)
}
func escapeError(dst []byte) []byte {
	return append(dst, '\\', 'u', 'F', 'F', 'F', 'D')
}

const (

	// U+2028 is LINE SEPARATOR.
	unicodeLineSeparator = '\u2028'
	// U+2029 is PARAGRAPH SEPARATOR.
	unicodeParagraphSeparator = '\u2029'
)

func Escape(dst []byte, s string, HTML bool) []byte {
	if len(s) == 0 {
		return dst
	}
	var (
		c, e  byte
		r     rune
		size  int
		start int
		i     int
	)
	if size = len(dst) + len(s); cap(dst) < size {
		buf := make([]byte, len(dst), size)
		copy(buf, dst)
		dst = buf
	}
escape:
	for start = i; 0 <= i && i < len(s); i++ {
		c = s[i]
		e = ToJSON(c)
		if e == utf8.RuneSelf {
			continue
		}
		if e == 0xff {
			r, size = utf8.DecodeRuneInString(s[i:])
			switch r {
			case utf8.RuneError:
				if size > 1 {
					i += size - 1
					continue
				}
				if size == 0 {
					continue
				}
				fallthrough
			case unicodeLineSeparator, unicodeParagraphSeparator:
				if 0 <= start && start < i {
					dst = append(dst, s[start:i]...)
				}
				dst = escapeUTF8(dst, r)
				i += size
				goto escape
			}
			i += size - 1
			continue
		}
		if e == 1 && !HTML {
			continue
		}
		if 0 <= start && start < i {
			dst = append(dst, s[start:i]...)
		}
		switch e {
		case '\\':
			dst = append(dst, '\\', c)
		case 0, 1:
			dst = escapeByte(dst, c)
		default:
			dst = append(dst, '\\', e)
		}
		i++
		goto escape
	}
	if 0 <= start && start < len(s) {
		dst = append(dst, s[start:]...)
	}
	return dst
}

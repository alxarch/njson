package strjson

import (
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

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
			switch utf8.EncodeRune(buf[:], r) {
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
		i, j int
		c    byte
		r    rune
	)
	for i, j = 0, 1; 0 <= i && i < len(s); i, j = i+j, 1 {
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
	return append(dst, '\\', 'u', 'F', 'F', 'F', 'D')
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

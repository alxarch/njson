package strjson

import (
	"strings"
	"unicode/utf8"
)

// Escaped returns the JSON escaped form of string.
// It allocates a new string so if possible use AppendEscaped for best performance.
func Escaped(s string, HTML bool, quoted bool) string {
	if len(s) == 0 {
		if quoted {
			return `""`
		}
		return ""
	}
	var (
		b    strings.Builder
		r    rune
		c, e byte
		size int
		pos  uint
		i    uint
	)
	b.Grow(3 * len(s) / 2)
	if quoted {
		b.WriteByte('"')
	}
escape:
	for pos = i; i < uint(len(s)); i++ {
		c = s[i]
		e = toJSON(c)
		if e == utf8.RuneSelf {
			continue
		}
		if e == 0xff {
			r, size = utf8.DecodeRuneInString(s[i:])
			switch r {
			case utf8.RuneError, uLineSeparator, uParagraphSeparator:
				if size == 0 {
					continue
				}
				if pos < i {
					b.WriteString(s[pos:i])
				}
				b.Write([]byte{'\\', 'u',
					toHex(byte(r >> 12)),
					toHex(byte(r>>8) & 0x0F),
					toHex(byte(r) >> 4),
					toHex(byte(r) & 0x0F),
				})
				i += uint(size)
				goto escape
			default:
				i += uint(size - 1)
				continue
			}
		}
		if e == 1 {
			if HTML {
				e = 0
			} else {
				continue
			}
		}

		if pos < i {
			b.WriteString(s[pos:i])
		}
		i++
		if e == '\\' {
			b.Write([]byte{'\\', c})
		} else if e == 0 {
			b.Write([]byte{'\\', 'u', '0', '0',
				toHex(c >> 4),
				toHex(c & 0x0F),
			})
		} else {
			b.Write([]byte{'\\', e})
		}
		goto escape
	}
	if pos < uint(len(s)) {
		b.WriteString(s[pos:])
	}
	if quoted {
		b.WriteByte('"')
	}
	return b.String()
}

// AppendEscaped appends the JSON escaped form of a string to a buffer.
func AppendEscaped(dst []byte, s string, HTML bool) []byte {
	if len(s) == 0 {
		return dst
	}
	var (
		c, e byte
		r    rune
		size int
		pos  uint
		i    uint
	)
	if size = len(dst) + len(s); cap(dst) < size {
		if buf := make([]byte, len(dst), size); len(buf) >= len(dst) {
			copy(buf[:len(dst)], dst)
			dst = buf
		}
	}

escape:
	for pos = i; i < uint(len(s)); i++ {
		c = s[i]
		e = toJSON(c)
		if e == utf8.RuneSelf {
			continue
		}
		if e == 0xff {
			r, size = utf8.DecodeRuneInString(s[i:])
			switch r {
			case utf8.RuneError, uLineSeparator, uParagraphSeparator:
				if size == 0 {
					continue
				}
				if pos < i {
					dst = append(dst, s[pos:i]...)
				}
				dst = escapeUTF8(dst, r)
				i += uint(size)
				goto escape
			default:
				i += uint(size - 1)
				continue
			}
		}
		if e == 1 {
			if HTML {
				e = 0
			} else {
				continue
			}
		}

		if pos < i {
			dst = append(dst, s[pos:i]...)
		}
		i++
		if e == '\\' {
			dst = append(dst, '\\', c)
		} else if e == 0 {
			dst = escapeByte(dst, c)
		} else {
			dst = append(dst, '\\', e)
		}
		goto escape
	}
	if pos < uint(len(s)) {
		return append(dst, s[pos:]...)
	}
	return dst
}

const (

	// U+2028 is LINE SEPARATOR.
	uLineSeparator = '\u2028'
	// U+2029 is PARAGRAPH SEPARATOR.
	uParagraphSeparator = '\u2029'
)

func escapeByte(dst []byte, c byte) []byte {
	return append(dst, '\\', 'u', '0', '0', toHex(c>>4), toHex(c&0x0F))
}

func escapeUTF8(dst []byte, r rune) []byte {
	return append(dst, '\\', 'u',
		toHex(byte(r>>12)),
		toHex(byte(r>>8)&0x0F),
		toHex(byte(r)>>4),
		toHex(byte(r)&0x0F),
	)
}

func NeedsEscape(s string) bool {
	for i := 0; i < len(s); i++ {
		if toJSON(s[i]) != utf8.RuneSelf {
			return true
		}
	}
	return false
}
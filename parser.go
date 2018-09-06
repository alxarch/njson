package njson

import (
	"strings"
	"sync"
)

// Parser is a JSON parser.
type Parser struct {
	nodes []Node
	safe  bool
	n     int
	err   error
}

// ParseUnsafe parses JSON from a slice of bytes without copying it to a string.
// The contents of the slice should not be modified while using the result node.
func (p *Parser) ParseUnsafe(data []byte) (*Node, []byte, error) {
	p.reset(false)
	s := b2s(data)
	n := p.node()
	n.values = n.values[:0]
	pos := p.parseValue(' ', s, -1, n)
	if p.err != nil {
		return nil, data, p.err
	}
	if 0 <= pos && pos < len(data) {
		return n, data[pos:], nil
	}
	return n, nil, nil
}

func (p *Parser) Parse(s string) (*Node, string, error) {
	p.reset(true)
	n := p.node()
	n.values = n.values[:0]
	pos := p.parseValue(' ', s, -1, n)
	if p.err != nil {
		return nil, s, p.err
	}
	if 0 <= pos && pos < len(s) {
		return n, s[pos:], nil
	}
	return n, "", nil
}

const minNumNodes = 64

func (p *Parser) reset(safe bool) {
	p.err = nil
	// If we don't use unsafe pkg we're safe anyway
	p.safe = safe || safebytes
	p.n = 0
}

func (p *Parser) node() (n *Node) {
	if 0 <= p.n && p.n < len(p.nodes) {
		n = &p.nodes[p.n]
		// n.safe = p.safe
		p.n++
		return
	}

	p.nodes = make([]Node, len(p.nodes)*2+1)
	if len(p.nodes) > 0 {
		n = &p.nodes[0]
		// n.safe = p.safe
		p.n = 1
	}
	return
}

func (p *Parser) parseValue(c byte, s string, pos int, n *Node) int {
	// Skip space if needed
	if bytemapIsSpace[c] == 1 {
		for pos++; 0 <= pos && pos < len(s); pos++ {
			if c = s[pos]; bytemapIsSpace[c] == 0 {
				break
			}
		}
	}
	switch c {
	case delimString:
		n.info = vString
		n.values = n.values[:0]
		n.safe = p.safe
		if pos++; 0 < pos && pos < len(s) {
			n.raw = s[pos:]
			end := strings.IndexByte(n.raw, delimString)
			if end--; 0 <= end && end < len(n.raw) {
				if n.raw[end] == delimEscape {
					end += 2
					for ; 0 <= end && end < len(n.raw); end++ {
						switch n.raw[end] {
						case delimString:
							n.raw = n.raw[:end]
							end++
							return end + pos
						case delimEscape:
							end++
						}
					}
				} else if end++; 0 <= end && end <= len(n.raw) {
					n.raw = n.raw[:end]
					end++
					return end + pos
				}
			} else if end == -1 {
				n.raw = ""
				return pos + 1
			}
		}
		return p.abort(pos-1, TypeString, nil, delimString)
	case delimBeginObject:
		n.info = vObject
		// Skip space after '{'
		for pos++; 0 <= pos && pos < len(s); pos++ {
			if c = s[pos]; bytemapIsSpace[c] == 0 {
				goto isEmptyObject
			}
		}
		return p.eof(TypeObject)
	isEmptyObject:
		if c == delimEndObject {
			n.values = n.values[:0]
			return pos + 1
		}
		v := p.node()
		if cap(n.values) > 0 {
			n.values = n.values[:cap(n.values)]
		} else {
			n.values = make([]*Node, 8)
		}

		i := 0
		var key string

	readObject:
		for pos++; 0 <= pos && pos < len(s) && c == delimString; pos++ {
			// This slices the string after the opening quote
			key = s[pos:]
			for end := 0; 0 <= end && end < len(key); end++ {
				switch key[end] {
				case delimString:
					v.key = key[:end]
					// Skip space after '"'
					for pos += end + 1; 0 <= pos && pos < len(s); pos++ {
						if c = s[pos]; bytemapIsSpace[c] == 0 {
							// goto used to return proper eof error without checks
							goto isKey
						}
					}
					break
				case delimEscape:
					end++
				}
			}
			return p.eof(TypeKey)
		isKey:
			// Check for ':'
			if c != delimNameSeparator {
				return p.abort(pos, TypeKey, c, delimNameSeparator)
			}

			// Skip space after ':'
			for pos++; 0 <= pos && pos < len(s); pos++ {
				if c = s[pos]; bytemapIsSpace[c] == 0 {
					break
				}
			}
			pos = p.parseValue(c, s, pos, v)
			if p.err != nil {
				return pos
			}

			// Skip space after value
			for ; 0 <= pos && pos < len(s); pos++ {
				if c = s[pos]; bytemapIsSpace[c] == 0 {
					break
				}
			}

			switch c {
			case delimValueSeparator:
				// Skip space after ','
				for pos++; 0 <= pos && pos < len(s); pos++ {
					if c = s[pos]; bytemapIsSpace[c] == 0 {
						// Append value
						n.append(v, i)
						i++
						v = p.node()
						continue readObject
					}
				}
				return p.eof(TypeObject)
			case delimEndObject:
				n.append(v, i)
				if i++; 0 <= i && i <= cap(n.values) {
					n.values = n.values[:i]
				}
				return pos + 1
			default:
				return p.abort(pos, TypeObject, c, []rune{delimValueSeparator, delimEndObject})
			}
		}

		return p.eof(TypeObject)
	case delimBeginArray:
		n.info = vArray
		// Skip space after '['
		for pos++; 0 <= pos && pos < len(s); pos++ {
			if c = s[pos]; bytemapIsSpace[c] == 0 {
				// goto used to return proper eof type without checking pos
				goto isEmptyArray
			}
		}
		return p.eof(TypeArray)
	isEmptyArray:
		if c == delimEndArray {
			n.values = n.values[:0]
			return pos + 1
		}
		if cap(n.values) == 0 {
			n.values = make([]*Node, 8)
		} else {
			n.values = n.values[:cap(n.values)]
		}

		v := p.node()
		i := 0
	readArray:
		pos = p.parseValue(c, s, pos, v)
		if p.err != nil {
			return pos
		}

		// Skip space after value
		for ; 0 <= pos && pos < len(s); pos++ {
			if c = s[pos]; bytemapIsSpace[c] == 0 {
				break
			}
		}

		switch c {
		case delimValueSeparator:
			// Skip space after ','
			for pos++; 0 <= pos && pos < len(s); pos++ {
				if c = s[pos]; bytemapIsSpace[c] == 0 {
					n.append(v, i)
					i++
					v = p.node()
					// goto used to return proper eof type without checking pos
					goto readArray
				}
			}
			return p.eof(TypeArray)
		case delimEndArray:
			n.append(v, i)
			if i++; 0 <= i && i <= cap(n.values) {
				n.values = n.values[:i]
			}
			return pos + 1
		default:
			return p.abort(pos, TypeArray, c, []rune{delimValueSeparator, delimEndArray})
		}
	case 'n':
		switch s = sliceAtN(s, pos, 4); s {
		case strNull:
			n.info, n.raw = vNull, strNull
			return pos + 4
		default:
			return p.abort(pos, TypeNull, s, strNull)
		}
	case 'f':
		switch s = sliceAtN(s, pos, 5); s {
		case strFalse:
			n.info, n.raw = vFalse, strFalse
			return pos + 5
		default:
			return p.abort(pos, TypeBoolean, s, strFalse)
		}
	case 't':
		switch s = sliceAtN(s, pos, 4); s {
		case strTrue:
			n.info, n.raw = vTrue, strTrue
			return pos + 4
		default:
			return p.abort(pos, TypeBoolean, s, strTrue)
		}
	case '-':
		if 0 <= pos && pos < len(s) {
			n.raw = s[pos:]
			n.info = vNumber | NumberSigned
			for i := 1; 0 < i && i < len(n.raw); i++ {
				if c = n.raw[i]; bytemapIsNumberEnd[c] == 1 {
					n.raw = n.raw[:i]
					return pos + i
				}
			}
			return pos + len(n.raw)
		}
		return p.eof(TypeNumber)
		// if n.raw, pos, n.info = scanNumberAt(c, s, pos); n.info == HasError {
		// 	return p.abort(pos, TypeNumber, n.raw, "valid number token")
		// }
		// return pos
	default:
		if bytemapIsDigit[c] == 1 {
			if 0 <= pos && pos < len(s) {
				n.raw = s[pos:]
				n.info = vNumber
				for i := 1; 0 < i && i < len(n.raw); i++ {
					if c = n.raw[i]; bytemapIsNumberEnd[c] == 1 {
						n.raw = n.raw[:i]
						return pos + i
					}
				}
				return pos + len(n.raw)
			}
			return p.eof(TypeNumber)
			// if n.raw, pos, n.info = scanNumberAt(c, s, pos); n.info == HasError {
			// 	return p.abort(pos, TypeNumber, n.raw, "valid number token")
			// }
			// return pos
		}
		if 0 <= pos && pos < len(s) {
			return p.abort(pos, TypeAnyValue, c, "any value")
		}
		return p.eof(TypeAnyValue)
	}

}

const (
	delimString         = '"'
	delimEscape         = '\\'
	delimBeginObject    = '{'
	delimEndObject      = '}'
	delimBeginArray     = '['
	delimEndArray       = ']'
	delimNameSeparator  = ':'
	delimValueSeparator = ','
)

func (p *Parser) abort(pos int, typ Type, got, want interface{}) int {
	p.err = abort(pos, typ, got, want)
	return pos
}
func (p *Parser) eof(typ Type) int {
	p.err = eof(typ)
	return -1
}

func sliceAtN(s string, pos, n int) string {
	if 0 <= pos && pos < len(s) {
		if s = s[pos:]; 0 <= n && n < len(s) {
			return s[:n]
		}
		return s
	}
	return ""
}

// func scanNumberAt(c byte, s string, pos int) (_ string, end int, inf Info) {
// 	if 0 <= pos && pos < len(s) {
// 		s = s[pos:]
// 	} else {
// 		return "", -1, HasError
// 	}
// 	inf = vNumberUint
// 	switch c {
// 	case '0':
// 		if len(s) > 1 && bytemapIsNumberEnd[s[1]] == 0 {
// 			end = 1
// 			c = s[1]
// 			goto decimal
// 		} else {
// 			return "0", pos + 1, vNumberUint
// 		}
// 	case '-':
// 		inf = vNumberInt
// 		fallthrough
// 	default:
// 		for end = 1; 0 < end && end < len(s); end++ {
// 			if c = s[end]; bytemapIsDigit[c] == 0 {
// 				if bytemapIsNumberEnd[c] == 1 {
// 					return s[:end], pos + end, inf
// 				}
// 				goto decimal
// 			}
// 		}
// 		goto done

// 	}
// decimal:
// 	if c == '.' {
// 		inf = vNumberFloat
// 		for end++; 0 < end && end < len(s); end++ {
// 			if c = s[end]; bytemapIsDigit[c] == 0 {
// 				if bytemapIsNumberEnd[c] == 1 {
// 					return s[:end], pos + end, inf
// 				}
// 				goto scientific
// 			}
// 		}
// 	}
// scientific:
// 	if c == 'e' || c == 'E' {
// 		inf = vNumberFloat
// 		if end++; 0 <= end && end < len(s) {
// 			if c = s[end]; c == '+' || c == '-' {
// 				end++
// 			}
// 		}
// 		for ; 0 <= end && end < len(s); end++ {
// 			if c = s[end]; bytemapIsDigit[c] == 0 {
// 				if bytemapIsNumberEnd[c] == 1 {
// 					return s[:end], pos + end, inf
// 				}
// 				end++
// 				goto done
// 			}
// 		}
// 	}
// done:
// 	if 0 <= end && end < len(s) {
// 		s = s[:end+1]
// 	}
// 	if bytemapIsDigit[c] == 0 {
// 		return s, pos + end, HasError
// 	}
// 	return s, pos + end, inf

// }

var pool = new(sync.Pool)

// Get returns a parser from a a pool.
// Put it back once you're done with Parser.Close()
func Get() *Parser {
	x := pool.Get()
	if x == nil {
		return &Parser{
			nodes: make([]Node, minNumNodes),
		}
	}
	return x.(*Parser)
}

// Close returns the parser to the pool.
func (p *Parser) Close() error {
	if p != nil {
		pool.Put(p)
	}
	return nil
}

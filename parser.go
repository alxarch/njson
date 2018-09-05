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
func (p *Parser) ParseUnsafe(data []byte) (n *Node, err error) {
	p.reset(false)
	s := b2s(data)
	var (
		c   byte
		pos int
	)
	for ; 0 <= pos && pos < len(s); pos++ {
		if c = s[pos]; bytemapIsSpace[c] == 0 {
			break
		}
	}
	n = p.node()
	n.next, n.value = nil, nil
	if pos = p.parseValue(c, s, pos, n); p.err != nil {
		return nil, p.err
	}
	for ; 0 <= pos && pos < len(s); pos++ {
		if c = s[pos]; bytemapIsSpace[c] == 0 {
			return nil, abort(pos, n.info.Type(), c, []rune{' ', '\t', '\n', '\r'})
		}
	}
	return n, nil
}

func (p *Parser) Parse(s string) (n *Node, err error) {
	p.reset(true)
	var (
		c   byte
		pos int
	)
	for ; 0 <= pos && pos < len(s); pos++ {
		if c = s[pos]; bytemapIsSpace[c] == 0 {
			break
		}
	}
	n = p.node()
	n.next, n.value = nil, nil
	if pos = p.parseValue(c, s, pos, n); p.err != nil {
		return nil, p.err
	}
	for ; 0 <= pos && pos < len(s); pos++ {
		if c = s[pos]; bytemapIsSpace[c] == 0 {
			return nil, abort(pos, n.info.Type(), c, []rune{' ', '\t', '\n', '\r'})
		}
	}
	return n, nil
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

	p.nodes = make([]Node, len(p.nodes)*3+minNumNodes)
	if len(p.nodes) > 0 {
		n = &p.nodes[0]
		// n.safe = p.safe
		p.n = 1
	}
	return
}

func (p *Parser) parseValue(c byte, s string, pos int, n *Node) int {
	switch c {
	case delimString:
		n.info, n.value, n.next, n.safe = vString, nil, nil, p.safe
		if pos++; 0 < pos && pos < len(s) {
			ss := s[pos:]
			end := strings.IndexByte(ss, delimString)
			if end--; 0 <= end && end < len(ss) {
				if ss[end] == delimEscape {
					end += 2
					for ; 0 <= end && end < len(ss); end++ {
						switch ss[end] {
						case delimString:
							n.raw = ss[:end]
							end++
							return end + pos
						case delimEscape:
							end++
						}
					}
				} else if end++; 0 <= end && end <= len(ss) {
					n.raw = ss[:end]
					end++
					return end + pos
				}
			} else if end == -1 {
				n.raw = ""
				return pos + 1
			}
		}
		p.err = abort(pos-1, TypeString, nil, delimString)
		return -1
	case delimBeginObject:
		for pos++; 0 <= pos && pos < len(s); pos++ {
			if bytemapIsSpace[s[pos]] == 0 {
				c = s[pos]
				break
			}
		}
		if c == delimEndObject {
			n.info, n.value, n.next = vObject, nil, nil
			return pos + 1
		}
		n.info, n.value, n.next = vObject, p.node(), nil
		n = n.value

		for pos++; 0 <= pos && pos < len(s); pos++ {
			if c != delimString {
				p.err = abort(pos, TypeKey, c, delimString)
				return pos
			}
			ss := s[pos:]
		readkey:
			for end := 0; 0 <= end && end < len(ss); end++ {
				switch ss[end] {
				case delimString:
					n.raw = ss[:end]
					pos += end + 1
					break readkey
				case delimEscape:
					end++
				}
			}
			for ; 0 <= pos && pos < len(s); pos++ {
				if c = s[pos]; bytemapIsSpace[c] == 0 {
					break
				}
			}
			if c != delimNameSeparator {
				p.err = abort(pos, TypeKey, c, delimNameSeparator)
				return pos
			}
			n.info, n.safe = vKey, p.safe
			for pos++; 0 <= pos && pos < len(s); pos++ {
				if c = s[pos]; bytemapIsSpace[c] == 0 {
					break
				}
			}
			n.value = p.node()
			if pos = p.parseValue(c, s, pos, n.value); p.err != nil {
				return pos
			}
			for ; 0 <= pos && pos < len(s); pos++ {
				if c = s[pos]; bytemapIsSpace[c] == 0 {
					break
				}
			}
			switch c {
			case delimValueSeparator:
				n.next = p.node()
				n = n.next
				for pos++; 0 <= pos && pos < len(s); pos++ {
					if c = s[pos]; bytemapIsSpace[c] == 0 {
						break
					}
				}
			case delimEndObject:
				n.next = nil
				return pos + 1
			default:
				p.err = abort(pos, TypeObject, c, []rune{delimValueSeparator, delimEndObject})
				return pos
			}
		}
		p.err = eof(TypeObject)
		return pos
	case delimBeginArray:
		for pos++; 0 <= pos && pos < len(s); pos++ {
			if bytemapIsSpace[s[pos]] == 0 {
				c = s[pos]
				break
			}
		}
		if c == delimEndArray {
			n.info, n.value, n.next = vArray, nil, nil
			return pos + 1
		}
		n.info, n.next, n.value = vArray, nil, p.node()
		n = n.value
		for {
			if pos = p.parseValue(c, s, pos, n); p.err != nil {
				return pos
			}
			for ; 0 <= pos && pos < len(s); pos++ {
				if bytemapIsSpace[s[pos]] == 0 {
					c = s[pos]
					break
				}
			}
			switch c {
			case delimValueSeparator:
				n.next = p.node()
				n = n.next
			case delimEndArray:
				n.next = nil
				return pos + 1
			default:
				p.err = abort(pos, TypeArray, c, []rune{delimValueSeparator, delimEndArray})
				return pos
			}

			for pos++; 0 <= pos && pos < len(s); pos++ {
				if bytemapIsSpace[s[pos]] == 0 {
					c = s[pos]
					break
				}
			}
		}
	case 'n':
		switch s = sliceAtN(s, pos, 4); s {
		case strNull:
			n.info, n.raw, n.value, n.next = vNull, strNull, nil, nil
			return pos + 4
		default:
			p.err = abort(pos, TypeNull, s, strNull)
			return pos
		}
	case 'f':
		switch s = sliceAtN(s, pos, 5); s {
		case strFalse:
			n.info, n.raw, n.value, n.next = vFalse, strFalse, nil, nil
			return pos + 5
		default:
			p.err = abort(pos, TypeBoolean, s, strFalse)
			return pos
		}
	case 't':
		switch s = sliceAtN(s, pos, 4); s {
		case strTrue:
			n.info, n.raw, n.value, n.next = vTrue, strTrue, nil, nil
			return pos + 4
		default:
			return p.abort(pos, TypeBoolean, s, strTrue)
		}
	case '-':
		if n.raw, pos, n.info = scanNumberAt(c, s, pos); n.info == HasError {
			return p.abort(pos, TypeNumber, n.raw, "valid number token")
		}
		return pos
	default:
		if bytemapIsDigit[c] == 1 {
			if n.raw, pos, n.info = scanNumberAt(c, s, pos); n.info == HasError {
				return p.abort(pos, TypeNumber, n.raw, "valid number token")
			}
			return pos
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

func scanNumberAt(c byte, s string, pos int) (_ string, end int, inf Info) {
	if 0 <= pos && pos < len(s) {
		s = s[pos:]
	} else {
		return "", -1, HasError
	}
	inf = vNumberUint
	switch c {
	case '0':
		if len(s) > 1 && bytemapIsNumberEnd[s[1]] == 0 {
			end = 1
			c = s[1]
			goto decimal
		} else {
			return "0", pos + 1, vNumberUint
		}
	case '-':
		inf = vNumberInt
		fallthrough
	default:
		for end = 1; 0 < end && end < len(s); end++ {
			if c = s[end]; bytemapIsDigit[c] == 0 {
				if bytemapIsNumberEnd[c] == 1 {
					return s[:end], pos + end, inf
				}
				goto decimal
			}
		}
		goto done

	}
decimal:
	if c == '.' {
		inf = vNumberFloat
		for end++; 0 < end && end < len(s); end++ {
			if c = s[end]; bytemapIsDigit[c] == 0 {
				if bytemapIsNumberEnd[c] == 1 {
					return s[:end], pos + end, inf
				}
				goto scientific
			}
		}
	}
scientific:
	if c == 'e' || c == 'E' {
		inf = vNumberFloat
		if end++; 0 <= end && end < len(s) {
			if c = s[end]; c == '+' || c == '-' {
				end++
			}
		}
		for ; 0 <= end && end < len(s); end++ {
			if c = s[end]; bytemapIsDigit[c] == 0 {
				if bytemapIsNumberEnd[c] == 1 {
					return s[:end], pos + end, inf
				}
				end++
				goto done
			}
		}
	}
done:
	if 0 <= end && end < len(s) {
		s = s[:end+1]
	}
	if bytemapIsDigit[c] == 0 {
		return s, pos + end, HasError
	}
	return s, pos + end, inf

}

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

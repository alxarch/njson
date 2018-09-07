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
	pos := p.parseValue(s, 0, n)
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
	pos := p.parseValue(s, 0, n)
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

	p.nodes = make([]Node, (len(p.nodes)*2)+minNumNodes)
	if len(p.nodes) > 0 {
		n = &p.nodes[0]
		// n.safe = p.safe
		p.n = 1
	}
	return
}

func (p *Parser) parseValue(s string, pos int, n *Node) int {
	var c byte
	for ; 0 <= pos && pos < len(s); pos++ {
		if c = s[pos]; bytemapIsSpace[c] == 1 {
			continue
		}
		if c == delimString {
			goto readString
		}
		if c == delimBeginObject {
			return p.parseObject(s, pos+1, n)
		}
		if c == delimBeginArray {
			return p.parseArray(s, pos+1, n)
		}
		if bytemapIsDigit[c] == 1 {
			n.info = vNumber
			goto readNumber
		}
		if c == 'n' {
			goto readNull
		}
		if c == '-' {
			n.info = vNumber | NumberSigned
			goto readNumber
		}
		if c == 'f' {
			goto readFalse
		}
		if c == 't' {
			goto readTrue
		}
		return p.abort(pos, TypeAnyValue, c, "any value")
	}
	return p.eof(TypeAnyValue)
readNumber:
	if 0 <= pos && pos < len(s) {
		s = s[pos:]
		for i := 0; 0 <= i && i < len(s); i++ {
			if c = s[i]; bytemapIsNumberEnd[c] == 1 {
				n.raw = s[:i]
				return pos + i
			}
		}
		n.raw = s
		return pos + len(n.raw)
	}
	return p.eof(TypeNumber)
readString:
	n.info = vString
	n.safe = p.safe
	if pos++; 0 < pos && pos < len(s) {
		s = s[pos:]
		pos++
		i := strings.IndexByte(s, delimString) - 1
		if 0 <= i && i < len(s) {
			if s[i] == delimEscape {
				for i += 2; 0 <= i && i < len(s); i++ {
					switch s[i] {
					case delimString:
						n.raw = s[:i]
						return pos + i
					case delimEscape:
						i++
					}
				}
			} else if i++; 0 <= i && i <= len(s) {
				n.raw = s[:i]
				return pos + i
			}
		}
		if i == -1 {
			n.raw = ""
			return pos
		}
	}
	return p.eof(TypeString)
readTrue:
	if 0 <= pos && pos < len(s) {
		if s = s[pos:]; len(s) >= 4 {
			if s = s[:4]; s == strTrue {
				n.info = vTrue
				n.raw = strTrue
				return pos + 4
			}
			return p.abort(pos, TypeBoolean, s, strTrue)
		}
	}
	return p.eof(TypeBoolean)
readFalse:
	if 0 <= pos && pos < len(s) {
		if s = s[pos:]; len(s) >= 5 {
			if s = s[:5]; s == strFalse {
				n.info = vFalse
				n.raw = strFalse
				return pos + 5
			}
			return p.abort(pos, TypeBoolean, s, strFalse)
		}
	}
	return p.eof(TypeBoolean)
readNull:
	if 0 <= pos && pos < len(s) {
		if s = s[pos:]; len(s) >= 4 {
			if s = s[:4]; s == strNull {
				n.info = vNull
				n.raw = strNull
				return pos + 4
			}
			return p.abort(pos, TypeNull, s, strNull)
		}
	}
	return p.eof(TypeNull)

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

func (p *Parser) parseArray(s string, pos int, n *Node) int {
	var (
		c         byte
		v         *Node
		numValues = 0
	)
	n.info = vArray
	// Skip space after '['
	for ; 0 <= pos && pos < len(s); pos++ {
		if c = s[pos]; bytemapIsSpace[c] == 0 {
			if c == delimEndArray {
				n.values = n.values[:0]
				return pos + 1
			}
			n.values = n.values[:cap(n.values)]
			v = p.node()
			goto readValue
		}
	}
	return p.eof(TypeArray)
readValue:
	pos = p.parseValue(s, pos, v)
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
		n.append(v, numValues)
		numValues++
		v = p.node()
		pos++
		goto readValue
	case delimEndArray:
		n.append(v, numValues)
		if numValues++; 0 <= numValues && numValues <= cap(n.values) {
			n.values = n.values[:numValues]
		}
		return pos + 1
	default:
		return p.abort(pos, TypeArray, c, []rune{delimValueSeparator, delimEndArray})
	}
}
func (p *Parser) parseObject(s string, pos int, n *Node) int {
	var (
		c       byte
		numKeys = 0
		v       *Node
	)
	for ; 0 <= pos && pos < len(s); pos++ {
		if c = s[pos]; bytemapIsSpace[c] == 0 {
			if c == delimEndObject {
				n.info = vObject
				n.values = n.values[:0]
				return pos + 1
			}
			if c == delimString {
				n.info = vObject
				n.values = n.values[:cap(n.values)]
				v = p.node()
				goto readKey
			}
			return p.abort(pos, TypeObject, c, []rune{delimEndObject, delimString})
		}
	}
	return p.eof(TypeObject)
readKey:
	if pos++; 0 <= pos && pos < len(s) {
		v.key = s[pos:]
		for i := 0; 0 <= i && i < len(v.key); i++ {
			switch v.key[i] {
			case delimString:
				v.key = v.key[:i]
				for pos += i + 1; 0 <= pos && pos < len(s); pos++ {
					if c = s[pos]; c == delimNameSeparator {
						goto readValue
					}
					if bytemapIsSpace[c] == 0 {
						break
					}
				}
				return p.abort(pos, TypeObject, c, delimNameSeparator)
			case delimEscape:
				i++
			}
		}
	}
	return p.eof(TypeObject)

readValue:
	// We're at ':' after key
	pos = p.parseValue(s, pos+1, v)
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
			if c = s[pos]; c == delimString {
				// Append value
				n.append(v, numKeys)
				numKeys++
				v = p.node()
				goto readKey
			}
			if bytemapIsSpace[c] == 1 {
				continue
			}
			return p.abort(pos, TypeObject, c, delimString)
		}
		return p.eof(TypeObject)
	case delimEndObject:
		n.append(v, numKeys)
		if numKeys++; 0 <= numKeys && numKeys <= cap(n.values) {
			n.values = n.values[:numKeys]
		}
		return pos + 1
	default:
		return p.abort(pos, TypeObject, c, []rune{delimValueSeparator, delimEndObject})
	}
}

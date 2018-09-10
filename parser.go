package njson

import (
	"math"
	"strings"
	"sync"
)

// Parser is a JSON parser.
type Parser struct {
	nodes  []Node
	unsafe Info
	n      int
	err    error
}

// ParseUnsafe parses JSON from a slice of bytes without copying it to a string.
// The contents of the slice should not be modified while using the result node.
func (p *Parser) ParseUnsafe(data []byte) (*Node, []byte, error) {
	p.reset()
	p.unsafe = Unsafe
	s := b2s(data)
	n := p.node()
	n.values = n.values[:0]
	pos := p.parseValue(s, 0, n)
	if p.err != nil {
		return nil, data, p.err
	}
	if pos < uint(len(data)) {
		return n, data[pos:], nil
	}
	return n, nil, nil
}

// Parse parses JSON from a string.
func (p *Parser) Parse(s string) (*Node, string, error) {
	p.reset()
	n := p.node()
	n.values = n.values[:0]
	pos := p.parseValue(s, 0, n)
	if p.err != nil {
		return nil, s, p.err
	}
	if pos < uint(len(s)) {
		return n, s[pos:], nil
	}
	return n, "", nil
}

const minNumNodes = 64

func (p *Parser) reset() {
	p.err = nil
	p.unsafe = 0
	p.n = 0
}

func (p *Parser) node() (n *Node) {
	if 0 <= p.n && p.n < len(p.nodes) {
		n = &p.nodes[p.n]
		p.n++
		return
	}

	p.nodes = make([]Node, (len(p.nodes)*2)+minNumNodes)
	if len(p.nodes) > 0 {
		n = &p.nodes[0]
		p.n = 1
	}
	return
}

func (p *Parser) parseValue(s string, pos uint, n *Node) uint {
	var c byte
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
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
				n.info = vNumber | p.unsafe
				goto readNumber
			}
			if c == 'n' {
				goto readNull
			}
			if c == '-' {
				const signedNumber = vNumber | NumberSigned
				n.info = signedNumber | p.unsafe
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
	}
	return p.eof(TypeAnyValue)
readNumber:
	if pos < uint(len(s)) {
		s = s[pos:]
		for i := uint(0); i < uint(len(s)); i++ {
			c = s[i]
			if bytemapIsNumberEnd[c] == 0 {
				continue
			}
			n.raw = s[:i]
			return pos + i
		}
		n.raw = s
		return pos + uint(len(s))
	}
	return p.eof(TypeNumber)
readString:
	n.info = vString | p.unsafe
	if pos++; pos < uint(len(s)) {
		// Slice after the opening quote
		s = s[pos:]
		pos++ // Early jump to the next character after the closing quote
		// Immediately decrement to check if previous byte is '\'
		i := strings.IndexByte(s, delimString) - 1
		if 0 <= i && i < len(s) { // Avoid bounds check and -1 result from IndexByte
			if s[i] == delimEscape {
				// Advance past '\' and '"' and scan the remaining string
				for i += 2; 0 <= i && i < len(s); i++ { // Avoid bounds check
					switch s[i] {
					case delimString:
						// Slice until the closing quote
						n.raw = s[:i]
						return pos + uint(i)
					case delimEscape:
						// Jump over the next character
						i++
					}
				}
			} else if i++; 0 <= i && i <= len(s) { // Avoid bounds check
				// Slice until the closing quote
				n.raw = s[:i]
				return pos + uint(i)
			}
		} else if i == -1 { // Empty string case
			n.raw = ""
			return pos
		}
	}
	return p.eof(TypeString)
readTrue:
	if pos < uint(len(s)) {
		if s = s[pos:]; len(s) >= 4 {
			if s = s[:4]; s == strTrue {
				n.info = vTrue
				// n.raw = strTrue
				return pos + 4
			}
			return p.abort(pos, TypeBoolean, s, strTrue)
		}
	}
	return p.eof(TypeBoolean)
readFalse:
	if pos < uint(len(s)) {
		if s = s[pos:]; len(s) >= 5 {
			if s = s[:5]; s == strFalse {
				n.info = vFalse
				// n.raw = strFalse
				return pos + 5
			}
			return p.abort(pos, TypeBoolean, s, strFalse)
		}
	}
	return p.eof(TypeBoolean)
readNull:
	if pos < uint(len(s)) {
		if s = s[pos:]; len(s) >= 4 {
			if s = s[:4]; s == strNull {
				n.info = vNull
				// n.raw = strNull
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

func (p *Parser) abort(pos uint, typ Type, got, want interface{}) uint {
	p.err = abort(int(pos), typ, got, want)
	return pos
}
func (p *Parser) eof(typ Type) uint {
	p.err = eof(typ)
	return math.MaxUint64
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

func (p *Parser) parseArray(s string, pos uint, n *Node) uint {
	var (
		c         byte
		v         *Node
		numValues = 0
	)
	n.info = vArray
	// Skip space after '['
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
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
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
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
func (p *Parser) parseObject(s string, pos uint, n *Node) uint {
	var (
		c       byte
		numKeys = 0
		v       *Node
	)
	// Skip space after opening '{'
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
			// Check for empty object
			if c == delimEndObject {
				n.info = vObject
				n.values = n.values[:0]
				return pos + 1
			}
			// Check for start of key
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

	// Current pos is at the opening quote of a key.
readKey:
	if pos++; pos < uint(len(s)) {
		// Slice after the opening quote
		v.key = s[pos:]
		pos++ // Early jump after the closing quote
		// Keys are usually small.
		// IndexByte seems to have a performance benefit only if the
		// byte we're looking for is more than 16 bytes away.
		// Since most keys are less than 16 bytes using a simple loop
		// actually improves throughput.
		for i := uint(0); i < uint(len(v.key)); i++ {
			switch v.key[i] {
			case delimString:
				// Slice until closing quote
				v.key = v.key[:i]
				// Skip space after closing quote
				for pos += i; pos < uint(len(s)); pos++ {
					c = s[pos]
					if c == delimNameSeparator {
						goto readValue
					}
					if bytemapIsSpace[c] == 0 {
						return p.abort(pos, TypeObject, c, delimNameSeparator)
					}
					// Space
				}
				return p.eof(TypeObject)
			case delimEscape:
				i++
			}
		}
		// end of input reached
	}
	return p.eof(TypeObject)

readValue:
	// We're at ':' after key
	pos = p.parseValue(s, pos+1, v)
	if p.err != nil {
		return pos
	}
	// Skip space after value
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
			break
		}
	}
	switch c {
	case delimValueSeparator:
		// Skip space after ','
		for pos++; pos < uint(len(s)); pos++ {
			c = s[pos]
			if c == delimString {
				// Append value
				n.append(v, numKeys)
				numKeys++
				v = p.node()
				goto readKey
			}
			if bytemapIsSpace[c] == 0 {
				return p.abort(pos, TypeObject, c, delimString)
			}
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

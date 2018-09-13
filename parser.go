package njson

import (
	"math"
	"strings"
	"sync"
)

// Parser is a JSON parser.
type Parser struct {
	nodes  []Node
	n      int
	times  int
	unsafe Info
	err    error
}

func (p *Parser) snapshot() []Node {
	nodes := p.nodes
	if 0 <= p.n && p.n < len(nodes) {
		nodes = nodes[:p.n]
	}
	return nodes
}

// ParseUnsafe parses JSON from a slice of bytes without copying it to a string.
// The contents of the slice should not be modified while using the result node.
func (p *Parser) ParseUnsafe(data []byte) (*Node, []byte, error) {
	p.Reset()
	p.unsafe = Unsafe
	s := b2s(data)
	n := p.node()
	if pos := p.parseValue(s, 0, n); pos <= uint(len(data)) {
		data = data[pos:]
	} else {
		data = nil
	}
	if p.err != nil {
		return nil, data, p.err
	}
	return n, data, nil
}

func (p *Parser) Reset() {
	p.n = 0
	p.err = nil
	p.unsafe = 0

}

// Parse parses JSON from a string.
func (p *Parser) Parse(s string) (*Node, string, error) {
	p.Reset()
	n := p.node()
	if pos := p.parseValue(s, 0, n); pos < uint(len(s)) {
		s = s[pos:]
	} else {
		s = ""
	}
	if p.err != nil {
		return nil, s, p.err
	}
	return n, "", nil
}

const minNumNodes = 64

func (p *Parser) reset() {
	p.err = nil
	p.unsafe = 0
	p.n = 0
}

func (p *Parser) node() *Node {
	if 0 <= p.n && p.n < len(p.nodes) {
		n := &p.nodes[p.n]
		p.n++
		return n
	}

	nodes := make([]Node, (2*len(p.nodes))+minNumNodes)
	p.nodes = nodes
	if len(p.nodes) > 0 {
		n := &p.nodes[0]
		p.n = 1
		return n
	}
	return nil
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
				goto readNumber
			}
			if c == 'n' {
				goto readNull
			}
			if c == '-' {
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
	n.info = vNumber | p.unsafe
	if pos < uint(len(s)) {
		s = s[pos:]
		for i := uint(0); i < uint(len(s)); i++ {
			c = s[i]
			if bytemapIsNumberEnd[c] == 0 {
				continue
			}
			n.raw = s[:i]
			pos += i
			goto done
		}
		n.raw = s
		pos += uint(len(s))
		goto done
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
						pos += uint(i)
						goto done
					case delimEscape:
						// Jump over the next character
						i++
					}
				}
			} else if i++; 0 <= i && i <= len(s) { // Avoid bounds check
				// Slice until the closing quote
				n.raw = s[:i]
				pos += uint(i)
				goto done
			}
		} else if i == -1 { // Empty string case
			n.raw = ""
			goto done
		}
	}
	return p.eof(TypeString)
readTrue:
	if pos < uint(len(s)) {
		if s = s[pos:]; len(s) >= 4 {
			if s = s[:4]; s == strTrue {
				n.set(vTrue, strTrue)
				pos += 4
				goto done
			}
			return p.abort(pos, TypeBoolean, s, strTrue)
		}
	}
	return p.eof(TypeBoolean)
readFalse:
	if pos < uint(len(s)) {
		if s = s[pos:]; len(s) >= 5 {
			if s = s[:5]; s == strFalse {
				n.set(vFalse, strFalse)
				pos += 5
				goto done
			}
			return p.abort(pos, TypeBoolean, s, strFalse)
		}
	}
	return p.eof(TypeBoolean)
readNull:
	if pos < uint(len(s)) {
		if s = s[pos:]; len(s) >= 4 {
			if s = s[:4]; s == strNull {
				n.set(vNull, strNull)
				pos += 4
				goto done
			}
			return p.abort(pos, TypeNull, s, strNull)
		}
	}
	return p.eof(TypeNull)
done:
	for i := range n.values {
		n.values[i] = KV{}
	}
	n.values = n.values[:0]
	return pos
}

func (n *Node) set(info Info, raw string) {
	n.info = info
	n.raw = raw
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
		c      byte
		v      *Node
		values []KV
	)
	n.info = vArray
	n.raw = ""
	// Skip space after '['
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
			if c == delimEndArray {
				for i := range n.values {
					n.values[i] = KV{}
				}
				n.values = n.values[:0]
				return pos + 1
			}
			values = n.values
			n.values = n.values[:0]
			goto readValue
		}
	}
	return p.eof(TypeArray)
readValue:
	v = p.node()
	pos = p.parseValue(s, pos, v)
	if p.err != nil {
		return pos
	}
	n.values = append(n.values, KV{"", v})

	// Skip space after value
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		switch c {
		case delimValueSeparator:
			// n.append(v, numValues)
			// numValues++
			// v = p.node()
			pos++
			goto readValue
		case delimEndArray:
			if i := len(n.values); 0 <= i && i < len(values) {
				values = values[:i]
				for i := range values {
					values[i] = KV{}
				}
			}
			return pos + 1
		default:
			if bytemapIsSpace[c] == 0 {
				return p.abort(pos, TypeArray, c, []rune{delimValueSeparator, delimEndArray})
			}
		}
	}
	return p.eof(TypeArray)

}
func (p *Parser) parseObject(s string, pos uint, n *Node) uint {
	var (
		c      byte
		v      *Node
		key    string
		values []KV
	)
	// Skip space after opening '{'
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
			// Check for empty object
			if c == delimEndObject {
				n.info = vObject
				n.raw = ""
				for i := range n.values {
					n.values[i] = KV{}
				}
				n.values = n.values[:0]
				return pos + 1
			}
			// Check for start of key
			if c == delimString {
				n.info = vObject
				values = n.values
				n.values = n.values[:0]
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
		key = s[pos:]
		pos++ // Early jump after the closing quote
		// Keys are usually small.
		// IndexByte seems to have a performance benefit only if the
		// byte we're looking for is more than 16 bytes away.
		// Since most keys are less than 16 bytes using a simple loop
		// actually improves throughput.
		for i := uint(0); i < uint(len(key)); i++ {
			switch key[i] {
			case delimString:
				// Slice until closing quote
				key = key[:i]
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
	v = p.node()
	// We're at ':' after key
	pos = p.parseValue(s, pos+1, v)
	if p.err != nil {
		return pos
	}
	n.values = append(n.values, KV{key, v})
	// Skip space after value
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		switch c {
		case delimValueSeparator:
			// Skip space after ','
			for pos++; pos < uint(len(s)); pos++ {
				c = s[pos]
				if c == delimString {
					// Append value
					goto readKey
				}
				if bytemapIsSpace[c] == 0 {
					return p.abort(pos, TypeObject, c, delimString)
				}
			}
			return p.eof(TypeObject)
		case delimEndObject:
			if i := len(n.values); 0 <= i && i < len(values) {
				values = values[i:]
				for i := range values {
					values[i] = KV{}
				}
			}
			return pos + 1
		default:
			if bytemapIsSpace[c] == 0 {
				return p.abort(pos, TypeObject, c, []rune{delimValueSeparator, delimEndObject})
			}
		}
	}
	return p.eof(TypeObject)
}

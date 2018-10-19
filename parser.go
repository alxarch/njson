package njson

import (
	"strings"
)

type parser struct {
	nodes  []node
	n      uint
	unsafe info
	err    error
}

// Parse parses a JSON string and returns the root node as a Partial.
func (d *Document) Parse(s string) (Node, string, error) {
	p := d.parser()
	id := p.n
	// n := p.node()
	if pos := p.parseValue(s, 0); pos <= uint(len(s)) {
		s = s[pos:]
	} else {
		s = ""
	}
	if p.err != nil {
		return Node{}, s, p.err
	}
	if nodes := p.update(d); len(nodes) > 0 {
		resetNodes(nodes)
	}
	d.get(id).info |= infRoot
	return Node{id, d.rev, d}, s, nil
}

// ParseUnsafe parses a JSON buffer without copying it to a string and returns the root node as a Partial.
// Make sure to call Document.Reset() or Document.Close() to avoid memory leaks.
func (d *Document) ParseUnsafe(b []byte) (Node, []byte, error) {
	p := d.parser()
	p.unsafe = infUnsafe
	id := p.n
	// n := p.node()
	if pos := p.parseValue(b2s(b), 0); pos <= uint(len(b)) {
		b = b[pos:]
	} else {
		b = nil
	}
	if p.err != nil {
		return Node{}, b, p.err
	}
	// Free references to previous JSON source
	if nodes := p.update(d); len(nodes) > 0 {
		resetNodes(nodes)
	}
	d.get(id).info |= infRoot
	return Node{id, d.rev, d}, b, nil
}

func (d *Document) parser() parser {
	return parser{
		nodes: d.nodes[:cap(d.nodes)],
		n:     uint(len(d.nodes)),
	}
}

func (p *parser) abort(pos uint, typ Type, got, want interface{}) uint {
	p.err = abort(int(pos), typ, got, want)
	return pos
}

// node returns a node pointer. The pointer is valid until the next call to node()
func (p *parser) node() *node {
	if p.n < uint(len(p.nodes)) {
		n := &p.nodes[p.n]
		p.n++
		return n
	}
	nodes := make([]node, 2*len(p.nodes)+1)
	copy(nodes, p.nodes)
	p.nodes = nodes
	if p.n < uint(len(p.nodes)) {
		n := &p.nodes[p.n]
		p.n++
		return n
	}
	return nil
}
func appendV(values []V, key string, id, i uint) []V {
	if i < uint(len(values)) {
		values[i] = V{id, key}
		return values
	}
	tmp := make([]V, 2*len(values)+1)
	copy(tmp, values)
	if i < uint(len(tmp)) {
		tmp[i] = V{id, key}
	}
	return tmp
}

func (p *parser) eof(typ Type) (pos uint) {
	p.err = eof(typ)
	return maxUint
}

func (p *parser) parseValue(s string, pos uint) uint {
	var (
		c    byte
		info = p.unsafe
	)

	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
			if c == delimString {
				goto readString
			}
			if c == delimBeginObject {
				return p.parseObject(s, pos+1)
			}
			if c == delimBeginArray {
				return p.parseArray(s, pos+1)
			}
			if bytemapIsDigit[c] == 1 {
				s = s[pos:]
				goto readNumber
			}
			if c == 'n' {
				s = s[pos:]
				goto readNull
			}
			if c == '-' {
				s = s[pos:]
				goto readNumber
			}
			if c == 'f' {
				s = s[pos:]
				goto readFalse
			}
			if c == 't' {
				s = s[pos:]
				goto readTrue
			}
			return p.abort(pos, TypeAnyValue, c, "any value")
		}
	}
	return p.eof(TypeAnyValue)
readNumber:
	info |= vNumber
	for i := uint(0); i < uint(len(s)); i++ {
		if bytemapIsNumberEnd[s[i]] == 0 {
			continue
		}
		s = s[:i]
		pos += i
		goto done
	}
	pos += uint(len(s))
	goto done
readString:
	info |= vString
	if pos++; pos < uint(len(s)) {
		// Slice after the opening quote
		s = s[pos:]
		pos++ // Early jump to the next character after the closing quote
		// Immediately decrement to check if previous byte is '\'
		i := strings.IndexByte(s, delimString) - 1
		if 0 <= i && i < len(s) {
			if s[i] == delimEscape {
				// Advance past '\' and '"' and scan the remaining string
				for i += 2; 0 <= i && i < len(s); i++ { // Avoid bounds check
					switch s[i] {
					case delimString:
						// Slice until the closing quote
						s = s[:i]
						pos += uint(i)
						goto done
					case delimEscape:
						// Jump over the next character
						i++
					}
				}
			} else if i++; 0 <= i && i <= len(s) { // Avoid bounds check
				// Slice until the closing quote
				s = s[:i]
				pos += uint(i)
				goto done
			}
		} else if i == -1 {
			s = ""
			goto done
		}
	}
	return p.eof(TypeString)
readTrue:
	info |= vBoolean
	if len(s) >= 4 {
		if s = s[:4]; s == strTrue {
			pos += 4
			goto done
		}
		return p.abort(pos, TypeBoolean, s, strTrue)
	}
	return p.eof(TypeBoolean)
readFalse:
	info |= vBoolean
	if len(s) >= 5 {
		if s = s[:5]; s == strFalse {
			pos += 5
			goto done
		}
		return p.abort(pos, TypeBoolean, s, strFalse)
	}
	return p.eof(TypeBoolean)
readNull:
	info |= vNull
	if len(s) >= 4 {
		if s = s[:4]; s == strNull {
			pos += 4
			goto done
		}
		return p.abort(pos, TypeNull, s, strNull)
	}
	return p.eof(TypeNull)
done:
	n := p.node()
	n.set(info, s)
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]
	return pos
}

func (p *parser) parseArray(s string, pos uint) uint {
	var (
		id = p.n
		n  = p.node()
		c  byte
		// v      *N
		values []V
		numV   uint
	)
	n.set(vArray, "")
	// Skip space after '['
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
			if c == delimEndArray {
				for i := range n.values {
					n.values[i] = V{}
				}
				n.values = n.values[:0]
				return pos + 1
			}

			values = n.values[:cap(n.values)]
			goto readValue
		}
	}
	return p.eof(TypeArray)
readValue:
	values = appendV(values, "", p.n, numV)
	numV++
	// pos = p.parseValue(s, pos, p.node())
	pos = p.parseValue(s, pos)
	if p.err != nil {
		return pos
	}

	// Skip space after value
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		switch c {
		case delimValueSeparator:
			pos++
			goto readValue
		case delimEndArray:
			if numV < uint(len(n.values)) {
				values = n.values[:numV]
				n.values = n.values[numV:]
				for i := range n.values {
					n.values[i] = V{}
				}
			} else if numV < uint(len(values)) {
				values = values[:numV]
			}
			if id < uint(len(p.nodes)) {
				p.nodes[id].values = values
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

func (p *parser) parseObject(s string, pos uint) uint {
	var (
		id     = p.n
		n      = p.node()
		c      byte
		key    string
		values []V
		numV   uint
		i      uint
	)
	n.set(vObject, "")
	// Skip space after opening '{'
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		switch c {
		case delimEndObject:
			// Zero out the values slice to release key strings
			for i := range n.values {
				n.values[i] = V{}
			}
			n.values = n.values[:0]
			return pos + 1
		case delimString:
			values = n.values[:cap(n.values)]
			goto readKey
		default:
			if bytemapIsSpace[c] == 0 {
				return p.abort(pos, TypeObject, c, []rune{delimEndObject, delimString})
			}
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
		for i = 0; i < uint(len(key)); i++ {
			switch key[i] {
			case delimString:
				// key = key[:i]
				// Slice until closing quote
				values = appendV(values, key[:i], p.n, numV)
				numV++
				pos += i
				// key = key[:i]
				// Skip space after closing quote
				for ; pos < uint(len(s)); pos++ {
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
	pos = p.parseValue(s, pos+1)
	if p.err != nil {
		return pos
	}
	// Skip space after value
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		switch c {
		case delimValueSeparator:
			// Skip space after ','
			for pos++; pos < uint(len(s)); pos++ {
				c = s[pos]
				if c == delimString {
					goto readKey
				}
				if bytemapIsSpace[c] == 0 {
					return p.abort(pos, TypeObject, c, delimString)
				}
			}
			return p.eof(TypeObject)
		case delimEndObject:
			if numV < uint(len(n.values)) {
				values = n.values[:numV]
				n.values = n.values[numV:]
				// Zero out unused values to release key strings
				for i := range n.values {
					n.values[i] = V{}
				}
			} else if numV < uint(len(values)) {
				values = values[:numV]
				// No need to zero out n.values because n will have no references after return
			}
			// Use id because n pointer might be invalid after a node() call
			if id < uint(len(p.nodes)) {
				n = &p.nodes[id]
				n.values = values
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

// Update document nodes and returned unused nodes
func (p *parser) update(d *Document) []node {
	if p.n < uint(len(d.nodes)) {
		nodes := d.nodes[p.n:]
		d.nodes = d.nodes[:p.n]
		return nodes
	} else if p.n <= uint(cap(p.nodes)) {
		d.nodes = p.nodes[:p.n]
	}
	return nil

}

// Garbage collect unused nodes' references to JSON source string
func resetNodes(nodes []node) {
	for i := range nodes {
		n := &nodes[i]
		n.raw = ""
		for i := range n.values {
			n.values[i] = V{}
		}
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

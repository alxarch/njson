package njson

import "strings"

type parser struct {
	nodes  []N
	n      uint
	unsafe Info
	err    error
}

// Parse parses a JSON string and returns the root node as a Partial.
func (d *Document) Parse(s string) (Partial, string, error) {
	p := d.parser()
	id := p.n
	n := p.node()
	if pos := p.parseValue(s, 0, n); pos <= uint(len(s)) {
		s = s[pos:]
	} else {
		s = ""
	}
	if p.err != nil {
		return Partial{}, s, p.err
	}
	if nodes := p.update(d); len(nodes) > 0 {
		resetNodes(nodes)
	}
	return Partial{id, d.rev, d}, s, nil
}

// ParseUnsafe parses a JSON buffer without copying it to a string and returns the root node as a Partial.
// Make sure to call Document.Reset() or Document.Close() to avoid memory leaks.
func (d *Document) ParseUnsafe(b []byte) (Partial, []byte, error) {
	p := d.parser()
	p.unsafe = Unsafe
	id := p.n
	n := p.node()
	if pos := p.parseValue(b2s(b), 0, n); pos <= uint(len(b)) {
		b = b[pos:]
	} else {
		b = nil
	}
	if p.err != nil {
		return Partial{}, b, p.err
	}
	// Free references to previous JSON source
	if nodes := p.update(d); len(nodes) > 0 {
		resetNodes(nodes)
	}
	return Partial{id, d.rev, d}, b, nil
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
func (p *parser) value(n *N, key string) *N {
start:
	if p.n < uint(len(p.nodes)) {
		v := &p.nodes[p.n]
		p.n++
		return v
	}
	nodes := make([]N, 2*len(p.nodes)+minNumNodes)
	copy(nodes, p.nodes)
	p.nodes = nodes
	goto start

}

func (p *parser) node() *N {
start:
	if p.n < uint(len(p.nodes)) {
		n := &p.nodes[p.n]
		p.n++
		return n
	}
	nodes := make([]N, 2*len(p.nodes)+minNumNodes)
	copy(nodes, p.nodes)
	p.nodes = nodes
	goto start
}
func (p *parser) eof(typ Type) (pos uint) {
	p.err = eof(typ)
	return maxUint
}

func (p *parser) parseValue(s string, pos uint, n *N) uint {
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
	if pos < uint(len(s)) {
		s = s[pos:]
		for i := uint(0); i < uint(len(s)); i++ {
			c = s[i]
			if bytemapIsNumberEnd[c] == 0 {
				continue
			}
			n.set(vNumber, s[:i])
			pos += i
			goto done
		}
		n.set(vNumber, s)
		pos += uint(len(s))
		goto done
	}
	return p.eof(TypeNumber)
readString:
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
						n.set(vString|p.unsafe, s[:i])
						pos += uint(i)
						goto done
					case delimEscape:
						// Jump over the next character
						i++
					}
				}
			} else if i++; 0 <= i && i <= len(s) { // Avoid bounds check
				// Slice until the closing quote
				n.set(vString|p.unsafe, s[:i])
				pos += uint(i)
				goto done
			}
		} else if i == -1 { // Empty string case
			n.set(vString|p.unsafe, "")
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
		n.values[i] = V{}
	}
	n.values = n.values[:0]
	return pos
}
func (p *parser) parseArray(s string, pos uint, n *N) uint {
	var (
		c      byte
		v      *N
		values []V
	)
	// Skip space after '['
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
			if c == delimEndArray {
				n.set(vArray, "")
				for i := range n.values {
					n.values[i] = V{}
				}
				n.values = n.values[:0]
				return pos + 1
			}

			n.info = vArray
			n.raw = ""
			values = n.values
			n.values = n.values[:0]
			goto readValue
		}
	}
	return p.eof(TypeArray)
readValue:
	v = p.value(n, "")
	pos = p.parseValue(s, pos, v)
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
			if i := len(n.values); 0 <= i && i < len(values) {
				values = values[i:]
				for i := range values {
					values[i] = V{}
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

func (p *parser) parseObject(s string, pos uint, n *N) uint {
	var (
		c      byte
		key    string
		v      *N
		values []V
	)
	n.set(vObject, "")
	// Skip space after opening '{'
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		switch c {
		case delimEndObject:
			// Zero out the values slice to
			for i := range n.values {
				n.values[i] = V{}
			}
			n.values = n.values[:0]
			return pos + 1
		case delimString:
			// Store the initial node values to zero out any unused values at the end.
			values = n.values
			n.values = n.values[:0]
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
		for i := uint(0); i < uint(len(key)); i++ {
			switch key[i] {
			case delimString:
				// Slice until closing quote
				key = key[:i]
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
	// Append a blank value to n with key
	v = p.value(n, key)
	// We're at ':' after key
	pos = p.parseValue(s, pos+1, v)
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
			if i := len(n.values); 0 <= i && i < len(values) {
				// Zero out Vs that were not used
				values = values[i:]
				for i := range values {
					values[i] = V{}
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

func (n *N) set(info Info, raw string) {
	n.info = info
	n.raw = raw
}

// Update document nodes and returned unused nodes
func (p *parser) update(d *Document) []N {
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
func resetNodes(nodes []N) {
	for i := range nodes {
		n := &nodes[i]
		n.raw = ""
		for i := range n.values {
			n.values[i] = V{}
		}
	}
}

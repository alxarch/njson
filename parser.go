package njson

import (
	"github.com/alxarch/njson/strjson"
	"strings"
)

type parser struct {
	values []value
	n      uint
	err    error
}

// Parse parses a single JSON value from a string and returns a Node and the remaining string
func (d *Document) Parse(s string) (Node, string, error) {
	p := d.parser()
	id := p.n
	pos := p.parseValue(s, 0)
	// Check for any parse errors
	switch p.err.(type) {
	case nil:
		// Update the document
		// Get the values arena from the parser
		d.values = p.values[:p.n]
		// Mark value as root
		d.get(id).flags |= flagRoot
		// Return any tail string after the value
		if pos < uint(len(s)) {
			return Node{id, d.rev, d}, s[pos:], nil
		}
		return Node{id, d.rev, d}, "", nil
	case UnexpectedEOF:
		// Return input as is. Caller can append more data and re-parse.
		return Node{}, s, p.err
	default:
		return Node{}, "", p.err

	}
}

func (d *Document) parser() parser {
	return parser{
		values: d.values[:cap(d.values)],
		n:      uint(len(d.values)),
	}
}

func (p *parser) abort(pos uint, typ Type, got, want interface{}) uint {
	p.err = abort(int(pos), typ, got, want)
	return pos
}

// node returns a node pointer. The pointer is valid until the next call to node()
func (p *parser) value() *value {
	if p.n < uint(len(p.values)) {
		n := &p.values[p.n]
		p.n++
		return n
	}
	values := make([]value, 2*len(p.values)+1)
	copy(values, p.values)
	p.values = values
	if p.n < uint(len(p.values)) {
		n := &p.values[p.n]
		p.n++
		return n
	}
	return nil
}
func appendChild(values []child, key string, id, i uint) []child {
	if i < uint(len(values)) {
		values[i] = child{id, key}
		return values
	}
	tmp := make([]child, 2*len(values)+1)
	copy(tmp, values)
	if i < uint(len(tmp)) {
		tmp[i] = child{id, key}
	}
	return tmp
}

func (p *parser) eof(typ Type, pos uint) uint {
	p.err = UnexpectedEOF(typ)
	return pos
}

func (p *parser) parseValue(s string, pos uint) uint {
	var (
		c   byte
		typ Type
		// We initialize the flag to JSON.
		// We read a string by jumping to the closing quote.
		// Since we do not check every character, we do not know if the string is 'safe' or not.
		f = flags(strjson.FlagJSON)
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
	return p.eof(TypeAnyValue, pos)
readNumber:
	typ = TypeNumber
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
	typ = TypeString
	if pos++; pos < uint(len(s)) {
		// Slice after the opening quote
		s = s[pos:]
		pos++ // Optimization: Early jump to the next character after the closing quote compiles to faster ASM

		// Find the closing quote (")
		i := strings.IndexByte(s, delimString)
		// Check its preceding byte to see if it is escaped
		if j := i-1; 0 <= j && j < len(s) && s[j] == delimEscape { // bounds check elision
			// Advance past '\' and '"' and scan the remaining string
			for i++; 0 <= i && i < len(s); i++ { // bounds check elision
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
			return p.eof(TypeString, pos-1)
		}
		// The end of the string is here
		if 0 <= i && i <= len(s) { // bounds check elision
			// Slice until the closing quote
			s = s[:i]
			// Advance past the closing quote (pos was already incremented)
			pos += uint(i)
			goto done
		}
	}
	return p.eof(TypeString, pos-1)
readTrue:
	typ = TypeBoolean
	if len(s) >= 4 {
		if s = s[:4]; s == strTrue {
			pos += 4
			goto done
		}
		return p.abort(pos, TypeBoolean, s, strTrue)
	}
	return p.eof(TypeBoolean, pos)
readFalse:
	typ = TypeBoolean
	if len(s) >= 5 {
		if s = s[:5]; s == strFalse {
			pos += 5
			goto done
		}
		return p.abort(pos, TypeBoolean, s, strFalse)
	}
	return p.eof(TypeBoolean, pos)
readNull:
	typ = TypeNull
	if len(s) >= 4 {
		if s = s[:4]; s == strNull {
			pos += 4
			goto done
		}
		return p.abort(pos, TypeNull, s, strNull)
	}
	return p.eof(TypeNull, pos)
done:
	n := p.value()
	n.set(typ, f, s)
	for i := range n.children {
		n.children[i] = child{}
	}
	n.children = n.children[:0]
	return pos
}

func (p *parser) parseArray(s string, pos uint) uint {
	var (
		id     = p.n
		v      = p.value()
		c      byte
		values []child
		numV   uint
	)
	v.set(TypeArray, flags(strjson.FlagJSON), "")
	// Skip space after '['
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		if bytemapIsSpace[c] == 0 {
			if c == delimEndArray {
				for i := range v.children {
					v.children[i] = child{}
				}
				v.children = v.children[:0]
				return pos + 1
			}

			values = v.children[:cap(v.children)]
			goto readValue
		}
	}
	return p.eof(TypeArray, pos)
readValue:
	values = appendChild(values, "", p.n, numV)
	numV++
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
			if numV < uint(len(v.children)) {
				values = v.children[:numV]
				v.children = v.children[numV:]
				for i := range v.children {
					v.children[i] = child{}
				}
			} else if numV < uint(len(values)) {
				values = values[:numV]
			}
			if id < uint(len(p.values)) {
				p.values[id].children = values
			}
			return pos + 1
		default:
			if bytemapIsSpace[c] == 0 {
				return p.abort(pos, TypeArray, c, []rune{delimValueSeparator, delimEndArray})
			}
		}
	}
	return p.eof(TypeArray, pos)

}

func (p *parser) parseObject(s string, pos uint) uint {
	var (
		id     = p.n
		v      = p.value()
		c      byte
		key    string
		values []child
		numV   uint
		i      uint
		// Mark all keys as safe JSON
		// If we encounter a backslash ('\\') on any key, we unset safe
		f = flags(strjson.FlagSafe | strjson.FlagJSON)
	)

	v.set(TypeObject, 0, "")
	// Skip space after opening '{'
	for ; pos < uint(len(s)); pos++ {
		c = s[pos]
		switch c {
		case delimEndObject:
			// Zero out the values slice to release key strings
			for i := range v.children {
				v.children[i] = child{}
			}
			v.children = v.children[:0]
			return pos + 1
		case delimString:
			values = v.children[:cap(v.children)]
			goto readKey
		default:
			if bytemapIsSpace[c] == 0 {
				return p.abort(pos, TypeObject, c, []rune{delimEndObject, delimString})
			}
		}
	}
	return p.eof(TypeObject, pos)

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
				// Slice until closing quote
				values = appendChild(values, key[:i], p.n, numV)
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
				return p.eof(TypeObject, pos)
			case delimEscape:
				// Unset safe flag
				f &^= flags(strjson.FlagSafe)
				i++
			}
		}
		// end of input reached
	}
	return p.eof(TypeObject, pos+i)

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
			return p.eof(TypeObject, pos)
		case delimEndObject:
			if numV < uint(len(v.children)) {
				values = v.children[:numV]
				v.children = v.children[numV:]
				// Zero out unused values to release key strings
				for i := range v.children {
					v.children[i] = child{}
				}
			} else if numV < uint(len(values)) {
				values = values[:numV]
				// No need to zero out n.values because n will have no references after return
			}
			// Use id because n pointer might be invalid after a node() call
			if id < uint(len(p.values)) {
				v = &p.values[id]
				v.children = values
				v.flags = f
			}
			return pos + 1
		default:
			if bytemapIsSpace[c] == 0 {
				return p.abort(pos, TypeObject, c, []rune{delimValueSeparator, delimEndObject})
			}
		}
	}
	return p.eof(TypeObject, pos)
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

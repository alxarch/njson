package njson

// parser is a JSON parser.
type parser struct {
	*Document
	depth  int
	parent uint16
	prev   uint16
	mode   Type
}

// Parse parses a JSON source string to a Document.
func (d *Document) Parse(src string) (root *Node, err error) {
	if d == nil {
		err = errNilDocument
		return
	}
	size := len(d.nodes)
	if size >= MaxDocumentSize {
		return nil, errDocumentMaxSize
	}
	n := len(src)
	if n == 0 {
		return nil, errEmptyJSON
	}
	if err = d.parse(src, uint16(size)); err != nil {
		d.nodes = d.nodes[:size]
	} else {
		root = d.Get(size)
	}
	return
}

// ParseUnsafe parses JSON from a buffer without copying it into a string.
// Any modifications to the buffer could mess the document's nodes validity.
// Use only when the buffer is not modified throughout the lifecycle of the document.
func (d *Document) ParseUnsafe(buf []byte) (root *Node, err error) {
	if d == nil {
		err = errNilDocument
		return
	}
	size := len(d.nodes)
	if size >= MaxDocumentSize {
		return nil, errDocumentMaxSize
	}
	n := len(buf)
	if n == 0 {
		return nil, errEmptyJSON
	}
	if err = d.parse(b2s(buf), uint16(size)); err != nil {
		d.nodes = d.nodes[:size]
	} else {
		root = d.Get(size)
	}
	return
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

func (d *Document) parser() parser {
	return parser{
		Document: d,
		parent:   MaxDocumentSize,
	}
}

func (p *parser) push(typ Type, n uint16) {
	p.depth = len(p.stack)
	p.stack = append(p.stack, n)
	p.mode = typ
	p.parent = n
}

func (p *parser) link(id uint16) {
	if int(p.prev) < len(p.nodes) {
		if p.prev == p.parent {
			p.nodes[p.prev].value = id
		} else {
			p.nodes[p.prev].next = id
		}
		p.prev = id

	}
}

// add adds a Node for Token returning the new node's id
func (p *parser) add(t Token) (id uint16) {
	id = uint16(len(p.nodes))
	p.nodes = append(p.nodes, Node{
		doc:    p.Document,
		id:     id,
		parent: p.parent,
		token:  t,
	})
	return

}

func (p *parser) pop() int {
	if 0 <= p.depth && p.depth < len(p.stack) {
		p.stack = p.stack[:p.depth]
		if p.depth--; 0 <= p.depth && p.depth < len(p.stack) {
			p.prev, p.parent = p.parent, p.stack[p.depth]
			if int(p.parent) < len(p.nodes) {
				p.mode = Type(p.nodes[p.parent].token.info)
			}
		}
	}
	return p.depth
}

// Parse parses a JSON string into a Document.
func (d *Document) parse(src string, root uint16) (err error) {
	var (
		p          = d.parser()
		next       uint16
		start, end int // token start, end
		num        uint64
		pos        = 0
		info       ValueInfo
		c          byte
	)

scanValue:
	info = ValueInfo(TypeAnyValue)
	for ; 0 <= pos && pos < len(src); pos++ {
		if c = src[pos]; isSpace(c) {
			continue
		}
		switch c {
		case delimString:
			info, num = ValueInfo(TypeString), 0
			pos++
			start = pos
			for ; 0 <= pos && pos < len(src); pos++ {
				switch c = src[pos]; c {
				case delimString:
					end = pos
					pos++
					goto scanEndValue
				case delimEscape:
					info |= ValueUnescaped
					pos++
				}
			}
			return errEOF
		case delimBeginObject:
			info = ValueInfo(TypeObject)
			switch next = p.add(tokenObject); next {
			case MaxDocumentSize:
				return errDocumentMaxSize
			case root:
				p.prev = next
			default:
				p.link(next)
			}
			p.push(TypeObject, next)
			for pos++; 0 <= pos && pos < len(src); pos++ {
				if c = src[pos]; isSpace(c) {
					continue
				}
				if c == delimEndObject {
					goto scanEndParent
				}
				goto scanKey
			}
		case delimBeginArray:
			info = ValueInfo(TypeArray)
			switch next = p.add(tokenArray); next {
			case MaxDocumentSize:
				return errDocumentMaxSize
			case root:
				p.prev = next
			default:
				p.link(next)
			}
			p.push(TypeArray, next)
			for pos++; 0 <= pos && pos < len(src); pos++ {
				if c = src[pos]; isSpace(c) {
					continue
				}
				if c == delimEndArray {
					goto scanEndParent
				}
				goto scanValue
			}
		case 'n':
			info = ValueInfo(TypeNull)
			if checkUllString(src[pos:]) {
				start, end, num = pos, pos+4, 0
				pos = end
				goto scanEndValue
			}
			return newParseError(pos, c, info)
		case 'f':
			info = ValueFalse
			if !checkAlseString(src[pos:]) {
				return newParseError(pos, c, info)
			}
			start, end, num = pos, pos+5, 0
			pos = end
			goto scanEndValue
		case 't':
			info = ValueTrue
			if checkRueString(src[pos:]) {
				start, end, num = pos, pos+4, 0
				pos = end
				goto scanEndValue
			}
			return newParseError(pos, c, info)
		case 'N':
			info = ValueNumberFloatReady
			if checkAnString(src[pos:]) {
				start, end = pos, pos+3
				pos = end
				num = uNaN
				goto scanEndValue
			}
			return newParseError(pos, c, info)
		case '-':
			start = pos
			info = ValueInfo(TypeNumber) | ValueNegative
			if pos++; 0 <= pos && pos < len(src) {
				c = src[pos]
				goto scanNumber
			}
			return errEOF
		default:
			if isDigit(c) {
				info = ValueInfo(TypeNumber)
				start = pos
				goto scanNumber
			}
		}
		return newParseError(pos, c, info)
	}
scanEndParent:
	pos++
	if p.pop() < 0 {
		goto done
	}
	goto scanMore
scanKey:
	for ; 0 <= pos && pos < len(src); pos++ {
		if c = src[pos]; isSpace(c) {
			continue
		}
		switch c {
		case delimString:
			info = ValueInfo(TypeKey)
			pos++
			start = pos
			for ; 0 <= pos && pos < len(src); pos++ {
				switch c = src[pos]; c {
				case delimString:
					end = pos
					for pos++; 0 <= pos && pos < len(src); pos++ {
						if c = src[pos]; isSpace(c) {
							continue
						}
						if c != delimNameSeparator {
							return newParseError(pos, c, info)
						}
						next = p.add(Token{info: info, src: src[start:end]})
						if root < next && next < MaxDocumentSize {
							if int(p.parent) < len(p.nodes) {
								if p.prev == p.parent {
									p.nodes[p.parent].value = next
									p.push(TypeKey, next)
								} else {
									p.nodes[p.parent].next = next
									p.parent = next
									if int(p.depth) < len(p.stack) {
										p.stack[p.depth] = next
									}
								}

							}
							p.prev = next
							pos++
							goto scanValue
						}
						switch next {
						case MaxDocumentSize:
							return errDocumentMaxSize
						default:
							return errPanic
						}
					}
					return errEOF
				case delimEscape:
					info |= ValueUnescaped
					pos++
				}
			}
			return errEOF
		default:
			return newParseError(pos, c, info)
		}
	}
	return errEOF
scanNumber:
	num = 0
	if c == '0' {
		if pos++; 0 <= pos && pos < len(src) {
			c = src[pos]
		}
	} else {
		for ; 0 <= pos && pos < len(src); pos++ {
			if c = src[pos]; isDigit(c) {
				num = num*10 + uint64(c-'0')
			} else {
				break
			}
		}
	}
	if pos == len(src) || isNumberEnd(c) {
		if info == ValueNegativeInteger {
			num = negative(num)
		}
		goto scanNumberEnd
	}
	num = uNaN
	if c == '.' {
		info |= ValueFloat
		for pos++; 0 <= pos && pos < len(src); pos++ {
			if c = src[pos]; !isDigit(c) {
				break
			}
		}
		if pos == len(src) || isNumberEnd(c) {
			goto scanNumberEnd
		}
	}
scanNumberScientific:
	switch c {
	case 'e', 'E':
		info |= ValueFloat
		for pos++; 0 <= pos && pos < len(src); pos++ {
			if c = src[pos]; isDigit(c) {
				continue
			}
			switch c {
			case '-', '+':
				c = src[pos-1]
				goto scanNumberScientific
			default:
				break
			}
		}
		if pos == len(src) || isNumberEnd(c) {
			goto scanNumberEnd
		}
	}
	return newParseError(pos, c, info)
scanNumberEnd:
	// check last part has at least 1 digit
	if c = src[pos-1]; isDigit(c) {
		end = pos
	} else {
		return newParseError(pos-1, c, info)
	}
scanEndValue:
	next = p.add(Token{info: info, src: src[start:end], num: num})
	switch next {
	case root:
		goto done
	case MaxDocumentSize:
		return errDocumentMaxSize
	default:
		p.link(next)
	}
scanMore:
	for ; 0 <= pos && pos < len(src); pos++ {
		if c = src[pos]; isSpace(c) {
			continue
		}
		switch c {
		case delimValueSeparator:
			switch p.mode {
			case TypeKey:
				pos++
				goto scanKey
			case TypeArray:
				pos++
				goto scanValue
			}
		case delimEndObject:
			switch p.mode {
			case TypeKey:
				p.pop()
				fallthrough
			case TypeObject:
				goto scanEndParent
			}
		case delimEndArray:
			if p.mode == TypeArray {
				goto scanEndParent
			}
		}
		return newParseError(pos, c, info)
	}
	return errEOF
done:
	// Check only space left in source
	for ; 0 <= pos && pos < len(src); pos++ {
		if c = src[pos]; isSpace(c) {
			continue
		}
		return newParseError(pos, c, 0)
	}
	return
}

var (
	tokenObject = Token{
		info: ValueInfo(TypeObject),
	}
	tokenArray = Token{
		info: ValueInfo(TypeArray),
	}
)

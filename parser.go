package njson

// parser is a JSON parser.
type parser struct {
	*Document
	n      uint16
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
	if d.n == MaxDocumentSize {
		return nil, errDocumentMaxSize
	}
	n := len(src)
	if n == 0 {
		return nil, errEmptyJSON
	}
	id, err := d.parse(src, n)
	if err != nil {
		d.nodes = d.nodes[:id]
	} else {
		root = &d.nodes[id]
	}
	d.stack = d.stack[:0]
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
	if d.n == MaxDocumentSize {
		return nil, errDocumentMaxSize
	}
	n := len(buf)
	if n == 0 {
		return nil, errEmptyJSON
	}
	id, err := d.parse(b2s(buf), n)
	if err != nil {
		d.nodes = d.nodes[:id]
	} else {
		root = &d.nodes[id]
	}
	d.stack = d.stack[:0]
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
		n:        MaxDocumentSize,
		parent:   MaxDocumentSize,
	}
}

func (p *parser) push(typ Type, n uint16) {
	// p.n initialized to MaxDocumentSize so p.n++ overflows it to 0
	p.n++
	p.stack = append(p.stack, n)
	p.mode = typ
	p.parent = n
}

func (p *parser) link(id uint16) {
	if p.prev == p.parent {
		p.nodes[p.prev].value = id
	} else {
		p.nodes[p.prev].next = id
	}
	p.prev = id
}

// add adds a Node for Token returning the new node's id
func (p *parser) add(t Token) (id uint16) {
	p.nodes = append(p.nodes, Node{
		doc:    p.Document,
		id:     p.Document.n,
		parent: p.parent,
		token:  t,
	})
	id = p.Document.n
	p.Document.n++
	return

}

func (p *parser) pop() {
	p.stack = p.stack[:p.n]
	p.n--
	p.prev = p.parent
	p.parent = p.stack[p.n]
	p.mode = p.nodes[p.parent].Type()
}

// Parse parses a JSON string into a Document.
func (d *Document) parse(src string, n int) (root uint16, err error) {
	var (
		p          = d.parser()
		next       uint16
		start, end int // token start, end
		num        uint64
		pos        = 0
		info       ValueInfo
		c          byte
	)
	root = d.n

scanValue:
	info = ValueInfo(TypeAnyValue)
	for ; pos < n; pos++ {
		if c = src[pos]; isSpace(c) {
			continue
		}
		switch c {
		case delimString:
			info, num = ValueInfo(TypeString), 0
			pos++
			start = pos
			for ; pos < n; pos++ {
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
			goto eof
		case delimBeginObject:
			info = ValueInfo(TypeObject)
			switch next = p.add(tokenObject); next {
			case MaxDocumentSize:
				goto max
			case root:
				p.prev = next
			default:
				p.link(next)
			}
			p.push(TypeObject, next)
			for pos++; pos < n; pos++ {
				if c = src[pos]; isSpace(c) {
					continue
				} else if c == delimEndObject {
					goto scanEndParent
				}
				goto scanKey
			}
		case delimBeginArray:
			info = ValueInfo(TypeArray)
			switch next = p.add(tokenArray); next {
			case MaxDocumentSize:
				goto max
			case root:
				p.prev = next
			default:
				p.link(next)
			}
			p.push(TypeArray, next)
			for pos++; pos < n; pos++ {
				if c = src[pos]; isSpace(c) {
					continue
				} else if c == delimEndArray {
					goto scanEndParent
				}
				goto scanValue
			}
		case 'n':
			info = ValueInfo(TypeNull)
			if start, end = pos, pos+4; end > n {
				goto eof
			}
			if !checkUllString(src[start:end]) {
				goto abort
			}
			pos, num = end, 0
			goto scanEndValue
		case 'f':
			info = ValueFalse
			if start, end = pos, pos+5; end > n {
				goto eof
			}
			if !checkAlseString(src[start:end]) {
				goto abort
			}
			pos, num = end, 0
			goto scanEndValue
		case 't':
			info = ValueTrue
			if start, end = pos, pos+4; end > n {
				goto eof
			}
			if !checkRueString(src[start:end]) {
				goto abort
			}
			pos, num = end, 0
			goto scanEndValue
		case 'N':
			info = ValueNumberFloatReady
			if start, end = pos, pos+3; end > n {
				goto eof
			}
			if !checkAnString(src[start:end]) {
				goto abort
			}
			pos, num = end, uNaN
			goto scanEndValue
		case '-':
			start = pos
			info = ValueInfo(TypeNumber) | ValueNegative
			if pos++; pos >= n {
				goto eof
			}
			c = src[pos]
			goto scanNumber
		default:
			if isDigit(c) {
				info = ValueInfo(TypeNumber)
				start = pos
				goto scanNumber
			}
			goto abort
		}
	}
eof:
	err = errEOF
	return
abort:
	err = newParseError(pos, c, info)
	return
max:
	p.Document.n = MaxDocumentSize
	p.nodes = p.nodes[:MaxDocumentSize]
	err = errDocumentMaxSize
	return
wtf:
	err = errPanic
	return
scanEndParent:
	pos++
	if p.n == 0 {
		goto done
	}
	p.pop()
	goto scanMore
scanKey:
	for ; pos < n; pos++ {
		if c = src[pos]; isSpace(c) {
			continue
		}
		switch c {
		case delimString:
			info = ValueInfo(TypeKey)
			pos++
			start = pos
			for ; pos < n; pos++ {
				switch c = src[pos]; c {
				case delimString:
					end = pos
					for pos++; pos < n; pos++ {
						if c = src[pos]; isSpace(c) {
							continue
						}
						if c != delimNameSeparator {
							goto abort
						}
						next = p.add(Token{info: info, src: src[start:end]})
						if root < next && next < MaxDocumentSize {
							if p.prev == p.parent {
								// First object key
								d.nodes[p.parent].value = next
							} else {
								d.nodes[p.parent].next = next
								// Pop last key
								p.pop()
							}
							p.prev = next
							p.push(TypeKey, next)
							pos++
							goto scanValue
						}
						switch next {
						case MaxDocumentSize:
							goto max
						default:
							goto wtf
						}
					}
					goto eof
				case delimEscape:
					info |= ValueUnescaped
					pos++
				}
			}
			goto eof
		default:
			goto abort
		}
	}
	goto eof
scanNumber:
	num = 0
	if c == '0' {
		if pos++; pos < n {
			c = src[pos]
		}
	} else {
		for ; pos < n; pos++ {
			if c = src[pos]; isDigit(c) {
				num = num*10 + uint64(c-'0')
			} else {
				break
			}
		}
	}
	if pos == n || isNumberEnd(c) {
		if info == ValueNegativeInteger {
			num = negative(num)
		}
		goto scanNumberEnd
	}
	num = uNaN
	if c == '.' {
		info |= ValueFloat
		for pos++; pos < n; pos++ {
			if c = src[pos]; !isDigit(c) {
				break
			}
		}
		if pos == n || isNumberEnd(c) {
			goto scanNumberEnd
		}
	}
scanNumberScientific:
	switch c {
	case 'e', 'E':
		info |= ValueFloat
		for pos++; pos < n; pos++ {
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
		if pos == n || isNumberEnd(c) {
			goto scanNumberEnd
		}
	}
	goto abort
scanNumberEnd:
	// check last part has at least 1 digit
	if c = src[pos-1]; isDigit(c) {
		end = pos
	} else {
		goto abort
	}
scanEndValue:
	next = p.add(Token{info: info, src: src[start:end], num: num})
	switch next {
	case root:
		goto done
	case MaxDocumentSize:
		goto max
	default:
		p.link(next)
	}
scanMore:
	for ; pos < n; pos++ {
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
		goto abort
	}
	goto eof
done:
	// Check only space left in source
	for ; pos < n; pos++ {
		if c = src[pos]; isSpace(c) {
			continue
		}
		info = 0
		goto abort
	}
	return
}

func checkUllString(data string) bool {
	_ = data[3]
	return data[1] == 'u' && data[2] == 'l' && data[3] == 'l'
}

func checkRueString(data string) bool {
	_ = data[3]
	return data[1] == 'r' && data[2] == 'u' && data[3] == 'e'
}
func checkAnString(data string) bool {
	_ = data[2]
	return data[1] == 'a' && data[2] == 'N'
}
func checkAlseString(data string) bool {
	_ = data[4]
	return data[1] == 'a' && data[2] == 'l' && data[3] == 's' && data[4] == 'e'
}

var (
	tokenObject = Token{
		info: ValueInfo(TypeObject),
	}
	tokenArray = Token{
		info: ValueInfo(TypeArray),
	}
)

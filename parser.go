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
	id, pos, err := d.parse(src, n)
	d.stack = d.stack[:0]
	switch err {
	case nil:
		root = &d.nodes[id]
	case errInvalidToken:
		err = ParseError(pos, src[pos])
		fallthrough
	default:
		d.nodes = d.nodes[:id]
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
	p.nodes[id].parent = p.parent
	switch p.mode {
	case TypeArray:
		if p.prev == p.parent {
			p.nodes[p.prev].value = id
		} else {
			p.nodes[p.prev].next = id
		}
	case TypeKey:
		p.nodes[p.parent].value = id
	default:
		return
	}
	p.prev = id
}

func (p *parser) pop() {
	p.stack = p.stack[:p.n]
	p.n--
	p.prev = p.parent
	p.parent = p.stack[p.n]
	p.mode = p.nodes[p.parent].Type()
}

// Parse parses a JSON string into a Document.
func (d *Document) parse(src string, n int) (root uint16, pos int, err error) {
	var (
		p          = d.parser()
		info       ValueInfo
		next       uint16
		start, end int // token start, end
		num        uint64
		c          byte
	)
	root = d.n

scanToken:
	for ; pos < n; pos++ {
		if c = src[pos]; IsSpaceASCII(c) {
			continue
		}
		switch c {
		case delimBeginObject:
			switch next = d.add(tokenObject); next {
			case MaxDocumentSize:
				goto max
			case root:
				p.prev = next
			default:
				p.link(next)
			}
			p.push(TypeObject, next)
			goto scanKey
		case delimEndObject:
			switch p.mode {
			case TypeKey:
				p.pop()
				fallthrough
			case TypeObject:
				if p.n == 0 {
					return
				}
				p.pop()
			default:
				goto abort
			}
		case delimBeginArray:
			switch next = d.add(tokenArray); next {
			case MaxDocumentSize:
				goto max
			case root:
				p.prev = next
			default:
				p.link(next)
			}
			p.push(TypeArray, next)
		case delimEndArray:
			if p.n == 0 {
				return
			}
			p.pop()
		case delimValueSeparator:
			switch p.mode {
			case TypeObject, TypeKey:
				goto scanKey
			case TypeArray:
			default:
				goto abort
			}
		case delimString:
			info, num = ValueInfo(TypeString), 0
			pos++
			start = pos
			for ; pos < n; pos++ {
				switch c = src[pos]; c {
				case delimString:
					end = pos
					pos++
					goto value
				case delimEscape:
					info |= ValueUnescaped
					pos++
				}
			}
			// will go to eof
		case 'n':
			if start, end = pos, pos+4; end > n {
				goto eof
			}
			if !checkUllString(src[start:end]) {
				goto abort
			}
			pos, num, info = end, 0, ValueInfo(TypeNull)
			goto value
		case 'f':
			if start, end = pos, pos+5; end > n {
				goto eof
			}
			if !checkAlseString(src[start:end]) {
				goto abort
			}
			pos, num, info = end, 0, ValueInfo(TypeBoolean)
			goto value
		case 't':
			if start, end = pos, pos+4; end > n {
				goto eof
			}
			if !checkRueString(src[start:end]) {
				goto abort
			}
			pos, num, info = end, 0, ValueInfo(TypeBoolean)
			goto value
		case 'N':
			if start, end = pos, pos+3; end > n {
				goto eof
			}
			if !checkAnString(src[start:end]) {
				goto abort
			}
			pos, num, info = end, uNaN, ValueNumberFloat
			goto value
		case '-':
			start = pos
			if pos++; pos >= n {
				goto eof
			}
			info = ValueInfo(TypeNumber) | ValueNegative
			goto scanNumber
		default:
			if IsDigit(c) {
				start = pos
				info = ValueInfo(TypeNumber)
				goto scanNumber
			}
			goto abort
		}
	}
eof:
	err = errEOF
	return
abort:
	err = errInvalidToken
	return
max:
	err = errDocumentMaxSize
	return
wtf:
	err = errPanic
	return
scanKey:
	for pos++; pos < n; pos++ {
		if c = src[pos]; IsSpaceASCII(c) {
			continue
		}
		switch c {
		case delimEndObject:
			if p.mode == TypeObject {
				goto scanToken
			}
			goto abort
		case delimString:
			info = ValueInfo(TypeKey)
			pos++
			start = pos
			for ; pos < n; pos++ {
				switch c = src[pos]; c {
				case delimString:
					end = pos
					pos++
					goto scanKeyEnd
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
scanKeyEnd:
	for ; pos < n; pos++ {
		if c = src[pos]; IsSpaceASCII(c) {
			continue
		}
		if c != delimNameSeparator {
			goto abort
		}
		next = d.add(Token{info: info, src: src[start:end]})
		if root < next && next < MaxDocumentSize {
			switch p.mode {
			case TypeObject:
				d.nodes[p.parent].value = next
			case TypeKey:
				d.nodes[p.parent].next = next
				p.pop()
			default:
				goto wtf
			}
			p.prev = next
			p.push(TypeKey, next)
			pos++
			goto scanToken
		}
		switch next {
		case MaxDocumentSize:
			goto max
		default:
			goto wtf
		}
	}
	goto eof

scanNumber:
	num = 0
	if c == '0' {
		if pos++; pos < n {
			c = src[pos]
		}
		goto scanNumberIntegralEnd
	}
	for ; pos < n; pos++ {
		if c = src[pos]; IsDigit(c) {
			num = num*10 + uint64(c-'0')
		} else {
			break
		}
	}
	goto scanNumberIntegralEnd
scanNumberIntegralEnd:
	if pos == n || IsNumberEnd(c) {
		if info == ValueNegativeInteger {
			num = negative(num)
		}
		goto scanNumberEnd
	}
	num = uNaN
	switch c {
	case 'E', 'e':
		info |= ValueFloat
		goto scanNumberScientific
	case '.':
		info |= ValueFloat
		pos++
	default:
		goto abort
	}
	for ; pos < n; pos++ {
		if c = src[pos]; !IsDigit(c) {
			break
		}
	}
	if pos == n || IsNumberEnd(c) {
		goto scanNumberEnd
	}
	switch c {
	case 'e', 'E':
		goto scanNumberScientific
	default:
		goto abort
	}
scanNumberScientific:
	for pos++; pos < n; pos++ {
		if c = src[pos]; IsDigit(c) {
			continue
		}
		if IsNumberEnd(c) {
			break
		}
		switch c {
		case '-', '+':
			switch c = src[pos-1]; c {
			case 'e', 'E':
				continue scanNumberScientific
			}
		}
		goto abort
	}
scanNumberEnd:
	// check last part has at least 1 digit
	if c = src[pos-1]; IsDigit(c) {
		end = pos
		goto value
	}
	goto abort
value:
	next = d.add(Token{info: info, src: src[start:end], num: num})
	switch next {
	case root:
		return
	case MaxDocumentSize:
		goto max
	default:
		p.link(next)
		goto scanToken
	}

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

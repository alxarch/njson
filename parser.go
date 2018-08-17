package njson

import (
	"errors"
	"math"
)

func Parse(src string) (*Document, error) {
	d := Document{}
	_, err := d.CreateNode(src)
	return &d, err
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

type Parser struct {
	stack  []uint16
	n      uint16
	mode   Type
	parent uint16
	prev   uint16
}

func (p *Parser) reset() {
	*p = Parser{
		stack:  p.stack[:0],
		n:      math.MaxUint16,
		parent: MaxDocumentSize,
	}
}

func (p *Parser) push(typ Type, n uint16) {
	p.n++
	p.stack = append(p.stack, n)
	p.mode = typ
	p.parent = n
}

func (p *Parser) link(d *Document, id uint16) {
	d.nodes[id].parent = p.parent
	if p.parent == MaxDocumentSize {
		p.parent = id
	} else {
		switch p.mode {
		case TypeArray:
			d.nodes[p.parent].size++
			if p.prev == p.parent {
				d.nodes[p.prev].value = id
			} else {
				d.nodes[p.prev].next = id
			}
		case TypeKey:
			d.nodes[p.parent].value = id
		default:
			return
		}
	}
	p.prev = id
}

func (p *Parser) pop(d *Document) {
	p.stack = p.stack[:p.n]
	p.n--
	p.prev = p.parent
	p.parent = p.stack[p.n]
	p.mode = d.nodes[p.parent].Type()
}

const (
	MaxDocumentSize = math.MaxUint16
)

var (
	errDocumentMaxSize = errors.New("Document max size")
)

func (p *Parser) Parse(src string, d *Document) (rootNode *Node, err error) {
	if d == nil {
		return nil, errors.New("Nil document")
	}
	var (
		c     byte
		n     = (len(src))
		pos   int
		start int // token start
		end   int // token end
		info  ValueInfo
		num   uint64
		next  uint16
		root  = d.n
	)

	if root == MaxDocumentSize {
		return nil, errDocumentMaxSize
	}

	p.reset()

scanToken:
	for ; pos < n; pos++ {
		if c = src[pos]; IsSpaceASCII(c) {
			continue
		}
		switch c {
		case delimBeginObject:
			if next = d.add(tokenObject); next == MaxDocumentSize {
				return nil, errDocumentMaxSize
			}
			p.link(d, next)
			p.push(TypeObject, next)
			goto scanKey
		case delimEndObject:
			switch p.mode {
			case TypeKey:
				p.pop(d)
				fallthrough
			case TypeObject:
				if p.n == 0 {
					goto done
				}
				p.pop(d)
			default:
				return nil, NewError(pos, ErrObjectEnd)
			}
		case delimBeginArray:
			if next = d.add(tokenArray); next == MaxDocumentSize {
				return nil, errDocumentMaxSize
			}
			p.link(d, next)
			p.push(TypeArray, next)
		case delimEndArray:
			if p.n == 0 {
				goto done
			}
			p.pop(d)
		case delimValueSeparator:
			switch p.mode {
			case TypeObject, TypeKey:
				goto scanKey
			case TypeArray:
			default:
				return nil, NewError(pos, ErrMore)
			}
		case delimString:
			info = ValueInfo(TypeString)
			start = pos
			for pos++; pos < n; pos++ {
				switch c = src[pos]; c {
				case delimString:
					switch next = d.add(Token{info: info, src: src[start : pos+1]}); next {
					case root:
						goto done
					case MaxDocumentSize:
						return nil, errDocumentMaxSize
					default:
						p.link(d, next)
						continue scanToken
					}
				case delimEscape:
					info |= ValueUnescaped
					pos++
				}
			}
		case 'n':
			if start, end = pos, pos+4; end > n {
				goto eof
			}
			if !checkUllString(src[start:end]) {
				return nil, NewError(pos, ErrNull)
			}
			switch next = d.add(tokenNull); next {
			case root:
				goto done
			case MaxDocumentSize:
				return nil, errDocumentMaxSize
			default:
				p.link(d, next)
				pos = end - 1
			}
		case 'f':
			if start, end = pos, pos+5; end > n {
				goto eof
			}
			if !checkAlseString(src[start:end]) {
				return nil, NewError(pos, ErrBoolean)
			}
			switch next = d.add(tokenFalse); next {
			case root:
				goto done
			case MaxDocumentSize:
				return nil, errDocumentMaxSize
			default:
				p.link(d, next)
				pos = end - 1
			}
		case 't':
			if start, end = pos, pos+4; end > n {
				goto eof
			}
			if !checkRueString(src[start:end]) {
				return nil, NewError(pos, ErrBoolean)
			}
			switch next = d.add(tokenTrue); next {
			case root:
				goto done
			case MaxDocumentSize:
				return nil, errDocumentMaxSize
			default:
				p.link(d, next)
				pos = end - 1
			}
		case 'N':
			if start, end = pos, pos+3; end > n {
				goto eof
			}
			if !checkAnString(src[start:end]) {
				return nil, NewError(pos, ErrNumber)
			}
			switch next = d.add(tokenNaN); next {
			case root:
				goto done
			case MaxDocumentSize:
				return nil, errDocumentMaxSize
			default:
				p.link(d, next)
				pos = end - 1
			}

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
			return nil, NewError(pos, ErrType)
		}
	}
eof:
	return nil, NewError(n, ErrEOF)
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
			return nil, NewError(pos, ErrObjectEnd)
		case delimString:
			info = ValueInfo(TypeKey)
			start = pos
			for pos++; pos < n; pos++ {
				switch c = src[pos]; c {
				case delimString:
					pos++
					end = pos
					goto scanKeyEnd
				case delimEscape:
					info |= ValueUnescaped
					pos++
				}
			}
			goto eof
		default:
			return nil, NewError(pos, ErrKey)
		}
	}
	goto eof
scanKeyEnd:
	for ; pos < n; pos++ {
		if c = src[pos]; IsSpaceASCII(c) {
			continue
		}
		if c != delimNameSeparator {
			return nil, NewError(pos, ErrKey)
		}
		if next = d.add(Token{info: info, src: src[start:end]}); next == MaxDocumentSize {
			return nil, errDocumentMaxSize
		}
		switch p.mode {
		case TypeObject:
			d.nodes[p.parent].value = next
		case TypeKey:
			d.nodes[p.parent].next = next
			p.pop(d)
		default:
			return nil, NewError(pos, ErrPanic)
		}
		p.prev = next
		d.nodes[p.parent].size++
		p.push(TypeKey, next)
		pos++
		goto scanToken
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
			num = ^(num - 1)
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
		goto scanNumberError
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
		goto scanNumberError
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
		goto scanNumberError
	}
scanNumberEnd:
	// check last part has at least 1 digit
	if c = src[pos-1]; IsDigit(c) {
		switch next = d.add(Token{info: info, src: src[start:pos], num: num}); next {
		case root:
			goto done
		case MaxDocumentSize:
			return nil, errDocumentMaxSize
		default:
			p.link(d, next)
			goto scanToken
		}
	}
scanNumberError:
	return nil, NewError(pos, ErrNumber)

done:
	return &d.nodes[root], nil

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
	tokenTrue = Token{
		info: ValueInfo(TypeBoolean) | ValueTrue,
		src:  strTrue,
	}
	tokenFalse = Token{
		info: ValueInfo(TypeBoolean),
		src:  strFalse,
	}
	tokenNaN = Token{
		info: ValueNumberFloat,
		src:  strNaN,
		num:  uNaN,
	}
	tokenNull = Token{
		info: ValueInfo(TypeNull),
		src:  strNull,
		num:  uNaN,
	}
)

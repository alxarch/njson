package njson

import (
	"errors"
	"math"
)

func Parse(src string) (*Document, error) {
	d := Document{}
	p := DocumentParser{}
	return &d, p.ParseDocument(src, &d)
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

type DocumentParser struct {
	stack  []uint16
	n      uint16
	mode   Type
	parent uint16
	prev   uint16
}

func (p *DocumentParser) reset() {
	*p = DocumentParser{
		stack:  p.stack[:0],
		n:      math.MaxUint16,
		parent: MaxDocumentSize,
	}
}

func (p *DocumentParser) push(typ Type, n uint16) {
	p.n++
	p.stack = append(p.stack, n)
	p.mode = typ
	p.parent = n
}

func (p *DocumentParser) link(d *Document, id uint16) {
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

func (p *DocumentParser) pop(d *Document) {
	p.stack = p.stack[:p.n]
	p.n--
	p.prev = p.parent
	p.parent = p.stack[p.n]
	p.mode = d.nodes[p.parent].Type()
}

func (p *DocumentParser) ParseDocument(src string, d *Document) error {
	d.Reset()
	_, err := p.Parse(src, d)
	return err
}

const (
	MaxDocumentSize = math.MaxUint16
)

var (
	errDocumentMaxSize = errors.New("Document max size")
)

func (p *DocumentParser) Parse(src string, d *Document) (root uint16, err error) {
	if d == nil {
		return 0, errors.New("Nil document")
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
	)

	if root = d.n; root == MaxDocumentSize {
		return root, errDocumentMaxSize
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
				return root, errDocumentMaxSize
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
					return
				}
				p.pop(d)
			default:
				return root, NewError(pos, ErrObjectEnd)
			}
		case delimBeginArray:
			if next = d.add(tokenArray); next == MaxDocumentSize {
				return root, errDocumentMaxSize
			}
			p.link(d, next)
			p.push(TypeArray, next)
		case delimEndArray:
			if p.n == 0 {
				return
			}
			p.pop(d)
		case delimValueSeparator:
			switch p.mode {
			case TypeObject, TypeKey:
				goto scanKey
			case TypeArray:
			default:
				return root, NewError(pos, ErrMore)
			}
		case delimString:
			info = ValueInfo(TypeString)
			start = pos
			for pos++; pos < n; pos++ {
				switch c = src[pos]; c {
				case delimString:
					switch next = d.add(Token{info: info, src: src[start : pos+1]}); next {
					case root:
						return
					case MaxDocumentSize:
						return root, errDocumentMaxSize
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
				return root, NewError(pos, ErrNull)
			}
			switch next = d.add(tokenNull); next {
			case root:
				return
			case MaxDocumentSize:
				return root, errDocumentMaxSize
			default:
				p.link(d, next)
				pos = end - 1
			}
		case 'f':
			if start, end = pos, pos+5; end > n {
				goto eof
			}
			if !checkAlseString(src[start:end]) {
				return root, NewError(pos, ErrBoolean)
			}
			switch next = d.add(tokenFalse); next {
			case root:
				return
			case MaxDocumentSize:
				return root, errDocumentMaxSize
			default:
				p.link(d, next)
				pos = end - 1
			}
		case 't':
			if start, end = pos, pos+4; end > n {
				goto eof
			}
			if !checkRueString(src[start:end]) {
				return root, NewError(pos, ErrBoolean)
			}
			switch next = d.add(tokenTrue); next {
			case root:
				return
			case MaxDocumentSize:
				return root, errDocumentMaxSize
			default:
				p.link(d, next)
				pos = end - 1
			}
		case 'N':
			if start, end = pos, pos+3; end > n {
				goto eof
			}
			if !checkAnString(src[start:end]) {
				return root, NewError(pos, ErrBoolean)
			}
			switch next = d.add(tokenNaN); next {
			case root:
				return
			case MaxDocumentSize:
				return root, errDocumentMaxSize
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
			return root, NewError(pos, ErrType)
		}
	}
eof:
	return root, NewError(n, ErrEOF)
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
			return root, NewError(pos, ErrObjectEnd)
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
			return root, NewError(pos, ErrKey)
		}
	}
	goto eof
scanKeyEnd:
	for ; pos < n; pos++ {
		if c = src[pos]; IsSpaceASCII(c) {
			continue
		}
		if c != delimNameSeparator {
			return root, NewError(pos, ErrKey)
		}
		if next = d.add(Token{info: info, src: src[start:end]}); next == MaxDocumentSize {
			return root, errDocumentMaxSize
		}
		switch p.mode {
		case TypeObject:
			d.nodes[p.parent].value = next
		case TypeKey:
			d.nodes[p.parent].next = next
			p.pop(d)
		default:
			return 0, NewError(pos, ErrPanic)
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
			continue
		}
		goto scanNumberIntegralEnd
	}
	goto eof
scanNumberIntegralEnd:
	if IsNumberEnd(c) {
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
		if c = src[pos]; IsDigit(c) {
			continue
		}
		if IsNumberEnd(c) {
			goto scanNumberEnd
		}
		switch c {
		case 'e', 'E':
			goto scanNumberScientific
		default:
			goto scanNumberError
		}
	}
	goto eof
scanNumberScientific:
	for pos++; pos < n; pos++ {
		if c = src[pos]; IsDigit(c) {
			continue
		}
		if IsNumberEnd(c) {
			goto scanNumberEnd
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
	goto eof
scanNumberEnd:
	// check last part has at least 1 digit
	if c = src[pos-1]; IsDigit(c) {
		switch next = d.add(Token{info: info, src: src[start:pos], num: num}); next {
		case root:
			return
		case MaxDocumentSize:
			return root, errDocumentMaxSize
		default:
			p.link(d, next)
			goto scanToken
		}
	}
scanNumberError:
	return root, NewError(pos, ErrNumber)

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
	_ = data[1]
	return data[0] == 'a' && data[1] == 'N'
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

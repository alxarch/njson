package scan

import "io"

func ScanToken(input []byte, atEOF bool) (pos int, token []byte, err error) {
	var (
		c byte
		n = len(input)
		t int // token start
	)

scanToken:
	for ; pos < n; pos++ {
		if c = input[pos]; IsSpaceASCII(c) {
			continue scanToken
		}
		switch c {
		case '{', '}', '[', ']', ',', ':':
			t = pos
			pos++
			token = input[t:pos]
			return
		case '"':
			t = pos
			for pos++; pos < n; pos++ {
				switch c = input[pos]; c {
				case '"':
					pos++
					token = input[t:pos]
					return
				case '\\':
					pos++
				}
			}
			goto eof
		case 'f':
			t = pos
			if pos += 5; pos < n {
				if token = input[t:pos]; token[1] == 'a' && token[2] == 'l' && token[3] == 's' && token[4] == 'e' {
					return
				}
				return t, nil, sliceError(pos, ErrBoolean)
			}
			goto eof
		case 'n':
			t = pos
			if pos += 4; pos < n {
				if token = input[t:pos]; token[1] == 'u' && token[2] == 'l' && token[3] == 'l' {
					return
				}
				return t, nil, sliceError(pos, ErrNull)
			}
			goto eof
		case 't':
			t = pos
			if pos += 4; pos < n {
				if token = input[t:pos]; token[1] == 'r' && token[2] == 'u' && token[3] == 'e' {
					return
				}
				return t, nil, sliceError(pos, ErrBoolean)
			}
			goto eof
		case '-':
			t = pos
			if pos++; pos < n {
				c = input[pos]
				goto number
			}
			goto eof
		default:
			if '0' <= c && c <= '9' {
				t = pos
				goto number
			}
			return pos, nil, sliceError(pos, ErrType)
		}
	}
eof:
	if atEOF {
		err = io.ErrUnexpectedEOF
	} else {
		n = pos
	}
	return
number:
	if c == '0' {
		if pos++; pos < n {
			c = input[pos]
		}
		goto scanNumberIntegralEnd
	}
	for ; pos < n; pos++ {
		if c = input[pos]; IsDigit(c) {
			continue
		}
		goto scanNumberIntegralEnd
	}
	goto eof
scanNumberIntegralEnd:
	if IsNumberEnd(c) {
		goto scanNumberEnd
	}
	switch c {
	case 'E', 'e':
		goto scanNumberScientific
	case '.':
		goto scanNumberDecimal
	}
	goto scanNumberError
scanNumberDecimal:
	for pos++; pos < n; pos++ {
		if c = input[pos]; IsDigit(c) {
			continue
		}
		if IsNumberEnd(c) {
			goto scanNumberEnd
		}
		switch c {
		case 'e', 'E':
			goto scanNumberScientific
		}
		goto scanNumberError
	}
	goto eof
scanNumberScientific:
	for pos++; pos < n; pos++ {
		if c = input[pos]; IsDigit(c) {
			continue
		}
		if IsNumberEnd(c) {
			goto scanNumberEnd
		}
		switch c {
		case '-', '+':
			switch c = input[pos-1]; c {
			case 'e', 'E':
				continue scanNumberScientific
			}
		}
		goto scanNumberError
	}
	goto eof
scanNumberEnd:
	if c = input[pos-1]; IsDigit(c) {
		token = input[t:pos]
		return
	}
scanNumberError:
	return pos, nil, sliceError(pos, ErrNumber)

}

// type Token struct {
// 	typ  byte
// 	data []byte
// }

// const (
// 	_                   = iota
// 	nextValueStart uint = 1 << iota
// 	nextObjectEnd
// 	nextKeyEnd
// 	nextPlusMinus
// 	nextNumberStart
// 	nextNumberDecimal
// )

// type Tokens []Token

// type stack []byte

// func (s stack) push(b byte) (_ stack, last int) {
// 	last = len(s)
// 	s = append(s, b)
// 	return s, last
// }

// func (s stack) pop() (_ stack, last int, b byte) {
// 	if last = len(s) - 1; last > 0 {
// 		b = s[last]
// 		s = s[:last]
// 	}
// 	return s, last, 0
// }

// func (tokens Tokens) Tokenize(data []byte) (Tokens, error) {
// 	var (
// 		n     = len(data)
// 		pos   = 0
// 		end = pos
// 		c     byte
// 		state = stack(make([]byte, 1, 64))
// 		last = 0
// 		start = 0
// 		tok   Token
// 	)

// scanValue:
// 	for ; pos < n; pos++ {
// 		c = data[pos]
// 		if IsSpaceASCII(c) {
// 			continue
// 		}
// 		switch c {
// 		case '{':
// 			state, last = state.push('}')
// 			tok = Token{typ: c}
// 			tokens = append(tokens, tok)
// 			goto scanObject
// 		case '[':
// 			state, last = state.push(']')
// 			tok = Token{typ: c}
// 			tokens = append(tokens, tok)
// 			goto scanMore
// 		case '"':
// 			start = pos
// 			for pos++; pos < n; pos++ {
// 				switch c = data[pos]; c {
// 				case '"':
// 					pos++
// 					tokens = append(tokens, Token{typ: '"', data: data[start:pos]})
// 					goto scanMore
// 				case '\\':
// 					pos++
// 				}
// 			}
// 		case 't':
// 			end = pos+4
// 			if end < n && data[pos+1] == 'r' && data[pos+2] == 'u' && data[pos+3]=='e' {
// 				tokens = append(tokens, Token{typ: 't'})
// 				pos+=3
// 				goto scanMore
// 			}
// 			return tokens, sliceError(pos, ErrBoolean)
// 		case 'f':
// 			end = pos+5
// 			if end < n && data[pos+1] == 'a' && data[pos+2] == 'l' && data[pos+3]=='s' && data[pos+4] == 'e' {
// 				tokens = append(tokens, Token{typ: 'f'})
// 				pos+=4
// 				goto scanMore
// 			}
// 			return tokens, sliceError(pos, ErrBoolean)
// 		case 'n':
// 			end = pos+4
// 			if end < n && data[pos+1] == 'u' && data[pos+2] == 'l' && data[pos+3]=='l' {
// 				tokens = append(tokens, Token{typ: 'n'})
// 				pos+=3
// 				goto scanMore
// 			}
// 			return tokens, sliceError(pos, ErrNull)
// 		default:
// 			if IsNumberStart(c) {
// 				goto scanNumber
// 			}
// 			return tokens, fmt.Errorf("?")
// 		}
// 	}
// 	goto eof
// scanObject:
// 	for ; pos < n; pos++ {
// 		if c = data[pos]; IsSpaceASCII(c) { continue }
// 		switch c {
// 		case '"':
// 			start = pos
// 			goto scanKey
// 		case '}':
// 			tokens = append(tokens, Token{typ: '}'})
// 			goto scanMore
// 		default:
// 			return tokens, fmt.Errorf("?")
// 	}
// 	goto eof
// scanKey:
// 	for pos++; pos < n; pos++ {
// 		switch c = data[pos]; c {
// 		case '"':
// 			tokens = append(tokens, Token{typ: ':', data: data[start:pos + 1]})
// 			for pos++; pos < n; pos++ {
// 				if c = data[pos]; IsSpaceASCII(c) { continue }
// 				if c != ':' { return tokens, errors.New(ErrKey)}
// 				goto scanValue
// 			}
// 			goto eof
// 		case '\\':
// 			pos++
// 		}
// 	}
// 	goto eof
// scanMore:
// 	for ; pos < n; pos++ {
// 		if c = data[pos]; IsSpaceASCII(c){ continue }
// 		switch c {
// 		case ',':
// 			goto hasMore
// 		case ']', '}':
// 			goto hasNoMore
// 		default:
// 			return tokens, errors.New("?")
// 		}
// 	}
// 	goto eof
// hasMore:
// 	switch state[last] {
// 	case '}':
// 		goto scanKey
// 	case ']':
// 		goto scanValue
// 	case 0:

// 	}
// scanArrayMore:
// 	for ; pos < n; pos++ {
// 		c = data[pos]
// 		if IsSpaceASCII(c) {
// 			continue
// 		}
// 		switch c {
// 		case ']':
// 		}
// 	}
// arrayEnd:
// 	{
// 		last := len(stack) - 1
// 		open := stack[last]
// 		switch c {
// 		case ']':
// 			if open != '[' {

// 			}
// 		}

// 	}
// 	if open != '[' {
// 		return tokens, fmt.Errorf("Unexpected array end")
// 	}
// 	stack = stack[:last]

// 	return tokens, sliceError(pos, ErrEOF)
// }

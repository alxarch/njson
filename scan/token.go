package scan

import "strconv"

type Token struct {
	typ  byte
	data []byte
	skip int
}

const (
	tokenKey    = ':'
	tokenString = '"'
	tokenMore   = ','
	tokenArray  = '['
	tokenObject = '{'
	tokenNumber = '1'
	tokenNull   = 'n'
	tokenTrue   = 't'
	tokenFalse  = 'f'
	tokenNaN    = 'N'
)

type Tokens []Token

func (tokens Tokens) AppendTo(dst []byte) []byte {
	for i := 0; i < len(tokens); i++ {
		switch token := &tokens[i]; token.typ {
		case tokenArray, tokenObject:
			dst = append(dst, token.typ)
			if skip := i + token.skip; i < skip && skip <= len(tokens) {
				dst = (tokens[i+1 : skip]).AppendTo(dst)
				i = skip - 1
				switch token.typ {
				case tokenObject:
					dst = append(dst, '}')
				case tokenArray:
					dst = append(dst, ']')
				}
			}
		case tokenKey:
			dst = append(dst, '"')
			dst = append(dst, token.data...)
			dst = append(dst, '"', ':')
		case tokenMore:
			dst = append(dst, ","...)
		case tokenNull:
			dst = append(dst, "null"...)
		case tokenNaN:
			dst = append(dst, "NaN"...)
		case tokenFalse:
			dst = append(dst, "false"...)
		case tokenTrue:
			dst = append(dst, "true"...)
		default:
			dst = append(dst, token.data...)
		}
	}
	return dst
}

type TokenError struct {
	Position int64
	Errno    Error
}

func (e TokenError) Error() string {
	data := make([]byte, 0, 64)
	data = append(data, "Token error at position "...)
	data = strconv.AppendInt(data, e.Position, 10)
	data = append(data, ':', ' ')
	data = append(data, e.Errno.String()...)
	return string(data)
}

func NewTokenError(pos int, errno Error) error {
	return TokenError{
		Position: int64(pos),
		Errno:    errno,
	}
}

func Tokenize(tokens Tokens, input []byte) (Tokens, int, error) {
	var (
		c     byte
		n     = len(input)
		nn    = len(tokens) - 1
		pos   = 0
		tmp   []byte
		stack = make([]int, 0, 32)
		p     int
		last  = -1   // last stack item
		token *Token //
		start int    // token start
	)

	// initialize stack
	for p = 0; p <= nn; p++ {
		switch token = &tokens[p]; token.typ {
		case tokenArray, tokenObject:
			if token.skip == -1 {
				last = len(stack)
				stack = append(stack, p)
			}
		}
	}

scanToken:
	for ; pos < n; pos++ {
		if c = input[pos]; IsSpaceASCII(c) {
			continue scanToken
		}
		switch c {
		case tokenKey:
			// Fetch last string token and convert it to key
			if nn < 0 {
				return tokens, pos, sliceError(pos, ErrKey)
			}
			if token = &tokens[nn]; token.typ != tokenString {
				return tokens, pos, NewTokenError(pos, ErrKey)
			}
			token.typ = tokenKey
			token.data = token.data[1 : len(token.data)-1]
		case tokenMore:
			nn = len(tokens)
			tokens = append(tokens, Token{
				typ: tokenMore,
			})
		case '}':
			if last == -1 {
				return tokens, pos, NewTokenError(pos, ErrObject)
			}
			p = stack[last]
			token = &tokens[p]
			if token.typ != tokenObject {
				return tokens, pos, NewTokenError(pos, ErrObject)
			}
			token.skip = len(tokens) - p
			stack = stack[:last]
			last--
		case ']':
			if last == -1 {
				return tokens, pos, NewTokenError(pos, ErrArray)
			}
			p = stack[last]
			token = &tokens[p]
			if token.typ != tokenArray {
				return tokens, pos, NewTokenError(pos, ErrArray)
			}
			token.skip = len(tokens) - p
			stack = stack[:last]
			last--
		case tokenArray:
			last = len(stack)
			nn = len(tokens)
			stack = append(stack, nn)
			tokens = append(tokens, Token{
				typ:  tokenArray,
				skip: -1,
			})
		case tokenObject:
			last = len(stack)
			nn = len(tokens)
			stack = append(stack, nn)
			tokens = append(tokens, Token{
				typ:  tokenObject,
				skip: -1,
			})
		case '"':
			start = pos
			for pos++; pos < n; pos++ {
				switch c = input[pos]; c {
				case '"':
					nn = len(tokens)
					tokens = append(tokens, Token{
						typ:  '"',
						data: input[start : pos+1],
					})
					continue scanToken
				case '\\':
					pos++
				}
			}
			goto eof
		case 'f':
			start = pos
			if pos += 5; pos < n {
				if tmp = input[start:pos]; tmp[1] == 'a' && tmp[2] == 'l' && tmp[3] == 's' && tmp[4] == 'e' {
					pos--
					nn = len(tokens)
					tmp = nil
					tokens = append(tokens, Token{
						typ: 'f',
					})
					continue scanToken
				}
				return tokens, pos, sliceError(pos, ErrBoolean)
			}
			goto eof
		case 'n':
			start = pos
			if pos += 4; pos < n {
				if tmp = input[start:pos]; tmp[1] == 'u' && tmp[2] == 'l' && tmp[3] == 'l' {
					pos--
					nn = len(tokens)
					tokens = append(tokens, Token{
						typ: 'n',
					})
					continue scanToken
				}
				return tokens, pos, sliceError(pos, ErrNull)
			}
			goto eof
		case 't':
			start = pos
			if pos += 4; pos < n {
				if tmp = input[start:pos]; tmp[1] == 'r' && tmp[2] == 'u' && tmp[3] == 'e' {
					pos--
					nn = len(tokens)
					tokens = append(tokens, Token{
						typ: 't',
					})
					continue scanToken
				}
				return tokens, pos, sliceError(pos, ErrBoolean)
			}
			goto eof
		case '-':
			start = pos
			if pos++; pos < n {
				c = input[pos]
				goto number
			}
			goto eof
		default:
			if '0' <= c && c <= '9' {
				start = pos
				goto number
			}
			return tokens, pos, NewTokenError(pos, ErrType)
		}
	}
eof:
	return tokens, pos, nil
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
		nn = len(tokens)
		tokens = append(tokens, Token{
			typ:  '1',
			data: input[start:pos],
		})
		pos--
		goto scanToken
	}
scanNumberError:
	return tokens, start, sliceError(pos, ErrNumber)

}

package scan

import "fmt"

type Slice struct {
	start, end int
}

func (s Slice) Error() string {
	if s.OK() {
		return ErrNone.String()
	}
	return fmt.Sprintf("Scan error at position %d: %s", s.end, Error(s.start).String())
}

func (s Slice) Err() error {
	if s.OK() {
		return nil
	}
	return s
}

func (s Slice) slice(data []byte) []byte {
	return data[s.start:s.end]
}

func (s Slice) OK() bool {
	return s.start >= 0
}
func (s Slice) Len() int {
	return s.start - s.start
}
func (s Slice) IsError(e Error) bool {
	return s.end == int(e)
}
func (s Slice) HasError() bool {
	return s.start < 0
}

func (s Slice) Slice(data []byte) []byte {
	n := len(data)
	if 0 <= s.start && s.start < n {
		if s.start <= s.end && s.end <= n {
			return s.slice(data)
		}
	}
	return nil
}

func NewError(pos int, e Error) error {
	return sliceError(pos, e)
}

func sliceError(pos int, e Error) (s Slice) {
	s.start = int(e)
	s.end = pos
	return s
}

func (s Slice) swap() (i Slice) {
	i.start, i.end = s.end, s.start
	return
}

func (s Slice) error(e Error) Slice {
	if s.end < s.start {
		s.start, s.end = s.end, s.start
	}
	s.start = int(e)
	return s
}

type Error int

func (e Error) String() string {
	switch e {
	case ErrNone:
		return "None"
	case ErrEOF:
		return "EOF"
	case ErrType:
		return "Type"
	case ErrNull:
		return "Null"
	case ErrKey:
		return "Key"
	case ErrEmpty:
		return "Empty"
	case ErrString:
		return "String"
	case ErrArray:
		return "Array"
	case ErrObject:
		return "Object"
	case ErrNumber:
		return "Number"
	case ErrBoolean:
		return "Boolean"
	case ErrPanic:
		return "Invalid scan state"
	default:
		return "Unknown error"
	}
}

const (
	ErrNone Error = 0 - iota
	ErrEOF
	ErrType
	ErrNull
	ErrKey
	ErrEmpty
	ErrString
	ErrArray
	ErrObject
	ErrObjectEnd
	ErrNumber
	ErrBoolean
	ErrPanic
)

func (s Slice) EOF(data []byte) (n int) {
	if n = len(data); s.end < n {
		n = s.end
	}
	return
}
func scanDigits(data []byte, pos int) Slice {
	var (
		c byte
		s = Slice{pos, pos}
	)
	pos = len(data) // reuse
	for ; s.end < pos; s.end++ {
		c = data[s.end]
		if IsDigit(c) {
			continue
		}
		if (IsNumberEnd(c) || c == '.' || c == 'e' || c == 'E') && s.end > s.start {
			return s
		}
		return s.error(ErrNumber)
	}
	return s.error(ErrEOF)
}

func (s Slice) ScanNumber(data []byte) (num Slice) {
	var (
		c     byte
		n     = s.EOF(data)
		state = scanNumberStart
	)

	for s.end = s.start; s.end < n; s.end++ {
		c = data[s.end]
		switch state {
		case scanNumberStart:
			if IsSpaceASCII(c) {
				continue
			}
			if !IsNumberStart(c) {
				return sliceError(s.end, ErrNumber)
			}
			state = scanNumberIntegral
			s.start = s.end
			if IsDigit(c) {
				s.end--
			}
		case scanNumberIntegral:
			if num = scanDigits(data, s.end); num.HasError() {
				return
			}
			if c == '0' && num.Len() > 1 {
				// 0xx number
				return num.error(ErrNumber)
			}
			if s.end = num.end; s.end < n {
				switch c = data[s.end]; c {
				case 'e', 'E':
					state = scanNumberPlusMinus
				case '.':
					state = scanNumberDecimal
				default:
					if IsNumberEnd(c) {
						return s
					}
					return s.error(ErrNumber)
				}
			}
		case scanNumberDecimal:
			if num = scanDigits(data, s.end); num.HasError() {
				return num
			}
			if s.end = num.end; s.end < n {
				switch c = data[s.end]; c {
				case 'e', 'E':
					state = scanNumberPlusMinus
				default:
					if IsNumberEnd(c) {
						return
					}
					return s.error(ErrNumber)
				}
			}
		case scanNumberPlusMinus:
			state = scanNumberScientific
			switch c {
			case '+', '-':
				s.end++
			}
			fallthrough
		case scanNumberScientific:
			if num = scanDigits(data, s.end); num.HasError() {
				return num
			}
			if s.end = num.end; s.end < n {
				c = data[s.end]
				if IsNumberEnd(c) {
					return s
				}
				return s.error(ErrNumber)
			}
		default:
			return s.error(ErrPanic)
		}
	}
	return s.error(ErrEOF)
}

const (
	_ = iota
	scanStringStart
	scanStringEnd
	scanKeyStart
	scanValue
	scanMore
	scanKeyEnd
	scanArrayStart
	scanArrayEnd
	scanObjectStart
	scanEmptyObject
	scanEmptyArray
	scanObjectEnd
	scanNumberStart
	scanNumberIntegral
	scanNumberDecimal
	scanNumberPlusMinus
	scanNumberScientific
)

func NewSlice(data []byte) (s Slice) {
	s.end = len(data)
	return
}

func (s Slice) Reset(data []byte) Slice {
	s.start, s.end = 0, len(data)
	return s
}

func String(data []byte) ([]byte, error) {
	s := NewSlice(data).ScanString(data)
	return s.Slice(data), s.Err()
}
func Any(data []byte) ([]byte, error) {
	s := NewSlice(data).ScanValue(data)
	return s.Slice(data), s.Err()
}

func Key(data []byte) ([]byte, error) {
	name, key := NewSlice(data).ScanKey(data)
	if key.IsError(ErrEmpty) || key.IsError(ErrObjectEnd) {
		return nil, nil
	}
	return name.Slice(data), key.Err()
}

func Object(data []byte) ([]byte, error) {
	s := NewSlice(data).ScanObject(data)
	return s.Slice(data), s.Err()
}

func Array(data []byte) ([]byte, error) {
	s := NewSlice(data).ScanArray(data)
	return s.Slice(data), s.Err()
}
func Number(data []byte) ([]byte, error) {
	s := NewSlice(data).ScanNumber(data)
	return s.Slice(data), s.Err()
}

func (s Slice) ScanToken(data []byte) Slice {
	if s.HasError() {
		return s
	}
	var (
		c byte
		n = s.EOF(data)
	)
	for ; s.start < n; s.start++ {
		c = data[s.start]
		if IsSpaceASCII(c) {
			continue
		}
		switch c {
		case '{', '}', '[', ']', ',', ':':
			s.end = s.start + 1
			return s
		case '"':
			return s.ScanString(data)
		case 't':
			return s.scanTrue(data)
		case 'n':
			return s.scanNull(data)
		case 'f':
			return s.scanFalse(data)
		default:
			if IsNumberStart(c) {
				return s.ScanNumber(data)
			}
			return s.error(ErrType)
		}
	}
	return s.error(ErrEOF)
}

func (s Slice) ScanString(data []byte) Slice {
	var (
		c     byte
		state = scanStringStart
		n     = s.EOF(data)
	)
	for s.end = s.start; s.end < n; s.end++ {
		c = data[s.end]
		switch state {
		case scanStringStart:
			if IsSpaceASCII(c) {
				continue
			}
			if c != '"' {
				return sliceError(s.end, ErrString)
			}
			s.start = s.end
			state = scanStringEnd
		case scanStringEnd:
			switch c {
			case '\\':
				s.end++
			case '"':
				s.end++
				return s
			}
		default:
			return sliceError(s.end, ErrPanic)
		}
	}
	return sliceError(s.end, ErrEOF)
}

func (s Slice) ScanKey(data []byte) (name, key Slice) {
	var (
		c     byte
		state = scanKeyStart
		n     = s.EOF(data)
	)
	for s.end = s.start; s.end < n; s.end++ {
		c = data[s.end]
		switch state {
		case scanKeyStart:
			if IsSpaceASCII(c) {
				continue
			}
			switch c {
			case '{':
				key.start = s.end
				state = scanObjectEnd
			case ',':
				key.start = s.end
				state = scanStringStart
			case '}':
				key = sliceError(s.end, ErrObjectEnd)
				return
			default:
				s = s.error(ErrKey)
				return s, s
			}
		case scanObjectEnd:
			if IsSpaceASCII(c) {
				continue
			}
			if c == '}' {
				name.start = key.start
				name.end = s.end + 1
				return name, s.error(ErrEmpty)
			}
			state = scanStringStart
			fallthrough
		case scanStringStart:
			if IsSpaceASCII(c) {
				continue
			}
			if c != '"' {
				s = s.error(ErrKey)
				return s, s
			}
			name.start = s.end + 1
			state = scanStringEnd
		case scanStringEnd:
			switch c {
			case '"':
				name.end = s.end
				state = scanKeyEnd
			case '\\':
				s.end++
			}
		case scanKeyEnd:
			if IsSpaceASCII(c) {
				continue
			}
			if c == ':' {
				key.end = s.end + 1
				return
			}
			return name, s.error(ErrKey)
		default:
			return name, s.error(ErrPanic)
		}
	}
	return name, s.error(ErrEOF)
}

func (s Slice) ScanObject(data []byte) (obj Slice) {
	var (
		c     byte
		state = scanObjectStart
		n     = s.EOF(data)
		v     Slice
	)
	for s.end = s.start; s.start < n; s.start++ {
		c = data[s.start]
		switch state {
		case scanObjectStart:
			if IsSpaceASCII(c) {
				continue
			}
			if c != '{' {
				return s.error(ErrObject)
			}
			obj.start = s.start
			state = scanEmptyObject
		case scanEmptyObject:
			if IsSpaceASCII(c) {
				continue
			}
			if c == '}' {
				obj.end = s.start + 1
				return
			}
			state = scanStringStart
			fallthrough
		case scanStringStart:
			if IsSpaceASCII(c) {
				continue
			}
			if c != '"' {
				return s.error(ErrKey)
			}
			state = scanStringEnd
		case scanStringEnd:
			switch c {
			case '"':
				state = scanKeyEnd
			case '\\':
				s.start++
			}
		case scanMore:
			if IsSpaceASCII(c) {
				continue
			}
			switch c {
			case ',':
				state = scanStringStart
			case '}':
				obj.end = s.start + 1
				return
			default:
				return s.error(ErrObject)
			}
		case scanKeyEnd:
			if IsSpaceASCII(c) {
				continue
			}
			if c != ':' {
				return s.error(ErrKey)
			}
			state = scanValue
		case scanValue:
			if IsSpaceASCII(c) {
				continue
			}
			switch c {
			case '"':
				v = s.ScanString(data)
			case '{':
				v = s.ScanObject(data)
			case '[':
				v = s.ScanArray(data)
			case 'f':
				v = s.scanFalse(data)
			case 't':
				v = s.scanTrue(data)
			case 'n':
				v = s.scanNull(data)
			default:
				if IsNumberStart(c) {
					v = s.ScanNumber(data)
				}
				return s.error(ErrType)
			}
			if v.HasError() {
				return v
			}
			s.start = v.end - 1
			state = scanMore
		default:
			return s.error(ErrPanic)
		}
	}
	return s.error(ErrEOF)
}

func (s Slice) ScanArray(data []byte) (arr Slice) {
	var (
		c     byte
		n     = s.EOF(data)
		v     Slice
		state = scanArrayStart
	)
	for ; s.start < n; s.start++ {
		c = data[s.start]
		switch state {
		case scanValue:
			v = s.ScanValue(data)
			if v.HasError() {
				return v
			}
			s.start = v.end - 1
			state = scanMore
		case scanMore:
			if IsSpaceASCII(c) {
				continue
			}
			switch c {
			case ',':
				state = scanValue
			case ']':
				arr.end = s.start + 1
				return
			default:
				return s.error(ErrArray)
			}
		case scanArrayStart:
			if IsSpaceASCII(c) {
				continue
			}
			if c != '[' {
				return s.error(ErrArray)
			}
			arr.start = s.start
			state = scanArrayEnd
		case scanArrayEnd:
			if IsSpaceASCII(c) {
				continue
			}
			if c == ']' {
				arr.end = s.start + 1
				return
			}
			state = scanValue
			// Better than fallthrough because switch evaluates in order
			s.start--
		default:
			return s.error(ErrPanic)
		}
	}
	return s.error(ErrEOF)
}

func (s Slice) scanTrue(data []byte) Slice {
	s.end = s.start + 4
	if s.end > len(data) {
		return s.error(ErrEOF)
	}
	if data[s.start] == 't' &&
		data[s.start+1] == 'r' &&
		data[s.start+2] == 'u' &&
		data[s.start+3] == 'e' {
		return s
	}
	return s.error(ErrBoolean)
}

func (s Slice) scanNull(data []byte) Slice {
	s.end = s.start + 4
	if s.end > len(data) {
		return s.error(ErrEOF)
	}
	if data[s.start] == 'n' &&
		data[s.start+1] == 'u' &&
		data[s.start+2] == 'l' &&
		data[s.start+3] == 'l' {
		return s
	}
	return s.error(ErrNull)
}

func (s Slice) scanFalse(data []byte) Slice {
	s.end = s.start + 5
	if s.end > len(data) {
		return s.error(ErrEOF)
	}
	if data[s.start] == 'f' &&
		data[s.start+1] == 'a' &&
		data[s.start+2] == 'l' &&
		data[s.start+3] == 's' &&
		data[s.start+4] == 'e' {
		return s
	}
	return s.error(ErrNull)
}

func (s Slice) ScanValue(data []byte) Slice {
	var (
		c byte
		n = s.EOF(data)
	)
	if s.HasError() {
		return s
	}
	for ; s.start < n; s.start++ {
		c = data[s.start]
		if IsSpaceASCII(c) {
			continue
		}
		switch c {
		case '"':
			return s.ScanString(data)
		case '{':
			return s.ScanObject(data)
		case '[':
			return s.ScanArray(data)
		case 'f':
			return s.scanFalse(data)
		case 't':
			return s.scanTrue(data)
		case 'n':
			return s.scanNull(data)
		default:
			if IsNumberStart(c) {
				return s.ScanNumber(data)
			}
			return s.error(ErrType)
		}
	}
	return s.error(ErrEOF)
}

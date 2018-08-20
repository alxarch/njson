package njson

import (
	"fmt"
	"strconv"
)

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
	case ErrArrayEnd:
		return "Unexpected array end"
	case ErrObjectEnd:
		return "Unexpected object end"
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
	ErrMore
	ErrNull
	ErrKey
	ErrEmpty
	ErrString
	ErrStringUnescape
	ErrArray
	ErrArrayEnd
	ErrObject
	ErrObjectEnd
	ErrNumber
	ErrBoolean
	ErrPanic
)

type TokenError struct {
	Position int64
	Errno    Error
	Token    byte
}

func (e TokenError) Error() string {
	data := make([]byte, 0, 64)
	data = append(data, "Token error at position "...)
	data = strconv.AppendInt(data, e.Position, 10)
	data = append(data, ':', ' ')
	data = append(data, e.Errno.String()...)
	return string(data)
}

func NewError(pos int, errno Error) error {
	return TokenError{
		Position: int64(pos),
		Errno:    errno,
	}
}

type typeError struct {
	Type Type
	Want Type
}

func TypeError(t, want Type) error {
	return typeError{t, want}
}
func (e typeError) Error() string {
	return fmt.Sprintf("Invalid type %s not in %v", e.Type, e.Want.Types())
}

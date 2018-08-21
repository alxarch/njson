package njson

import (
	"errors"
	"fmt"
	"strconv"
)

type errMsg string

func (e errMsg) Error() string {
	return string(e)
}
func (e errMsg) String() string {
	return string(e)
}

var (
	errDocumentMaxSize = errors.New("Document max size")
	errNilDocument     = errors.New("Nil document")
	errPanic           = errors.New("Invalid parser state")
	errEOF             = errors.New("Unexpected end of input")
	errEmptyJSON       = errors.New("Empty JSON source")
	errInvalidToken    = errors.New("Invalid token")
)

type parseError struct {
	pos int64
	c   byte
	// str  string
	// want string
	// typ Type
}

func (p parseError) Error() string {
	buf := make([]byte, 0, 64)
	buf = append(buf, "Invalid token '"...)
	buf = append(buf, p.c)
	buf = append(buf, "' at position "...)
	buf = strconv.AppendInt(buf, p.pos, 10)
	// if p.typ != 0 {
	// 	buf = append(buf, " while scanning "...)
	// 	buf = append(buf, p.typ.String()...)
	// }
	return string(buf)
}
func ParseError(pos int, c byte) error {
	return parseError{int64(pos), c}
}

type typeError struct {
	Type Type
	Want Type
}

func TypeError(t, want Type) error {
	return typeError{t, want}
}
func (e typeError) Error() string {
	if e.Want&e.Type != 0 {
		return fmt.Sprintf("Invalid value for type %s", e.Type)
	}
	return fmt.Sprintf("Invalid type %s not in %v", e.Type, e.Want.Types())
}

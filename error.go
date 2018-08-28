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
	errInvalidNode     = errors.New("Invalid node")
)

type parseError struct {
	pos  int
	c    byte
	info ValueInfo
	// str  string
	// want string
	// typ Type
}

func (p parseError) Error() (msg string) {
	buf := blankBuffer(minBufferSize)
	buf = append(buf[:0], "Invalid token '"...)
	buf = append(buf, p.c)
	buf = append(buf, "' at position "...)
	buf = strconv.AppendInt(buf, int64(p.pos), 10)
	if p.info != 0 {
		buf = append(buf, " while scanning for "...)
		buf = append(buf, p.info.Type().String()...)
	}
	msg = string(buf)
	putBuffer(buf)
	return
}

func newParseError(pos int, c byte, i ValueInfo) error {
	return parseError{pos, c, i}
}

type typeError struct {
	Type Type
	Want Type
}

// newTypeError returns a type mismatch error.
func newTypeError(t, want Type) error {
	return typeError{t, want}
}

func (e typeError) Error() string {
	if e.Want&e.Type != 0 {
		return fmt.Sprintf("Invalid value for type %s", e.Type)
	}
	return fmt.Sprintf("Invalid type %s not in %v", e.Type, e.Want.Types())
}

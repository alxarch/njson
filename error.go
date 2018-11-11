package njson

import "fmt"

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

// ParseError signifies an invalid token in JSON data
type ParseError struct {
	got  interface{}
	want interface{}
	pos  int
	typ  Type
}

// Type returns type of value that was being parsed when the error ocurred.
func (e *ParseError) Type() Type {
	return e.typ
}

// Pos returns the offset at which the error ocurred
func (e *ParseError) Pos() int {
	return e.pos
}

func (e *ParseError) Error() string {
	if e == nil {
		return fmt.Sprintf("%v", error(nil))
	}
	return fmt.Sprintf("Invalid token %q != %q at position %d while scanning %s", e.got, e.want, e.pos, e.typ.String())
}

// UnexpectedEOF signifies incomplete JSON data
type UnexpectedEOF Type

func (e UnexpectedEOF) Error() string {
	return fmt.Sprintf("Unexpected end of input while scanning %s", Type(e).String())
}

func abort(pos int, typ Type, got interface{}, want interface{}) error {
	return &ParseError{
		pos:  pos,
		typ:  typ,
		got:  got,
		want: want,
	}
}

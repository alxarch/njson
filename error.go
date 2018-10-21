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

type parseError struct {
	got  interface{}
	want interface{}
	pos  int
	typ  Type
}

func (e *parseError) Error() string {
	if e == nil {
		return fmt.Sprintf("%v", error(nil))
	}
	if e.pos == -1 {
		return fmt.Sprintf("Unexpected end of input while scanning %s", e.typ.String())
	}
	if e.got == nil {
		return fmt.Sprintf("Invalid parser state at position %d %v", e.pos, e.want)
	}
	if e.want != nil {
		return fmt.Sprintf("Invalid token %q != %q at position %d while scanning %s", e.got, e.want, e.pos, e.typ.String())
	}
	return fmt.Sprintf("Invalid token %q at position %d while scanning %s", e.got, e.pos, e.typ.String())
}

func eof(typ Type) error {
	return &parseError{
		pos: -1,
		typ: typ,
	}
}
func abort(pos int, typ Type, got interface{}, want interface{}) error {
	return &parseError{
		pos:  pos,
		typ:  typ,
		got:  got,
		want: want,
	}
}

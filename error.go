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

func (e parseError) Err() error {
	if e.pos == -1 {
		return fmt.Errorf("Unexpected end of input while scanning %s %q %q", e.typ.String(), e.got, e.want)
	}
	if e.got == nil {
		return nil
	}
	if e.typ == TypeInvalid {
		return fmt.Errorf("Invalid parser state at position %d %v %v", e.pos, e.got, e.want)
	}
	if e.want != nil {
		return fmt.Errorf("Invalid token %q != %q at position %d while scanning %s", e.got, e.want, e.pos, e.typ.String())
	}
	return fmt.Errorf("Invalid token %q at position %d while scanning %s", e.got, e.pos, e.typ.String())
}

func eof(typ Type) error {
	return parseError{
		pos: -1,
		typ: typ,
	}.Err()
}
func abort(pos int, typ Type, got interface{}, want interface{}) error {
	return parseError{
		pos:  pos,
		typ:  typ,
		got:  got,
		want: want,
	}.Err()
}

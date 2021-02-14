package njson

import "fmt"

func errNodeInvalid(typ Type, op string, n Node) Node {
	if err := n.Err(); err != nil {
		return n
	}
	return errNode(fmt.Errorf("method %s.%s() called on an invalid node", typ, op))
}
func errNodeLocked(typ Type, op string) Node {
	return errNode(fmt.Errorf("method %s.%s() called during iteration", typ, op))
}

func errNodeType(typ Type, op string, got Type) Node {
	return errNode(fmt.Errorf("method %s.%s() called on a %s node", typ, op, got))
}

// newTypeError returns a type mismatch error.
func newTypeError(t, want Type) error {
	return &TypeError{Type: t, Want: want}
}

type TypeError struct {
	Type Type
	Want Type
}

func (e *TypeError) Error() string {
	if e.Type == TypeInvalid && e.Want == TypeAnyValue {
		return "value type is not valid"

	}
	return fmt.Sprintf("value type is %s, expecting %v", e.Type, e.Want.Types())
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
	return fmt.Sprintf("invalid token %q != %q at position %d while scanning %s", e.got, e.want, e.pos, e.typ.String())
}

// UnexpectedEOF signifies incomplete JSON data
type UnexpectedEOF Type

func (e UnexpectedEOF) Error() string {
	return fmt.Sprintf("unexpected end of input while scanning %s", Type(e).String())
}

func abort(pos int, typ Type, got interface{}, want interface{}) error {
	return &ParseError{
		pos:  pos,
		typ:  typ,
		got:  got,
		want: want,
	}
}

type KeyError struct {
	Key string
}

func (e *KeyError) Error() string {
	return fmt.Sprintf("key %q not found in JSON object", e.Key)
}

package njson

import (
	"encoding"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/alxarch/njson/numjson"
	"github.com/alxarch/njson/strjson"
)

// Node is a reference to a node in a JSON Document.
// It is a versioned reference to avoid document manipulation after reset.
type Node struct {
	doc *Document
	id  uint
	rev uint
}

// Type returns a Node's type.
func (n Node) Type() Type {
	if v := n.value(); v != nil {
		return v.typ
	}
	return TypeInvalid
}

func (n Node) Err() error {
	if d := n.doc; d != nil {
		if d.rev == n.rev && n.id < uint(len(d.values)) {
			return nil
		}
		return d.err
	}
	return ErrNodeInvalid
}

// Inspect returns a node's raw string and type
// Object and Array values return an empty string.
func (n Node) Inspect() (string, Type) {
	if v := n.value(); v != nil {
		return v.raw, v.typ
	}
	return "", TypeInvalid
}

// Document returns a node's document.
func (n Node) Document() *Document {
	if d := n.doc; d != nil && d.rev == n.rev {
		return d
	}
	return nil
}

func (n Node) IsValid() bool {
	return n.value() != nil
}

func (n Node) Object() Object {
	return Object(n)
}
func (n Node) Array() Array {
	return Array(n)
}

func (n Node) ToString() (string, Type) {
	if v := n.value(); v != nil {
		switch v.typ {
		case TypeString:
			s, f := v.str()
			if f.IsGoSafe() {
				return s, TypeString
			}
			if f.IsJSON() && strings.IndexByte(s, '\\') == -1 {
				v.flags |= flags(strjson.FlagSafe)
				return s, TypeString
			}
			return strjson.Unescaped(s), TypeString
		case TypeNumber, TypeBoolean, TypeNull:
			return v.raw, v.typ
		case TypeObject, TypeArray:
			return "", v.typ
		default:
			return "", TypeInvalid
		}
	}
	// fallback to an invalid empty string
	return "", TypeInvalid
}

func (v *value) str() (string, strjson.Flags) {
	return v.raw, strjson.Flags(v.flags)
}

// Number parses a node's value as numjson.Number
// If the parsing fails, or the node is not a JSON number, an invalid numjson.Number is returned
func (n Node) Number() numjson.Number {
	if s, typ := n.Inspect(); typ == TypeNumber {
		if num, err := numjson.Parse(s); err == nil {
			return num
		}
	}
	return numjson.Number{}
}

func (n Node) Boolean() Const {
	if v := n.value(); v != nil && v.typ == TypeBoolean {
		return Const(v.raw)
	}
	return ""
}

// Const is a constant JSON value
type Const string

const (
	True  Const = "true"
	False Const = "false"
	Null  Const = "null"
)

func (c Const) IsTrue() bool {
	return c == True
}
func (c Const) IsFalse() bool {
	return c == False
}
func (c Const) IsNull() bool {
	return c == Null
}

func (n Node) IsNull() bool {
	if v := n.value(); v != nil {
		return v.typ == TypeNull
	}
	return false
}

// Unmarshaler is the interface implemented by types that can unmarshal from a Node.
type Unmarshaler interface {
	UnmarshalNodeJSON(n Node) error
}

func (n Node) value() *value {
	if d := n.doc; d != nil && d.rev == n.rev && n.id < uint(len(d.values)) {
		return &d.values[n.id]
	}
	return nil
}

// AppendJSON appends a node's JSON data to a byte slice.
func (n Node) AppendJSON(dst []byte) ([]byte, error) {
	if v := n.value(); v != nil {
		return n.doc.appendJSON(dst, v)
	}
	return nil, ErrNodeInvalid
}

// ToNumber converts a node's value to numjson.Number
// If a node's value cannot be converted to a number an error is returned
func (n Node) ToNumber() (numjson.Number, error) {
	if v := n.value(); v != nil {
		return v.toNumber()
	}
	return numjson.Number{}, ErrNodeInvalid
}

var ErrNodeInvalid = errors.New("node is not valid")

func (v *value) toNumber() (numjson.Number, error) {
	switch v.typ {
	case TypeNumber, TypeString:
		return numjson.Parse(v.raw)
	case TypeBoolean:
		if v.raw == "true" {
			return numjson.Int64(1), nil
		}
		return numjson.Int64(0), nil
	default:
		return numjson.Number{}, newTypeError(v.typ, TypeNumber|TypeString|TypeBoolean)
	}
}

// ToBool converts a Node to bool.
func (n Node) ToBool() (bool, bool) {
	if v := n.value(); v != nil && v.typ == TypeBoolean {
		switch v.raw {
		case strTrue:
			return true, true
		case strFalse:
			return false, true
		}
	}
	return false, false
}

// TypeError returns an error for a type not matching a Node's type.
func (n Node) TypeError(want Type) error {
	return newTypeError(n.Type(), want)
}

// Lookup finds a node by path
func (n Node) Lookup(path ...string) Node {
	return n.with(n.Document().lookup(n.id, path))
}

// ToInterface converts a Node to a generic interface{}.
func (n Node) ToInterface() (interface{}, bool) {
	return n.Document().toInterface(n.id)
}

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 2048)
	},
}

// Appender is a Marshaler interface for buffer append workflows.
type Appender interface {
	AppendJSON([]byte) ([]byte, error)
}

// PrintJSON is a helper to write an Appender to an io.Writer
func PrintJSON(w io.Writer, a Appender) (n int, err error) {
	b := bufferPool.Get().([]byte)
	if b, err = a.AppendJSON(b[:0]); err == nil {
		n, err = w.Write(b)
	}
	bufferPool.Put(b)
	return
}

// PrintJSON writes JSON to an io.Writer.
func (n Node) PrintJSON(w io.Writer) (int, error) {
	return PrintJSON(w, n)
}

// WrapUnmarshalJSON wraps a call to the json.Unmarshaler interface
func (n Node) WrapUnmarshalJSON(u json.Unmarshaler) (err error) {
	v := n.value()
	if v == nil {
		return ErrNodeInvalid
	}

	switch v.typ {
	case TypeArray:
		if len(v.children) == 0 {
			return u.UnmarshalJSON([]byte{delimBeginArray, delimEndArray})
		}
	case TypeObject:
		if len(v.children) == 0 {
			return u.UnmarshalJSON([]byte{delimBeginObject, delimEndObject})
		}
	case TypeString:
		if v.raw == "" {
			return u.UnmarshalJSON([]byte{delimString, delimString})
		}
	case TypeInvalid:
		return newTypeError(TypeInvalid, TypeAnyValue)
	}
	data := bufferPool.Get().([]byte)
	data, err = n.AppendJSON(data[:0])
	if err == nil {
		err = u.UnmarshalJSON(data)
	}
	bufferPool.Put(data)
	return
}

// WrapUnmarshalText wraps a call to the encoding.TextUnmarshaler interface
func (n Node) WrapUnmarshalText(u encoding.TextUnmarshaler) (err error) {
	if v := n.value(); v != nil {
		switch t := v.typ; t {
		case TypeString:
			buf := bufferPool.Get().([]byte)
			buf = append(buf[:0], v.raw...)
			err = u.UnmarshalText(buf)
			bufferPool.Put(buf)
			return
		default:
			return newTypeError(t, TypeString)
		}
	}
	return newTypeError(TypeInvalid, TypeString)
}

// With returns a document node for id.
func (n Node) with(id uint) Node {
	n.id = id
	return n
}

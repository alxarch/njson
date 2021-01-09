package njson

import (
	"encoding"
	"encoding/json"
	"io"
	"strconv"
	"sync"

	"github.com/alxarch/njson/numjson"
	"github.com/alxarch/njson/strjson"
)

// Node is a reference to a node in a JSON Document.
// It is a versioned reference to avoid document manipulation after reset.
type Node struct {
	id  uint
	rev uint
	doc *Document
}

// Type returns a Node's type.
func (n Node) Type() Type {
	if v := n.value(); v != nil {
		return v.typ
	}
	return TypeInvalid
}

// Document returns a node's document.
func (n Node) Document() *Document {
	if d := n.doc; d != nil && d.rev == n.rev {
		return d
	}
	// Unlink invalid Document reference
	n.doc = nil
	return nil
}

func (n Node) IsZero() bool {
	return n == Node{}
}


// Unmarshaler is the interface implemented by types that can unmarshal from a Node.
type Unmarshaler interface {
	UnmarshalNodeJSON(n Node) error
}

func (n Node) value() *value {
	if n.doc != nil && n.doc.rev == n.rev {
		if n.id < uint(len(n.doc.values)) {
			return &n.doc.values[n.id]
		}
	} else {
		// Unlink invalid Document reference
		n.doc = nil
	}
	return nil
}
func (n Node) Object() Object {
	return Object(n)
}
func (n Node) Array() Array {
	return Array(n)
}
// AppendJSON appends a node's JSON data to a byte slice.
func (n Node) AppendJSON(dst []byte) ([]byte, error) {
	if v := n.value(); v != nil {
		return n.doc.appendJSON(dst, v)
	}
	return nil, &typeError{TypeInvalid, TypeAnyValue}
}

// Raw returns the JSON string of a Node's value.
// Object and Array values return an empty string.
func (n Node) Raw() string {
	if n := n.value(); n != nil {
		return n.raw
	}
	return ""
}

func (n Node) ToString() (string, bool) {
	if v := n.value(); v != nil && v.typ == TypeString {
		if v.flags.IsSimpleString() {
			return v.raw, true
		}
		return strjson.Unescaped(v.raw), true
	}
	return "", false
}

// Data returns a node's raw string and type
func (n Node) Data() (string, Type) {
	if v := n.value(); v != nil {
		return v.raw, v.typ
	}
	return "", TypeInvalid
}

func (n Node) Text() (string, Type) {
	if v := n.value(); v != nil {
		switch v.typ {
		case TypeString:
			if v.flags.IsSimpleString() || v.flags.IsUnescaped() {
				return v.raw, TypeString
			}
			v.raw = strjson.Unescaped(v.raw)
			v.flags |= flagUnescapedString
			return v.raw, TypeString
		case TypeNumber, TypeBoolean:
			return v.raw, v.typ
		default:
			return "", v.typ
		}
	}
	return "", TypeInvalid
}


// ToFloat converts a node's value to float64.
func (n Node) ToFloat() (float64, bool) {
	if n := n.value(); n != nil {
		f := numjson.ParseFloat(n.raw)
		return f, f == f
	}
	return 0, false
}

// ToInt converts a node's value to int64.
func (n Node) ToInt() (int64, bool) {
	if n := n.value(); n != nil {
		return numjson.ParseInt(n.raw)
	}
	return 0, false
}

// ToUint converts a node's  value to uint64.
func (n Node) ToUint() (uint64, bool) {
	if n := n.value(); n != nil {
		return numjson.ParseUint(n.raw)
	}
	return 0, false
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
	return typeError{n.Type(), want}
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
		return typeError{TypeInvalid, TypeAnyValue}
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
		return typeError{TypeInvalid, TypeAnyValue}
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


// SetInt sets a Node's value to an integer.
func (n Node) SetInt(i int64) {
	if v := n.value(); v != nil {
		v.reset(TypeNumber, v.flags&flagRoot, strconv.FormatInt(i, 10), v.children[:0])
	}

}

// SetUint sets a Node's value to an unsigned integer.
func (n Node) SetUint(u uint64) {
	if v := n.value(); v != nil {
		v.reset(TypeNumber, v.flags&flagRoot, strconv.FormatUint(u, 10), v.children[:0])
	}

}

// SetFloat sets a Node's value to a float number.
func (n Node) SetFloat(f float64) {
	if v := n.value(); v != nil {
		v.reset(TypeNumber, v.flags&flagRoot, numjson.FormatFloat(f, 64), v.children[:0])
	}
}

// SetString sets a Node's value to a string escaping invalid JSON characters.
func (n Node) SetString(s string) {
	n.SetStringRaw(strjson.Escaped(s, false, false))
}

// SetStringHTML sets a Node's value to a string escaping invalid JSON and unsafe HTML characters.
func (n Node) SetStringHTML(s string) {
	n.SetStringRaw(strjson.Escaped(s, true, false))
}

// SetStringRaw sets a Node's value to a string without escaping.
// The provided string *must* not contain any JSON invalid characters,
// otherwise JSON output from this Node will be invalid.
func (n Node) SetStringRaw(s string) {
	if v := n.value(); v != nil {
		v.reset(TypeString, v.flags&flagRoot, s, v.children[:0])
	}
}

// SetFalse sets a Node's value to false.
func (n Node) SetFalse() {
	if v := n.value(); v != nil {
		v.reset(TypeBoolean, v.flags&flagRoot, strFalse, v.children[:0])
	}
}

// SetTrue sets a Node's value to true.
func (n Node) SetTrue() {
	if v := n.value(); v != nil {
		v.reset(TypeBoolean, v.flags&flagRoot, strTrue, v.children[:0])
	}
}

// SetNull sets a Node's value to null.
func (n Node) SetNull() {
	if v := n.value(); v != nil {
		v.reset(TypeNull, v.flags&flagRoot, strNull, v.children[:0])
	}
}

// With returns a document node for id.
func (n Node) with(id uint) Node {
	n.id = id
	return n
}

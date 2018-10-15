package njson

import (
	"encoding"
	"encoding/json"
	"io"
	"sync"
)

// Node is a reference to a node in a JSON Document.
// It is a versioned reference to avoid document manipulation after reset.
type Node struct {
	id  uint
	rev uint
	doc *Document
}

type Unmarshaler interface {
	UnmarshalNodeJSON(n Node) error
}

func (n Node) ID() uint {
	return n.id
}

func (n Node) With(id uint) Node {
	n.id = id
	return n
}

func (n Node) Document() *Document {
	if n.doc != nil && n.doc.rev == n.rev {
		return n.doc
	}
	return nil
}

func (n Node) get() *N {
	return n.Document().get(n.id)
}

func (n Node) AppendJSON(dst []byte) ([]byte, error) {
	return n.Document().AppendJSON(dst, n.id)
}

func (n Node) Raw() string              { return n.get().Raw() }
func (n Node) Unescaped() string        { return n.get().Unescaped() }
func (n Node) ToFloat() (float64, bool) { return n.get().ToFloat() }
func (n Node) ToInt() (int64, bool)     { return n.get().ToInt() }
func (n Node) ToUint() (uint64, bool)   { return n.get().ToUint() }
func (n Node) ToBool() (bool, bool)     { return n.get().ToBool() }
func (n Node) Type() Type               { return n.get().Type() }
func (n Node) Bytes() []byte            { return n.get().Bytes() }
func (n Node) Values() IterV            { return n.get().Values() }

func (n Node) TypeError(want Type) error {
	return n.get().TypeError(want)
}

// Lookup finds a node by path
func (n Node) Lookup(path ...string) Node {
	return n.With(n.Document().Lookup(n.id, path))
}

func (n Node) ToInterface() (interface{}, bool) {
	return n.Document().ToInterface(n.id)
}

var bufferpool = &sync.Pool{
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
	b := bufferpool.Get().([]byte)
	if b, err = a.AppendJSON(b[:0]); err == nil {
		n, err = w.Write(b)
	}
	bufferpool.Put(b)
	return
}

func (n Node) PrintJSON(w io.Writer) (int, error) {
	return PrintJSON(w, n)
}

// WrapUnmarshalJSON wraps a call to the json.Unmarshaler interface
func (n Node) WrapUnmarshalJSON(u json.Unmarshaler) (err error) {
	node := n.get()
	if node == nil {
		return node.TypeError(TypeAnyValue)
	}

	switch node.info.Type() {
	case TypeArray:
		if len(node.values) == 0 {
			return u.UnmarshalJSON([]byte{delimBeginArray, delimEndArray})
		}
	case TypeObject:
		if len(node.values) == 0 {
			return u.UnmarshalJSON([]byte{delimBeginObject, delimEndObject})
		}
	case TypeInvalid:
		return node.TypeError(TypeAnyValue)
	default:
		return u.UnmarshalJSON(s2b(node.raw))
	}
	data := bufferpool.Get().([]byte)
	data, err = n.Document().AppendJSON(data[:0], n.id)
	if err == nil {
		err = u.UnmarshalJSON(data)
	}
	bufferpool.Put(data)
	return
}

// WrapUnmarshalText wraps a call to the encoding.TextUnmarshaler interface
func (n Node) WrapUnmarshalText(u encoding.TextUnmarshaler) (err error) {
	node := n.get()
	if node != nil && node.info.IsString() {
		return u.UnmarshalText(node.Bytes())
	}
	return node.TypeError(TypeAnyValue)
}

func (n Node) Get(key string) Node {
	return n.With(n.get().Get(key))
}
func (n Node) Index(i int) Node {
	return n.With(n.get().Index(i))
}

func (n Node) Set(key string, id uint) { n.get().Set(key, id) }
func (n Node) Slice(i, j int)          { n.get().Slice(i, j) }
func (n Node) Del(key string)          { n.get().Del(key) }
func (n Node) SetInt(i int64)          { n.get().SetInt(i) }
func (n Node) SetUint(u uint64)        { n.get().SetUint(u) }
func (n Node) SetFloat(f float64)      { n.get().SetFloat(f) }
func (n Node) SetString(s string)      { n.get().SetString(s) }
func (n Node) SetStringHTML(s string)  { n.get().SetStringHTML(s) }
func (n Node) SetStringRaw(s string)   { n.get().SetStringRaw(s) }
func (n Node) SetFalse()               { n.get().SetFalse() }
func (n Node) SetTrue()                { n.get().SetTrue() }
func (n Node) SetNull()                { n.get().SetNull() }

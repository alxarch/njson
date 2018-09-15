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

func (n Node) Raw() string {
	return n.get().Raw()
}
func (n Node) Unescaped() string {
	return n.get().Unescaped()
}
func (n Node) ToFloat() (float64, bool) {
	return n.get().ToFloat()
}
func (p Node) ToInt() (int64, bool) {
	if n := p.get(); n != nil {
		return n.ToInt()
	}
	return 0, false
}
func (p Node) ToUint() (uint64, bool) {
	if n := p.get(); n != nil {
		return n.ToUint()
	}
	return 0, false
}
func (p Node) ToBool() (bool, bool) {
	if n := p.get(); n != nil {
		return n.ToBool()
	}
	return false, false
}

func (p Node) TypeError(want Type) error {
	return p.get().TypeError(want)
}

func (n Node) Values() IterV {
	return n.get().Values()
}

// Lookup finds a node by path
func (p Node) Lookup(path []string) Node {
	return p.With(p.Document().Lookup(p.id, path))
}
func (p Node) Type() Type {
	return p.get().Type()
}
func (p Node) Bytes() []byte {
	if n := p.get(); n != nil {
		return n.Bytes()
	}
	return nil
}

func (p Node) ToInterface() (interface{}, bool) {
	return p.Document().ToInterface(p.id)
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

func (p Node) PrintJSON(w io.Writer) (n int, err error) {
	return PrintJSON(w, p)
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

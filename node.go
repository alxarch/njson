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

// Unmarshaler is the interface implemented by types that can unmarshal from a Node.
type Unmarshaler interface {
	UnmarshalNodeJSON(n Node) error
}

// ID returns a node's id.
func (n Node) ID() uint {
	return n.id
}

// With returns a document node for id.
func (n Node) With(id uint) Node {
	n.id = id
	return n
}

// Document returns a node's document.
func (n Node) Document() *Document {
	if n.doc != nil && n.doc.rev == n.rev {
		return n.doc
	}
	return nil
}

func (n Node) get() *N {
	return n.Document().get(n.id)
}

// AppendJSON appends a node's JSON data to a byte slice.
func (n Node) AppendJSON(dst []byte) ([]byte, error) {
	return n.Document().AppendJSON(dst, n.id)
}

// Raw return the JSON string of a Node's value.
func (n Node) Raw() string { return n.get().Raw() }

// Unescaped unescapes a String Node's value.
func (n Node) Unescaped() string { return n.get().Unescaped() }

// ToFloat converts a node's value to float64.
func (n Node) ToFloat() (float64, bool) { return n.get().ToFloat() }

// ToInt converts a node's value to int64.
func (n Node) ToInt() (int64, bool) { return n.get().ToInt() }

// ToUint converts a node's  value to uint64.
func (n Node) ToUint() (uint64, bool) { return n.get().ToUint() }

// ToBool converts a Node to bool.
func (n Node) ToBool() (bool, bool) { return n.get().ToBool() }

// Type returnsa a Node's type.
func (n Node) Type() Type { return n.get().Type() }

// Bytes returns a Node's JSON string as bytes.
// The slice is NOT a copy of the string's data and SHOULD not be modified.
func (n Node) Bytes() []byte { return n.get().Bytes() }

// Values returns a value iterator over an Array or Object values.
func (n Node) Values() IterV { return n.get().Values() }

// TypeError returns an error for a type not matching a Node's type.
func (n Node) TypeError(want Type) error {
	return n.get().TypeError(want)
}

// Lookup finds a node by path
func (n Node) Lookup(path ...string) Node {
	return n.With(n.Document().Lookup(n.id, path))
}

// ToInterface converts a Node to a generic interface{}.
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

// PrintJSON writes JSON to an io.Writer.
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

// Get gets a Node by key.
func (n Node) Get(key string) Node {
	return n.With(n.get().Get(key))
}

// Index gets the Node at offset i of an Array.
// If the index is out of bounds the returned node's id
// will be MaxID and the Node will behave as empty.
func (n Node) Index(i int) Node {
	return n.With(n.get().Index(i))
}

// Set assigns a Node to the key of an Object Node.
// Since most keys need no escaping it doesn't escape the key.
// If the key needs escaping use strjson.Escaped.
func (n Node) Set(key string, value Node) { n.get().Set(key, value.ID()) }

// Slice reslices an Array node.
func (n Node) Slice(i, j int) { n.get().Slice(i, j) }

// Replace replaces the value at offset i of an Array node.
func (n Node) Replace(i int, r Node) { n.get().Replace(i, r.ID()) }

// Remove removes the value at offset i of an Array or Object node.
func (n *Node) Remove(i int) { n.get().Remove(i) }

// Del finds a key in an Object node's values and removes it.
func (n Node) Del(key string) { n.get().Del(key) }

// SetInt sets a Node's value to an integer.
func (n Node) SetInt(i int64) { n.get().SetInt(i) }

// SetUint sets a Node's value to an unsigned integer.
func (n Node) SetUint(u uint64) { n.get().SetUint(u) }

// SetFloat sets a Node's value to a float number.
func (n Node) SetFloat(f float64) { n.get().SetFloat(f) }

// SetString sets a Node's value to a string escaping invalid JSON characters.
func (n Node) SetString(s string) { n.get().SetString(s) }

// SetStringHTML sets a Node's value to a string escaping invalid JSON and unsafe HTML characters.
func (n Node) SetStringHTML(s string) { n.get().SetStringHTML(s) }

// SetStringRaw sets a Node's value to a string without escaping.
// Unless the provided string is guaranteed to not contain any JSON invalid characters,
// JSON output from this Node will be invalid.
func (n Node) SetStringRaw(s string) { n.get().SetStringRaw(s) }

// SetFalse sets a Node's value to false.
func (n Node) SetFalse() { n.get().SetFalse() }

// SetTrue sets a Node's value to true.
func (n Node) SetTrue() { n.get().SetTrue() }

// SetNull sets a Node's value to null.
func (n Node) SetNull() { n.get().SetNull() }

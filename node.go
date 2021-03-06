package njson

import (
	"encoding"
	"encoding/json"
	"io"
	"math"
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
	// Unlink invalid Document reference
	n.doc = nil
	return nil
}

func (n Node) get() *node {
	if n.doc != nil && n.doc.rev == n.rev {
		if n.id < uint(len(n.doc.nodes)) {
			return &n.doc.nodes[n.id]
		}
	} else {
		// Unlink invalid Document reference
		n.doc = nil
	}
	return nil
}

// AppendJSON appends a node's JSON data to a byte slice.
func (n Node) AppendJSON(dst []byte) ([]byte, error) {
	if nn := n.get(); nn != nil {
		return n.doc.appendJSON(dst, nn)
	}
	return nil, &typeError{TypeInvalid, TypeAnyValue}
}

// Raw returns the JSON string of a Node's value.
// Object and Array nodes return an empty string.
func (n Node) Raw() string {
	if n := n.get(); n != nil {
		return n.raw
	}
	return ""
}

// Unescaped unescapes the value of a String Node.
func (n Node) Unescaped() string {
	if n := n.get(); n != nil && n.info.IsString() {
		return strjson.Unescaped(n.raw)
	}
	return ""
}

// Data returns a node's raw string and type
func (n Node) Data() (string, Type) {
	if n := n.get(); n != nil {
		return n.raw, n.info.Type()
	}
	return "", TypeInvalid
}

// ToFloat converts a node's value to float64.
func (n Node) ToFloat() (float64, bool) {
	if n := n.get(); n != nil {
		f := numjson.ParseFloat(n.raw)
		return f, f == f
	}
	return 0, false
}

// ToInt converts a node's value to int64.
func (n Node) ToInt() (int64, bool) {
	if n := n.get(); n != nil {
		f := numjson.ParseFloat(n.raw)
		return int64(f), math.MinInt64 <= f && f < math.MaxInt64 && math.Trunc(f) == f
	}
	return 0, false
}

// ToUint converts a node's  value to uint64.
func (n Node) ToUint() (uint64, bool) {
	if n := n.get(); n != nil {
		f := numjson.ParseFloat(n.raw)
		return uint64(f), 0 <= f && f < math.MaxUint64 && math.Trunc(f) == f
	}
	return 0, false
}

// ToBool converts a Node to bool.
func (n Node) ToBool() (bool, bool) {
	if n := n.get(); n != nil && n.info.IsBoolean() {
		switch n.raw {
		case strTrue:
			return true, true
		case strFalse:
			return false, true
		}
	}
	return false, false
}

// Type returnsa a Node's type.
func (n Node) Type() Type {
	if n := n.get(); n != nil {
		return n.info.Type()
	}
	return TypeInvalid
}

// Values returns a value iterator over an Array or Object values.
func (n Node) Values() IterV {
	if node := n.get(); node != nil {
		return IterV{node: n, values: node.values}
	}
	return IterV{}
}

// TypeError returns an error for a type not matching a Node's type.
func (n Node) TypeError(want Type) error {
	return typeError{n.Type(), want}
}

// Lookup finds a node by path
func (n Node) Lookup(path ...string) Node {
	return n.With(n.Document().lookup(n.id, path))
}

// ToInterface converts a Node to a generic interface{}.
func (n Node) ToInterface() (interface{}, bool) {
	return n.Document().toInterface(n.id)
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
		return typeError{TypeInvalid, TypeAnyValue}
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
	case TypeString:
		if node.raw == "" {
			return u.UnmarshalJSON([]byte{delimString, delimString})
		}
	case TypeInvalid:
		return typeError{TypeInvalid, TypeAnyValue}
	}
	data := bufferpool.Get().([]byte)
	data, err = n.AppendJSON(data[:0])
	if err == nil {
		err = u.UnmarshalJSON(data)
	}
	bufferpool.Put(data)
	return
}

// WrapUnmarshalText wraps a call to the encoding.TextUnmarshaler interface
func (n Node) WrapUnmarshalText(u encoding.TextUnmarshaler) (err error) {
	if node := n.get(); node != nil {
		switch t := node.info.Type(); t {
		case TypeString:
			buf := bufferpool.Get().([]byte)
			buf = append(buf[:0], node.raw...)
			err = u.UnmarshalText(buf)
			bufferpool.Put(buf)
			return
		default:
			return newTypeError(t, TypeString)
		}
	}
	return newTypeError(TypeInvalid, TypeString)
}

// Get gets a Node by key.
// If the key is not found the returned node's id
// will be MaxID and the Node will behave as empty.
func (n Node) Get(key string) Node {
	if nn := n.get(); nn != nil && nn.info.IsObject() {
		for i := range nn.values {
			if key == nn.values[i].key {
				n.id = nn.values[i].id
				return n
			}
		}
	}
	n.id = maxUint
	return n
}

// Index gets the Node at offset i of an Array.
// If the index is out of bounds it sets the id to MaxUint
// and the Node will behave as empty but still have a usable
// reference to Document.
func (n Node) Index(i int) Node {
	if nn := n.get(); nn != nil && nn.info.IsArray() && 0 <= i && i < len(nn.values) {
		n.id = nn.values[i].id
		return n
	}
	n.id = maxUint
	return n
}

// Set assigns a Node to the key of an Object Node.
// Since most keys need no escaping it doesn't escape the key.
// If the key needs escaping use strjson.Escaped.
func (n Node) Set(key string, value Node) {
	if nn := n.get(); nn != nil && nn.info.IsObject() {
		// Make a copy of the value if it's not Orphan to avoid recursion infinite loops.
		id := n.doc.copyOrAdopt(value.Document(), value.ID(), n.id)
		if id < maxUint {
			// copyOrAdopt might grow nodes array invalidating nn pointer
			nn = &n.doc.nodes[n.id]
			var v *V
			for i := range nn.values {
				v = &nn.values[i]
				if v.key == key {
					v.id = id
					return
				}
			}
			nn.values = append(nn.values, V{
				id:  id,
				key: key,
			})
		}
	}
}

// Append appends a node id to an Array node's values.
func (n Node) Append(values ...Node) {
	if len(values) == 0 {
		return
	}
	if nn := n.get(); nn != nil && nn.info.IsArray() {
		vv := nn.values
		for _, v := range values {
			vv = append(vv, V{
				id:  n.doc.copyOrAdopt(v.Document(), v.ID(), n.id),
				key: "",
			})
		}
		// copyOrAdopt might grow nodes array invalidating nn pointer
		n.doc.nodes[n.id].values = vv
	}
}

// TODO: Splice, Prepend

// Slice reslices an Array node.
func (n Node) Slice(i, j int) {
	if n := n.get(); n != nil && n.info.IsArray() && 0 <= i && i < j && j < len(n.values) {
		n.values = n.values[i:j]
	}
}

// Replace replaces the value at offset i of an Array node.
func (n Node) Replace(i int, value Node) {
	if nn := n.get(); nn != nil && nn.info.IsArray() && 0 <= i && i < len(nn.values) {
		// Make a copy of the value if it's not Orphan to avoid recursion infinite loops.
		id := n.doc.copyOrAdopt(value.Document(), value.ID(), n.id)
		if id < maxUint {
			// copyOrAdopt might grow nodes array invalidating nn pointer
			n.doc.nodes[n.id].values[i] = V{id, ""}
		}
	}
}

// Remove removes the value at offset i of an Array node.
func (n Node) Remove(i int) {
	if n := n.get(); n != nil && n.info.IsArray() && 0 <= i && i < len(n.values) {
		if j := i + 1; 0 <= j && j < len(n.values) {
			copy(n.values[i:], n.values[j:])
		}
		if j := len(n.values) - 1; 0 <= j && j < len(n.values) {
			n.values[j] = V{}
			n.values = n.values[:j]
		}
	}
}

// Strip recursively deletes a key from a node.
func (n Node) Strip(key string) {
	if nn := n.get(); nn != nil && nn.info.IsObject() {
		for i := range nn.values {
			v := &nn.values[i]
			if key == v.key {
				if j := len(nn.values) - 1; 0 <= j && j < len(nn.values) {
					nn.values[i] = nn.values[j]
					nn.values[j] = V{}
					nn.values = nn.values[:j]
					for j := i; 0 <= j && j < len(nn.values); j++ {
						n.With(nn.values[j].id).Strip(key)
					}
				}
				return
			}
			n.With(v.id).Strip(key)
		}
	}

}

// Del finds a key in an Object node's values and removes it.
// It does not keep the order of keys.
func (n Node) Del(key string) {
	if n := n.get(); n != nil && n.info.IsObject() {
		for i := range n.values {
			if n.values[i].key == key {
				if j := len(n.values) - 1; 0 <= j && j < len(n.values) {
					n.values[i] = n.values[j]
					n.values[j] = V{}
					n.values = n.values[:j]
				}
				return
			}
		}
	}
}

// SetInt sets a Node's value to an integer.
func (n Node) SetInt(i int64) {
	if n := n.get(); n != nil {
		n.reset(vNumber|n.info.Flags(), strconv.FormatInt(i, 10), n.values[:0])
	}

}

// SetUint sets a Node's value to an unsigned integer.
func (n Node) SetUint(u uint64) {
	if n := n.get(); n != nil {
		n.reset(vNumber|n.info.Flags(), strconv.FormatUint(u, 10), n.values[:0])
	}

}

// SetFloat sets a Node's value to a float number.
func (n Node) SetFloat(f float64) {
	if n := n.get(); n != nil {
		n.reset(vNumber|n.info.Flags(), numjson.FormatFloat(f, 64), n.values[:0])
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
// Unless the provided string is guaranteed to not contain any JSON invalid characters,
// JSON output from this Node will be invalid.
func (n Node) SetStringRaw(s string) {
	if n := n.get(); n != nil {
		n.reset(vString|n.info.Flags(), s, n.values[:0])
	}
}

// SetFalse sets a Node's value to false.
func (n Node) SetFalse() {
	if n := n.get(); n != nil {
		n.reset(vBoolean|n.info.Flags(), strFalse, n.values[:0])
	}
}

// SetTrue sets a Node's value to true.
func (n Node) SetTrue() {
	if n := n.get(); n != nil {
		n.reset(vBoolean|n.info.Flags(), strTrue, n.values[:0])
	}
}

// SetNull sets a Node's value to null.
func (n Node) SetNull() {
	if n := n.get(); n != nil {
		n.reset(vNull|n.info.Flags(), strNull, n.values[:0])
	}
}

// IterV is an iterator over a node's values.
type IterV struct {
	V
	index  int
	values []V
	node   Node
}

// Reset resets the iterator.
func (i *IterV) Reset() {
	i.index = 0
	i.id = maxUint
}

// Close closes the iterator unlinking the values slice.
func (i *IterV) Close() {
	*i = IterV{
		index: -1,
		V: V{
			id: maxUint,
		},
	}
}

// Next increments the iteration cursor and checks if the iterarion finished.
func (i *IterV) Next() bool {
	if 0 <= i.index && i.index < len(i.values) {
		i.V = i.values[i.index]
		i.index++
		return true
	}
	i.id = maxUint
	// Set index to -1 so every Next() returns false until Reset() is called.
	i.index = -1
	return false
}

// Len returns the length of the values.
func (i *IterV) Len() int {
	return len(i.values)
}

// Index returns the current iteration index.
// Before Next() is called for the first time it returns -1.
// After the iteration has finished it returns -2.
func (i *IterV) Index() int {
	return i.index - 1
}

// Value returns the node of the current iteration value.
func (i *IterV) Value() Node {
	return i.node.With(i.id)
}

package njson

import (
	"io"
	"sync"
)

// Node is a document node.
type Node struct {
	Token
	doc    *Document
	id     uint16
	next   uint16
	parent uint16
	value  uint16
}

// IsRoot checks whether a node is a root node.
func (n *Node) IsRoot() bool {
	return n != nil && (n.id == 0 || n.parent == MaxDocumentSize)
}

// IsDocumentRoot checks whether a node is the root of it's document.
func (n *Node) IsDocumentRoot() bool {
	return n != nil && n.id == 0
}

// IsValid checks if a node belongs to it's document
func (n *Node) IsValid() bool {
	return n != nil && n.doc != nil
}

// AppendJSON implements the Appender interface.
func (n *Node) AppendJSON(data []byte) ([]byte, error) {
	if n == nil {
		return data, nil
	}
	switch n.info.Type() {
	case TypeObject:
		data = append(data, delimBeginObject)
		for n = n.Value(); n != nil; n = n.Next() {
			data, _ = n.AppendJSON(data)
			if n.next != 0 {
				data = append(data, delimValueSeparator)
			}
		}
		data = append(data, delimEndObject)
	case TypeString:
		data = append(data, delimString)
		data = append(data, n.src...)
		data = append(data, delimString)
	case TypeKey:
		data = append(data, delimString)
		data = append(data, n.src...)
		data = append(data, delimString)
		data = append(data, delimNameSeparator)
		data, _ = n.Value().AppendJSON(data)
	case TypeArray:
		data = append(data, delimBeginArray)
		for n = n.Value(); n != nil; n = n.Next() {
			data, _ = n.AppendJSON(data)
			if n.next != 0 {
				data = append(data, delimValueSeparator)
			}
		}
		data = append(data, delimEndArray)
	default:
		data = append(data, n.src...)
	}
	return data, nil
}

// Prev returns the Node's previous sibling.
func (n *Node) Prev() (p *Node) {
	if p = n.Parent(); p == nil || p.value == n.id {
		return nil
	}
	for p = p.Value(); p != nil && p.next != n.id; p = p.Next() {
	}
	return
}

// Parent returns the parent node.
func (n *Node) Parent() *Node {
	if n.IsRoot() || n.doc == nil {
		return nil
	}
	return n.doc.Get(n.parent)
}

// Next returns the next sibling of a Node.
// If the Node is an object key it's the next key.
// If the Node is an array element it's the next element.
func (n *Node) Next() *Node {
	if n == nil || n.next == 0 || n.doc == nil {
		return nil
	}
	// Use GetCheck to avoid document mismatch
	return n.doc.Get(n.next)
}

// Value returns a Node holding the value of a Node.
// This is the first key of an object Node, the first element
// of an array Node or the value of a key Node.
// For all other types it's nil.
func (n *Node) Value() *Node {
	if n == nil || n.value == 0 || n.doc == nil {
		return nil
	}
	return n.doc.Get(n.value)
}

// Index returns the i-th element of an Array node
func (n *Node) Index(i int) (v *Node) {
	if n.IsArray() && i >= 0 {
		for v = n.Value(); v != nil && i > 0; v, i = v.Next(), i-1 {
		}
	}
	return
}

// IndexKey returns the key Node of an object.
func (n *Node) IndexKey(key string) (v *Node) {
	if n.IsObject() {
		for v = v.Value(); v != nil; v = v.Next() {
			if v.info == ValueInfo(TypeKey) {
				if v.src == key {
					return
				}
			} else if v.Unescaped() == key {
				return
			}
		}
	}
	return nil
}

// // IndexKeyUnescaped returns the key Node of an object without unescaping.
// func (n *Node) IndexKeyUnescaped(key string) (v *Node) {
// 	if n.IsObject() {
// 		for v = n.Value(); v != nil; v = v.Next() {
// 			if v.src == key {
// 				return
// 			}
// 		}
// 	}
// 	return nil
// }

// ToInterface converts a node to any combatible go value (many allocations on large trees).
func (n *Node) ToInterface() (interface{}, bool) {
	switch n.Type() {
	case TypeObject:
		m := make(map[string]interface{}, n.Len())
		ok := false
		for n = n.Value(); n != nil; n = n.Next() {
			if m[n.Unescaped()], ok = n.Value().ToInterface(); !ok {
				return nil, false
			}
		}
		return m, true
	case TypeArray:
		s := make([]interface{}, n.Len())
		j := 0
		ok := false
		for n = n.Value(); n != nil; n, j = n.Next(), j+1 {
			if s[j], ok = n.Value().ToInterface(); !ok {
				return nil, false
			}
		}
		return s, true
	case TypeString, TypeKey:
		return n.Unescaped(), true
	case TypeBoolean:
		switch n.info {
		case ValueTrue:
			return true, true
		case ValueFalse:
			return false, true
		default:
			return nil, false
		}
	case TypeNull:
		return nil, true
	case TypeNumber:
		return n.ToFloat()
	default:
		return nil, false
	}

}

// IsObject checks if a Node is a JSON Object
func (n *Node) IsObject() bool {
	return n != nil && n.info == ValueInfo(TypeObject)
}

// IsArray checks if a Node is a JSON Array
func (n *Node) IsArray() bool {
	return n != nil && n.info == ValueInfo(TypeArray)
}

// IsNull checks if a Node is JSON Null
func (n *Node) IsNull() bool {
	return n != nil && n.info == ValueInfo(TypeNull)
}

// IsKey checks if a Node is a JSON Object's key
func (n *Node) IsKey() bool {
	return n != nil && n.info&(ValueInfo(TypeKey)) != 0
}
func (n *Node) IsString() bool {
	return n != nil && n.info&(ValueInfo(TypeString)) != 0
}

// IsValue checks if a Node is any JSON value (ie String, Boolean, Number, Array, Object, Null)
func (n *Node) IsValue() bool {
	return n != nil && n.info&ValueInfo(TypeAnyValue) != 0
}

// TypeError creates an error for the Node's type.
func (n *Node) TypeError(want Type) error {
	return TypeError(n.Type(), want)
}

// PrintJSON implements the Printer interface
func (n *Node) PrintJSON(w io.Writer) (int, error) {
	return PrintJSON(w, n)
}

// Appender is a Marshaler interface for buffer append workflows.
type Appender interface {
	AppendJSON([]byte) ([]byte, error)
}

var bufferpool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 4096)
	},
}

// PrintJSON is a helper to write an Appender to an io.Writer
func PrintJSON(w io.Writer, a Appender) (n int, err error) {
	b := bufferpool.Get().([]byte)
	if b, err = a.AppendJSON(b); err == nil {
		n, err = w.Write(b)
	}
	bufferpool.Put(b[:0])
	return
}

// Printer is a Marshaler interface for io.Writer workflows.
type Printer interface {
	PrintJSON(w io.Writer) (int, error)
}

// Unmarshaler unmarshals from a Node
type Unmarshaler interface {
	UnmarshalNodeJSON(*Node) error
}

// Len gets the number of children a node has.
func (n *Node) Len() (i int) {
	if n != nil && n.info&ValueInfo(TypeSized) != 0 {
		for n = n.Value(); n != nil; n, i = n.Next(), i+1 {
		}
	}
	return
}

// UnescapedBytes returns a byte slice of the unescaped form of Node
func (n *Node) UnescapedBytes() []byte {
	if n == nil {
		return nil
	}
	if !n.info.NeedsEscape() {
		return s2b(n.src)
	}
	if n.doc != nil {
		if 0 < n.extra && n.extra < n.doc.n {
			return s2b(n.doc.nodes[n.extra].src)
		}
		b := make([]byte, len(n.src))
		b = b[:Unescape(b, n.src)]
		n.extra = n.doc.add(Token{
			src: string(b),
		})
		return b
	}
	return AppendEscaped(nil, n.src)
}

func (n *Node) unescaped() string {
	if 0 < n.extra && n.extra < n.doc.n {
		return n.doc.nodes[n.extra].src
	}
	b := make([]byte, len(n.src))
	b = b[:Unescape(b, n.src)]
	s := string(b)
	n.extra = n.doc.add(Token{
		src: s,
	})
	return s
}

// Unescaped returns the unescaped string form of the Node
func (n *Node) Unescaped() string {
	if n == nil {
		return ""
	}
	if !n.info.NeedsEscape() {
		return n.src
	}
	if n.doc != nil {
		return n.unescaped()
	}
	return string(AppendUnescaped(nil, n.src))
}

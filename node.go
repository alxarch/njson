package njson

import "sync"

type Node struct {
	Token
	doc    *Document
	id     uint16
	next   uint16
	parent uint16
	value  uint16
}

func (n *Node) IsRoot() bool {
	return n.id == 0
}

var bufferpool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func (n *Node) AppendTo(data []byte) []byte {
	if n == nil {
		return data
	}
	switch n.info.Type() {
	case TypeObject:
		data = append(data, delimBeginObject)
		for n = n.Value(); n != nil; n = n.Next() {
			data = n.AppendTo(data)
			if n.next != 0 {
				data = append(data, delimValueSeparator)
			}
		}
		data = append(data, delimEndObject)
	case TypeKey:
		data = append(data, n.src...)
		data = append(data, delimNameSeparator)
		data = n.Value().AppendTo(data)
	case TypeArray:
		data = append(data, delimBeginArray)
		for n = n.Value(); n != nil; n = n.Next() {
			data = n.AppendTo(data)
			if n.next != 0 {
				data = append(data, delimValueSeparator)
			}
		}
		data = append(data, delimEndArray)
	default:
		data = append(data, n.src...)
	}
	return data
}

func (n *Node) Prev() (p *Node) {
	if p = p.Parent(); p == nil || p.value == n.id {
		return nil
	}
	for p = p.Value(); p != nil; p = p.Next() {
		if p.next == n.id {
			return p
		}
	}
	return nil
}

func (n *Node) Parent() *Node {
	if n.parent == MaxDocumentSize {
		return nil
	}
	return n.doc.Get(n.parent)
}

// Next returns the next sibling of a Node.
// If the Node is an object key it's the next key.
// If the Node is an array element it's the next element.
func (n *Node) Next() *Node {
	return n.doc.get(n.next)
}

// Value returns a Node holding the value of a Node.
// This is the first key of an object Node, the first element
// of an array Node or the value of a key Node.
// For all other types it's nil.
func (n *Node) Value() *Node {
	return n.doc.get(n.value)
}

func (n *Node) Index(i uint16) (v *Node) {
	if n.Type() == TypeArray && i < n.size {
		for v = n.Value(); v != nil && i > 0; v, i = v.Next(), i-1 {
		}
		return
	}
	return nil
}

func (n *Node) FindKey(key string) (v *Node) {
	if n.Type() == TypeObject {
		for v = v.Value(); v != nil; v = v.Next() {
			if v.unquote() == key {
				return
			}
		}
	}
	return nil
}

func (n *Node) FindKeyJSON(key string) (v *Node) {
	if n.Type() == TypeObject {
		for v = n.Value(); v != nil; v = v.Next() {
			if v.ToJSON() == key {
				return
			}
		}
	}
	return nil
}

func (n *Node) ToInterface() (interface{}, bool) {
	switch n.Type() {
	case TypeObject:
		m := make(map[string]interface{}, n.size)
		ok := false
		for n = n.Value(); n != nil; n = n.Next() {
			if m[n.Unescaped()], ok = n.Value().ToInterface(); !ok {
				return nil, false
			}
		}
		return m, true
	case TypeArray:
		s := make([]interface{}, n.size)
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
		switch n.src {
		case strTrue:
			return true, true
		case strFalse:
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

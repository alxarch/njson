package njson

import "github.com/alxarch/njson/strjson"

type Object Node

func (o Object) Node() Node {
	return Node(o)
}

func (o Object) Document() *Document {
	return o.doc
}

// Get gets a value Node by key.
// If the key is not found the returned node is zero.
func (o Object) Get(key string) Node {
	n := Node(o)
	if v := n.value(); v != nil && v.typ == TypeObject {
		if c := v.get(key); c != nil {
			return n.with(c.id)
		}
	}
	return Node{}
}

// Set assigns a Node to the key of an Object Node.
func (o Object) Set(key string, value Node) Node {
	n := Node(o)
	if v := n.value(); v != nil && v.typ == TypeObject {
		// Make a copy of the value if it's not Orphan to avoid cyclic references and infinite loops.
		id, ok := n.doc.copyNode(value, n.id)
		if !ok {
			return Node{}
		}
		// copyNode might grow values array invalidating value pointer
		// so we need to 'refresh' the value
		v = &n.doc.values[n.id]
		if c := v.get(key); c != nil {
			c.id = id
			return n.with(id)
		}
		v.children = append(v.children, child{
			id:  id,
			key: key,
		})
		return n.with(id)
	}
	return Node{}
}
func (o Object) Each(fn func(key string, value Node) bool) {
	n := Node(o)
	if v := n.value(); v != nil && v.typ == TypeObject {
		for i := range v.children {
			c := &v.children[i]
			if !fn(c.key, n.with(c.id)) {
				return
			}
		}
	}
}

// Strip recursively deletes a key from a node.
func (o Object) Strip(key string) (total int) {
	n := Node(o)
	if v := n.value(); v != nil && v.typ == TypeObject {
		if _, ok := v.del(key); ok {
			total++
		}
		for i := range v.children {
			c := &v.children[i]
			total += Object(n.with(c.id)).Strip(key)
		}
	}
	return
}

// Del finds a key in an Object node's values and removes it.
// It does not keep the order of keys.
func (o Object) Del(key string) Node {
	n := Node(o)
	if v := n.value(); v != nil && v.typ == TypeObject {
		if id, ok := v.del(key); ok {
			el := n.with(id)
			if v := el.value(); v != nil {
				v.flags |= flagRoot
				return el
			}
		}
	}
	return Node{}
}

// Len return the number of keys in the object.
func (o Object) Len() int {
	n := Node(o)
	if v := n.value(); v != nil && v.typ == TypeObject {
		return len(v.children)
	}
	return -1
}
func (o Object) Iter() ObjectIterator {
	if v := Node(o).value(); v != nil {
		return ObjectIterator{
			key:  "",
			node: o.doc.Node(maxUint),
			iter: v.Iter(),
		}
	}
	return ObjectIterator{}
}

type ObjectIterator struct {
	key  string
	node Node
	iter iterator
}

func (i *ObjectIterator) Key() string {
	return i.key
}
func (i *ObjectIterator) Node() Node {
	return i.node
}

func (i *ObjectIterator) Next() bool {
	if next := (child{}); i.iter.Next(&next) {
		i.key = next.key
		i.node.id = next.id
		return true
	}
	return false
}
func (i *ObjectIterator) Close() {
	i.node = Node{}
	i.iter.Done()
}

func (v *value) del(key string) (uint, bool) {
	if !v.unlocked() {
		return 0, false
	}
	if !strjson.Flags(v.flags).IsGoSafe() && strjson.NeedsEscape(key) {
		v.unescapeKeys()
	}
	if i := len(v.children) - 1; 0 <= i && i < len(v.children) {
		children, last := v.children[:i], &v.children[i]
		if last.key == key {
			v.children = children
			return last.id, true
		}
		for i := range children {
			if c := &children[i]; c.key == key {
				id := c.id
				*c = *last
				v.children = children
				return id, true
			}
		}
	}
	return 0, false
}

func (v *value) get(key string) *child {
	if !strjson.Flags(v.flags).IsGoSafe() {
		for i := uint(0); i < uint(len(key)); i++ {
			if strjson.NeedsEscapeByte(key[i]) {
				v.unescapeKeys()
				break
			}
		}
	}
	for i := range v.children {
		if c := &v.children[i]; c.key == key {
			return c
		}
	}
	return nil
}
func (v *value) index(key string) *child {
	switch len(key) {
	case 1:
		// Fast path for small indexes
		if len(key) == 1 {
			if i := uint(key[0] - '0'); i < uint(len(v.children)) && i <= 9 {
				return &v.children[i]
			}
		}
		return nil
	case 0:
		return nil
	default:
		// We inline index parsing to avoid extra function call
		var index, i, digit uint
		// stop parsing index early
		cutoff := uint(len(v.children))
		for i = uint(0); i < uint(len(key)); i++ {
			// byte will roll on underflow
			digit = uint(key[i] - '0')
			if digit <= 9 && index < cutoff {
				index = index*10 + digit
			} else {
				return nil
			}
		}
		if index < uint(len(v.children)) {
			return &v.children[index]
		}
		return nil
	}
}

func (v *value) unescapeKeys() {
	for i := range v.children {
		c := &v.children[i]
		c.key = strjson.Unescaped(c.key)
	}
	v.flags = 0
}

package njson

import (
	"errors"
	"fmt"
)

type Object Node

func (o Object) Node() Node {
	return Node(o)
}

// Get gets a value Node by key.
// If the key is not found the returned node is zero.
func (o Object) Get(key string) Node {
	const opName = "Get"
	n := Node(o)
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeObject, opName, n)
	}
	if v.typ != TypeObject {
		return errNodeType(TypeObject, opName, v.typ)
	}
	if c := v.get(key); c != nil {
		return n.with(c.id)
	}
	return errNode(&KeyError{Key: key})
}

// Set assigns a Node to the key of an Object Node.
func (o Object) Set(key string, value Node) Node {
	const opName = "Set"
	n := Node(o)
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeObject, opName, n)
	}
	if v.typ != TypeObject {
		return errNodeType(TypeObject, opName, v.typ)
	}
	if v.locked() {
		return errNodeLocked(TypeObject, opName)
	}
	// Make a copy of the value if it's not Orphan to avoid cyclic references and infinite loops.
	id, ok := n.doc.copyNode(value, n.id)
	if !ok {
		return errNodeInvalid(TypeAnyValue, opName, value)
	}
	// copyNode might alloc values array invalidating value pointer
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

func (o Object) Pop() (string, Node) {
	n := Node(o)
	if v := n.value(); v != nil && v.typ == TypeObject {
		var lastChild child
		if last := len(v.children) - 1; 0 <= last && last < len(v.children) {
			lastChild, v.children = v.children[last], v.children[:last]
			return lastChild.key, n.with(lastChild.id)
		}
	}
	return "", Node{}
}
func (o Object) Clear() Object {
	n := Node(o)
	v := n.value()
	if v == nil {
		return o
	}
	if v.typ != TypeObject {
		return errNode(newTypeError(v.typ, TypeObject)).Object()
	}
	if v.locked() {
		return errNode(ErrObjectLocked).Object()
	}
	for i := range v.children {
		c := &v.children[i]
		n.doc.get(c.id).flags |= flagRoot
		*c = child{}
	}
	v.children = v.children[:0]
	return o
}

// Strip recursively deletes a key from an object node.
func (o Object) Strip(key string) (int, error) {
	n := Node(o)
	if v := n.value(); v != nil && v.typ == TypeObject {
		locked := maxUint
		total, ok := n.strip(key, &locked)
		if !ok {
			return total, &ImmutableNodeError{Node: n.with(locked)}
		}
		return total, nil
	}
	return 0, n.TypeError(TypeObject)
}

func (n Node) strip(key string, locked *uint) (int, bool) {
	v := n.value()
	if v == nil {
		return 0, false
	}
	total := 0
	switch v.typ {
	case TypeObject:
		if v.locked() {
			*locked = n.id
			return 0, false
		}
		index := -1
		for i := range v.children {
			c := &v.children[i]
			if c.key == key {
				index = i
				continue
			}
			n := n.with(c.id)
			d, ok := n.strip(key, locked)
			if !ok {
				return 0, false
			}
			total += d
		}
		if id, ok := v.remove(index); ok {
			n.doc.values[id].flags |= flagRoot
			total++
		}
		return total, true
	case TypeArray:
		if v.locked() {
			*locked = n.id
			return 0, false
		}
		for i := range v.children {
			c := &v.children[i]
			n := n.with(c.id)
			d, ok := n.strip(key, locked)
			if !ok {
				return 0, false
			}
			total += d
		}
		return total, true
	default:
		return 0, true
	}
}

type ImmutableNodeError struct {
	Node Node
}

func (e *ImmutableNodeError) Error() string {
	return fmt.Sprintf("%s node is not mutable", e.Node.Type())
}

var ErrObjectLocked = errors.New("cannot modify an object during an iteration")

// Del finds a key in an Object node's values and removes it
// If stable is true, the operation will preserver key order.
func (o Object) Del(key string, stable bool) Node {
	const opName = "Del"
	n := o.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeObject, opName, n)
	}
	if v.typ != TypeObject {
		return errNodeType(TypeObject, opName, v.typ)
	}
	if v.locked() {
		return errNodeLocked(TypeObject, opName)
	}
	if id, ok := v.del(key, stable); ok {
		el := n.with(id)
		if v := el.value(); v != nil {
			v.flags |= flagRoot
		}
		return el
	}
	return errNode(fmt.Errorf("key %q not found in %s", key, TypeObject))
}

// Len return the number of keys in the object.
func (o Object) Len() int {
	n := Node(o)
	if v := n.value(); v != nil && v.typ == TypeObject {
		return len(v.children)
	}
	return -1
}

func (o Object) Iterate() ObjectIterator {
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

func (i *ObjectIterator) Len() int {
	return i.iter.Len()
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

func (o Object) IsMutable() bool {
	if v := Node(o).value(); v != nil {
		return v.unlocked()
	}
	return false
}

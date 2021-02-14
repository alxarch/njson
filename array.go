package njson

import (
	"errors"
	"fmt"
	"sort"
)

type Array Node

func (a Array) Len() int {
	if v := Node(a).value(); v != nil && v.typ == TypeArray {
		return len(v.children)
	}
	return -1
}

func (a Array) Node() Node {
	return Node(a)
}

func (a Array) IsMutable() bool {
	if v := Node(a).value(); v != nil {
		return v.unlocked()
	}
	return false
}

func (a Array) Iterate() ArrayIterator {
	if v := Node(a).value(); v != nil {
		return ArrayIterator{
			// initialize node to invalid
			node: a.doc.Node(maxUint),
			iter: v.Iter(),
		}
	}
	return ArrayIterator{}
}

type ArrayIterator struct {
	node Node
	iter iterator
}

func (i *ArrayIterator) Node() Node {
	return i.node
}

func (i *ArrayIterator) Next() bool {
	next := child{}
	if i.iter.Next(&next) {
		i.node.id = next.id
		return true
	}
	i.Close()
	return false
}

func (i *ArrayIterator) Len() int {
	return i.iter.Len()
}

func (i *ArrayIterator) Close() {
	i.node = Node{}
	i.iter.Done()
}

// Get gets the Node at offset i of an Array.
// If the index is out of bounds it sets the id to MaxUint
// and the Node will behave as empty but still have a usable
// reference to Document.
func (a Array) Get(i int) Node {
	const opName = "Get"
	n := a.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeArray, opName, n)
	}
	if v.typ != TypeArray {
		return errNodeType(TypeArray, opName, v.typ)
	}
	if 0 <= i && i < len(v.children) {
		return a.Node().with(v.children[i].id)
	}
	return errNode(ErrInvalidIndex)
}

// Insert inserts a node at offset i of an Array node.
func (a Array) Set(i int, el Node) Node {
	const opName = "Set"
	n := a.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeArray, opName, n)
	}
	if v.typ != TypeArray {
		return errNodeType(TypeArray, opName, v.typ)
	}
	if v.locked() {
		return errNodeLocked(TypeArray, opName)
	}
	id, ok := a.doc.copyNode(el, a.id)
	if !ok {
		return errNodeInvalid(TypeAnyValue, opName, el)
	}
	if 0 <= i && i < len(v.children) {
		// copyNode might alloc values array invalidating v pointer
		a.doc.values[a.id].children[i].id = id
		return a.Node().with(id)
	}
	return errNode(ErrInvalidIndex)
}

// Append appends a Node to an Array node's values.
func (a Array) Push(el Node) Node {
	const opName = "Push"
	n := a.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeArray, opName, n)
	}
	if v.typ != TypeArray {
		return errNodeType(TypeArray, opName, v.typ)
	}
	if v.locked() {
		return errNodeLocked(TypeArray, opName)
	}
	id, ok := a.doc.copyNode(el, a.id)
	if !ok {
		return errNodeInvalid(TypeAnyValue, opName, el)
	}
	// copyNode might alloc values array invalidating v pointer
	v = &a.doc.values[a.id]
	v.children = append(v.children, child{
		id: id,
	})
	return n.with(id)
}

func (a Array) Pop() Node {
	const opName = "Pop"
	n := a.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeArray, opName, n)
	}
	if v.typ != TypeArray {
		return errNodeType(TypeArray, opName, v.typ)
	}
	if v.locked() {
		return errNodeLocked(TypeArray, opName)
	}
	if i := len(v.children) - 1; 0 <= i && i < len(v.children) {
		var id uint
		id, v.children = v.children[i].id, v.children[:i]
		el := n.with(id)
		if v := el.value(); v != nil {
			v.flags &= flagRoot
		}
		return el
	}
	return errNode(ErrInvalidIndex)
}

// Replace replaces the value at offset i of an Array node.
func (a Array) Replace(i int, el Node) Node {
	const opName = "Replace"
	n := a.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeArray, opName, n)
	}
	if v.typ != TypeArray {
		return errNodeType(TypeArray, opName, v.typ)
	}
	if v.locked() {
		return errNodeLocked(TypeArray, opName)
	}
	// Make a copy of the value if it's not Orphan to avoid recursion infinite loops.
	id, ok := a.doc.copyNode(el, a.id)
	if !ok {
		return errNodeInvalid(TypeAnyValue, opName, el)
	}
	if 0 <= i && i < len(v.children) {
		v.children[i] = child{id, ""}
		return n.with(id)
	}
	return errNode(ErrInvalidIndex)
}

var ErrInvalidIndex = errors.New("index out of bounds")

// Remove removes the value at offset i of an Array node.
func (a Array) Remove(i int) Node {
	const opName = "Remove"
	n := a.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeArray, opName, n)
	}
	if v.typ != TypeArray {
		return errNodeType(TypeArray, opName, v.typ)
	}
	if v.locked() {
		return errNodeLocked(TypeArray, opName)
	}
	if id, ok := v.remove(i); ok {
		// Mark node as root since it's removed from it's parent
		a.doc.values[id].flags |= flagRoot
		return n.with(id)
	}
	if len(v.children) == 0 {
		return errNode(fmt.Errorf("called %s.%s() on an empty %s", TypeArray, opName, TypeArray))
	}
	return errNode(fmt.Errorf("called %s.%s(i) with invalid i, 0 <= i < %d", TypeArray, opName, len(v.children)))
}

// Insert inserts a node at offset i of an Array node.
func (a Array) Insert(i int, el Node) Node {
	const opName = "Insert"
	n := a.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeArray, opName, n)
	}
	if v.typ != TypeArray {
		return errNodeType(TypeArray, opName, v.typ)
	}
	if v.locked() {
		return errNodeLocked(TypeArray, opName)
	}

	id, ok := a.doc.copyNode(el, a.id)
	if !ok {
		return errNodeInvalid(TypeAnyValue, opName, el)
	}
	// copyNode might alloc values array invalidating v pointer
	v = &a.doc.values[a.id]
	if 0 <= i && i < len(v.children) {
		children := append(v.children, child{})
		copy(children[i+1:], children[i:])
		children[i].id = id
		v.children = children
		return Node(a).with(id)
	}
	return errNode(ErrInvalidIndex)
}

// Sort sorts an Array using a callback.
func (a Array) Sort(less func(a, b Node) bool) Array {
	const opName = "Sort"
	n := a.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeArray, opName, n).Array()
	}
	if v.typ != TypeArray {
		return errNodeType(TypeArray, opName, v.typ).Array()
	}
	if v.locked() {
		return errNodeLocked(TypeArray, opName).Array()
	}
	sort.Slice(v.children, func(i, j int) bool {
		return less(n.with(v.children[i].id), n.with(v.children[j].id))
	})
	return a
}

func (a Array) Clear() Array {
	const opName = "Clear"
	n := a.Node()
	v := n.value()
	if v == nil {
		return errNodeInvalid(TypeArray, opName, n).Array()
	}
	if v.typ != TypeArray {
		return errNodeType(TypeArray, opName, v.typ).Array()
	}
	if v.locked() {
		return errNodeLocked(TypeArray, opName).Array()
	}

	for i := range v.children {
		c := &v.children[i]
		if v := a.doc.get(c.id); v != nil {
			v.flags |= flagRoot
		}
		*c = child{}
	}
	v.children = v.children[:0]
	return a
}

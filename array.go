package njson

import (
	"math"
	"sort"
	"unsafe"
)

type Array Node

func (a Array) Len() int {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray {
		return len(v.children)
	}
	return -1
}
func (a Array) Node() Node {
	return Node(a)
}

func (a Array) Document() *Document {
	return a.doc
}

func (a Array) Iter() ArrayIterator {
	if v := Node(a).value(); v != nil {
		return ArrayIterator{
			node: a.doc.Node(maxUint),
			iter: v.Iter(),
		}
	}
	return ArrayIterator{}
}

type ArrayIterator struct {
	node Node
	iter childIter
}

func (i *ArrayIterator) Node() Node {
	return i.node
}
func (i *ArrayIterator) Next() bool {
	if next := (child{}); i.iter.Next(&next) {
		i.node.id = next.id
		return true
	}
	i.Close()
	return false
}

func (i *ArrayIterator) Close() {
	i.node = Node{}
	i.iter.Done()
}

type childIter struct {
	// A pointer to the next iterator element
	cur   *child
	end   uintptr
	count *uint16
}

const sizeOfChild = unsafe.Sizeof(child{})

var staticChild = &child{}

func (v *value) Iter() childIter {
	switch v.typ {
	case TypeArray, TypeObject:
		if v.locks == math.MaxUint16 {
			panic("too many iterators")
		}
		v.locks++
		cur := staticChild
		offset := uintptr(0)
		if len(v.children) > 0 {
			cur, offset = &v.children[0], uintptr(len(v.children))*sizeOfChild
		}
		return childIter{
			cur:   cur,
			end:   uintptr(unsafe.Pointer(cur)) + offset,
			count: &v.locks,
		}
	default:
		return childIter{}
	}
}

func (i *childIter) Next(p *child) bool {
	if i.cur == nil {
		return false
	}
	next := uintptr(unsafe.Pointer(i.cur)) + sizeOfChild
	if next <= i.end {
		*p = *i.cur
		i.cur = (*child)(unsafe.Pointer(next))
		return true
	}
	i.Done()
	return false
}

func (i *childIter) Done() {
	if i.count != nil {
		// unlock mutations
		*i.count--
	}
	// clear references
	*i = childIter{}
}

func (a Array) Each(fn func(i int, el Node) bool) {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray {
		for i := 0; 0 <= i && i < len(v.children); i++ {
			id := v.children[i].id
			if !fn(i, n.with(id)) {
				return
			}
		}
	}
}

// Get gets the Node at offset i of an Array.
// If the index is out of bounds it sets the id to MaxUint
// and the Node will behave as empty but still have a usable
// reference to Document.
func (a Array) Get(i int) Node {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray && 0 <= i && i < len(v.children) {
		return n.with(v.children[i].id)
	}
	return Node{}
}

// Insert inserts a node at offset i of an Array node.
func (a Array) Set(i int, el Node) Node {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray && v.locks == 0 && 0 <= i && i < len(v.children) {
		id, ok := n.doc.copyNode(el, n.id)
		if !ok {
			return Node{}
		}
		// copyNode might grow values array invalidating v pointer
		n.doc.values[n.id].children[i].id = id
		return n.with(id)
	}
	return Node{}
}

// Append appends a Node to an Array node's values.
func (a Array) Push(element Node) Node {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray {
		children := v.children
		id, ok := n.doc.copyNode(element, n.id)
		if !ok {
			return Node{}
		}
		// copyNode might grow values array invalidating v pointer
		n.doc.values[n.id].children = append(children, child{
			id: id,
		})
		return n.with(id)
	}
	return Node{}
}

func (a Array) Pop() Node {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray {
		if i := len(v.children) - 1; 0 <= i && i < len(v.children) {
			var id uint
			id, v.children = v.children[i].id, v.children[:i]
			el := n.with(id)
			if v := el.value(); v != nil {
				v.flags &= flagRoot
				return el
			}
		}
	}
	return Node{}
}

// TODO: Splice, Prepend

// Slice reslices an Array node.
func (a Array) Slice(i, j int) {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray && 0 <= i && i < j && j < len(v.children) {
		v.children = v.children[i:j]
	}
}

// Replace replaces the value at offset i of an Array node.
func (a Array) Replace(i int, value Node) Node {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray && 0 <= i && i < len(v.children) {
		// Make a copy of the value if it's not Orphan to avoid recursion infinite loops.
		if id, ok := n.doc.copyNode(value, n.id); ok {
			// copyNode might grow values array invalidating v pointer
			n.doc.values[n.id].children[i] = child{id, ""}
			return n.with(id)
		}
	}
	return Node{}
}

// Remove removes the value at offset i of an Array node.
func (a Array) Remove(i int) Node {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray && 0 <= i && i < len(v.children) {
		id := v.children[i].id
		if j := i + 1; 0 <= j && j < len(v.children) {
			copy(v.children[i:], v.children[j:])
		}
		if j := len(v.children) - 1; 0 <= j && j < len(v.children) {
			v.children[j] = child{}
			v.children = v.children[:j]
		}
		// Mark node as root since it's removed from it's parent
		n.doc.values[id].flags |= flagRoot
		return n.with(id)
	}
	return Node{}
}

// Insert inserts a node at offset i of an Array node.
func (a Array) Insert(i int, el Node) Node {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray && 0 <= i && i < len(v.children) {
		children := make([]child, len(v.children)+1)
		copy(children, v.children[:i])
		id, ok := n.doc.copyNode(el, n.id)
		if !ok {
			return Node{}
		}
		children[i].id = id
		copy(children[i+1:], v.children[i:])
		// copyNode might grow values array invalidating v pointer
		n.doc.values[n.id].children = children
		return n.with(id)
	}
	return Node{}
}

// Sort sorts an Array using a callback.
func (a Array) Sort(less func(a, b Node) bool) {
	n := Node(a)
	if v := n.value(); v != nil && v.typ == TypeArray {
		sort.Slice(v.children, func(i, j int) bool {
			a, b := n.with(v.children[i].id), n.with(v.children[j].id)
			return less(a, b)
		})
	}
}

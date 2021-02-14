package njson

import (
	"unsafe"
)

// HERE BE DRAGONS
// We use pointer tricks for inlined fast iterators.
type iterator struct {
	// A pointer to the next iterator element
	cur *child
	end uintptr
	v   *value
}

const sizeOfChild = unsafe.Sizeof(child{})

var staticChild = &child{}

func (v *value) Iter() iterator {
	switch v.typ {
	case TypeArray, TypeObject:
		if !v.lock() {
			return iterator{}
		}
		// initialize to staticChild so that empty array iterators work
		cur := staticChild
		offset := uintptr(0)
		if len(v.children) > 0 {
			cur, offset = &v.children[0], uintptr(len(v.children))*sizeOfChild
		}
		return iterator{
			cur: cur,
			end: uintptr(unsafe.Pointer(cur)) + offset,
			v:   v,
		}
	default:
		return iterator{}
	}
}

// Next advances the iterator and assigns the next child to p
// It is inlined by the compiler for fast iterations
func (i *iterator) Next(p *child) bool {
	if i.cur == nil {
		return false
	}
	next := uintptr(unsafe.Pointer(i.cur)) + sizeOfChild
	if next <= i.end {
		*p = *i.cur
		//nolint:unsafeptr
		i.cur = (*child)(unsafe.Pointer(next))
		return true
	}
	i.Done()
	return false
}
func (i *iterator) Len() int {
	if i.v != nil {
		return len(i.v.children)
	}
	return -1
}

func (i *iterator) Done() {
	v := i.v
	// clear references
	*i = iterator{}
	// unlock mutations
	if v != nil {
		v.unlock()
	}
}

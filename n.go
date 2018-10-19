package njson

import (
	"github.com/alxarch/njson/strjson"
)

// node is a JSON document node.
// node's data is only valid until Document.Close() or Document.Reset().
type node struct {
	info   info
	raw    string
	values []V
}

// V is a node's value referencing it's id.
// Array node's values have empty string keys.
type V struct {
	id  uint
	key string
}

// Key returns a value's key
func (v *V) Key() string {
	return v.key
}

// ID returns a value's id
func (v *V) ID() uint {
	return v.id
}

// Unescaped unescapes the value of a String node.
// The returned string is safe to use even if ParseUnsafe was used.
func (n *node) Unescaped() string {
	if n != nil && n.info.IsString() {
		return strjson.Unescaped(n.Safe())
	}
	return ""
}

// Safe returns a safe copy of a node's value JSON string.
// Object and Array nodes return an empty string.
// The returned string is safe to use even if ParseUnsafe was used.
func (n *node) Safe() string {
	if n == nil || n.raw == "" {
		return ""
	}
	if n.info.IsSafe() {
		return n.raw
	}
	n.raw = scopy(n.raw)
	n.info &^= infUnsafe
	return n.raw
}

// Bytes returns a node's JSON string as bytes
func (n *node) Bytes() []byte {
	if n != nil {
		return s2b(n.raw)
	}
	return nil
}

const maxUint = ^(uint(0))

// Index gets the id of an Array node's values at position i
func (n *node) Index(i int) uint {
	if n != nil && n.info.IsArray() && 0 <= i && i < len(n.values) {
		v := &n.values[i]
		return v.id
	}
	return maxUint
}

// Get finds a key in an Object node's values and returns it's id.
func (n *node) Get(key string) uint {
	if n != nil && n.info.IsObject() {
		var v *V
		for i := range n.values {
			v = &n.values[i]
			if v.key == key {
				return v.id
			}
		}
	}
	return maxUint
}

func (n *node) Type() Type {
	if n != nil {
		return n.info.Type()
	}
	return TypeInvalid
}

// TypeError creates an error for the Node's type.
func (n *node) TypeError(want Type) error {
	return newTypeError(n.Type(), want)
}

func (n *node) set(inf info, raw string) {
	n.info = inf
	n.raw = raw
}

func (n *node) reset(inf info, raw string, values []V) {
	for i := range n.values {
		n.values[i] = V{}
	}
	*n = node{
		info:   inf,
		raw:    raw,
		values: values,
	}
}

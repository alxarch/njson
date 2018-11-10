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
		return strjson.Unescaped(n.raw)
	}
	return ""
}

const maxUint = ^(uint(0))

func (n *node) Type() Type {
	if n == nil {
		return TypeInvalid
	}
	return n.info.Type()
}

// TypeError creates an error for the Node's type.
func (n *node) TypeError(want Type) error {
	return typeError{n.Type(), want}
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

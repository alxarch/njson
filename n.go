package njson

import (
	"math"

	"github.com/alxarch/njson/numjson"
	"github.com/alxarch/njson/strjson"
)

// N is a JSON document node.
// Do not store pointers to a N.
// N's data is only valid until Document.Close() or Document.Reset().
type N struct {
	info   Info
	raw    string
	values []V
}

// V is a node's value referencing it's id.
// Array node's values have empty string keys.
type V struct {
	ID  uint
	Key string
}

// Unescaped unescapes the value of a String node.
// The returned string is safe to use even if ParseUnsafe was used.
func (n *N) Unescaped() string {
	if n != nil && n.info.IsString() {
		return strjson.Unescaped(n.Safe())
	}
	return ""
}

// Safe returns a safe copy of a node's value JSON string.
// Object and Array nodes return an empty string.
// The returned string is safe to use even if ParseUnsafe was used.
func (n *N) Safe() string {
	if n.info.Safe() {
		return n.raw
	}
	n.raw = scopy(n.raw)
	n.info &^= Unsafe
	return n.raw
}

// Raw returns the JSON string of a node's value.
// Object and Array nodes return an empty string.
// The returned string is NOT safe to use if ParseUnsafe was used.
func (n *N) Raw() string {
	return n.raw
}
func (n *N) Bytes() []byte {
	return s2b(n.raw)
}

// Len returns the length of a node's values.
func (n *N) Len() int {
	return len(n.values)
}

// Values returns a node's values.
// Only non-empty Object and Array nodes have values.
func (n *N) Values() []V {
	return n.values
}

// Append appends a node id to an Array node's values.
func (n *N) Append(id uint) {
	if n != nil && n.info == vArray {
		n.values = append(n.values, V{id, ""})
	}
}

// Set sets an Object's node key to a node id.
// Since most keys need no escaping it doesn't escape the key.
// If the key needs escaping use strjson.Escaped.
func (n *N) Set(key string, id uint) {
	if n != nil && n.info == vObject {
		for i := range n.values {
			if n.values[i].Key == key {
				n.values[i].ID = id
				return
			}
		}
		n.values = append(n.values, V{id, key})
	}
}

// Replace replaces the id of a value at offset i of an Array node.
func (n *N) Replace(i int, id uint) {
	if n != nil && n.info == vArray && 0 <= i && i < len(n.values) {
		n.values[i] = V{id, ""}
	}
}

// Remove removes the value at offset i of an Array or Object node.
func (n *N) Remove(i int) {
	if n != nil && n.info == vArray && 0 <= i && i < len(n.values) {
		if j := i + 1; 0 <= j && j < len(n.values) {
			copy(n.values[i:], n.values[j:])
		}
		if j := len(n.values) - 1; 0 <= j && j < len(n.values) {
			n.values[j] = V{}
			n.values = n.values[:j]
		}
	}
}

const maxUint = ^(uint(0))

// Get finds a key in an Object node's values and returns it's id.
func (n *N) Get(key string) uint {
	if n != nil && n.info == vObject {
		for i := range n.values {
			if n.values[i].Key == key {
				return n.values[i].ID
			}
		}
	}
	return maxUint
}

// Del finds a key in an Object node's values and removes it.
func (n *N) Del(key string) {
	if n != nil && n.info == vObject {
		for i := range n.values {
			if n.values[i].Key == key {
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

func (n *N) Type() Type {
	if n != nil {
		return n.info.Type()
	}
	return TypeInvalid
}

// TypeError creates an error for the Node's type.
func (n *N) TypeError(want Type) error {
	return newTypeError(n.Type(), want)
}

func (n *N) ToUint() (uint64, bool) {
	f := numjson.ParseFloat(n.raw)
	return uint64(f), 0 <= f && f < math.MaxUint64 && math.Trunc(f) == f
}
func (n *N) ToInt() (int64, bool) {
	f := numjson.ParseFloat(n.raw)
	return int64(f), math.MinInt64 <= f && f < math.MaxInt64 && math.Trunc(f) == f
}
func (n *N) ToFloat() (float64, bool) {
	f := numjson.ParseFloat(n.raw)
	return f, f == f
}

func (n *N) ToBool() (bool, bool) {
	if n.info == vBoolean {
		switch n.raw {
		case strTrue:
			return true, true
		case strFalse:
			return false, true
		}
	}
	return false, false
}

package njson

import (
	"math"
	"strconv"

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

type IterV struct {
	*V
	index  int
	values []V
}

func (i *IterV) Reset() {
	i.index = 0
	i.V = nil
}
func (i *IterV) Close() {
	i.values = nil
	i.V = nil
}
func (i *IterV) Next() bool {
	if 0 <= i.index && i.index < len(i.values) {
		i.V = &i.values[i.index]
		i.index++
		return true
	}
	i.V = nil
	i.index = -1
	// i.values = nil
	return false
}
func (i *IterV) Len() int {
	return len(i.values)
}

func (i *IterV) Index() int {
	return i.index
}

func (n *N) Values() IterV {
	if n != nil {
		return IterV{values: n.values}
	}
	return IterV{}
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
	if n == nil {
		return ""
	}
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
	if n != nil {
		return n.raw
	}
	return ""
}

// Bytes returns a node's JSON string as bytes
func (n *N) Bytes() []byte {
	if n != nil {
		return s2b(n.raw)
	}
	return nil
}

// Len returns the length of a node's values.
func (n *N) Len() int {
	if n != nil {
		return len(n.values)
	}
	return 0
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
		var v *V
		for i := range n.values {
			v = &n.values[i]
			if v.key == key {
				v.id = id
				return
			}
		}
		n.values = append(n.values, V{id, key})
	}
}

// Slice reslices the node's values.
func (n *N) Slice(i, j int) {
	if n != nil && n.info == vArray && 0 <= i && i < j && j < len(n.values) {
		n.values = n.values[i:j]
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

// Del finds a key in an Object node's values and removes it.
func (n *N) Del(key string) {
	if n != nil && n.info == vObject {
		for i := range n.values {
			if n.values[i].key == key {
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
	if n != nil {
		f := numjson.ParseFloat(n.raw)
		return uint64(f), 0 <= f && f < math.MaxUint64 && math.Trunc(f) == f
	}
	return 0, false
}
func (n *N) ToInt() (int64, bool) {
	if n != nil {
		f := numjson.ParseFloat(n.raw)
		return int64(f), math.MinInt64 <= f && f < math.MaxInt64 && math.Trunc(f) == f
	}
	return 0, false
}
func (n *N) ToFloat() (float64, bool) {
	if n != nil {
		f := numjson.ParseFloat(n.raw)
		return f, f == f
	}
	return 0, false
}

func (n *N) ToBool() (bool, bool) {
	if n != nil && n.info == vBoolean {
		switch n.raw {
		case strTrue:
			return true, true
		case strFalse:
			return false, true
		}
	}
	return false, false
}

func (n *N) SetStringRaw(s string) {
	if n == nil {
		return
	}
	n.info = vString
	n.raw = s
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]
}
func (n *N) SetStringHTML(s string) {
	if n == nil {
		return
	}
	n.info = vString
	n.raw = strjson.Escaped(s, true, false)
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]
}
func (n *N) SetString(s string) {
	if n == nil {
		return
	}
	n.info = vString
	n.raw = strjson.Escaped(s, false, false)
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]
}
func (n *N) SetFalse() {
	if n == nil {
		return
	}
	n.info = vBoolean
	n.raw = strFalse
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]
}
func (n *N) SetNull() {
	if n == nil {
		return
	}
	n.info = vNull
	n.raw = strNull
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]
}
func (n *N) SetTrue() {
	if n == nil {
		return
	}
	n.info = vBoolean
	n.raw = strTrue
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]
}

func (n *N) SetFloat(f float64) {
	if n == nil {
		return
	}
	n.info = vNumber
	n.raw = numjson.FormatFloat(f, 64)
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]
}

func (n *N) SetUint(u uint64) {
	if n == nil {
		return
	}
	n.info = vNumber
	n.raw = strconv.FormatUint(u, 10)
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]

}
func (n *N) SetInt(i int64) {
	if n == nil {
		return
	}
	n.info = vNumber
	n.raw = strconv.FormatInt(i, 10)
	for i := range n.values {
		n.values[i] = V{}
	}
	n.values = n.values[:0]
}

func (n *N) set(info Info, raw string) {
	n.info = info
	n.raw = raw
}

package njson

import (
	"sync"

	"github.com/alxarch/njson/strjson"

	"github.com/alxarch/njson/numjson"
)

// Document is a JSON document.
type Document struct {
	values []value
	rev    uint // document revision incremented on every Reset/Close invalidating values
}

// value is a JSON document value.
// value's data is only valid until Document.Close() or Document.Reset().
type value struct {
	typ   Type
	flags flags
	// locks is the number of iterator locks.
	// all mutations on the value should fail if it's not zero.
	locks    uint16
	raw      string
	children []child
}

// child is a node's value referencing it's id.
// Array node's values have empty string keys.
type child struct {
	id  uint
	key string
}

// Key returns a value's key
func (v *child) Key() string {
	return v.key
}

// ID returns a value's id
func (v *child) ID() uint {
	return v.id
}

const maxUint = ^(uint(0))

func (v *value) set(typ Type, f flags, raw string) {
	v.typ = typ
	v.flags = f
	v.raw = raw
}

func (v *value) reset(typ Type, f flags, raw string, values []child) {
	for i := range v.children {
		v.children[i] = child{}
	}
	*v = value{
		typ:      typ,
		flags:    f,
		raw:      raw,
		children: values,
	}
}

// lookup finds a node's id by path.
func (d *Document) lookup(id uint, path []string) uint {
	var (
		c *child
		v *value
	)
lookup:
	for _, key := range path {
		if v = d.get(id); v != nil {
			switch v.typ {
			case TypeObject:
				for i := range v.children {
					c = &v.children[i]
					if c.key == key {
						id = c.id
						continue lookup
					}
				}
			case TypeArray:
				i := 0
				for _, c := range []byte(key) {
					if c -= '0'; c <= 9 {
						i = i*10 + int(c)
					} else {
						return maxUint
					}
				}
				if 0 <= i && i < len(v.children) {
					c = &v.children[i]
					id = c.id
					continue lookup
				}
			}
		}
		return maxUint
	}
	return id
}

// Null adds a new Null node to the document.
func (d *Document) Null() Node {
	id := uint(len(d.values))
	v := d.grow()
	v.reset(TypeNull, flagRoot, strNull, v.children[:0])
	return d.Node(id)
}

// False adds a new Boolean node with it's value set to false to the document.
func (d *Document) False() Node {
	id := uint(len(d.values))
	v := d.grow()
	v.reset(TypeBoolean, flagRoot, strFalse, v.children[:0])
	return d.Node(id)
}

// True adds a new Boolean node with it's value set to true to the document.
func (d *Document) True() Node {
	id := uint(len(d.values))
	v := d.grow()
	v.reset(TypeBoolean, flagRoot, strTrue, v.children[:0])
	return d.Node(id)
}

// TextRaw adds a new String node to the document.
func (d *Document) TextRaw(s string) Node {
	id := uint(len(d.values))
	v := d.grow()
	v.reset(TypeString, flagRoot, s, v.children[:0])
	return d.Node(id)

}

// Text adds a new String node to the document escaping JSON unsafe characters.
func (d *Document) Text(s string) Node {
	return d.TextRaw(strjson.Escaped(s, false, false))
}

// TextHTML adds a new String node to the document escaping HTML and JSON unsafe characters.
func (d *Document) TextHTML(s string) Node {
	return d.TextRaw(strjson.Escaped(s, true, false))
}

// Object adds a new empty Object node to the document.
func (d *Document) Object() Object {
	id := uint(len(d.values))
	v := d.grow()
	v.reset(TypeObject, flagRoot, "", v.children[:0])
	return Object(d.Node(id))
}

// Array adds a new empty Array node to the document.
func (d *Document) Array() Array {
	id := uint(len(d.values))
	v := d.grow()
	v.reset(TypeArray, flagRoot, "", v.children[:0])
	return Array(d.Node(id))
}

// Number adds a new Number node to the document.
func (d *Document) Number(f float64) Node {
	id := uint(len(d.values))
	v := d.grow()
	v.reset(TypeNumber, flagRoot, numjson.FormatFloat(f, 64), v.children[:0])
	return d.Node(id)
}

// RawNumber adds a new Number node to the document.
func (d *Document) RawNumber(num string) Node {
	id := uint(len(d.values))
	v := d.grow()
	v.reset(TypeNumber, flagRoot, num, v.children[:0])
	return d.Node(id)
}

// Reset resets the document to empty and releases all strings.
func (d *Document) Reset() {
	for i := range d.values {
		n := &d.values[i]
		for j := range n.children {
			n.children[j] = child{}
		}
		n.raw = ""
	}
	d.values = d.values[:0]
	// Invalidate any partials
	d.rev++
}

// get finds a node by id.
func (d *Document) get(id uint) *value {
	if d != nil && id < uint(len(d.values)) {
		return &d.values[id]
	}
	return nil
}

// Node returns a node with id set to id.
func (d *Document) Node(id uint) Node {
	return Node{id, d.rev, d}
}

// Root returns the document root node.
func (d *Document) Root() Node {
	return Node{0, d.rev, d}
}

// toInterface converts a node to any compatible go value (many allocations on large trees).
func (d *Document) toInterface(id uint) (interface{}, bool) {
	n := d.get(id)
	if n == nil {
		return nil, false
	}
	switch n.typ {
	case TypeObject:
		return d.toInterfaceMap(n.children)
	case TypeArray:
		return d.toInterfaceSlice(n.children)
	case TypeString:
		return strjson.Unescaped(n.raw), true
	case TypeBoolean:
		switch n.raw {
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
		f := numjson.ParseFloat(n.raw)
		return f, f == f
	default:
		return nil, false
	}
}

func (d *Document) toInterfaceMap(values []child) (interface{}, bool) {
	var (
		m  = make(map[string]interface{}, len(values))
		ok bool
	)
	for _, v := range values {
		m[v.key], ok = d.toInterface(v.id)
		if !ok {
			return nil, false
		}
	}
	return m, true
}

func (d *Document) toInterfaceSlice(values []child) ([]interface{}, bool) {
	var (
		x  = make([]interface{}, len(values))
		ok bool
	)
	for i, v := range values {
		x[i], ok = d.toInterface(v.id)
		if !ok {
			return nil, false
		}
	}
	return x, true
}

// AppendJSON appends the JSON data of the document root node to a byte slice.
func (d *Document) AppendJSON(dst []byte) ([]byte, error) {
	return d.appendJSON(dst, d.get(0))
}

func (d *Document) appendJSONEscaped(dst []byte, v *value) ([]byte, error) {
	switch v.typ {
	case TypeObject:
		dst = append(dst, delimBeginObject)
		var err error
		more := ""
		for _, child := range v.children {
			dst = append(dst, more...)
			dst = append(dst, delimString)
			dst = strjson.AppendEscaped(dst, child.key, true)
			dst = append(dst, delimString, delimNameSeparator)
			dst, err = d.appendJSON(dst, d.get(child.id))
			if err != nil {
				return dst, err
			}
			more = ","
		}
		dst = append(dst, delimEndObject)
		return dst, nil
	case TypeString:
		dst = append(dst, delimString)
		dst = strjson.AppendEscaped(dst, v.raw, true)
		dst = append(dst, delimString)
		return dst, nil
	default:
		return dst, newTypeError(v.typ, TypeObject|TypeString)
	}
}

func (d *Document) appendJSON(dst []byte, v *value) ([]byte, error) {
	if v == nil {
		return dst, newTypeError(TypeInvalid, TypeAnyValue)
	}
	if v.flags.IsUnescaped() {
		return d.appendJSONEscaped(dst, v)
	}
	switch v.typ {
	case TypeString:
		dst = append(dst, delimString)
		dst = append(dst, v.raw...)
		dst = append(dst, delimString)
	case TypeObject:
		dst = append(dst, delimBeginObject)
		var err error
		more := ""
		for _, child := range v.children {
			dst = append(dst, more...)
			dst = append(dst, delimString)
			dst = append(dst, child.key...)
			dst = append(dst, delimString, delimNameSeparator)
			dst, err = d.appendJSON(dst, d.get(child.id))
			if err != nil {
				return dst, err
			}
			more = ","
		}
		dst = append(dst, delimEndObject)
	case TypeArray:
		dst = append(dst, delimBeginArray)
		var err error
		for i, child := range v.children {
			if i > 0 {
				dst = append(dst, delimValueSeparator)
			}
			dst, err = d.appendJSON(dst, d.get(child.id))
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, delimEndArray)
	case TypeBoolean, TypeNull, TypeNumber:
		dst = append(dst, v.raw...)
	default:
		return dst, newTypeError(TypeInvalid, TypeAnyValue)
	}
	return dst, nil
}

func (d *Document) copyValue(other *Document, v *value) uint {
	id := uint(len(d.values))
	cp := d.grow()
	children := cp.children[:cap(cp.children)]
	*cp = value{
		typ:   v.typ,
		flags: v.flags,
		raw:   v.raw,
	}
	numChildren := uint(0)
	for i := range v.children {
		c := &v.children[i]
		vc := other.get(c.id)
		if vc != nil {
			children = appendChild(children, c.key, d.copyValue(other, vc), numChildren)
			numChildren++
		}
	}
	d.values[id].children = children[:numChildren]
	return id
}

func (d *Document) copyNode(n Node, parent uint) (uint, bool) {
	v := n.value()
	if v == nil {
		return 0, false
	}
	if n.doc == d && n.id != parent && v.flags.IsRoot() && n.id != 0 {
		v.flags &^= flagRoot
		return n.id, true
	}
	return d.copyValue(n.doc, v), true
}

// grow adds a value to the document.
// It does not reset its data.
// It is inlined by the compiler.
func (d *Document) grow() *value {
	if cap(d.values) > len(d.values) {
		d.values = d.values[:len(d.values)+1]
	} else {
		d.values = append(d.values, value{})
	}
	return &d.values[len(d.values)-1]
}

// Pool is a pool of document objects
type Pool struct {
	docs sync.Pool
}

const minNumNodes = 64

// Get returns a blank document from the pool
func (pool *Pool) Get() *Document {
	if x := pool.docs.Get(); x != nil {
		return x.(*Document)
	}
	d := Document{
		values: make([]value, 0, minNumNodes),
	}
	return &d
}

// Put returns a document to the pool
func (pool *Pool) Put(d *Document) {
	if d == nil {
		return
	}
	// // Free all heap pointers
	// for i := range d.values {
	// 	n := &d.values[i]
	// 	n.raw = ""
	// 	for i := range n.values {
	// 		n.values[i] = V{}
	// 	}
	// }
	d.Reset()
	pool.docs.Put(d)
}

var defaultPool Pool

// Blank returns a blank document from the default pool.
// Use Document.Close() to reset and return the document to the pool.
func Blank() *Document {
	return defaultPool.Get()
}

// Close resets and returns the document to the default pool to be reused.
func (d *Document) Close() {
	defaultPool.Put(d)
}

package njson

import (
	"github.com/alxarch/njson/strjson"
	"math"
	"sync"

	"github.com/alxarch/njson/numjson"
)

// Document is a JSON document.
type Document struct {
	values []value
	rev    uint // document revision incremented on every Clear/Close invalidating values
	pool   *Pool
	err    error
}

func errNode(err error) Node {
	return Node{
		doc: &Document{err: err},
	}
}

// value is a JSON document value.
// value's data is only valid until Document.Close() or Document.Clear().
type value struct {
	typ   Type
	flags flags
	// locks holds the number of locks for Array, Object
	// all mutations on the value should fail if it's not zero.
	locks uint16
	// raw holds the JSON literal for String, Number, Bool, Null
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
		flags:    f | flags(strjson.FlagJSON),
		raw:      raw,
		children: values,
	}
}

func (v *value) remove(index int) (uint, bool) {
	if 0 <= index && index < len(v.children) {
		id, _ := v.children[index].id, copy(v.children[index:], v.children[index+1:])
		v.children = v.children[:len(v.children)-1]
		return id, true
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

func (v *value) del(key string, stable bool) (uint, bool) {
	if !strjson.Flags(v.flags).IsGoSafe() && strjson.NeedsEscape(key) {
		v.unescapeKeys()
	}
	if stable {
		for i := range v.children {
			c := &v.children[i]
			if c.key == key {
				return v.remove(i)
			}
		}
		return 0, false
	}
	if i := len(v.children) - 1; 0 <= i && i < len(v.children) {
		children, last := v.children[:i], v.children[i]
		if last.key == key {
			v.children = children
			return last.id, true
		}
		for i := range children {
			if c := &children[i]; c.key == key {
				id := c.id
				*c = last
				v.children = children
				return id, true
			}
		}
	}
	return 0, false
}

// lookup finds a node's id by path.
func (d *Document) lookup(id uint, path []string) uint {
	var (
		c *child
		v *value
	)
	for _, key := range path {
		if v = d.get(id); v != nil {
			switch v.typ {
			case TypeObject:
				if c = v.get(key); c != nil {
					id = c.id
					continue
				}
			case TypeArray:
				if c = v.index(key); c != nil {
					id = c.id
					continue
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
	v := d.alloc()
	v.reset(TypeNull, flagNew, strNull, v.children[:0])
	return d.Node(id)
}

// False adds a new Boolean node with it's value set to false to the document.
func (d *Document) False() Node {
	id := uint(len(d.values))
	v := d.alloc()
	v.reset(TypeBoolean, flagNew, strFalse, v.children[:0])
	return d.Node(id)
}

// True adds a new Boolean node with it's value set to true to the document.
func (d *Document) True() Node {
	id := uint(len(d.values))
	v := d.alloc()
	v.reset(TypeBoolean, flagNew, strTrue, v.children[:0])
	return d.Node(id)
}

// NewString adds a new String node to the document escaping JSON unsafe characters.
func (d *Document) NewString(s string) Node {
	return d.NewStringJSON(strjson.FromString(s).Escape(false))
}

// NewStringSafe adds a new String node to the document without escaping JSON unsafe characters.
func (d *Document) NewStringSafe(s string) Node {
	return d.NewStringJSON(strjson.FromSafeString(s).Escape(false))
}

// NewStringHTML adds a new String node to the document escaping JSON and HTML unsafe characters.
func (d *Document) NewStringHTML(s string) Node {
	return d.NewStringJSON(strjson.FromHTML(s).Escape(true))
}

func (d *Document) NewStringJSON(s strjson.String) Node {
	id := uint(len(d.values))
	v := d.alloc()
	v.reset(TypeString, flagNew|flags(s.Flags()), s.Value, v.children[:0])
	return d.Node(id)
}

// NewObject adds a new empty Object node to the document.
func (d *Document) NewObject() Object {
	id := uint(len(d.values))
	v := d.alloc()
	v.reset(TypeObject, flagNew, "", v.children[:0])
	return Object(d.Node(id))
}

// NewArray adds a new empty Array node to the document.
func (d *Document) NewArray() Array {
	id := uint(len(d.values))
	v := d.alloc()
	v.reset(TypeArray, flagNew, "", v.children[:0])
	return Array(d.Node(id))
}

// NewFloat adds a new decimal Number node to the document.
func (d *Document) NewFloat(f float64) Node {
	return d.NewNumber(numjson.Float64(f))
}

// NewUint adds a new decimal Number node to the document.
func (d *Document) NewUint(u uint64) Node {
	return d.NewNumber(numjson.Uint64(u))
}

// NewInt adds a new integer Number node to the document.
func (d *Document) NewInt(i int64) Node {
	return d.NewNumber(numjson.Int64(i))
}

// NewNumber adds a new Number node to the document.
func (d *Document) NewNumber(num numjson.Number) Node {
	if num := num.String(); num != "" {
		return d.NewNumberString(num)
	}
	return Node{}
}

// NewNumberString adds a new Number literal to the document.
// Use this for big numbers that cannot be represented as numjson.Number.
func (d *Document) NewNumberString(num string) Node {
	id := uint(len(d.values))
	v := d.alloc()
	v.reset(TypeNumber, flagRoot, num, v.children[:0])
	return d.Node(id)
}

// Reset resets the document to empty
func (d *Document) Reset() {
	d.values = d.values[:0]
	// Invalidate all node references
	d.rev++
}

// Clear resets the document to empty and releases all strings.
func (d *Document) Clear() {
	// We free all heap references
	for i := range d.values {
		n := &d.values[i]
		for j := range n.children {
			n.children[j] = child{}
		}
		n.raw = ""
	}
	d.Reset()
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
	return Node{d, id, d.rev}
}

// Root returns the document root node.
func (d *Document) Root() Node {
	return Node{d, 0, d.rev}
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
		num, err := numjson.Parse(n.raw)
		if err != nil {
			return nil, false
		}
		return num.Value(), true
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
	if !strjson.Flags(v.flags).IsJSONSafe() {
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
	cp := d.alloc()
	children := cp.children[:cap(cp.children)]
	*cp = value{
		typ:   v.typ,
		flags: v.flags,
		raw:   copyRaw(v.raw),
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

// This protects from copying locks
func copyRaw(raw string) string {
	// raw Len is zero even if locks exist
	if len(raw) == 0 {
		return ""
	}
	return raw
}

const zeroID uint = 0

func (d *Document) copyNode(n Node, parent uint) (uint, bool) {
	v := n.value()
	if v == nil {
		return 0, false
	}
	// For nodes of the same document we can skip copying in some cases.
	if n.doc == d && n.id != parent && n.id != zeroID {
		// Since immutable nodes cannot change we can reuse their id without copy.
		// Nodes created with the NewArray or NewObject  methods have a 'root' flag set.
		// The first time they are assigned as a child of another node, we unset the flag and use their id directly.
		// This allows building 'deep' node trees without unnecessary copying.
		// This only happens once for each node so cyclic assignments always result in copying.
		if v.typ.IsImmutable() || v.flags.IsRoot() {
			v.flags &^= flagRoot
			return n.id, true
		}
	}
	return d.copyValue(n.doc, v), true
}

// alloc adds a value to the document.
// It does not reset its data.
// It is inlined by the compiler.
func (d *Document) alloc() *value {
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
		pool:   pool,
	}
	return &d
}

// Put returns a document to the pool
func (pool *Pool) Put(d *Document) {
	if d == nil || d.pool != pool {
		return
	}
	d.Clear()
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
	if d != nil && d.pool != nil {
		d.pool.Put(d)
	}
}

func (d *Document) Pool() *Pool {
	if d != nil {
		return d.pool
	}
	return nil
}

func (v *value) lock() bool {
	if v.locks < math.MaxUint16 {
		v.locks++
		return true
	}
	return false
}

func (v *value) unlock() bool {
	if v.locks > 0 {
		v.locks--
		return true
	}
	return false
}

func (v *value) unlocked() bool {
	return v.locks == 0
}
func (v *value) locked() bool {
	return v.locks != 0
}

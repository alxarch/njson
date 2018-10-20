package njson

import (
	"sync"

	"github.com/alxarch/njson/strjson"

	"github.com/alxarch/njson/numjson"
)

// Document is a JSON document.
type Document struct {
	nodes []node
	rev   uint // document revision incremented on every Reset/Close invalidating nodes
}

// lookup finds a node's id by path.
func (d *Document) lookup(id uint, path []string) uint {
	var (
		v *V
		n *node
	)
lookup:
	for _, key := range path {
		if n = d.get(id); n != nil {
			switch n.info.Type() {
			case TypeObject:
				for i := range n.values {
					v = &n.values[i]
					if v.key == key {
						id = v.id
						continue lookup
					}
				}
			case TypeArray:
				i := 0
				for _, c := range []byte(key) {
					if c -= '0'; 0 <= c && c <= 9 {
						i = i*10 + int(c)
					} else {
						return maxUint
					}
				}
				if 0 <= i && i < len(n.values) {
					v = &n.values[i]
					id = v.id
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
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vNull|infRoot, strNull, n.values[:0])
	return d.Node(id)
}

// False adds a new Boolean node with it's value set to false to the document.
func (d *Document) False() Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vBoolean|infRoot, strFalse, n.values[:0])
	return d.Node(id)
}

// True adds a new Boolean node with it's value set to true to the document.
func (d *Document) True() Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vBoolean|infRoot, strTrue, n.values[:0])
	return d.Node(id)
}

// TextRaw adds a new String node to the document.
func (d *Document) TextRaw(s string) Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vString|infRoot, s, n.values[:0])
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
func (d *Document) Object() Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vObject|infRoot, "", n.values[:0])
	return d.Node(id)
}

// Array adds a new empty Array node to the document.
func (d *Document) Array() Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vArray|infRoot, "", n.values[:0])
	return d.Node(id)
}

// Number adds a new Number node to the document.
func (d *Document) Number(f float64) Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vNumber|infRoot, numjson.FormatFloat(f, 64), n.values[:0])
	return d.Node(id)
}

// Reset resets the document to empty.
func (d *Document) Reset() {
	d.nodes = d.nodes[:0]
	// Invalidate any partials
	d.rev++
}

// get finds a node by id.
func (d *Document) get(id uint) *node {
	if d != nil && id < uint(len(d.nodes)) {
		return &d.nodes[id]
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

// toInterface converts a node to any combatible go value (many allocations on large trees).
func (d *Document) toInterface(id uint) (interface{}, bool) {
	n := d.get(id)
	if n == nil {
		return nil, false
	}
	switch n.info.Type() {
	case TypeObject:
		ok := false
		m := make(map[string]interface{}, len(n.values))
		for _, v := range n.values {

			if m[v.key], ok = d.toInterface(v.id); !ok {
				return nil, false
			}
		}
		return m, true
	case TypeArray:
		s := make([]interface{}, len(n.values))
		if len(n.values) == len(s) {
			ok := false
			// Avoid bounds check
			s = s[:len(n.values)]
			for i, v := range n.values {
				if s[i], ok = d.toInterface(v.id); !ok {
					return nil, false

				}
			}
		}
		return s, true
	case TypeString:
		return n.Unescaped(), true
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

// AppendJSON appends the JSON data of the document root node to a byte slice.
func (d *Document) AppendJSON(dst []byte) ([]byte, error) {
	return d.appendJSON(dst, 0)
}

func (d *Document) appendJSON(dst []byte, id uint) ([]byte, error) {
	n := d.get(id)
	if n == nil {
		return dst, newTypeError(TypeInvalid, TypeAnyValue)
	}
	switch n.info.Type() {
	case TypeObject:
		dst = append(dst, delimBeginObject)
		var err error
		for i, v := range n.values {
			if i > 0 {
				dst = append(dst, delimValueSeparator)
			}
			dst = append(dst, delimString)
			dst = append(dst, v.key...)
			dst = append(dst, delimString, delimNameSeparator)
			dst, err = d.appendJSON(dst, v.id)
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, delimEndObject)
	case TypeArray:
		dst = append(dst, delimBeginArray)
		var err error
		for i, v := range n.values {
			if i > 0 {
				dst = append(dst, delimValueSeparator)
			}
			dst, err = d.appendJSON(dst, v.id)
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, delimEndArray)
	case TypeString:
		dst = append(dst, delimString)
		dst = append(dst, n.raw...)
		dst = append(dst, delimString)
	default:
		dst = append(dst, n.raw...)
	}
	return dst, nil

}

func (d *Document) ncopy(other *Document, n *node) uint {
	id := uint(len(d.nodes))
	cp := d.grow()
	values := cp.values[:cap(cp.values)]
	*cp = node{
		raw:  n.raw,
		info: n.info,
	}
	numV := uint(0)
	for i := range n.values {
		v := &n.values[i]
		n := other.get(v.id)
		if n != nil {
			values = appendV(values, v.key, d.ncopy(other, n), numV)
			numV++
		}
	}
	d.nodes[id].values = values[:numV]
	return id
}

func (d *Document) ncopysafe(other *Document, n *node) uint {
	id := uint(len(d.nodes))
	cp := d.grow()
	values := cp.values[:cap(cp.values)]
	*cp = node{
		raw:  scopy(n.raw),
		info: n.info,
	}
	numV := uint(0)
	for i := range n.values {
		v := &n.values[i]
		n := other.get(v.id)
		if n != nil {
			values = appendV(values, scopy(v.key), d.ncopysafe(other, n), numV)
			numV++
		}
	}
	d.nodes[id].values = values[:numV]
	return id
}

func (d *Document) copyOrAdopt(other *Document, id, to uint) uint {
	n := other.get(id)
	if n == nil {
		return maxUint
	}
	if other == d {
		if id != to && n.info.IsRoot() {
			n.info &^= infRoot
			return id
		}

	} else if !n.info.IsSafe() {
		return d.ncopysafe(other, n)
	}
	return d.ncopy(other, n)
}

func (d *Document) grow() (n *node) {
	if len(d.nodes) < cap(d.nodes) {
		d.nodes = d.nodes[:len(d.nodes)+1]
	} else {
		d.nodes = append(d.nodes, node{})
	}
	return &d.nodes[len(d.nodes)-1]
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
		nodes: make([]node, 0, minNumNodes),
	}
	return &d
}

// Put returns a document to the pool
func (pool *Pool) Put(d *Document) {
	if d == nil {
		return
	}
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

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

// Lookup finds a node's id by path.
func (d *Document) Lookup(id uint, path []string) uint {
	var (
		v *V
		n *node
	)
lookup:
	for _, key := range path {
		if n = d.get(id); n != nil {
			switch n.info {
			case vObject:
				for i := range n.values {
					v = &n.values[i]
					if v.key == key {
						id = v.id
						continue lookup
					}
				}
			case vArray:
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
	n.reset(vNull|Orphan, strNull, n.values[:0])
	return Node{id, d.rev, d}
}

// False adds a new Boolean node with it's value set to false to the document.
func (d *Document) False() Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vBoolean|Orphan, strFalse, n.values[:0])
	return Node{id, d.rev, d}
}

// True adds a new Boolean node with it's value set to true to the document.
func (d *Document) True() Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vBoolean|Orphan, strTrue, n.values[:0])
	return Node{id, d.rev, d}
}

// TextRaw adds a new String node to the document.
func (d *Document) TextRaw(s string) Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vString|Orphan, s, n.values[:0])
	return Node{id, d.rev, d}

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
	n.reset(vObject|Orphan, "", n.values[:0])
	return Node{id, d.rev, d}
}

// Array adds a new empty Array node to the document.
func (d *Document) Array() Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vArray|Orphan, "", n.values[:0])
	return Node{id, d.rev, d}
}

// Number adds a new Number node to the document.
func (d *Document) Number(f float64) Node {
	id := uint(len(d.nodes))
	n := d.grow()
	n.reset(vNumber|Orphan, numjson.FormatFloat(f, 64), n.values[:0])
	return Node{id, d.rev, d}
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

// With returns a node with id set to id.
func (d *Document) With(id uint) Node {
	return Node{id, d.rev, d}
}

var docs sync.Pool

const minNumNodes = 64

// Blank returns a blank document from a pool.
// Use Document.Close() to reset and return the document to the pool.
func Blank() *Document {
	if x := docs.Get(); x != nil {
		return x.(*Document)
	}
	d := Document{
		nodes: make([]node, 0, minNumNodes),
	}
	return &d
}

// Close returns the document to the pool to be reused.
func (d *Document) Close() {
	d.Reset()
	docs.Put(d)
}

// ToInterface converts a node to any combatible go value (many allocations on large trees).
func (d *Document) ToInterface(id uint) (interface{}, bool) {
	n := d.get(id)
	if n == nil {
		return nil, false
	}
	switch n.info.Type() {
	case TypeObject:
		ok := false
		m := make(map[string]interface{}, len(n.values))
		for _, v := range n.values {

			if m[v.key], ok = d.ToInterface(v.id); !ok {
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
				if s[i], ok = d.ToInterface(v.id); !ok {
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

// AppendJSON appends the JSON data of a specific node id to a byte slice.
func (d *Document) AppendJSON(dst []byte, id uint) ([]byte, error) {
	n := d.get(id)
	if n == nil {
		return dst, newTypeError(TypeInvalid, TypeAnyValue)
	}
	switch n.info {
	case vObject:
		dst = append(dst, delimBeginObject)
		var err error
		for i, v := range n.values {
			if i > 0 {
				dst = append(dst, delimValueSeparator)
			}
			dst = append(dst, delimString)
			dst = append(dst, v.key...)
			dst = append(dst, delimString, delimNameSeparator)
			dst, err = d.AppendJSON(dst, v.id)
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, delimEndObject)
	case vArray:
		dst = append(dst, delimBeginArray)
		var err error
		for i, v := range n.values {
			if i > 0 {
				dst = append(dst, delimValueSeparator)
			}
			dst, err = d.AppendJSON(dst, v.id)
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, delimEndArray)
	case vString:
		dst = append(dst, delimString)
		dst = append(dst, n.raw...)
		dst = append(dst, delimString)
	default:
		dst = append(dst, n.raw...)
	}
	return dst, nil

}

func (d *Document) copy(other *Document, id uint) uint {
	if d == nil {
		return maxUint
	}
	n := other.get(id)
	if n == nil {
		return maxUint
	}
	if n.info.IsOrhpan() && other == d {
		n.info &^= Orphan
		return id
	}
	id = uint(len(d.nodes))
	cp := d.grow()
	values := cp.values[:cap(cp.values)]
	*cp = node{
		raw:  n.raw,
		info: n.info,
	}
	n = cp
	numV := uint(0)
	for i := range n.values {
		v := &n.values[i]
		if id := d.copy(other, v.id); id < maxUint {
			values = appendV(values, v.key, id, numV)
			numV++
		}
	}
	d.nodes[id].values = values[:numV]
	return id
}

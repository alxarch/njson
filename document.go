package njson

import (
	"sync"

	"github.com/alxarch/njson/strjson"

	"github.com/alxarch/njson/numjson"
)

// Document is a JSON document.
type Document struct {
	nodes []N
	rev   uint // document revision incremented on every Reset/Close invalidating partials
}

func (d *Document) Lookup(id uint, path []string) uint {
	var (
		v *V
		n *N
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

// Null adds a new Null node to the document and returns it's id
func (d *Document) Null() uint {
	id := uint(len(d.nodes))
	d.nodes = append(d.nodes, N{
		info: vNull,
	})
	return id
}

// False adds a new Boolean node with it's value set to false to the document and returns it's id
func (d *Document) False() uint {
	id := uint(len(d.nodes))
	d.nodes = append(d.nodes, N{
		info: vBoolean,
		raw:  strFalse,
	})
	return id
}

// True adds a new Boolean node with it's value set to true to the document and returns it's id
func (d *Document) True() uint {
	id := uint(len(d.nodes))
	d.nodes = append(d.nodes, N{
		info: vBoolean,
		raw:  strTrue,
	})
	return id
}

// TextRaw adds a new String node to the document and returns it's id.
func (d *Document) TextRaw(s string) uint {
	id := uint(len(d.nodes))
	d.nodes = append(d.nodes, N{
		info: vString,
		raw:  s,
	})
	return id

}

// Text adds a new String node to the document escaping JSON unsafe characters and returns it's id.
func (d *Document) Text(s string) uint {
	id := uint(len(d.nodes))
	d.nodes = append(d.nodes, N{
		info: vString,
		raw:  strjson.Escaped(s, false, false),
	})
	return id
}

// TextHTML adds a new String node to the document escaping HTML and JSON unsafe characters and returns it's id.
func (d *Document) TextHTML(s string) uint {
	id := uint(len(d.nodes))
	d.nodes = append(d.nodes, N{
		info: vString,
		raw:  strjson.Escaped(s, true, false),
	})
	return id
}

// Object adds a new empty Object node to the document and returns it's id.
func (d *Document) Object() uint {
	id := uint(len(d.nodes))
	d.nodes = append(d.nodes, N{
		info: vObject,
	})
	return id
}

// NewObject adds a new empty Object node to the document and returns a pointer to the node.
func (d *Document) NewObject() Node {
	return Node{d.Object(), d.rev, d}
}

// Array adds a new empty Array node to the document and returns it's id.
func (d *Document) Array() uint {
	id := uint(len(d.nodes))
	d.nodes = append(d.nodes, N{
		info: vArray,
	})
	return id
}

// NewArray adds a new empty Array node to the document and returns a pointer to the node.
func (d *Document) NewArray() Node {
	return Node{d.Array(), d.rev, d}
}

// Number adds a new Number node to the document and returns it's id.
func (d *Document) Number(f float64) uint {
	id := uint(len(d.nodes))
	d.nodes = append(d.nodes, N{
		info: vNumber,
		raw:  numjson.FormatFloat(f, 64),
	})
	return id
}

// Reset resets the document to empty.
func (d *Document) Reset() {
	d.nodes = d.nodes[:0]
	// Invalidate any partials
	d.rev++
}

// get finds a node by id.
func (d *Document) get(id uint) *N {
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
		nodes: make([]N, 0, minNumNodes),
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

// Root returns the root partial of a Document
func (d *Document) root() *N {
	if len(d.nodes) > 0 {
		return &d.nodes[0]
	}
	return nil
}

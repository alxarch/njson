package njson

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"

	"github.com/alxarch/njson/strjson"

	"github.com/alxarch/njson/numjson"
)

// Document is a JSON document.
type Document struct {
	nodes []N
	rev   uint // document revision incremented on every Reset/Close invalidating partials
}

func (d *Document) Lookup(id uint, path []string) (uint, bool) {
lookup:
	for _, key := range path {
		if n := d.N(id); n != nil {
			switch n.info {
			case vObject:
				for _, v := range n.values {
					if v.Key == key {
						id = v.ID
						continue lookup
					}
				}
			case vArray:
				if i, err := strconv.Atoi(key); err == nil && 0 <= i && i < len(n.values) {
					id = n.values[i].ID
					continue lookup
				}
			}
		}
		return id, false
	}
	return id, true
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

// N finds a node by id.
// The returned node is only valid until Document.Close() or Document.Reset().
func (d *Document) N(id uint) *N {
	if d != nil && id < uint(len(d.nodes)) {
		return &d.nodes[id]
	}
	return nil
}

// With returns a node with id set to id.
func (d *Document) With(id uint) Node {
	return Node{id, d.rev, d}
}

var docs = new(sync.Pool)

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
	n := d.N(id)
	if n == nil {
		return nil, false
	}
	switch n.info.Type() {
	case TypeObject:
		ok := false
		m := make(map[string]interface{}, len(n.values))
		for _, v := range n.values {

			if m[v.Key], ok = d.ToInterface(v.ID); !ok {
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
				if s[i], ok = d.ToInterface(v.ID); !ok {
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

func (p Node) AppendJSON(dst []byte) ([]byte, error) {
	if d := p.Document(); d != nil {
		return d.AppendJSON(dst, p.id)
	}
	return dst, errors.New("Nil document")
}

func (d *Document) AppendJSON(dst []byte, id uint) ([]byte, error) {
	n := d.N(id)
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
			dst = append(dst, v.Key...)
			dst = append(dst, delimString, delimNameSeparator)
			dst, err = d.AppendJSON(dst, v.ID)
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
			dst, err = d.AppendJSON(dst, v.ID)
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
func (d *Document) Root() *N {
	if len(d.nodes) > 0 {
		return &d.nodes[0]
	}
	return nil
}

type Unmarshaler interface {
	UnmarshalNodeJSON(n Node) error
}

// WrapUnmarshalJSON wraps a call to the json.Unmarshaler interface
func (n Node) WrapUnmarshalJSON(u json.Unmarshaler) (err error) {
	node := n.N()
	if node == nil {
		return node.TypeError(TypeAnyValue)
	}

	switch node.info.Type() {
	case TypeArray:
		if len(node.values) == 0 {
			return u.UnmarshalJSON([]byte{delimBeginArray, delimEndArray})
		}
	case TypeObject:
		if len(node.values) == 0 {
			return u.UnmarshalJSON([]byte{delimBeginObject, delimEndObject})
		}
	case TypeInvalid:
		return node.TypeError(TypeAnyValue)
	default:
		return u.UnmarshalJSON(s2b(node.raw))
	}
	data := bufferpool.Get().([]byte)
	data, err = n.Document().AppendJSON(data[:0], n.id)
	if err == nil {
		err = u.UnmarshalJSON(data)
	}
	bufferpool.Put(data)
	return
}

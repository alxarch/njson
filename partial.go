package njson

import (
	"encoding/json"
	"errors"
)

// Partial is a partial JSON Document.
// It is a versioned reference to a node to avoid document manipulation after reset.
type Partial struct {
	ID  uint
	rev uint
	doc *Document
}

func (p Partial) Document() *Document {
	if p.doc != nil && p.doc.rev == p.rev {
		return p.doc
	}
	return nil
}
func (p Partial) Node() *N {
	return p.Document().Get(p.ID)
}

// Lookup finds a node by path
func (p Partial) Lookup(path []string) (Partial, bool) {
	if id, ok := p.Document().Lookup(p.ID, path); ok {
		p.ID = id
		return p, true
	}
	return p, false
}
func (p Partial) ToInterface() (interface{}, bool) {
	return p.Document().ToInterface(p.ID)
}

// AppendJSON appends the JSON string of a node to a byte slice
func (p Partial) AppendJSON(dst []byte) ([]byte, error) {
	if doc := p.Document(); doc != nil {
		return doc.AppendJSON(dst, p.ID)
	}
	return dst, errors.New("Nil document")
}

type PartialUnmarshaler interface {
	UnmarshalJSONPartial(Partial) error
}

// WrapUnmarshalJSON wraps a call to the json.Unmarshaler interface
func (p Partial) WrapUnmarshalJSON(u json.Unmarshaler) (err error) {
	n := p.Node()
	if n == nil {
		return n.TypeError(TypeAnyValue)
	}

	switch n.info.Type() {
	case TypeArray:
		if len(n.values) == 0 {
			return u.UnmarshalJSON([]byte{delimBeginArray, delimEndArray})
		}
	case TypeObject:
		if len(n.values) == 0 {
			return u.UnmarshalJSON([]byte{delimBeginObject, delimEndObject})
		}
	case TypeInvalid:
		return n.TypeError(TypeAnyValue)
	default:
		return u.UnmarshalJSON(s2b(n.raw))
	}
	data := bufferpool.Get().([]byte)
	data, err = p.AppendJSON(data[:0])
	if err == nil {
		err = u.UnmarshalJSON(data)
	}
	bufferpool.Put(data)
	return
}

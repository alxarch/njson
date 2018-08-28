package unjson

import (
	"encoding"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/alxarch/njson"
)

var (
	typNodeUnmarshaler = reflect.TypeOf((*njson.Unmarshaler)(nil)).Elem()
	typJSONUnmarshaler = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	typTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// Unmarshaler is a type specific decoder
type Unmarshaler interface {
	Unmarshal(x interface{}, n *njson.Node) error
	unmarshaler // disallow external implementations
}

type unmarshaler interface {
	unmarshal(v reflect.Value, n *njson.Node) error
}

type typeUnmarshaler struct {
	unmarshaler
	typ reflect.Type // PtrTo(typ)
}

// Unmarshal implements the Unmarshaler interface.
// It handles the case of a x being a nil pointer by creating a new blank value.
func (c *typeUnmarshaler) Unmarshal(x interface{}, n *njson.Node) error {
	if x == nil {
		return errInvalidValueType
	}
	v := reflect.ValueOf(x)
	if v.Type() != c.typ {
		return errInvalidValueType
	}
	if v.IsNil() {
		if n.IsNull() {
			return nil
		}
		v.Set(reflect.New(c.typ.Elem()))
	}
	return c.unmarshal(v.Elem(), n)
}

var (
	errInvalidValueType = errors.New("Invalid value type")
	errInvalidType      = errors.New("Invalid type")
)

// njsonUnmarshaler implements the Unmarshaler interface for types implementing njson.Unmarshaler
type njsonUnmarshaler struct{}

var _ Unmarshaler = njsonUnmarshaler{}

func (njsonUnmarshaler) Unmarshal(x interface{}, n *njson.Node) error {
	if x, ok := x.(njson.Unmarshaler); ok {
		return x.UnmarshalNodeJSON(n)
	}
	return errInvalidValueType
}

func (njsonUnmarshaler) unmarshal(v reflect.Value, tok *njson.Node) error {
	return v.Interface().(njson.Unmarshaler).UnmarshalNodeJSON(tok)
}

// jsonUnmarshaler implements the Decoder interface for types implementing json.Unmarshaller
type jsonUnmarshaler struct{}

var _ Unmarshaler = jsonUnmarshaler{}

func (jsonUnmarshaler) Unmarshal(x interface{}, n *njson.Node) (err error) {
	if u, ok := x.(json.Unmarshaler); ok {
		return n.WrapUnmarshalJSON(u)
	}
	return errInvalidValueType
}

func (jsonUnmarshaler) unmarshal(v reflect.Value, n *njson.Node) (err error) {
	return n.WrapUnmarshalJSON(v.Interface().(json.Unmarshaler))
}

func TypeUnmarshaler(typ reflect.Type, options Options) (Unmarshaler, error) {
	options = options.normalize()
	if typ == nil {
		return interfaceCodec{options}, nil
	}
	return newTypeUnmarshaler(typ, options)
}

func newTypeUnmarshaler(typ reflect.Type, options Options) (*typeUnmarshaler, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	if typ.Kind() != reflect.Ptr {
		return nil, errInvalidType
	}
	c := typeUnmarshaler{typ: typ}
	switch {
	case typ.Implements(typNodeUnmarshaler):
		c.unmarshaler = njsonUnmarshaler{}
	case typ.Implements(typJSONUnmarshaler):
		c.unmarshaler = jsonUnmarshaler{}
	case typ.Implements(typTextUnmarshaler):
		c.unmarshaler = textCodec{}
	default:
		typ = typ.Elem()
		d, err := newUnmarshaler(typ, options)
		if err != nil {
			return nil, err
		}
		c.unmarshaler = d
	}
	return &c, nil
}

func newUnmarshaler(typ reflect.Type, options Options) (unmarshaler, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	switch {
	case typ.Implements(typNodeUnmarshaler):
		return njsonUnmarshaler{}, nil
	case typ.Implements(typJSONUnmarshaler):
		return jsonUnmarshaler{}, nil
	case typ.Implements(typTextUnmarshaler):
		return textCodec{}, nil
	}
	return newCodec(typ, options)
}

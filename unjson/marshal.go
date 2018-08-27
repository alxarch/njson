package unjson

import (
	"encoding"
	"encoding/json"
	"reflect"

	"github.com/alxarch/njson"
)

// Marshaler is a type specific encoder
type Marshaler interface {
	MarshalTo(out []byte, x interface{}) ([]byte, error)
	marshaler // disallow external implementations
}

type marshaler interface {
	marshal(out []byte, v reflect.Value) ([]byte, error)
}

type typeMarshaler struct {
	marshaler
	typ reflect.Type
}

var (
	typNodeMarshaler = reflect.TypeOf((*njson.Appender)(nil)).Elem()
	typJSONMarshaler = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	typTextMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func (m *typeMarshaler) MarshalTo(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return out, errInvalidValueType
	}
	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Type() != m.typ {
		return out, errInvalidValueType
	}
	return m.marshal(out, v)
}

func TypeMarshaler(typ reflect.Type, options Options) (Marshaler, error) {
	options = options.normalize()
	if typ == nil {
		return interfaceCodec{options}, nil
	}

	return cachedMarshaler(typ, options)
}

func newTypeMarshaler(typ reflect.Type, options Options) (*typeMarshaler, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	m := typeMarshaler{}
	if typ.Kind() == reflect.Ptr {
		m.typ = typ.Elem()
	} else {
		m.typ = typ
		typ = reflect.PtrTo(typ)
	}
	switch {
	case m.typ.Implements(typNodeMarshaler):
		m.marshaler = njsonMarshaler{}
	case m.typ.Implements(typJSONMarshaler):
		m.marshaler = jsonMarshaler{}
	default:
		d, err := newMarshaler(m.typ, options)
		if err != nil {
			return nil, err
		}
		m.marshaler = d
	}
	return &m, nil
}

func newMarshaler(typ reflect.Type, options Options) (marshaler, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	switch {
	case typ.Implements(typNodeMarshaler):
		return njsonMarshaler{}, nil
	case typ.Implements(typJSONMarshaler):
		return jsonMarshaler{}, nil
	case typ.Implements(typTextMarshaler):
		return textCodec{}, nil
	}

	return newCodec(typ, options)
}

type njsonMarshaler struct{}

func (njsonMarshaler) marshal(out []byte, v reflect.Value) ([]byte, error) {
	return v.Interface().(njson.Appender).AppendJSON(out)
}

type jsonMarshaler struct{}

func (jsonMarshaler) marshal(out []byte, v reflect.Value) (b []byte, err error) {
	b, err = v.Interface().(json.Marshaler).MarshalJSON()
	if err == nil && b != nil {
		out = append(out, b...)
	}
	return out, err
}

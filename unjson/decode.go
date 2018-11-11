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

// Decoder is a type specific decoder
type Decoder interface {
	Decode(x interface{}, n njson.Node) error
	decoder // disallow external implementations
}

type decoder interface {
	decode(v reflect.Value, n njson.Node) error
}

type typeDecoder struct {
	decoder
	typ reflect.Type // PtrTo(typ)
}

// Decode implements the Decoder interface.
// It handles the case of a x being a nil pointer by creating a new blank value.
func (c *typeDecoder) Decode(x interface{}, n njson.Node) error {
	if x == nil {
		return errInvalidValueType
	}
	v := reflect.ValueOf(x)
	if v.Type() != c.typ {
		return errInvalidValueType
	}
	if v.IsNil() {
		if n.Type() == njson.TypeNull {
			return nil
		}
		v.Set(reflect.New(c.typ.Elem()))
	}
	return c.decode(v.Elem(), n)
}

var (
	errInvalidValueType = errors.New("Invalid value type")
	errValue            = errors.New("Unsupported value")
	errInvalidType      = errors.New("Invalid type")
	errNilNode          = errors.New("Nil JSON node")
)

// njsonDecoder implements the Decoder interface for types implementing njson.Unmarshaler
type njsonDecoder struct{}

func (njsonDecoder) Decode(x interface{}, n njson.Node) error {
	if x, ok := x.(njson.Unmarshaler); ok {
		return x.UnmarshalNodeJSON(n)
	}
	return errInvalidValueType
}

func (njsonDecoder) decode(v reflect.Value, n njson.Node) error {
	return v.Interface().(njson.Unmarshaler).UnmarshalNodeJSON(n)
}

// jsonDecoder implements the Decoder interface for types implementing json.Unmarshaller
type jsonDecoder struct{}

func (jsonDecoder) Decode(x interface{}, n njson.Node) (err error) {

	if u, ok := x.(json.Unmarshaler); ok {
		return n.WrapUnmarshalJSON(u)
	}
	return errInvalidValueType
}

func (jsonDecoder) decode(v reflect.Value, n njson.Node) (err error) {
	return n.WrapUnmarshalJSON(v.Interface().(json.Unmarshaler))
}

// TypeDecoder creates a new decoder for a type.
func TypeDecoder(typ reflect.Type, tag string) (Decoder, error) {
	if typ == nil {
		return interfaceDecoder{}, nil
	}
	if tag == "" {
		tag = defaultTag
	}
	options := Options{Tag: tag}
	return newTypeDecoder(typ, &options)
}

func newTypeDecoder(typ reflect.Type, options *Options) (*typeDecoder, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	if typ.Kind() != reflect.Ptr {
		return nil, errInvalidType
	}
	c := typeDecoder{typ: typ}
	switch {
	case typ.Implements(typNodeUnmarshaler):
		c.decoder = njsonDecoder{}
	case typ.Implements(typJSONUnmarshaler):
		c.decoder = jsonDecoder{}
	case typ.Implements(typTextUnmarshaler):
		c.decoder = textDecoder{}
	default:
		typ = typ.Elem()
		d, err := newDecoder(typ, options, cache{})
		if err != nil {
			return nil, err
		}
		c.decoder = d
	}
	return &c, nil
}

func newDecoder(typ reflect.Type, options *Options, codecs cache) (decoder, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	switch {
	case typ.Implements(typNodeUnmarshaler):
		return njsonDecoder{}, nil
	case typ.Implements(typJSONUnmarshaler):
		return jsonDecoder{}, nil
	case typ.Implements(typTextUnmarshaler):
		return textDecoder{}, nil
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return newPtrDecoder(typ, options, codecs)
	case reflect.Struct:
		return newStructCodec(typ, options, codecs)
	case reflect.Slice:
		return newSliceDecoder(typ, options, codecs)
	case reflect.Map:
		return newMapDecoder(typ, options, codecs)
	case reflect.Interface:
		if typ.NumMethod() == 0 {
			return interfaceDecoder{}, nil
		}
		return nil, errInvalidType
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intDecoder{}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uintDecoder{}, nil
	case reflect.Float32, reflect.Float64:
		return floatDecoder{}, nil
	case reflect.Bool:
		return boolDecoder{}, nil
	case reflect.String:
		return stringDecoder{}, nil
	default:
		return nil, errInvalidType
	}
}

type sliceDecoder struct {
	typ     reflect.Type
	decoder decoder
}

func newSliceDecoder(typ reflect.Type, options *Options, codecs cache) (*sliceDecoder, error) {
	sd := sliceDecoder{
		typ: typ,
	}
	codecs[typ] = &sd
	dec, err := codecs.decoder(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	sd.decoder = dec
	return &sd, nil
}

func (d *sliceDecoder) decode(v reflect.Value, n njson.Node) (err error) {
	switch n.Type() {
	case njson.TypeNull:
		if !v.IsNil() {
			v.SetLen(0)
		}
	case njson.TypeArray:
		var (
			values = n.Values()
			size   = values.Len()
		)

		if v.IsNil() || v.Cap() < size {
			v.Set(reflect.MakeSlice(d.typ, size, size))
		} else {
			v.SetLen(size)
		}
		for values.Next() {
			err = d.decoder.decode(v.Index(values.Index()), n.With(values.ID()))
			if err != nil {
				v.SetLen(values.Index())
				break
			}
		}
	default:
		return n.TypeError(njson.TypeArray | njson.TypeNull)

	}
	return nil
}

type mapDecoder struct {
	typ       reflect.Type
	keys      decoder
	decoder   decoder
	zeroKey   reflect.Value
	zeroValue reflect.Value
}

func newMapDecoder(typ reflect.Type, options *Options, codecs cache) (*mapDecoder, error) {
	key := typ.Key()
	el := typ.Elem()
	md := mapDecoder{
		typ:       typ,
		zeroKey:   reflect.Zero(key),
		zeroValue: reflect.Zero(el),
	}
	if key.Implements(typTextUnmarshaler) {
		md.keys = textDecoder{}
	} else if key.Kind() == reflect.String {
		md.keys = stringDecoder{}
	} else {
		return nil, errInvalidType
	}
	// First cache the decoder to avoid recursion issues
	codecs[typ] = &md
	dec, err := codecs.decoder(el, options)
	if err != nil {
		return nil, err
	}
	md.decoder = dec
	return &md, nil
}
func (d *mapDecoder) decode(v reflect.Value, n njson.Node) (err error) {
	switch n.Type() {
	case njson.TypeNull:
		return
	case njson.TypeObject:
		val := reflect.New(d.typ.Elem()).Elem()
		for i := n.Values(); i.Next(); {
			val.Set(d.zeroValue)
			err = d.decoder.decode(val, n.With(i.ID()))
			if err != nil {
				return
			}
			v.SetMapIndex(reflect.ValueOf(i.Key()), val)
		}
		return
	default:
		return n.TypeError(njson.TypeObject | njson.TypeNull)
	}
}

type ptrDecoder struct {
	decoder decoder
	zero    reflect.Value
	typ     reflect.Type
}

func newPtrDecoder(typ reflect.Type, options *Options, codecs cache) (*ptrDecoder, error) {
	pd := ptrDecoder{
		typ:  typ.Elem(),
		zero: reflect.Zero(typ),
	}
	codecs[typ] = &pd
	dec, err := codecs.decoder(pd.typ, options)
	if err != nil {
		return nil, err
	}
	pd.decoder = dec
	return &pd, nil
}

func (d *ptrDecoder) decode(v reflect.Value, n njson.Node) error {
	switch n.Type() {
	case njson.TypeNull:
		v.Set(d.zero)
		return nil
	default:
		if v.IsNil() {
			v.Set(reflect.New(d.typ))
		}
		return d.decoder.decode(v.Elem(), n)
	}
}

type stringDecoder struct{}

func (stringDecoder) decode(v reflect.Value, n njson.Node) error {
	v.SetString(n.Unescaped())
	return nil
}

type interfaceDecoder struct{}

func (interfaceDecoder) Decode(x interface{}, n njson.Node) error {
	if x, ok := x.(*interface{}); ok {
		if *x, ok = n.ToInterface(); !ok {
			return n.TypeError(njson.TypeAnyValue)
		}
		return nil
	}
	return n.TypeError(njson.TypeAnyValue)
}

func (interfaceDecoder) decode(v reflect.Value, n njson.Node) error {
	if !v.CanAddr() {
		return errInvalidValueType
	}
	if x, ok := n.ToInterface(); ok {
		xx := v.Addr().Interface().(*interface{})
		*xx = x
		return nil
	}
	return n.TypeError(njson.TypeAnyValue)
}

type boolDecoder struct{}

func (boolDecoder) decode(v reflect.Value, n njson.Node) (err error) {
	if b, ok := n.ToBool(); ok {
		v.SetBool(b)
		return nil
	}
	return n.TypeError(njson.TypeBoolean)

}

type uintDecoder struct{}

func (uintDecoder) decode(v reflect.Value, n njson.Node) (err error) {
	if u, ok := n.ToUint(); ok {
		v.SetUint(u)
		return nil
	}
	return n.TypeError(njson.TypeNumber)
}

type intDecoder struct{}

func (intDecoder) decode(v reflect.Value, n njson.Node) (err error) {
	if i, ok := n.ToInt(); ok {
		v.SetInt(i)
		return nil
	}
	return n.TypeError(njson.TypeNumber)
}

type floatDecoder struct{}

func (floatDecoder) decode(v reflect.Value, n njson.Node) (err error) {
	if f, ok := n.ToFloat(); ok {
		v.SetFloat(f)
		return nil
	}
	return n.TypeError(njson.TypeNumber)
}

type textDecoder struct{}

func (textDecoder) decode(v reflect.Value, n njson.Node) error {
	if n.Type() == njson.TypeString {
		return n.WrapUnmarshalText(v.Interface().(encoding.TextUnmarshaler))
	}
	return n.TypeError(njson.TypeString)
}

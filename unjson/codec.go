package unjson

import (
	"encoding"
	"reflect"
	"strconv"

	"github.com/alxarch/njson"
	"github.com/alxarch/njson/strjson"
)

type codec interface {
	marshaler
	unmarshaler
}

func newCodec(typ reflect.Type, options Options) (codec, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return newPtrCodec(typ, options)
	case reflect.Struct:
		return newStructCodec(typ, options)
	case reflect.Slice:
		return newSliceCodec(typ, options)
	case reflect.Map:
		return newMapCodec(typ, options)
	case reflect.Interface:
		if typ.NumMethod() == 0 {
			return interfaceCodec{options}, nil
		}
		return nil, errInvalidType
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intCodec{}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uintCodec{}, nil
	case reflect.Float32, reflect.Float64:
		return floatCodec{options.FloatPrecision}, nil
	case reflect.Bool:
		return boolCodec{}, nil
	case reflect.String:
		return stringCodec{}, nil
	default:
		return nil, errInvalidType
	}

}

type stringCodec struct{}

var _ codec = stringCodec{}

func (stringCodec) unmarshal(v reflect.Value, n *njson.Node) (err error) {
	s := n.Unescaped()
	v.SetString(s)
	return
}

func (stringCodec) marshal(b []byte, v reflect.Value) ([]byte, error) {
	b = append(b, delimString)
	b = strjson.Escape(b, v.String())
	b = append(b, delimString)
	return b, nil
}

type boolCodec struct{}

var _ codec = boolCodec{}

func (boolCodec) unmarshal(v reflect.Value, n *njson.Node) (err error) {
	if b, ok := n.ToBool(); ok {
		v.SetBool(b)
		return nil
	}
	return n.TypeError(njson.TypeBoolean)
}

func (boolCodec) marshal(b []byte, v reflect.Value) ([]byte, error) {
	if v.Bool() {
		return append(b, strTrue...), nil
	}
	return append(b, strFalse...), nil
}

type uintCodec struct{}

var _ codec = uintCodec{}

func (uintCodec) unmarshal(v reflect.Value, n *njson.Node) (err error) {
	if u, ok := n.ToUint(); ok {
		v.SetUint(u)
		return nil
	}
	return n.TypeError(njson.TypeNumber)
}

func (uintCodec) marshal(b []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendUint(b, v.Uint(), 10), nil
}

type intCodec struct{}

var _ codec = intCodec{}

func (intCodec) unmarshal(v reflect.Value, n *njson.Node) (err error) {
	if i, ok := n.ToInt(); ok {
		v.SetInt(i)
		return nil
	}
	return n.TypeError(njson.TypeNumber)
}

func (intCodec) marshal(b []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendInt(b, v.Int(), 10), nil
}

type floatCodec struct{ precision int }

var _ codec = floatCodec{}

func (c floatCodec) marshal(out []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendFloat(out, v.Float(), 'f', c.precision, 64), nil
}

func (floatCodec) unmarshal(v reflect.Value, n *njson.Node) (err error) {
	if f, ok := n.ToFloat(); ok {
		v.SetFloat(f)
		return nil
	}
	return n.TypeError(njson.TypeNumber)
}

type interfaceCodec struct {
	options Options
}

var _ codec = interfaceCodec{}

func (interfaceCodec) unmarshal(v reflect.Value, n *njson.Node) error {
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

func (c interfaceCodec) MarshalTo(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return append(out, strNull...), nil
	}
	return c.marshal(out, reflect.ValueOf(x))
}

func (c interfaceCodec) marshal(b []byte, v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return append(b, strNull...), nil
	}
	return MarshalTo(b, v.Interface())
}

// func (d interfaceCodec) UnmarshalFromString(x interface{}, src string) (err error) {
// 	p := njson.BlankDocument()
// 	p.Reset()
// 	root, err := p.Parse(src)
// 	if err == nil {
// 		err = d.Unmarshal(x, root)
// 	}
// 	p.Close()
// 	return
// }

func (interfaceCodec) Unmarshal(x interface{}, n *njson.Node) error {
	if x, ok := x.(*interface{}); ok {
		if *x, ok = n.ToInterface(); !ok {
			return n.TypeError(njson.TypeAnyValue)
		}
		return nil
	}
	return n.TypeError(njson.TypeAnyValue)
}

type textCodec struct{}

var _ codec = textCodec{}

func (textCodec) unmarshal(v reflect.Value, n *njson.Node) error {
	if n.IsString() {
		return v.Interface().(encoding.TextUnmarshaler).UnmarshalText(n.Bytes())
	}
	return n.TypeError(njson.TypeString)
}

func (textCodec) marshal(out []byte, v reflect.Value) (text []byte, err error) {
	text, err = v.Interface().(encoding.TextMarshaler).MarshalText()
	if err == nil {
		out = append(out, delimString)
		out = append(out, text...)
		out = append(out, delimString)
	}
	return out, err
}

type mapCodec struct {
	typ        reflect.Type
	keyDecoder unmarshaler
	keyZero    reflect.Value
	valZero    reflect.Value
	decoder    unmarshaler
	encoder    marshaler
}

var _ codec = (*mapCodec)(nil)

func newMapCodec(typ reflect.Type, options Options) (*mapCodec, error) {
	if typ.Kind() != reflect.Map {
		return nil, errInvalidType
	}

	var keys unmarshaler
	if typ.Key().Implements(typTextUnmarshaler) {
		keys = textCodec{}
	} else if typ.Key().Kind() == reflect.String {
		keys = stringCodec{}
	} else {
		return nil, errInvalidType
	}
	dec, err := newUnmarshaler(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	enc, err := newMarshaler(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	return &mapCodec{
		typ:        typ,
		keyZero:    reflect.Zero(typ.Key()),
		valZero:    reflect.Zero(typ.Elem()),
		keyDecoder: keys,
		decoder:    dec,
		encoder:    enc,
	}, nil
}

func (d *mapCodec) marshal(out []byte, v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return append(out, strNull...), nil
	}
	out = append(out, delimBeginObject)
	var err error
	for i, key := range v.MapKeys() {
		if i > 0 {
			out = append(out, delimValueSeparator)
		}
		out = append(out, delimString)
		out = append(out, v.String()...)
		out = append(out, delimString, delimNameSeparator)
		out, err = d.encoder.marshal(out, v.MapIndex(key))
		if err != nil {
			return out, err
		}
	}
	out = append(out, delimEndObject)
	return out, nil

}

func (d *mapCodec) unmarshal(v reflect.Value, n *njson.Node) (err error) {
	switch n.Type() {
	case njson.TypeNull:
		return
	case njson.TypeObject:
		// key := reflect.New(d.typ.Key()).Elem()
		val := reflect.New(d.typ.Elem()).Elem()
		for _, n := range n.Values() {
			// key.Set(d.keyZero)
			// k := njson.Node{raw: n.key}
			// err = d.keyDecoder.unmarshal(key, &k)
			// if err != nil {
			// 	return
			// }
			val.Set(d.valZero)
			err = d.decoder.unmarshal(val, n)
			if err != nil {
				return
			}
			v.SetMapIndex(reflect.ValueOf(n.Key()), val)
		}
		return
	default:
		return n.TypeError(njson.TypeObject | njson.TypeNull)
	}
}

type ptrCodec struct {
	decoder unmarshaler
	encoder marshaler
	zero    reflect.Value
	typ     reflect.Type
}

var _ codec = (*ptrCodec)(nil)

func newPtrCodec(typ reflect.Type, options Options) (*ptrCodec, error) {
	dec, err := newUnmarshaler(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	enc, err := newMarshaler(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	return &ptrCodec{
		typ:     typ.Elem(),
		decoder: dec,
		encoder: enc,
		zero:    reflect.Zero(typ),
	}, nil
}

func (d *ptrCodec) marshal(b []byte, v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return append(b, strNull...), nil
	}
	return d.encoder.marshal(b, v.Elem())
}

func (d *ptrCodec) unmarshal(v reflect.Value, n *njson.Node) error {
	switch n.Type() {
	case njson.TypeNull:
		v.Set(d.zero)
		return nil
	default:
		if v.IsNil() {
			v.Set(reflect.New(d.typ))
		}
		return d.decoder.unmarshal(v.Elem(), n)
	}
}

type sliceCodec struct {
	typ     reflect.Type
	decoder unmarshaler
	encoder marshaler
}

var _ codec = (*sliceCodec)(nil)

func newSliceCodec(typ reflect.Type, options Options) (*sliceCodec, error) {
	dec, err := newUnmarshaler(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	enc, err := newMarshaler(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	return &sliceCodec{
		typ:     typ,
		decoder: dec,
		encoder: enc,
	}, nil
}

func (d sliceCodec) marshal(out []byte, v reflect.Value) ([]byte, error) {
	out = append(out, delimBeginArray)
	if !v.IsNil() {
		var (
			err error
			n   = v.Len()
			i   = 0
		)
		for ; i < n; i++ {
			if i > 0 {
				out = append(out, delimValueSeparator)
			}
			out, err = d.encoder.marshal(out, v.Index(i))
			if err != nil {
				return out, err
			}
		}
	}
	out = append(out, delimEndArray)
	return out, nil
}

func (d sliceCodec) unmarshal(v reflect.Value, n *njson.Node) (err error) {
	switch n.Type() {
	case njson.TypeNull:
		if !v.IsNil() {
			v.SetLen(0)
		}
	case njson.TypeArray:
		size := n.Len()
		if v.IsNil() || v.Cap() < size {
			v.Set(reflect.MakeSlice(d.typ, size, size))
		} else {
			v.SetLen(size)
		}

		for i, n := range n.Values() {
			err = d.decoder.unmarshal(v.Index(i), n)
			if err != nil {
				v.SetLen(i)
				break
			}
		}
	default:
		return n.TypeError(njson.TypeArray | njson.TypeNull)

	}
	return nil
}

const (
	delimString         = '"'
	delimBeginObject    = '{'
	delimEndObject      = '}'
	delimBeginArray     = '['
	delimEndArray       = ']'
	delimNameSeparator  = ':'
	delimValueSeparator = ','
)

const (
	strFalse = "false"
	strTrue  = "true"
	strNull  = "null"
)

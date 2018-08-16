package njson

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"
)

type CodecOptions struct {
	TagOptions
	OmitEmpty bool // Force omitempty on all fields
}

const defaultTag = "json"

var (
	defaultOptions = CodecOptions{
		TagOptions: TagOptions{
			Key: defaultTag,
			AutoName: func(name string) string {
				return name
			},
		},
	}
)

type codec interface {
	encoder
	decoder
}

func DefaultOptions() CodecOptions {
	return defaultOptions
}

type TagOptions struct {
	Key      string              // Tag key to use for encoder/decoder
	AutoName func(string) string // Field name to json key converter
}

func (o TagOptions) key() (key string) {
	if key = o.Key; key == "" {
		key = defaultTag
	}
	return
}

func (o TagOptions) tag(field reflect.StructField) (tag string, omitempty bool, ok bool) {
	if tag, ok = field.Tag.Lookup(o.key()); ok {
		if i := strings.IndexByte(tag, ','); i != -1 {
			omitempty = strings.Index(tag[i:], "omitempty") > 0
			tag = tag[:i]
		}
	} else if o.AutoName == nil {
		tag = field.Name
	} else {
		tag = o.AutoName(field.Name)
	}
	return
}

type stringCodec struct{}

var _ codec = stringCodec{}

func (stringCodec) decode(v reflect.Value, n *Node) (err error) {
	s := n.Unescaped()
	v.SetString(s)
	return
}

func (stringCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	b = append(b, delimString)
	b = EscapeString(b, v.String())
	b = append(b, delimString)
	return b, nil
}

type boolCodec struct{}

var _ codec = boolCodec{}

func (boolCodec) decode(v reflect.Value, n *Node) (err error) {
	switch n.src {
	case strFalse:
		v.SetBool(false)
		return nil
	case strTrue:
		v.SetBool(true)
		return nil
	default:
		return errInvalidNodeType
	}
}

func (boolCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	if v.Bool() {
		return append(b, strTrue...), nil
	}
	return append(b, strFalse...), nil
}

type uintCodec struct{}

var _ codec = uintCodec{}

func (uintCodec) decode(v reflect.Value, n *Node) (err error) {
	if u, ok := n.ToUint(); ok {
		v.SetUint(u)
		return nil
	}
	return errInvalidNodeType
}

func (uintCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendUint(b, v.Uint(), 10), nil
}

type intCodec struct{}

var _ codec = intCodec{}

func (intCodec) decode(v reflect.Value, n *Node) (err error) {
	if i, ok := n.ToInt(); ok {
		v.SetInt(i)
		return nil
	}
	return errInvalidNodeType
}

func (intCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendInt(b, v.Int(), 10), nil
}

type floatCodec struct{}

var _ codec = floatCodec{}

func (floatCodec) encode(out []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendFloat(out, v.Float(), 'f', 6, 64), nil
}

func (floatCodec) decode(v reflect.Value, n *Node) (err error) {
	if f, ok := n.ToFloat(); ok {
		v.SetFloat(f)
		return nil
	}
	return errInvalidNodeType
}

type interfaceCodec struct{}

var _ codec = interfaceCodec{}

func (interfaceCodec) decode(v reflect.Value, n *Node) error {
	if !v.CanAddr() {
		return errInvalidValueType
	}
	if x, ok := n.ToInterface(); ok {
		xx := v.Addr().Interface().(*interface{})
		*xx = x
		return nil
	}
	return errInvalidNodeType
}

func (c interfaceCodec) Encode(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return append(out, strNull...), nil
	}
	return c.encode(out, reflect.ValueOf(x))
}

func (c interfaceCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return append(b, strNull...), nil
	}
	switch v = v.Elem(); v.Kind() {
	case reflect.String:
		b = append(b, delimString)
		b = EscapeString(b, v.String())
		b = append(b, delimString)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b = strconv.AppendInt(b, v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b = strconv.AppendUint(b, v.Uint(), 10)
	case reflect.Bool:
		b = strconv.AppendBool(b, v.Bool())
	case reflect.Slice:
		b = append(b, delimBeginArray)
		var err error
		var vv reflect.Value
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				b = append(b, delimValueSeparator)
			}
			vv = v.Index(i)
			if vv.CanInterface() {
				b, err = c.encode(b, reflect.ValueOf(vv.Interface()))
			} else {
				err = errInvalidValueType
			}
			if err != nil {
				return b, err
			}
		}
		b = append(b, delimEndArray)
	default:
		return b, errInvalidValueType
	}
	return b, nil

}

func (d interfaceCodec) DecodeString(x interface{}, src string) (err error) {
	p := parsePool.Get().(*parsePair)
	p.Reset()
	if _, err = p.Parse(src, &p.Document); err == nil {
		err = d.Decode(x, p.Get(0))
	}
	parsePool.Put(p)
	return
}

func (interfaceCodec) Decode(x interface{}, n *Node) error {
	if x, ok := x.(*interface{}); ok {
		if *x, ok = n.ToInterface(); !ok {
			return errInvalidNodeType
		}
		return nil
	}
	return errInvalidValueType
}

type textCodec struct{}

var _ codec = textCodec{}

func (textCodec) decode(v reflect.Value, n *Node) error {
	if n.info.IsQuoted() {
		return v.Interface().(encoding.TextUnmarshaler).UnmarshalText(n.UnescapedBytes())
	}
	return errInvalidNodeType
}

func (textCodec) encode(out []byte, v reflect.Value) (text []byte, err error) {
	text, err = v.Interface().(encoding.TextMarshaler).MarshalText()
	if err == nil {
		out = append(out, text...)
	}
	return out, err
}

type mapCodec struct {
	typ        reflect.Type
	keyDecoder decoder
	keyZero    reflect.Value
	valZero    reflect.Value
	decoder    decoder
	encoder    encoder
}

var _ codec = (*mapCodec)(nil)

func newMapCodec(typ reflect.Type, options CodecOptions) (*mapCodec, error) {
	if typ.Kind() != reflect.Map {
		return nil, errInvalidType
	}

	var keys decoder
	if typ.Key().Implements(typTextUnmarshaler) {
		keys = textCodec{}
	} else if typ.Key().Kind() == reflect.String {
		keys = stringCodec{}
	} else {
		return nil, errInvalidType
	}
	dec, err := newDecoder(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	enc, err := newEncoder(typ.Elem(), options)
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

func (d *mapCodec) encode(out []byte, v reflect.Value) ([]byte, error) {
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
		out, err = d.encoder.encode(out, v.MapIndex(key))
		if err != nil {
			return out, err
		}
	}
	out = append(out, delimEndObject)
	return out, nil

}

func (d *mapCodec) decode(v reflect.Value, n *Node) (err error) {
	switch n.Type() {
	case TypeNull:
		return
	case TypeObject:
		key := reflect.New(d.typ.Key()).Elem()
		val := reflect.New(d.typ.Elem()).Elem()
		for n = n.Value(); n != nil; n = n.Next() {
			key.Set(d.keyZero)
			err = d.keyDecoder.decode(key, n)
			if err != nil {
				return
			}
			val.Set(d.valZero)
			err = d.decoder.decode(val, n.Value())
			if err != nil {
				return
			}
			v.SetMapIndex(key, val)
		}
		return
	default:
		return errInvalidNodeType
	}
}

type ptrCodec struct {
	decoder decoder
	encoder encoder
	zero    reflect.Value
	typ     reflect.Type
}

var _ codec = (*ptrCodec)(nil)

func newPtrCodec(typ reflect.Type, options CodecOptions) (*ptrCodec, error) {
	dec, err := newDecoder(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	return &ptrCodec{
		typ:     typ.Elem(),
		decoder: dec,
		zero:    reflect.Zero(typ),
	}, nil
}

func (d *ptrCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return append(b, strNull...), nil
	}
	return d.encoder.encode(b, v.Elem())
}

func (d *ptrCodec) decode(v reflect.Value, n *Node) error {
	switch n.Type() {
	case TypeNull:
		v.Set(d.zero)
		return nil
	default:
		if v.IsNil() {
			v.Set(reflect.New(d.typ))
		}
		return d.decoder.decode(v.Elem(), n)
	}
}

type sliceCodec struct {
	typ     reflect.Type
	decoder decoder
	encoder encoder
}

var _ codec = (*sliceCodec)(nil)

func newSliceCodec(typ reflect.Type, options CodecOptions) (*sliceCodec, error) {
	dec, err := newDecoder(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	enc, err := newEncoder(typ.Elem(), options)
	if err != nil {
		return nil, err
	}
	return &sliceCodec{
		typ:     typ,
		decoder: dec,
		encoder: enc,
	}, nil
}

func (d sliceCodec) encode(out []byte, v reflect.Value) ([]byte, error) {
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
			out, err = d.encoder.encode(out, v.Index(i))
			if err != nil {
				return out, err
			}
		}
	}
	out = append(out, delimEndArray)
	return out, nil
}

func (d sliceCodec) decode(v reflect.Value, n *Node) (err error) {
	switch n.Type() {
	case TypeNull:
		if !v.IsNil() {
			v.SetLen(0)
		}
	case TypeArray:
		size := n.Len()
		if v.IsNil() || v.Cap() < size {
			v.Set(reflect.MakeSlice(d.typ, size, size))
		} else {
			v.SetLen(size)
		}

		for i, next := 0, n.Value(); next != nil && i < size; i, next = i+1, next.Next() {
			err = d.decoder.decode(v.Index(i), next)
			if err != nil {
				v.SetLen(i)
				break
			}
		}
	default:
		return errInvalidNodeType
	}
	return nil
}

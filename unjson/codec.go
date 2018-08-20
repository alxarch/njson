package unjson

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/alxarch/njson"
)

type CodecOptions struct {
	FieldParser        // If nil DefaultFieldParser is used
	FloatPrecision int // strconv.FormatFloat precision for encoder
	OmitMethod     string
}

func (o CodecOptions) ParseField(f reflect.StructField) (name string, omiempty, ok bool) {
	if o.FieldParser == nil {
		return defaultFieldParser.ParseField(f)
	}
	return o.FieldParser.ParseField(f)
}

func (o CodecOptions) normalize() CodecOptions {
	if o.FieldParser == nil {
		o.FieldParser = defaultFieldParser
	}
	if o.FloatPrecision <= 0 {
		o.FloatPrecision = defaultOptions.FloatPrecision
	}
	if o.OmitMethod == "" {
		o.OmitMethod = defaultOmitMethod
	}
	return o
}

const (
	defaultTag        = "json"
	defaultOmitMethod = "Omit"
)

var (
	defaultOptions = CodecOptions{
		FieldParser:    fieldParser{defaultTag, false},
		FloatPrecision: 6,
		OmitMethod:     defaultOmitMethod,
	}
)

type codec interface {
	encoder
	decoder
}

func newCodec(typ reflect.Type, options CodecOptions) (codec, error) {
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

var defaultFieldParser = NewFieldParser(defaultTag, false)

func DefaultFieldParser() FieldParser {
	return defaultFieldParser
}
func DefaultOptions() CodecOptions {
	return defaultOptions
}

type FieldParser interface {
	ParseField(f reflect.StructField) (name string, omitempty, ok bool)
}

type fieldParser struct {
	Key       string // Tag key to use for encoder/decoder
	OmitEmpty bool   // Force omitempty on all fields
}

func NewFieldParser(key string, omitempty bool) FieldParser {
	if key == "" {
		key = defaultTag
	}
	return fieldParser{key, omitempty}
}

func (o fieldParser) ParseField(field reflect.StructField) (tag string, omitempty bool, ok bool) {
	omitempty = o.OmitEmpty
	if tag, ok = field.Tag.Lookup(o.Key); ok {
		if i := strings.IndexByte(tag, ','); i != -1 {
			if !omitempty {
				omitempty = strings.Index(tag[i:], "omitempty") > 0
			}
			tag = tag[:i]
		}
	} else {
		tag = field.Name
	}
	return
}

type stringCodec struct{}

var _ codec = stringCodec{}

func (stringCodec) decode(v reflect.Value, n *njson.Node) (err error) {
	s := n.Unescaped()
	v.SetString(s)
	return
}

func (stringCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	b = append(b, delimString)
	b = njson.EscapeString(b, v.String())
	b = append(b, delimString)
	return b, nil
}

type boolCodec struct{}

var _ codec = boolCodec{}

func (boolCodec) decode(v reflect.Value, n *njson.Node) (err error) {
	if b, ok := n.ToBool(); ok {
		v.SetBool(b)
		return nil
	}
	return n.TypeError(njson.TypeBoolean)
}

func (boolCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	if v.Bool() {
		return append(b, strTrue...), nil
	}
	return append(b, strFalse...), nil
}

type uintCodec struct{}

var _ codec = uintCodec{}

func (uintCodec) decode(v reflect.Value, n *njson.Node) (err error) {
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

func (intCodec) decode(v reflect.Value, n *njson.Node) (err error) {
	if i, ok := n.ToInt(); ok {
		v.SetInt(i)
		return nil
	}
	return errInvalidNodeType
}

func (intCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendInt(b, v.Int(), 10), nil
}

type floatCodec struct{ precision int }

var _ codec = floatCodec{}

func (c floatCodec) encode(out []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendFloat(out, v.Float(), 'f', c.precision, 64), nil
}

func (floatCodec) decode(v reflect.Value, n *njson.Node) (err error) {
	if f, ok := n.ToFloat(); ok {
		v.SetFloat(f)
		return nil
	}
	return errInvalidNodeType
}

type interfaceCodec struct {
	options CodecOptions
}

var _ codec = interfaceCodec{}

func (interfaceCodec) decode(v reflect.Value, n *njson.Node) error {
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
	return MarshalTo(b, v.Interface())
	// switch v = v.Elem(); v.Kind() {
	// case reflect.String:
	// 	b = append(b, delimString)
	// 	b = njson.EscapeString(b, v.String())
	// 	b = append(b, delimString)
	// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	// 	b = strconv.AppendInt(b, v.Int(), 10)
	// case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	// 	b = strconv.AppendUint(b, v.Uint(), 10)
	// case reflect.Bool:
	// 	b = strconv.AppendBool(b, v.Bool())
	// default:
	// 	enc, err := cachedEncoder(v.Type(), c.options)
	// 	if err != nil {
	// 		return b, err
	// 	}
	// 	return enc.encode(b, v)
	// }
	// return b, nil

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

func (interfaceCodec) Decode(x interface{}, n *njson.Node) error {
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

func (textCodec) decode(v reflect.Value, n *njson.Node) error {
	if n.IsQuoted() {
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

func (d *mapCodec) decode(v reflect.Value, n *njson.Node) (err error) {
	switch n.Type() {
	case njson.TypeNull:
		return
	case njson.TypeObject:
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
	enc, err := newEncoder(typ.Elem(), options)
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

func (d *ptrCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return append(b, strNull...), nil
	}
	return d.encoder.encode(b, v.Elem())
}

func (d *ptrCodec) decode(v reflect.Value, n *njson.Node) error {
	switch n.Type() {
	case njson.TypeNull:
		v.Set(d.zero)
		return nil
	default:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
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

func (d sliceCodec) decode(v reflect.Value, n *njson.Node) (err error) {
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

var bufferpool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

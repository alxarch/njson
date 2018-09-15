package unjson

import (
	"encoding"
	"encoding/json"
	"math"
	"reflect"
	"strconv"

	"github.com/alxarch/njson"
	"github.com/alxarch/njson/strjson"
)

// Encoder is a type specific encoder
type Encoder interface {
	Encode(out []byte, x interface{}) ([]byte, error)
	encoder // disallow external implementations
}

type encoder interface {
	encode(out []byte, v reflect.Value) ([]byte, error)
}

type typeEncoder struct {
	encoder
	typ reflect.Type
}

var (
	typAppender      = reflect.TypeOf((*njson.Appender)(nil)).Elem()
	typJSONMarshaler = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	typTextMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func (m *typeEncoder) Encode(out []byte, x interface{}) ([]byte, error) {
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
	return m.encode(out, v)
}

func TypeEncoder(typ reflect.Type, options Options) (Encoder, error) {
	if typ == nil {
		return interfaceEncoder{}, nil
	}
	options = options.normalize()

	return cachedEncoder(typ, &options)
}

func newTypeEncoder(typ reflect.Type, options *Options) (*typeEncoder, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	m := typeEncoder{}
	if typ.Kind() == reflect.Ptr {
		m.typ = typ.Elem()
	} else {
		m.typ = typ
		typ = reflect.PtrTo(typ)
	}
	switch {
	case m.typ.Implements(typAppender):
		m.encoder = njsonEncoder{}
	case m.typ.Implements(typJSONMarshaler):
		m.encoder = jsonEncoder{}
	case m.typ.Implements(typTextMarshaler):
		m.encoder = textEncoder{}
	default:
		d, err := newEncoder(m.typ, options)
		if err != nil {
			return nil, err
		}
		m.encoder = d
	}
	return &m, nil
}

func newEncoder(typ reflect.Type, options *Options) (encoder, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	switch {
	case typ.Implements(typAppender):
		return njsonEncoder{}, nil
	case typ.Implements(typJSONMarshaler):
		return jsonEncoder{}, nil
	case typ.Implements(typTextMarshaler):
		return textEncoder{}, nil
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return newPtrEncoder(typ, options)
	case reflect.Struct:
		return cachedCodec(typ, options)
	case reflect.Slice:
		return newSliceEncoder(typ, options)
	case reflect.Map:
		return newMapEncoder(typ, options)
	case reflect.Interface:
		if typ.NumMethod() == 0 {
			return interfaceEncoder{}, nil
		}
		return nil, errInvalidType
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder{}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uintEncoder{}, nil
	case reflect.Float32:
		return newFloatEncoder(32, options)
	case reflect.Float64:
		return newFloatEncoder(32, options)
	case reflect.Bool:
		return boolEncoder{}, nil
	case reflect.String:
		return stringEncoder(options.HTML), nil
	default:
		return nil, errInvalidType
	}
}

type njsonEncoder struct{}

func (njsonEncoder) encode(out []byte, v reflect.Value) ([]byte, error) {
	return v.Interface().(njson.Appender).AppendJSON(out)
}

type jsonEncoder struct{}

func (jsonEncoder) encode(out []byte, v reflect.Value) (b []byte, err error) {
	b, err = v.Interface().(json.Marshaler).MarshalJSON()
	if err == nil && b != nil {
		out = append(out, b...)
	}
	return out, err
}

type stringEncoder bool

func (HTML stringEncoder) encode(b []byte, v reflect.Value) ([]byte, error) {
	b = append(b, delimString)
	b = strjson.AppendEscaped(b, v.String(), bool(HTML))
	b = append(b, delimString)
	return b, nil
}

type interfaceEncoder struct {
	// options *Options
}

func (c interfaceEncoder) Encode(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return append(out, strNull...), nil
	}
	return c.encode(out, reflect.ValueOf(x))
}

func (c interfaceEncoder) encode(b []byte, v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return append(b, strNull...), nil
	}
	return MarshalTo(b, v.Interface())
}

type textEncoder struct{}

func (textEncoder) encode(out []byte, v reflect.Value) (text []byte, err error) {
	text, err = v.Interface().(encoding.TextMarshaler).MarshalText()
	if err == nil {
		out = append(out, delimString)
		out = append(out, text...)
		out = append(out, delimString)
	}
	return out, err
}

type mapEncoder struct {
	typ     reflect.Type
	encoder encoder
	keys    encoder
}

func newMapEncoder(typ reflect.Type, options *Options) (mapEncoder, error) {
	if typ.Kind() != reflect.Map {
		return mapEncoder{}, errInvalidType
	}

	var keys encoder
	if typ.Key().Implements(typTextMarshaler) {
		keys = textEncoder{}
	} else if typ.Key().Kind() == reflect.String {
		keys = stringEncoder(options.HTML)
	} else {
		return mapEncoder{}, errInvalidType
	}
	enc, err := newEncoder(typ.Elem(), options)
	if err != nil {
		return mapEncoder{}, err
	}
	return mapEncoder{
		typ:     typ,
		keys:    keys,
		encoder: enc,
	}, nil
}

func (d mapEncoder) encode(out []byte, v reflect.Value) ([]byte, error) {
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

type ptrEncoder struct {
	encoder encoder
}

func newPtrEncoder(typ reflect.Type, options *Options) (ptrEncoder, error) {
	enc, err := newEncoder(typ.Elem(), options)
	if err != nil {
		return ptrEncoder{}, err
	}
	return ptrEncoder{
		// typ:     typ.Elem(),
		encoder: enc,
	}, nil
}

func (d ptrEncoder) encode(b []byte, v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return append(b, strNull...), nil
	}
	return d.encoder.encode(b, v.Elem())
}

type sliceEncoder struct {
	encoder encoder
}

func newSliceEncoder(typ reflect.Type, options *Options) (sliceEncoder, error) {
	enc, err := newEncoder(typ.Elem(), options)
	if err != nil {
		return sliceEncoder{}, err
	}
	return sliceEncoder{
		encoder: enc,
	}, nil
}

func (e sliceEncoder) encode(out []byte, v reflect.Value) ([]byte, error) {
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
			out, err = e.encoder.encode(out, v.Index(i))
			if err != nil {
				return out, err
			}
		}
	}
	out = append(out, delimEndArray)
	return out, nil
}

type uintEncoder struct{}

func (uintEncoder) encode(b []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendUint(b, v.Uint(), 10), nil
}

type intEncoder struct{}

func (intEncoder) encode(b []byte, v reflect.Value) ([]byte, error) {
	return strconv.AppendInt(b, v.Int(), 10), nil
}

type boolEncoder struct{}

func (boolEncoder) encode(b []byte, v reflect.Value) ([]byte, error) {
	if v.Bool() {
		return append(b, strTrue...), nil
	}
	return append(b, strFalse...), nil
}

type floatEncoder struct {
	bits      int
	precision int
	allowInf  bool
	allowNan  bool
}

func newFloatEncoder(bits int, options *Options) (floatEncoder, error) {
	e := floatEncoder{bits, -1, options.AllowInf, options.AllowNaN}
	// if options.FloatPrecision > 0 {
	// 	e.precision = options.FloatPrecision
	// }
	return e, nil
}

func (e floatEncoder) encode(out []byte, v reflect.Value) ([]byte, error) {
	f := v.Float()
	if math.IsInf(f, 0) {
		if e.allowNan {
			if math.IsInf(f, -1) {
				return append(out, "-Inf"...), nil
			}
			return append(out, "+Inf"...), nil
		}
		return out, errValue
	}
	if math.IsNaN(f) {
		if e.allowNan {
			return append(out, "NaN"...), nil
		}
		return out, errValue
	}
	abs := math.Abs(f)
	fmt := byte('f')
	if (e.bits == 64 && (abs < 1e-6 || abs >= 1e21)) ||
		(e.bits == 32 && (float32(abs) < 1e-6 || float32(abs) > 1e21)) {
		fmt = 'e'
	}
	out = strconv.AppendFloat(out, f, fmt, e.precision, e.bits)
	if fmt == 'e' {
		if i := len(out) - 4; 0 <= i && i < len(out) {
			if buf := out[i:]; len(buf) == 4 && buf[0] == 'e' && buf[1] == '-' && buf[2] == '0' {
				buf[2] = buf[3]
				if i = len(out) - 1; 0 <= i && i < len(out) {
					out = out[:i]
				}
			}
		}
	}
	return out, nil
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

package njson

import (
	"encoding"
	"encoding/json"
	"reflect"
	"sync"
)

func Marshal(x interface{}) ([]byte, error) {
	return MarshalTo(nil, x)
}

func MarshalTo(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return append(out, strNull...), nil
	}
	enc, err := cachedEncoder(reflect.TypeOf(x), defaultOptions)
	if err != nil {
		return nil, err
	}
	return enc.Encode(out, x)
}

// Encoder is a type specific encoder
type Encoder interface {
	Encode(out []byte, x interface{}) ([]byte, error)
	encoder // disallow external implementations
}

type Marshaler interface {
	AppendJSON([]byte) ([]byte, error)
}

type Omiter interface {
	Omit() bool
}

type encoder interface {
	encode(out []byte, v reflect.Value) ([]byte, error)
}

type typeEncoder struct {
	encoder
	typ reflect.Type
}

var (
	typMarshaler     = reflect.TypeOf((*Marshaler)(nil)).Elem()
	typOmiter        = reflect.TypeOf((*Omiter)(nil)).Elem()
	typJSONMarshaler = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	typTextMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func (c *typeEncoder) Encode(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return out, errInvalidValueType
	}
	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Type() != c.typ {
		return out, errInvalidValueType
	}
	return c.encode(out, v)
}

func TypeEncoder(typ reflect.Type, options CodecOptions) (Encoder, error) {
	options = options.normalize()
	if typ == nil {
		return interfaceCodec{options}, nil
	}

	return cachedEncoder(typ, options)
}

func newTypeEncoder(typ reflect.Type, options CodecOptions) (*typeEncoder, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	c := typeEncoder{}
	if typ.Kind() == reflect.Ptr {
		c.typ = typ.Elem()
	} else {
		c.typ = typ
		typ = reflect.PtrTo(typ)
	}
	switch {
	case c.typ.Implements(typMarshaler):
		c.encoder = customEncoder{}
	case c.typ.Implements(typJSONMarshaler):
		c.encoder = customJSONEncoder{}
	default:
		d, err := newEncoder(c.typ, options)
		if err != nil {
			return nil, err
		}
		c.encoder = d
	}
	return &c, nil
}

func newEncoder(typ reflect.Type, options CodecOptions) (encoder, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	switch {
	case typ.Implements(typMarshaler):
		return customEncoder{}, nil
	case typ.Implements(typJSONMarshaler):
		return customJSONEncoder{}, nil
	case typ.Implements(typTextMarshaler):
		return textCodec{}, nil
	}

	return newCodec(typ, options)
}

var (
	encoderCacheLock sync.RWMutex
	encoderCache     = map[cacheKey]Encoder{}
)

func cachedEncoder(typ reflect.Type, options CodecOptions) (d Encoder, err error) {
	if typ == nil {
		return interfaceCodec{options}, nil
	}
	key := cacheKey{typ, options}
	encoderCacheLock.RLock()
	d, ok := encoderCache[key]
	encoderCacheLock.RUnlock()
	if ok {
		return
	}
	if d, err = newTypeEncoder(typ, DefaultOptions()); err != nil {
		return
	}
	encoderCacheLock.Lock()
	encoderCache[key] = d
	encoderCacheLock.Unlock()
	return
}

type customEncoder struct{}

func (customEncoder) encode(out []byte, v reflect.Value) ([]byte, error) {
	return v.Interface().(Marshaler).AppendJSON(out)
}

type customJSONEncoder struct{}

func (customJSONEncoder) encode(out []byte, v reflect.Value) (b []byte, err error) {
	b, err = v.Interface().(json.Marshaler).MarshalJSON()
	if err == nil && b != nil {
		out = append(out, b...)
	}
	return out, err
}

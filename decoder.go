package njson

import (
	"encoding"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
)

func Unmarshal(data []byte, x interface{}) error {
	return UnmarshalFromString(string(data), x)
}

func UnmarshalFromString(s string, x interface{}) error {
	if x == nil {
		return errInvalidValueType
	}
	dec, err := cachedDecoder(reflect.TypeOf(x))
	if err != nil {
		return err
	}
	return dec.DecodeString(x, s)
}

type Unmarshaler interface {
	UnmarshalNodeJSON(*Node) error
}

var (
	typUnmarshaler     = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
	typJSONUnmarshaler = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	typTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// Decoder is a type specific decoder
type Decoder interface {
	Decode(x interface{}, n *Node) error
	DecodeString(x interface{}, src string) error
	decoder // disallow external implementations
}

type decoder interface {
	decode(v reflect.Value, n *Node) error
}

type typeDecoder struct {
	decoder
	typ reflect.Type // PtrTo(typ)
}

func (c *typeDecoder) Decode(x interface{}, n *Node) error {
	if x == nil {
		return errInvalidValueType
	}
	v := reflect.ValueOf(x)
	if v.Type() != c.typ {
		return errInvalidValueType
	}
	if v.IsNil() {
		if n.Type() == TypeNull {
			return nil
		}
		v.Set(reflect.New(c.typ.Elem()))
	}
	return c.decode(v.Elem(), n)
}

func (c *typeDecoder) DecodeString(x interface{}, src string) (err error) {
	p := parsePool.Get().(*parsePair)
	p.Reset()
	if _, err = p.Parse(src, &p.Document); err == nil {
		err = c.Decode(x, p.Get(0))
	}
	parsePool.Put(p)
	return
}

var (
	errInvalidValueType = errors.New("Invalid value type")
	errInvalidNodeType  = errors.New("Invalid node type")
	errInvalidType      = errors.New("Invalid type")
)

// customDecoder implements the Decoder interface for types implementing Unmarshaller
type customDecoder struct{}

func (customDecoder) Decode(x interface{}, n *Node) error {
	if x, ok := x.(Unmarshaler); ok {
		return x.UnmarshalNodeJSON(n)
	}
	return errInvalidValueType
}

func (customDecoder) DecodeString(x interface{}, src string) (err error) {
	if x, ok := x.(Unmarshaler); ok {
		p := parsePool.Get().(*parsePair)
		p.Reset()
		if _, err = p.Parse(src, &p.Document); err == nil {
			err = x.UnmarshalNodeJSON(p.Get(0))
		}
		parsePool.Put(p)
		return
	}
	return errInvalidValueType
}

func (customDecoder) decode(v reflect.Value, tok *Node) error {
	return v.Interface().(Unmarshaler).UnmarshalNodeJSON(tok)
}

// customJSONDecoder implements the Decoder interface for types implementing json.Unmarshaller
type customJSONDecoder struct{}

func (customJSONDecoder) Decode(x interface{}, n *Node) (err error) {
	if u, ok := x.(json.Unmarshaler); ok {
		if n.src != "" {
			return u.UnmarshalJSON(s2b(n.src))
		}
		b := bufferpool.Get().([]byte)
		b = n.AppendTo(b[:0])
		err = u.UnmarshalJSON(b)
		bufferpool.Put(b)
		return
	}
	return errInvalidValueType
}

func (customJSONDecoder) DecodeString(x interface{}, src string) error {
	if x, ok := x.(json.Unmarshaler); ok {
		return x.UnmarshalJSON(s2b(src))
	}
	return errInvalidValueType
}

func (customJSONDecoder) decode(v reflect.Value, n *Node) (err error) {
	if n.src != "" {
		return v.Interface().(json.Unmarshaler).UnmarshalJSON(s2b(n.src))
	}
	b := bufferpool.Get().([]byte)
	b = n.AppendTo(b[:0])
	err = v.Interface().(json.Unmarshaler).UnmarshalJSON(b)
	bufferpool.Put(b)
	return
}

func TypeDecoder(typ reflect.Type, options *CodecOptions) (Decoder, error) {
	if options == nil {
		return cachedDecoder(typ)
	}
	if typ == nil {
		return interfaceCodec{}, nil
	}
	return newTypeDecoder(typ, *options)
}

func newTypeDecoder(typ reflect.Type, options CodecOptions) (*typeDecoder, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	c := typeDecoder{}
	if typ.Kind() == reflect.Ptr {
		c.typ = typ
		typ = typ.Elem()
	} else {
		c.typ = reflect.PtrTo(typ)
	}
	switch {
	case c.typ.Implements(typUnmarshaler):
		c.decoder = customDecoder{}
	case c.typ.Implements(typJSONUnmarshaler):
		c.decoder = customJSONDecoder{}
	default:
		d, err := newDecoder(typ, options)
		if err != nil {
			return nil, err
		}
		c.decoder = d
	}
	return &c, nil
}

func newDecoder(typ reflect.Type, options CodecOptions) (decoder, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	switch {
	case typ.Implements(typUnmarshaler):
		return customJSONDecoder{}, nil
	case typ.Implements(typJSONUnmarshaler):
		return customJSONDecoder{}, nil
	case typ.Implements(typTextUnmarshaler):
		return textCodec{}, nil
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
			return interfaceCodec{}, nil
		}
		return nil, errInvalidType
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intCodec{}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uintCodec{}, nil
	case reflect.Float32, reflect.Float64:
		return floatCodec{}, nil
	case reflect.Bool:
		return boolCodec{}, nil
	case reflect.String:
		return stringCodec{}, nil
	default:
		return nil, errInvalidType
	}
}

var (
	decoderCacheLock sync.RWMutex
	decoderCache     = map[reflect.Type]Decoder{}
)

func cachedDecoder(typ reflect.Type) (d Decoder, err error) {
	if typ == nil {
		return interfaceCodec{}, nil
	}
	decoderCacheLock.RLock()
	d, ok := decoderCache[typ]
	decoderCacheLock.RUnlock()
	if ok {
		return
	}
	if d, err = newTypeDecoder(typ, DefaultOptions()); err != nil {
		return
	}
	decoderCacheLock.Lock()
	decoderCache[typ] = d
	decoderCacheLock.Unlock()
	return
}

type parsePair struct {
	Document
	DocumentParser
}

var parsePool = sync.Pool{
	New: func() interface{} {
		p := parsePair{
			Document: Document{
				nodes: make([]Node, 0, 64),
			},
			DocumentParser: DocumentParser{
				stack: make([]uint16, 0, 64),
			},
		}
		return &p
	},
}

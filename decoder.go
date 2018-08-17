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
	dec, err := cachedDecoder(reflect.TypeOf(x), defaultOptions)
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

// Decode implements the Decoder interface.
// It handles the case of a x being a nil pointer by creating a new blank value.
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

func TypeDecoder(typ reflect.Type, options CodecOptions) (Decoder, error) {
	options = options.normalize()
	if typ == nil {
		return interfaceCodec{options}, nil
	}
	return newTypeDecoder(typ, options)
}

func newTypeDecoder(typ reflect.Type, options CodecOptions) (*typeDecoder, error) {
	if typ == nil {
		return nil, errInvalidType
	}
	if typ.Kind() != reflect.Ptr {
		return nil, errInvalidType
	}
	c := typeDecoder{typ: typ}
	switch {
	case typ.Implements(typUnmarshaler):
		c.decoder = customDecoder{}
	case typ.Implements(typJSONUnmarshaler):
		c.decoder = customJSONDecoder{}
	default:
		typ = typ.Elem()
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
	return newCodec(typ, options)
}

type cacheKey struct {
	reflect.Type
	CodecOptions
}

var (
	decoderCacheLock sync.RWMutex
	decoderCache     = map[cacheKey]Decoder{}
)

func cachedDecoder(typ reflect.Type, options CodecOptions) (d Decoder, err error) {
	if typ == nil {
		return interfaceCodec{options}, nil
	}
	key := cacheKey{typ, options}
	decoderCacheLock.RLock()
	d, ok := decoderCache[key]
	decoderCacheLock.RUnlock()
	if ok {
		return
	}
	if d, err = newTypeDecoder(typ, options); err != nil {
		return
	}
	decoderCacheLock.Lock()
	decoderCache[key] = d
	decoderCacheLock.Unlock()
	return
}

type parsePair struct {
	Document
	Parser
}

var parsePool = sync.Pool{
	New: func() interface{} {
		p := parsePair{
			Document: Document{
				nodes: make([]Node, 0, 64),
			},
			Parser: Parser{
				stack: make([]uint16, 0, 64),
			},
		}
		return &p
	},
}

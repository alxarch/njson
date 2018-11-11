package unjson

import (
	"reflect"
	"sync"
)

type cacheKey struct {
	typ     reflect.Type
	options Options
}

var (
	unmarshalCacheLock sync.RWMutex
	unmarshalCache     = map[cacheKey]Decoder{}
	marshalCacheLock   sync.RWMutex
	marshalCache       = map[cacheKey]Encoder{}
)

type cache map[reflect.Type]interface{}

func (c cache) codec(typ reflect.Type) *structCodec {
	if x := c[typ]; x != nil {
		if c, ok := x.(*structCodec); ok {
			return c
		}
	}
	return nil
}
func (c cache) encoder(typ reflect.Type, options *Options) (encoder, error) {
	if x := c[typ]; x != nil {
		if e, ok := x.(encoder); ok {
			return e, nil
		}
	}
	enc, err := newEncoder(typ, options, c)
	if err != nil {
		return nil, err
	}
	c[typ] = enc
	return enc, nil
}
func (c cache) decoder(typ reflect.Type, options *Options) (decoder, error) {
	if x := c[typ]; x != nil {
		if e, ok := x.(decoder); ok {
			return e, nil
		}
	}
	dec, err := newDecoder(typ, options, c)
	if err != nil {
		return nil, err
	}
	c[typ] = dec
	return dec, nil
}

func cachedDecoder(typ reflect.Type, options *Options) (u Decoder, err error) {
	if typ == nil {
		return interfaceDecoder{}, nil
	}
	key := cacheKey{typ, defaultOptions}
	if options == nil {
		options = &defaultOptions
	} else {
		key.options = *options
	}
	unmarshalCacheLock.RLock()
	u, ok := unmarshalCache[key]
	unmarshalCacheLock.RUnlock()
	if ok {
		return
	}
	if u, err = newTypeDecoder(typ, options); err != nil {
		return
	}
	unmarshalCacheLock.Lock()
	unmarshalCache[key] = u
	unmarshalCacheLock.Unlock()
	return
}

func cachedEncoder(typ reflect.Type, options *Options) (m Encoder, err error) {
	if typ == nil {
		return interfaceEncoder{}, nil
	}
	key := cacheKey{typ, defaultOptions}
	if options == nil {
		options = &defaultOptions
	} else {
		key.options = *options
	}

	marshalCacheLock.RLock()
	m, ok := marshalCache[key]
	marshalCacheLock.RUnlock()
	if ok {
		return
	}
	if m, err = newTypeEncoder(typ, options); err != nil {
		return
	}
	marshalCacheLock.Lock()
	marshalCache[key] = m
	marshalCacheLock.Unlock()
	return
}

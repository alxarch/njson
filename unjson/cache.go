package unjson

import (
	"reflect"
	"sync"
)

// Cache is an Encoder/Decoder cache for specific options
type Cache struct {
	Options
	mu       sync.RWMutex
	decoders map[reflect.Type]Decoder
	encoders map[reflect.Type]Encoder
}

var defaultCache Cache

// Decoder returns a Decoder for the type using the tag key from cache.Options
func (c *Cache) Decoder(typ reflect.Type) (dec Decoder, err error) {
	c.mu.RLock()
	dec = c.decoders[typ]
	c.mu.RUnlock()
	if dec != nil {
		return
	}
	dec, err = NewTypeDecoder(typ, c.Tag)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if d := c.decoders[typ]; d != nil {
		c.mu.Unlock()
		return d, nil
	}
	if c.decoders == nil {
		c.decoders = make(map[reflect.Type]Decoder)
	}
	c.decoders[typ] = dec
	c.mu.Unlock()
	return
}

// Encoder returns an Encoder for the type using cache.Options
func (c *Cache) Encoder(typ reflect.Type) (enc Encoder, err error) {
	c.mu.RLock()
	enc = c.encoders[typ]
	c.mu.RUnlock()
	if enc != nil {
		return
	}
	enc, err = NewTypeEncoder(typ, c.Options)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if e := c.encoders[typ]; e != nil {
		c.mu.Unlock()
		return e, nil
	}
	if c.encoders == nil {
		c.encoders = make(map[reflect.Type]Encoder)
	}
	c.encoders[typ] = enc
	c.mu.Unlock()
	return
}

// cache is used when creating new encoders/decoders to not recalculate stuff and avoid recursion issues.
type cache map[cacheKey]interface{}

// cacheKey is based on both Type and hint so types with different hints get correct encoders
type cacheKey struct {
	typ   reflect.Type
	hints hint
}

func (c cache) codec(typ reflect.Type) *structCodec {
	key := cacheKey{typ, 0}
	if x := c[key]; x != nil {
		if c, ok := x.(*structCodec); ok {
			return c
		}
	}
	return nil
}
func cacheType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array:
		return cacheType(typ.Elem())
	case reflect.Struct, reflect.Map:
		return true
	default:
		return false
	}
}
func (c cache) encoder(typ reflect.Type, options *Options, hints hint) (encoder, error) {
	key := cacheKey{typ, hints}
	if x := c[key]; x != nil {
		if e, ok := x.(encoder); ok {
			return e, nil
		}
	}
	enc, err := newEncoder(typ, options, hints, c)
	if err != nil {
		return nil, err
	}
	c[key] = enc
	return enc, nil
}

func (c cache) decoder(typ reflect.Type, options *Options) (decoder, error) {
	key := cacheKey{typ, 0}
	if x := c[key]; x != nil {
		if e, ok := x.(decoder); ok {
			return e, nil
		}
	}
	dec, err := newDecoder(typ, options, c)
	if err != nil {
		return nil, err
	}
	c[key] = dec
	return dec, nil
}

package unjson

import (
	"reflect"
	"sync"
)

type cacheKey struct {
	typ     reflect.Type
	options uint64
}

var (
	unmarshalCacheLock sync.RWMutex
	unmarshalCache     = map[cacheKey]Decoder{}
	marshalCacheLock   sync.RWMutex
	marshalCache       = map[cacheKey]Encoder{}
	codecCacheLock     sync.RWMutex
	codecCache         = map[cacheKey]dencoder{}
)

type dencoder interface {
	encoder
	decoder
}

func cachedCodec(typ reflect.Type, options *Options) (c dencoder, err error) {
	if typ == nil || typ.Kind() != reflect.Struct {
		return nil, errInvalidType
	}
	key := cacheKey{typ, defaultOptionsHash}
	if options == nil {
		options = &defaultOptions
	} else {
		key.options = options.hash()
	}
	codecCacheLock.RLock()
	c, ok := codecCache[key]
	codecCacheLock.RUnlock()
	if ok {
		return
	}
	if c, err = newStructCodec(typ, options); err != nil {
		return
	}
	codecCacheLock.Lock()
	codecCache[key] = c
	codecCacheLock.Unlock()
	return
}

func cachedDecoder(typ reflect.Type, options *Options) (u Decoder, err error) {
	if typ == nil {
		return interfaceDecoder{}, nil
	}
	key := cacheKey{typ, defaultOptionsHash}
	if options == nil {
		options = &defaultOptions
	} else {
		key.options = options.hash()
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
	key := cacheKey{typ, defaultOptionsHash}
	if options == nil {
		options = &defaultOptions
	} else {
		key.options = options.hash()
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

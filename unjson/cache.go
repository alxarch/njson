package unjson

import (
	"reflect"
	"sync"
)

type cacheKey struct {
	reflect.Type
	Options
}

var (
	unmarshalCacheLock sync.RWMutex
	unmarshalCache     = map[cacheKey]Unmarshaler{}
	marshalCacheLock   sync.RWMutex
	marhsalCache       = map[cacheKey]Marshaler{}
)

func cachedUnmarshaler(typ reflect.Type, options Options) (u Unmarshaler, err error) {
	if typ == nil {
		return interfaceCodec{options}, nil
	}
	key := cacheKey{typ, options}
	unmarshalCacheLock.RLock()
	u, ok := unmarshalCache[key]
	unmarshalCacheLock.RUnlock()
	if ok {
		return
	}
	if u, err = newTypeUnmarshaler(typ, options); err != nil {
		return
	}
	unmarshalCacheLock.Lock()
	unmarshalCache[key] = u
	unmarshalCacheLock.Unlock()
	return
}

func cachedMarshaler(typ reflect.Type, options Options) (m Marshaler, err error) {
	if typ == nil {
		return interfaceCodec{options}, nil
	}
	key := cacheKey{typ, options}
	marshalCacheLock.RLock()
	m, ok := marhsalCache[key]
	marshalCacheLock.RUnlock()
	if ok {
		return
	}
	if m, err = newTypeMarshaler(typ, DefaultOptions()); err != nil {
		return
	}
	marshalCacheLock.Lock()
	marhsalCache[key] = m
	marshalCacheLock.Unlock()
	return
}

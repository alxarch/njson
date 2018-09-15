// Package unjson uses reflection to marshal/unmarshal JSON from njson.Node input
// It's NOT a 'drop-in' replacement for "encoding/json".
package unjson

import (
	"reflect"

	"github.com/alxarch/njson"
)

func Unmarshal(data []byte, x interface{}) error {
	return UnmarshalFromString(string(data), x)
}

func UnmarshalUnsafe(data []byte, x interface{}) (err error) {
	if x == nil {
		return errInvalidValueType
	}
	d, err := cachedDecoder(reflect.TypeOf(x), nil)
	if err != nil {
		return
	}
	p := njson.Blank()
	n, data, err := p.ParseUnsafe(data)
	if err == nil {
		err = d.Decode(x, n)
	}
	p.Close()
	return
}
func UnmarshalFromNode(n njson.Node, x interface{}) error {
	if x == nil {
		return errInvalidValueType
	}
	dec, err := cachedDecoder(reflect.TypeOf(x), nil)
	if err != nil {
		return err
	}
	return dec.Decode(x, n)
}

func UnmarshalFromString(s string, x interface{}) (err error) {
	if x == nil {
		return errInvalidValueType
	}
	d, err := cachedDecoder(reflect.TypeOf(x), nil)
	if err != nil {
		return
	}
	p := njson.Blank()
	n, s, err := p.Parse(s)
	if err == nil {
		err = d.Decode(x, n)
	}
	p.Close()
	return
}

func Marshal(x interface{}) ([]byte, error) {
	return MarshalTo(nil, x)
}

func MarshalTo(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return append(out, strNull...), nil
	}
	m, err := cachedEncoder(reflect.TypeOf(x), nil)
	if err != nil {
		return nil, err
	}
	return m.Encode(out, x)
}

// Package unjson uses reflection to marshal/unmarshal JSON from njson.Node input
// It's NOT a 'drop-in' replacement for "encoding/json".
package unjson

import (
	"reflect"

	"github.com/alxarch/njson"
)

// Unmarshal behaves like json.Unmarshal.
// It copies the data to string.
// To avoid allocations use UnmarshalFromString or UnmarshalFromNode
func Unmarshal(data []byte, x interface{}) error {
	return UnmarshalFromString(string(data), x)
}

// UnmarshalFromNode unmarshals from an njson.Node
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

// UnmarshalFromString unmarshals from a JSON string
func UnmarshalFromString(s string, x interface{}) (err error) {
	if x == nil {
		return errInvalidValueType
	}
	d, err := cachedDecoder(reflect.TypeOf(x), nil)
	if err != nil {
		return
	}
	p := njson.Blank()
	n, _, err := p.Parse(s)
	if err == nil {
		err = d.Decode(x, n)
	}
	p.Close()
	return
}

// Marshal behaves like json.Marshal
// To avoid allocations use AppendJSON
func Marshal(x interface{}) ([]byte, error) {
	return AppendJSON(nil, x)
}

// AppendJSON appends the JSON encoding of a value to a buffer
func AppendJSON(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return append(out, strNull...), nil
	}
	m, err := cachedEncoder(reflect.TypeOf(x), nil)
	if err != nil {
		return nil, err
	}
	return m.Encode(out, x)
}

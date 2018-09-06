// Package unjson uses reflection to marshal/unmarshal JSON from njson.Node input
// It's NOT a 'drop-in' replacement for "encoding/json".
// Specifically due to the nature of the njson parser it does not (yet) support NewEncoder/NewDecoder
// for streaming JSON objects over an io.Reader.
// At best one could use it with newline delimited JSON streams (http://ndjson.org/) in combination with a bufio.Scanner
// Performance is better than "encoding/json" but for best results use njson command to generate UnmarshalNodeJSON methods.
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
	u, err := cachedUnmarshaler(reflect.TypeOf(x), defaultOptions)
	if err != nil {
		return
	}
	d := njson.Get()
	n, data, err := d.ParseUnsafe(data)
	if err == nil {
		err = u.Unmarshal(x, n)
	}
	d.Close()
	return
}
func UnmarshalFromNode(n *njson.Node, x interface{}) error {
	if x == nil {
		return errInvalidValueType
	}
	u, err := cachedUnmarshaler(reflect.TypeOf(x), defaultOptions)
	if err != nil {
		return err
	}
	return u.Unmarshal(x, n)
}

func UnmarshalFromString(s string, x interface{}) (err error) {
	if x == nil {
		return errInvalidValueType
	}
	u, err := cachedUnmarshaler(reflect.TypeOf(x), defaultOptions)
	if err != nil {
		return
	}
	d := njson.Get()
	n, s, err := d.Parse(s)
	if err == nil {
		err = u.Unmarshal(x, n)
	}
	d.Close()
	return
}

func Marshal(x interface{}) ([]byte, error) {
	return MarshalTo(nil, x)
}

func MarshalTo(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return append(out, strNull...), nil
	}
	m, err := cachedMarshaler(reflect.TypeOf(x), defaultOptions)
	if err != nil {
		return nil, err
	}
	return m.MarshalTo(out, x)
}

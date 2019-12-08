// Package unjson uses reflection to marshal/unmarshal JSON from `njson.Node` input
//
// It is combatible with the default `encoding/json` package but the API is different
// so it's *not* a 'drop-in' replacement.
//
// ## Decoding
//
// Values are decoded from JSON by `Decoder` instances.
// `Decoder` iterates over an `njson.Node` instance to decode a value
// instead of reading directly from an `io.Reader` as in `encoding/json`.
// This decouples decoding and parsing of a JSON document.
// A parsed JSON document can be thus decoded multiple times to produce
// different values or using different decoding options.
//
// To decode streams of JSON values from an `io.Reader` use `LineDecoder`.
//
// To decode a value from JSON you need to create a `Decoder` for it's type.
// A package-wide decoder registry using the default `json` tag is used to
// decode values with the `Unmarshal` method.
//
//
// ## Encoding
//
// Values are encoded to JSON by `Encoder` instances.
// `Encoder` instances provide an API that appends to a byte slice
// instead of writing to an `io.Writer` as in `encoding/json`.
// To encode streams of JSON values to an `io.Writer` use `LineEncoder`.
package unjson

import (
	"reflect"

	"github.com/alxarch/njson"
)

// Unmarshal is a drop-in replacement for json.Unmarshal.
//
// It delegates to `UnmarshalFromJSON` by allocating a new string.
// To avoid allocations use `UnmarshalFromString` or `UnmarshalFromNode`
func Unmarshal(data []byte, x interface{}) error {
	return UnmarshalFromString(string(data), x)
}

// UnmarshalFromNode unmarshals from an njson.Node
//
// It uses a package-wide cache of `Decoder` instances using the default options.
// In order to use custom options for a `Decoder` and avoid lock congestion
// on the registry, create a local `Decoder` instance with `NewTypeDecoder`
func UnmarshalFromNode(n njson.Node, x interface{}) error {
	if x == nil {
		return errInvalidValueType
	}
	dec, err := defaultCache.Decoder(reflect.TypeOf(x))
	if err != nil {
		return err
	}
	return dec.Decode(x, n)
}

// UnmarshalFromString unmarshals from a JSON string
//
// It borrows a blank `njson.Document` from `njson.Blank` to parse
// the JSON string and a `Decoder` with the default options from package-wide
// cache of `Dec`Decoder` instances.
func UnmarshalFromString(s string, x interface{}) (err error) {
	if x == nil {
		return errInvalidValueType
	}
	d, err := defaultCache.Decoder(reflect.TypeOf(x))
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

// Marshal is a drop-in replacement for json.Marshal.
//
// It delegates to `AppendJSON`, allocating a new buffer.
// To avoid allocations use `AppendJSON` and provide a buffer yourself.
func Marshal(x interface{}) ([]byte, error) {
	return AppendJSON(nil, x)
}

// AppendJSON appends the JSON encoding of a value to a buffer
//
// It uses a package-wide cache of `Encoder` instances using the default options.
// In order to use custom options for an `Encoder` and avoid lock congestion
// on the registry, create a local `Encoder` instance with `NewTypeEncoder`
func AppendJSON(out []byte, x interface{}) ([]byte, error) {
	if x == nil {
		return append(out, strNull...), nil
	}
	m, err := defaultCache.Encoder(reflect.TypeOf(x))
	if err != nil {
		return nil, err
	}
	return m.Encode(out, x)
}

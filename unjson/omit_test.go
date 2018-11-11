package unjson

import (
	"reflect"
	"testing"
)

func TestOmit(t *testing.T) {
	type Foo struct {
		Bar     string            `json:"bar,omitempty"`
		Baz     []string          `json:"baz,omitempty"`
		Foo     bool              `json:"foo,omitempty"`
		Int     int               `json:"int,omitempty"`
		Float   float64           `json:"float,omitempty"`
		Uint    uint              `json:"uint,omitempty"`
		Map     map[string]string `json:"map,omitempty"`
		Pointer *Foo              `json:"pointer,omitempty"`
	}
	foo := new(Foo)
	enc, err := TypeEncoder(reflect.TypeOf(foo), DefaultOptions())
	assertNoError(t, err)
	data, err := enc.Encode(nil, foo)
	assertNoError(t, err)
	assertEqual(t, string(data), `{}`)

}

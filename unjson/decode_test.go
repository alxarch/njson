package unjson

import (
	"reflect"
	"testing"

	"github.com/alxarch/njson"
)

func Test_codec_SelfRef(t *testing.T) {
	type A struct {
		B *A
	}
	a := new(A)
	typ := reflect.TypeOf(a)
	c, err := newStructCodec(typ, &defaultOptions, cache{})
	assertNoError(t, err)
	d := njson.Document{}
	node, _, _ := d.Parse(`{"B": {}}`)
	v := reflect.ValueOf(a)
	c.decode(v.Elem(), node)
	assert(t, a.B != nil, "Nil B")

}

func Test_codec_SelfSlice(t *testing.T) {
	type A struct {
		B []A
	}
	a := new(A)
	typ := reflect.TypeOf(a)
	c, err := newStructCodec(typ, &defaultOptions, cache{})
	assertNoError(t, err)
	d := njson.Document{}
	node, _, _ := d.Parse(`{"B": []}`)
	v := reflect.ValueOf(a)
	c.decode(v.Elem(), node)
	assert(t, a.B != nil, "Nil B")

}

func Test_codec_SelfRefMap(t *testing.T) {
	type A map[string]A
	a := A{}
	typ := reflect.TypeOf(&a)
	c, err := newDecoder(typ, &defaultOptions, cache{})
	assertNoError(t, err)
	d := njson.Document{}
	node, _, _ := d.Parse(`{"B": {}}`)
	v := reflect.ValueOf(&a)
	c.decode(v, node)
	_, ok := a["B"]
	assert(t, ok, "Nil B")

}

package unjson

import (
	"testing"
)

func TestMarshal(t *testing.T) {
	v := struct {
		Foo string
		Bar string
	}{"foo", "bar"}
	data, err := AppendJSON(nil, v)
	assertNoError(t, err)
	assertEqual(t, string(data), `{"Foo":"foo","Bar":"bar"}`)
}

func TestMarshalPtr(t *testing.T) {
	v := struct {
		Foo string `json:"foo"`
		Bar string
	}{"foo", "bar"}
	data, err := AppendJSON(nil, &v)
	assertNoError(t, err)
	assertEqual(t, string(data), `{"foo":"foo","Bar":"bar"}`)
}

func TestMarshalInterfaceField(t *testing.T) {
	v := struct {
		Foo string
		Bar interface{}
	}{"foo", "bar"}
	data, err := AppendJSON(nil, &v)
	assertNoError(t, err)
	assertEqual(t, string(data), `{"Foo":"foo","Bar":"bar"}`)
}

func TestMarshalInterface(t *testing.T) {
	v := struct {
		Foo string
		Bar interface{}
	}{"foo", "bar"}
	var x interface{} = &v

	data, err := AppendJSON(nil, &x)
	assertNoError(t, err)
	assertEqual(t, string(data), `{"Foo":"foo","Bar":"bar"}`)
}

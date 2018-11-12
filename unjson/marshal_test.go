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
func TestMarshalBasicTypes(t *testing.T) {
	v := struct {
		String string
		Uint   uint
		Int    int
		Bool   bool
		Float  float64
		Null   interface{}
		Map    map[string]string
		Slice  []int
	}{"foo", 1, -1, true, 0.02, nil, map[string]string{"foo": "bar"}, []int{1, 2, 3}}
	data, err := AppendJSON(nil, v)
	assertNoError(t, err)
	assertEqual(t, string(data), `{"String":"foo","Uint":1,"Int":-1,"Bool":true,"Float":0.02,"Null":null,"Map":{"foo":"bar"},"Slice":[1,2,3]}`)
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
	x = map[string]interface{}{
		"foo": "foo",
		"bar": []int{1, 2, 3},
	}
	data, err = Marshal(x)
	assertNoError(t, err)
	switch string(data) {
	case `{"foo":"foo","bar":[1,2,3]}`:
	case `{"bar":[1,2,3],"foo":"foo"}`:
	default:
		t.Fatalf("Invalid data %s", data)
	}
}

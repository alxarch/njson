package njson_test

import (
	"encoding/json"
	"testing"

	"github.com/alxarch/njson"
)

func TestMarshal(t *testing.T) {
	v := struct {
		Foo string
		Bar string
	}{"foo", "bar"}
	data, err := njson.MarshalTo(nil, v)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	switch string(data) {
	case `{"Bar":"bar","Foo":"foo"}`, `{"Foo":"foo","Bar":"bar"}`:
	default:
		expect, _ := json.Marshal(v)
		t.Errorf("Invalid marshal:\nactual: %s\nexpect: %s", data, expect)
		return
	}
}

func TestMarshalPtr(t *testing.T) {
	v := struct {
		Foo string
		Bar string
	}{"foo", "bar"}
	data, err := njson.MarshalTo(nil, &v)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	switch string(data) {
	case `{"Bar":"bar","Foo":"foo"}`, `{"Foo":"foo","Bar":"bar"}`:
	default:
		expect, _ := json.Marshal(v)
		t.Errorf("Invalid marshal:\nactual: %s\nexpect: %s", data, expect)
		return
	}
}

func TestMarshalInterfaceField(t *testing.T) {
	v := struct {
		Foo string
		Bar interface{}
	}{"foo", "bar"}
	data, err := njson.MarshalTo(nil, &v)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	switch string(data) {
	case `{"Bar":"bar","Foo":"foo"}`, `{"Foo":"foo","Bar":"bar"}`:
	default:
		expect, _ := json.Marshal(v)
		t.Errorf("Invalid marshal:\nactual: %s\nexpect: %s", data, expect)
		return
	}
}

func TestMarshalInterface(t *testing.T) {
	v := struct {
		Foo string
		Bar interface{}
	}{"foo", "bar"}
	var x interface{} = &v

	data, err := njson.MarshalTo(nil, &x)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	switch string(data) {
	case `{"Bar":"bar","Foo":"foo"}`, `{"Foo":"foo","Bar":"bar"}`:
	default:
		expect, _ := json.Marshal(&v)
		t.Errorf("Invalid marshal:\nactual: %s\nexpect: %s", data, expect)
		return
	}
}

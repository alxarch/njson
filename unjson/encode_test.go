package unjson

import (
	"errors"
	"reflect"
	"testing"
)

type customJSON struct{}

func (customJSON) MarshalJSON() ([]byte, error) {
	return []byte(`"custom"`), nil
}
func (customJSON) UnmarshalJSON(data []byte) error {
	if string(data) != `"custom"` {
		return errors.New("Invalid data " + string(data))

	}
	return nil
}

func TestCustomJSON(t *testing.T) {
	v := customJSON{}
	data, err := Marshal(v)
	assertNoError(t, err)
	assertEqual(t, string(data), `"custom"`)
	err = Unmarshal([]byte(`"custom"`), &v)
	assertNoError(t, err)
	err = Unmarshal([]byte(`"foo"`), &v)
	assert(t, err != nil, "Custom json unmarshaler didn't propagate error")
}

type customText string

func (t customText) MarshalText() ([]byte, error) {
	return []byte(string(t)), nil
}
func (t *customText) UnmarshalText(data []byte) error {
	*t = customText(string(data))
	return nil
}

func TestCustomText(t *testing.T) {
	v := customText("foo")
	dec, err := NewTypeDecoder(reflect.TypeOf(&v), "json")
	assertNoError(t, err)
	_, ok := dec.(textDecoder)
	assertEqual(t, ok, true)

	data, err := Marshal(v)
	assertNoError(t, err)
	assertEqual(t, string(data), `"foo"`)
	err = Unmarshal([]byte(`"bar"`), &v)
	assertNoError(t, err)
	assertEqual(t, string(v), `bar`)
	enc, err := NewTypeEncoder(reflect.TypeOf(v), DefaultOptions())
	assertNoError(t, err)
	// _, ok := dec.(textEncoder)
	// assertEqual(t, ok, true)
	data, err = enc.Encode(nil, customText("baz"))
	assertNoError(t, err)
	assertEqual(t, string(data), `"baz"`)

}

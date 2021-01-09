package unjson

import (
	"encoding/json"
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

type customJSONUnmarshal struct {
	Foo string `json:"foo"`
}

func (c *customJSONUnmarshal) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &c.Foo)
}

func TestDecodeJSONUnmarshaler(t *testing.T) {

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

func testDecoder(t *testing.T, input string, want interface{}, dec decoder, wantErr bool) {
	t.Helper()
	d := njson.Document{}
	n, _, err := d.Parse(input)
	assertNoError(t, err)
	w := reflect.ValueOf(want)
	v := reflect.New(w.Type())
	err = dec.decode(v.Elem(), n)
	if wantErr {
		assert(t, err != nil, "decode(%T): Expected decode error %s %v", dec, input, want)
	} else {
		assert(t, err == nil, "decode(%T): Unexpected decode error %s %s %v", dec, err, input, want)
		assertEqual(t, v.Elem().Interface(), want)
	}
}

func TestDecoderBool(t *testing.T) {
	testDecoder(t, "true", true, boolDecoder{}, false)
	testDecoder(t, "false", false, boolDecoder{}, false)
	testDecoder(t, "{}", false, boolDecoder{}, true)
	testDecoder(t, `null`, false, boolDecoder{}, false)
}
func TestDecoderFloat(t *testing.T) {
	testDecoder(t, "-1.2", float64(-1.2), floatDecoder{}, false)
	testDecoder(t, "1.2", float64(1.2), floatDecoder{}, false)
	testDecoder(t, "42", float64(42.0), floatDecoder{}, false)
	testDecoder(t, "0", float64(0.0), floatDecoder{}, false)
	testDecoder(t, `null`, float64(0), floatDecoder{}, false)

	testDecoder(t, "-1.2", float32(-1.2), floatDecoder{}, false)
	testDecoder(t, "1.2", float32(1.2), floatDecoder{}, false)
	testDecoder(t, "42", float32(42.0), floatDecoder{}, false)
	testDecoder(t, "0", float32(0.0), floatDecoder{}, false)
}

func TestDecoderInt(t *testing.T) {
	testDecoder(t, "42.0", int64(42.0), intDecoder{}, false)
	testDecoder(t, "-42", int64(-42), intDecoder{}, false)
	testDecoder(t, "0", int64(0.0), intDecoder{}, false)
	testDecoder(t, "-1.2", int64(0), intDecoder{}, true)
	testDecoder(t, "1.2", int64(0), intDecoder{}, true)

	testDecoder(t, "42.0", int32(42.0), intDecoder{}, false)
	testDecoder(t, "-42", int32(-42), intDecoder{}, false)
	testDecoder(t, "0", int32(0.0), intDecoder{}, false)
	testDecoder(t, "-1.2", int32(0), intDecoder{}, true)
	testDecoder(t, "1.2", int32(0), intDecoder{}, true)

	testDecoder(t, "42.0", int16(42.0), intDecoder{}, false)
	testDecoder(t, "-42", int16(-42), intDecoder{}, false)
	testDecoder(t, "0", int16(0.0), intDecoder{}, false)
	testDecoder(t, "-1.2", int16(0), intDecoder{}, true)
	testDecoder(t, "1.2", int16(0), intDecoder{}, true)

	testDecoder(t, "42.0", int8(42.0), intDecoder{}, false)
	testDecoder(t, "-42", int8(-42), intDecoder{}, false)
	testDecoder(t, "0", int8(0.0), intDecoder{}, false)
	testDecoder(t, "-1.2", int8(0), intDecoder{}, true)
	testDecoder(t, "1.2", int8(0), intDecoder{}, true)

	testDecoder(t, "42.0", int(42.0), intDecoder{}, false)
	testDecoder(t, "-42", int(-42), intDecoder{}, false)
	testDecoder(t, "0", int(0.0), intDecoder{}, false)
	testDecoder(t, "-1.2", int(0), intDecoder{}, true)
	testDecoder(t, "1.2", int(0), intDecoder{}, true)
	testDecoder(t, `null`, int(0), intDecoder{}, false)
}

func TestDecoderUint(t *testing.T) {
	testDecoder(t, "42.0", uint64(42.0), uintDecoder{}, false)
	testDecoder(t, "-42", uint64(0), uintDecoder{}, true)
	testDecoder(t, "0", uint64(0.0), uintDecoder{}, false)
	testDecoder(t, "-1.2", uint64(0), uintDecoder{}, true)
	testDecoder(t, "1.2", uint64(0), uintDecoder{}, true)

	testDecoder(t, "42.0", uint32(42.0), uintDecoder{}, false)
	testDecoder(t, "-42", uint32(0), uintDecoder{}, true)
	testDecoder(t, "0", uint32(0.0), uintDecoder{}, false)
	testDecoder(t, "-1.2", uint32(0), uintDecoder{}, true)
	testDecoder(t, "1.2", uint32(0), uintDecoder{}, true)

	testDecoder(t, "42.0", uint16(42.0), uintDecoder{}, false)
	testDecoder(t, "-42", uint16(0), uintDecoder{}, true)
	testDecoder(t, "0", uint16(0.0), uintDecoder{}, false)
	testDecoder(t, "-1.2", uint16(0), uintDecoder{}, true)
	testDecoder(t, "1.2", uint16(0), uintDecoder{}, true)

	testDecoder(t, "42.0", uint8(42.0), uintDecoder{}, false)
	testDecoder(t, "-42", uint8(0), uintDecoder{}, true)
	testDecoder(t, "0", uint8(0.0), uintDecoder{}, false)
	testDecoder(t, "-1.2", uint8(0), uintDecoder{}, true)
	testDecoder(t, "1.2", uint8(0), uintDecoder{}, true)

	testDecoder(t, "42.0", uint(42.0), uintDecoder{}, false)
	testDecoder(t, "-42", uint(0), uintDecoder{}, true)
	testDecoder(t, "0", uint(0.0), uintDecoder{}, false)
	testDecoder(t, "-1.2", uint(0), uintDecoder{}, true)
	testDecoder(t, "1.2", uint(0), uintDecoder{}, true)
	testDecoder(t, `null`, uint(0), uintDecoder{}, false)
}

func TestDecoderString(t *testing.T) {
	testDecoder(t, `"foo"`, "foo", stringDecoder{}, false)
	testDecoder(t, `"foo\u0032"`, "foo\x32", stringDecoder{}, false)
	testDecoder(t, `"\u4E16\u754C"`, "世界", stringDecoder{}, false)
	testDecoder(t, `null`, "", stringDecoder{}, false)
	testDecoder(t, `4`, "", stringDecoder{}, true)
	testDecoder(t, `{}`, "", stringDecoder{}, true)
}

func TestDecoderPtr(t *testing.T) {
	want := "foo"
	typ := reflect.PtrTo(reflect.TypeOf(want))
	dec, err := newPtrDecoder(typ, nil, cache{})
	assertNoError(t, err)
	testDecoder(t, `"foo"`, &want, dec, false)
	testDecoder(t, `null`, (*string)(nil), dec, false)

}

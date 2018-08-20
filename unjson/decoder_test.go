package unjson_test

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/alxarch/njson"
	"github.com/alxarch/njson/unjson"
)

func TestDecoderNilPointer(t *testing.T) {
	src := `{"Foo":1,"Bar":2,"Baz":3}`
	type A struct{ Foo, Bar, Baz int }
	var a *A = nil
	err := unjson.UnmarshalFromString(src, &a)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if a == nil {
		t.Errorf("Nil unmarshall")
		return

	}
	if a.Foo != 1 {
		t.Errorf("Invalid unmarshal Foo: %d", a.Foo)
		return
	}
	if a.Bar != 2 {
		t.Errorf("Invalid unmarshal Bar: %d", a.Bar)
		return
	}
	if a.Baz != 3 {
		t.Errorf("Invalid unmarshal Baz: %d", a.Baz)
		return
	}
}
func TestDecoderEmptyPointer(t *testing.T) {
	src := `{"Foo":1,"Bar":2,"Baz":3}`
	type A struct{ Foo, Bar, Baz int }
	a := &A{}
	err := unjson.UnmarshalFromString(src, &a)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if a == nil {
		t.Errorf("Nil unmarshall")
		return

	}
	if a.Foo != 1 {
		t.Errorf("Invalid unmarshal Foo: %d", a.Foo)
		return
	}
	if a.Bar != 2 {
		t.Errorf("Invalid unmarshal Bar: %d", a.Bar)
		return
	}
	if a.Baz != 3 {
		t.Errorf("Invalid unmarshal Baz: %d", a.Baz)
		return
	}
}

func TestUnmarshalInterface(t *testing.T) {
	src := `{"foo":1,"bar":2,"baz":3}`
	var v interface{}
	err := unjson.UnmarshalFromString(src, &v)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if m, ok := v.(map[string]interface{}); !ok {
		t.Errorf("Unexpected type: %v", v)
		return
	} else {
		for k, n := range map[string]int{
			"foo": 1,
			"bar": 2,
			"baz": 3,
		} {
			if f, ok := m[k].(float64); !ok {
				t.Errorf("Invalid decode %q not set %v", k, m[k])
			} else if f != float64(n) {
				t.Errorf("Invalid decode: %q %f", k, f)
			}
		}
	}

}
func TestUnmarshalMapInterface(t *testing.T) {
	v := map[string]interface{}{}
	src := `{"foo":1,"bar":2,"baz":3}`
	err := unjson.UnmarshalFromString(src, &v)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

}
func TestUnmarshalMap(t *testing.T) {
	v := map[string]int{}
	src := `{"foo":1,"bar":2,"baz":3}`
	if err := unjson.UnmarshalFromString(src, &v); err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	for k, n := range map[string]int{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	} {
		if nn, ok := v[k]; !ok {
			t.Errorf("Invalid decode %q not set %v", k, v)
		} else if nn != n {
			t.Errorf("Invalid decode: %q %d", k, nn)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	src := `{"foo":["bar"],"bar":23,"baz":{"foo":21.2, "bar": null}}`
	v := struct {
		Foo []string `json:"foo"`
		Bar *float64 `json:"bar"`
		Baz struct {
			Foo float64     `json:"foo"`
			Bar interface{} `json:"bar"`
		} `json:"baz"`
	}{}
	err := unjson.UnmarshalFromString(src, &v)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if len(v.Foo) != 1 || v.Foo[0] != "bar" {
		t.Errorf("Invalid decode foo: %v", v.Foo)
	}
	if *v.Bar != 23 {
		t.Errorf("Invalid decode bar: %f", *v.Bar)
	}
	if v.Baz.Foo != 21.2 {
		t.Errorf("Invalid decode baz: %f", v.Baz.Foo)
	}

}

type medium struct {
	Person struct {
		ID string `json:"id"`
	} `json:"person"`
	Email  string `json:"string"`
	Gender string `json:"gender"`
}

var mediumJSON string

func init() {
	data, err := ioutil.ReadFile("testdata/medium.min.json")
	if err != nil {
		panic(err)
	}
	mediumJSON = string(data)
}
func BenchmarkUnmarshal(b *testing.B) {
	mediumJSONBytes := []byte(mediumJSON)
	b.Run("json", func(b *testing.B) {
		m := medium{}
		for i := 0; i < b.N; i++ {
			if err := json.Unmarshal(mediumJSONBytes, &m); err != nil {
				b.Errorf("UnexpectedError: %s", err)
			}
		}
	})
	dec, err := unjson.TypeDecoder(reflect.TypeOf(medium{}), unjson.DefaultOptions())
	if err != nil {
		b.Errorf("UnexpectedError: %s", err)
	}
	b.Run("njson", func(b *testing.B) {
		m := medium{}
		p := njson.Parser{}
		doc := njson.Document{}
		var err error
		for i := 0; i < b.N; i++ {
			doc.Reset()
			if _, err = p.Parse(mediumJSON, &doc); err != nil {
				b.Errorf("UnexpectedError: %s", err)
			} else if err = dec.Decode(&m, doc.Get(0)); err != nil {
				b.Errorf("UnexpectedError: %s", err)
			}
		}
	})
}

func TestUnmarshalEmbeddedFields(t *testing.T) {
	type A struct {
		Foo string
	}
	type B struct {
		A
		Bar string
	}
	b := B{}
	if err := unjson.UnmarshalFromString(`{"Foo":"foo","Bar":"bar"}`, &b); err != nil {
		t.Errorf("Unexcpected error: %s", err)
		return
	}
	if b.Foo != "foo" {
		t.Errorf("Invalid b.Foo: %q", b.Foo)
		return
	}
	if b.Bar != "bar" {
		t.Errorf("Invalid b.Bar: %q", b.Bar)
		return
	}

}

func TestUnmarshalFromString(t *testing.T) {
	type args struct {
		s string
		x interface{}
	}
	var (
		f     float64
		empty interface{}
	)
	tests := []struct {
		name    string
		args    args
		wantErr bool
		check   interface{}
	}{
		{"float64", args{"1.2", f}, true, nil},
		{"float64", args{"1.2", empty}, true, nil},
		{"float64", args{"1.2", &f}, false, 1.2},
		// {"float64", args{"NaN", &f}, false, math.NaN()},
		{"float64", args{"0", &f}, false, 0.0},
		{"float64", args{"-1", &f}, false, -1.0},
		{"float64", args{"{}", &f}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := unjson.UnmarshalFromString(tt.args.s, tt.args.x); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalFromString(%v) Unexpected error: %v", tt.args.x, err)
				return
			}
			if tt.check == nil {
				return
			}
			v := reflect.ValueOf(tt.args.x)
			if !v.IsValid() {
				return
			}
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if !reflect.DeepEqual(v.Interface(), tt.check) {
				t.Errorf("UnmarshalFromString() %v != %v", v.Interface(), tt.check)
				return
			}
		})
	}
}

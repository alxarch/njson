package unjson

import (
	"reflect"
	"testing"

	"github.com/alxarch/njson"
)

func TestUnmarshalBasic(t *testing.T) {
	src := `{"Foo":1,"Bar":2,"Baz":3}`
	type A struct{ Foo, Bar, Baz int }
	a := A{}

	dec, err := NewTypeDecoder(reflect.TypeOf(&a), "")
	assertNoError(t, err)
	d := njson.Blank()
	defer d.Close()
	n, _, err := d.Parse(src)
	assertNoError(t, err)
	err = dec.Decode(&a, n)
	assertNoError(t, err)
	assertEqual(t, a, A{
		Foo: 1,
		Bar: 2,
		Baz: 3,
	})
}

func TestDecoderNilPointer(t *testing.T) {
	src := `{"Foo":1,"Bar":2,"Baz":3}`
	type A struct{ Foo, Bar, Baz int }
	var a *A
	err := UnmarshalFromString(src, &a)
	assertNoError(t, err)
	assert(t, a != nil, "Nil unmarshall")
	assertEqual(t, *a, A{
		Foo: 1,
		Bar: 2,
		Baz: 3,
	})
}
func TestDecoderEmptyPointer(t *testing.T) {
	src := `{"Foo":1,"Bar":2,"Baz":3}`
	type A struct{ Foo, Bar, Baz int }
	a := &A{}
	err := UnmarshalFromString(src, &a)
	assertNoError(t, err)
	assert(t, a != nil, "Nil unmarshall")
	assertEqual(t, *a, A{
		Foo: 1,
		Bar: 2,
		Baz: 3,
	})
}

func TestUnmarshalInterface(t *testing.T) {
	src := `{"foo":1,"bar":2,"baz":3}`
	var v interface{}
	err := UnmarshalFromString(src, &v)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	m, ok := v.(map[string]interface{})
	assert(t, ok, "Unexpected type: %v", v)
	assertEqual(t, m, map[string]interface{}{
		"foo": 1.0,
		"bar": 2.0,
		"baz": 3.0,
	})

}
func TestUnmarshalMapInterface(t *testing.T) {
	v := map[string]interface{}{}
	src := `{"foo":1,"bar":2,"baz":3}`
	err := UnmarshalFromString(src, &v)
	assertNoError(t, err)

}
func TestUnmarshalMap(t *testing.T) {
	v := map[string]int{}
	src := `{"foo":1,"bar":2,"baz":3}`
	err := UnmarshalFromString(src, &v)
	assertNoError(t, err)
	assertEqual(t, v, map[string]int{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	})
}

func TestUnmarshal(t *testing.T) {
	src := `{"foo":["bar"],"bar":23,"baz":{"foo":21.2, "bar": null}}`
	type baz struct {
		Foo float64     `json:"foo"`
		Bar interface{} `json:"bar"`
	}
	type foo struct {
		Foo []string `json:"foo"`
		Bar *float64 `json:"bar"`
		Baz baz      `json:"baz"`
	}
	v := foo{}
	err := UnmarshalFromString(src, &v)
	assertNoError(t, err)
	f := 23.0
	assertEqual(t, v, foo{
		Foo: []string{"bar"},
		Bar: &f,
		Baz: baz{
			Foo: 21.2,
			Bar: (interface{})(nil),
		},
	})

}

type medium struct {
	Person struct {
		ID string `json:"id"`
	} `json:"person"`
	Email  string `json:"string"`
	Gender string `json:"gender"`
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

	err := UnmarshalFromString(`{"Foo":"foo","Bar":"bar"}`, &b)
	assertNoError(t, err)
	assertEqual(t, b, B{
		A: A{
			Foo: "foo",
		},
		Bar: "bar",
	})
}

func TestUnmarshalFromString(t *testing.T) {
	type args struct {
		s string
		x interface{}
	}
	var (
		f64   float64
		f32   float32
		u64   uint64
		u32   uint32
		i64   int64
		i32   int32
		empty interface{}
	)
	tests := []struct {
		name    string
		args    args
		wantErr bool
		check   interface{}
	}{
		{"float64", args{"1.2", f64}, true, nil},
		{"float64", args{"1.2", empty}, true, nil},
		{"float64", args{"1.2", &f64}, false, 1.2},
		// {"float64", args{"NaN", &f}, false, math.NaN()},
		{"float64", args{"0", &f64}, false, 0.0},
		{"float64", args{"-1", &f64}, false, -1.0},
		{"float64", args{"{}", &f64}, true, nil},

		{"float32", args{"1.2", f32}, true, nil},
		{"float32", args{"1.2", empty}, true, nil},
		{"float32", args{"1.2", &f32}, false, float32(1.2)},
		// {"float32", args{"NaN", &f}, false, math.NaN()},
		{"float32", args{"0", &f32}, false, float32(0.0)},
		{"float32", args{"-1", &f32}, false, float32(-1.0)},
		{"float32", args{"{}", &f32}, true, nil},

		{"uint64", args{"1.2", u64}, true, nil},
		{"uint64", args{"1.2", empty}, true, nil},
		{"uint64", args{"1.2", &u64}, true, nil},
		{"uint64", args{"0", &u64}, false, uint64(0)},
		{"uint64", args{"42", &u64}, false, uint64(42)},
		{"uint64", args{"-1", &u64}, true, nil},
		{"uint64", args{"{}", &u64}, true, nil},

		{"uint32", args{"1.2", u32}, true, nil},
		{"uint32", args{"1.2", empty}, true, nil},
		{"uint32", args{"1.2", &u32}, true, nil},
		{"uint32", args{"0", &u32}, false, uint32(0)},
		{"uint32", args{"42", &u32}, false, uint32(42)},
		{"uint32", args{"-1", &u32}, true, nil},
		{"uint32", args{"{}", &u32}, true, nil},

		{"int32", args{"1.2", i32}, true, nil},
		{"int32", args{"1.2", empty}, true, nil},
		{"int32", args{"1.2", &i32}, true, nil},
		{"int32", args{"0", &i32}, false, int32(0)},
		{"int32", args{"42", &i32}, false, int32(42)},
		{"int32", args{"-1", &i32}, false, int32(-1)},
		{"int32", args{"{}", &i32}, true, nil},

		{"int64", args{"1.2", i64}, true, nil},
		{"int64", args{"1.2", empty}, true, nil},
		{"int64", args{"1.2", &i64}, true, nil},
		{"int64", args{"0", &i64}, false, int64(0)},
		{"int64", args{"42", &i64}, false, int64(42)},
		{"int64", args{"-1", &i64}, false, int64(-1)},
		{"int64", args{"{}", &i64}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := UnmarshalFromString(tt.args.s, tt.args.x); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalFromString(%T) Unexpected error: %v", tt.args.x, err)
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

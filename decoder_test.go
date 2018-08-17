package njson_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/alxarch/njson"
)

func TestUnmarshalInterface(t *testing.T) {
	src := `{"foo":1,"bar":2,"baz":3}`
	var v interface{}
	err := njson.UnmarshalFromString(src, &v)
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
	err := njson.UnmarshalFromString(src, &v)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

}
func TestUnmarshalMap(t *testing.T) {
	v := map[string]int{}
	src := `{"foo":1,"bar":2,"baz":3}`
	if err := njson.UnmarshalFromString(src, &v); err != nil {
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
	err := njson.UnmarshalFromString(src, &v)
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
	dec, err := njson.TypeDecoder(reflect.TypeOf(medium{}), njson.DefaultOptions())
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
	if err := njson.UnmarshalFromString(`{"Foo":"foo","Bar":"bar"}`, &b); err != nil {
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

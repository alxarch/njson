package njsonutil

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/alxarch/njson"
)

var (
	// typErr                 = reflect.TypeOf((*error)(nil)).Elem()
	// typNodePtr             = reflect.TypeOf((*Node)(nil)).Elem()
	typUnmarshaler       = reflect.TypeOf((*njson.Unmarshaler)(nil)).Elem()
	typUnmarshalerMethod = typUnmarshaler.Method(0)
)

type T struct {
	value reflect.Value
	check func(err error) error
	input func() ([]byte, error)
	tag   string
}

type TestOption interface {
	set(t *T)
}

type testOption func(t *T)

func (f testOption) set(t *T) {
	f(t)
}

func CustomTag(tag string) TestOption {
	if tag == "" {
		tag = "json"
	}
	return testOption(func(t *T) {
		t.tag = tag
	})
}

func CustomInput(custom func() ([]byte, error)) TestOption {
	return testOption(func(t *T) {
		t.input = custom
	})
}

func ReadInput(r io.Reader) TestOption {
	return testOption(func(t *T) {
		t.input = func() ([]byte, error) {
			return ioutil.ReadAll(r)
		}
	})
}
func defaultInput(x interface{}) TestOption {
	return testOption(func(t *T) {
		t.input = func() ([]byte, error) {
			return json.Marshal(x)
		}
	})
}
func NullInput() TestOption {
	return Input([]byte(`null`))
}

func Input(b []byte) TestOption {
	return testOption(func(t *T) {
		t.input = func() ([]byte, error) {
			return b, nil
		}
	})
}

func UnmarshalTest(x interface{}, options ...TestOption) func(t *testing.T) {
	if x == nil {
		return nil
	}
	test := T{
		value: reflect.ValueOf(x),
	}
	options = append([]TestOption{defaultInput(x)}, options...)
	for _, option := range options {
		if option != nil {
			option.set(&test)
		}
	}
	typ := test.value.Type()
	method, err := CheckImplementsTagged(reflect.PtrTo(typ), typUnmarshalerMethod, test.tag)
	if err != nil {
		return func(t *testing.T) {
			t.Error(err)
		}
	}

	return func(t *testing.T) {
		var (
			data []byte
			err  error
			d    = njson.BlankDocument()
			n    *njson.Node
			v    = reflect.New(typ)
		)
		defer d.Close()

		if data, err = test.input(); err != nil {
			t.Errorf("Unexpected input error: %s", err)
			return
		}
		if n, err = d.Parse(string(data)); err != nil {
			t.Errorf("Unexpected parse error: %s", err)
			return
		}

		results := v.MethodByName(method).Call([]reflect.Value{
			reflect.ValueOf(n),
		})

		if err, hasError := results[0].Interface().(error); hasError && err != nil {
			t.Errorf("[%s.%s] Unmarshal error: %s", reflect.TypeOf(v), method, err)
		} else if !reflect.DeepEqual(x, v.Elem().Interface()) {
			t.Errorf("[%s.%s] Invalid unmarshal:\nexpect: %v\nactual: %v", reflect.TypeOf(v), method, test.value, v.Elem())
		}
	}

}

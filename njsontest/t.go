package njsontest

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/alxarch/njson"
)

var (
	typUnmarshaler = reflect.TypeOf((*njson.Unmarshaler)(nil)).Elem()
	typAppender    = reflect.TypeOf((*njson.Appender)(nil)).Elem()
)

// T is a test case for njson.
type T struct {
	check  func(err error) error
	json   func() ([]byte, error)
	method string
}

func newT(x interface{}, options []Option) (t T) {
	options = append([]Option{
		defaultJSON(x),
	}, options...)
	for _, option := range options {
		if option != nil {
			option.set(&t)
		}
	}
	return
}

// AppendJSON check a value implementing njson.Appender interface.
func AppendJSON(x interface{}, options ...Option) func(t *testing.T) {
	if x == nil {
		return nil
	}
	test := newT(x, options)
	method := typAppender.Method(0)
	if test.method != "" {
		method.Name = test.method
	}

	return func(t *testing.T) {
		v := reflect.ValueOf(x)
		if err := checkMethod(v, method); err != nil {
			t.Errorf("Wrong value type: %s", err)
		}
		expect, err := test.json()
		if err != nil {
			t.Errorf("[%s.%s] Unexpected JSON data error: %s", v.Type(), method.Name, err)
			return
		}
		results := v.MethodByName(method.Name).Call([]reflect.Value{
			reflect.ValueOf(([]byte)(nil)),
		})
		if err := checkError(results[1], test.check); err != nil {
			t.Errorf("[%s.%s] Append error: %s", v.Type(), method.Name, err)
			return
		}
		if test.check == nil {
			actual, _ := results[0].Interface().([]byte)
			if !bytes.Equal(actual, expect) {
				t.Errorf("[%s.%s] Unexpected result:\nexpect: %s\nactual: %s", v.Type(), method.Name, expect, actual)
			}

		}
	}

}

// Unmarshal is a test case for an njson.Unmarshaler
func Unmarshal(x interface{}, options ...Option) func(t *testing.T) {
	if x == nil {
		return nil
	}
	test := newT(x, options)
	method := typUnmarshaler.Method(0)
	if test.method != "" {
		method.Name = test.method

	}
	original := reflect.ValueOf(x)
	typ := original.Type()
	if typ.Kind() == reflect.Ptr {
		original = original.Elem()
		typ = typ.Elem()
	}

	return func(t *testing.T) {
		var (
			data []byte
			err  error
			d    = njson.Get()
			n    *njson.Node
			v    = reflect.New(typ)
		)
		defer d.Close()
		if err := checkMethod(v, method); err != nil {
			t.Error(err)
			return
		}

		if data, err = test.json(); err != nil {
			t.Errorf("Unexpected input error: %s", err)
			return
		}
		if n, _, err = d.Parse(string(data)); err != nil {
			t.Errorf("Unexpected parse error: %s", err)
			return
		}

		results := v.MethodByName(method.Name).Call([]reflect.Value{
			reflect.ValueOf(n),
		})
		if err := checkError(results[0], test.check); err != nil {
			t.Errorf("[%s.%s] Unmarshal error: %s", typ, method.Name, err)
			return
		}
		if test.check == nil && !reflect.DeepEqual(original.Interface(), v.Elem().Interface()) {
			t.Errorf("[%s.%s] Invalid unmarshal:\nexpect: %v\nactual: %v\ninput: %s",
				reflect.TypeOf(v), method.Name, original, v.Elem(), data)
		}
	}

}

func checkError(v reflect.Value, check func(error) error) error {
	if v.IsNil() {
		if check == nil {
			return nil
		}
		return fmt.Errorf("Invalid error value %v", v)
	}
	if !v.CanInterface() {
		return fmt.Errorf("Invalid error value %v", v)
	}
	err, hasError := v.Interface().(error)
	if !hasError {
		if check == nil {
			return nil
		}
		return fmt.Errorf("Invalid error value %v", v)
	}
	if check == nil {
		return err
	}
	return check(err)
}

func checkMethod(v reflect.Value, method reflect.Method) error {
	if !v.IsValid() {
		return fmt.Errorf("Invalid value")
	}
	m := v.MethodByName(method.Name)
	if !m.IsValid() {
		return fmt.Errorf("Type %s doesn't have a method named %s", v.Type(), method.Name)
	}
	fn := m.Type()
	if fn == nil || !fn.ConvertibleTo(method.Type) {
		return fmt.Errorf("Type %s has wrong type of %s method:\n%s\n%s", v.Type(), method.Name, fn, method.Type)
	}
	return nil
}

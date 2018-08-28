package njson_test

import (
	"bytes"
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/alxarch/njson"
)

func TestNodeToBool(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()
	if n, err := d.Parse("true"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if v, ok := n.ToBool(); !ok {
		t.Errorf("Unexpected conversion %v", n)
	} else if !v {
		t.Errorf("Unexpected conversion %v", n)
	}
	if n, err := d.Parse("false"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if v, ok := n.ToBool(); !ok {
		t.Errorf("Unexpected conversion %v", n)
	} else if v {
		t.Errorf("Unexpected conversion %v", n)
	}
	if v, ok := ((*njson.Node)(nil)).ToBool(); ok {
		t.Errorf("Unexpected conversion %v", v)
	} else if v {
		t.Errorf("Unexpected conversion %v", v)
	}
	if v, ok := new(njson.Node).ToBool(); ok {
		t.Errorf("Unexpected conversion %v", v)
	} else if v {
		t.Errorf("Unexpected conversion %v", v)
	}
}

func TestNodeToFloat(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()
	n, err := d.Parse("1.2")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 1.2 {
		t.Errorf("Unexpected conversion %f", f)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 1.2 {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, err := d.Parse("NaN"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if !math.IsNaN(f) {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, err := d.Parse("0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 0 {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, err := d.Parse("-17"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != -17 {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, err := d.Parse("-a7"); err == nil || n != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if n, err := d.Parse("true"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToFloat(); ok {
		t.Errorf("Unexpected conversion error")
	}
	if v, ok := ((*njson.Node)(nil)).ToFloat(); ok {
		t.Errorf("Unexpected conversion %v", v)
	} else if v != 0 {
		t.Errorf("Unexpected conversion %v", v)
	}
}

func TestNodeToInt(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()
	if n, err := d.Parse("1.2"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToInt(); ok {
		t.Errorf("Unexpected conversion ok")
	}
	if n, err := d.Parse("-1.0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != -1 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, err := d.Parse("NaN"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToInt(); ok {
		t.Errorf("Unexpected conversion error")
	}

	if n, err := d.Parse("0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 0 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, err := d.Parse("-17"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != -17 {
		t.Errorf("Unexpected conversion %d", f)
	}
	if n, err := d.Parse("42.0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 42 {
		t.Errorf("Unexpected conversion %d", f)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 42 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, err := d.Parse("-a7"); err == nil || n != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if v, ok := ((*njson.Node)(nil)).ToInt(); ok {
		t.Errorf("Unexpected conversion %v", v)
	} else if v != 0 {
		t.Errorf("Unexpected conversion %v", v)
	}
}

func TestNodeToUint(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()
	n, err := d.Parse("1.2")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if _, ok := n.ToUint(); ok {
		t.Errorf("Unexpected conversion ok")
	}

	if n, err := d.Parse("NaN"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToUint(); ok {
		t.Errorf("Unexpected conversion error")
	}

	if n, err := d.Parse("0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToUint(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 0 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, err := d.Parse("-17"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToUint(); ok {
		t.Errorf("Unexpected conversion ok")
	}

	if n, err := d.Parse("42.0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToUint(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 42 {
		t.Errorf("Unexpected conversion %d", f)
	} else if f, ok := n.ToUint(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 42 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, err := d.Parse("-a7"); err == nil || n != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if v, ok := ((*njson.Node)(nil)).ToUint(); ok {
		t.Errorf("Unexpected conversion %v", v)
	} else if v != 0 {
		t.Errorf("Unexpected conversion %v", v)
	}
}

type customJSONUnmarshaler struct {
	Foo int
}

func (c *customJSONUnmarshaler) UnmarshalJSON(data []byte) error {
	v := []int{0}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	if len(v) > 0 {
		c.Foo = v[0]
	}
	return nil
}
func TestNode_WrapUnmarshalJSON(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()

	{
		c := customJSONUnmarshaler{}
		if n, err := d.ParseUnsafe([]byte("[42]")); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if err := n.WrapUnmarshalJSON(&c); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if c.Foo != 42 {
			t.Errorf("Unexpected value: %d", c.Foo)
		}

	}
	{
		c := customJSONUnmarshaler{}
		if n, err := d.ParseUnsafe([]byte("[]")); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if err := n.WrapUnmarshalJSON(&c); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if c.Foo != 0 {
			t.Errorf("Unexpected value: %d", c.Foo)
		}

	}
}

func TestNode_Unescaped(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()
	if n, err := d.Parse(`"foo"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.Unescaped(); s != "foo" {
		t.Errorf("Unexpected value: %s", s)
	}
	if n, err := d.Parse(`42`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.Unescaped(); s != "42" {
		t.Errorf("Unexpected value: %s", s)
	}
	if n, err := d.Parse(`"foo\n"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.Unescaped(); s != "foo\n" {
		t.Errorf("Unexpected value: %s", s)
	} else if s := n.Unescaped(); s != "foo\n" {
		t.Errorf("Unexpected value: %s", s)
	}
	if n, err := d.Parse(`"foo\n"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.UnescapedBytes(); string(s) != "foo\n" {
		t.Errorf("Unexpected value: %s", s)
	} else if s := n.UnescapedBytes(); string(s) != "foo\n" {
		t.Errorf("Unexpected value: %s", s)
	}
	if n, err := d.Parse(`"foo"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.UnescapedBytes(); string(s) != "foo" {
		t.Errorf("Unexpected value: %s", s)
	}
}

func TestNode_ToInterface(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()

	if n, err := d.Parse(`"foo"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, "foo") {
		t.Errorf("Unexpected value: %v", x)
	}

	if n, err := d.Parse(`42`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, 42.0) {
		t.Errorf("Unexpected value: %v", x)
	}

	if n, err := d.Parse(`["foo"]`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, []interface{}{"foo"}) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, err := d.Parse(`{}`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, map[string]interface{}{}) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, err := d.Parse(`{"answer": 42}`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, map[string]interface{}{"answer": 42.0}) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, err := d.Parse(`true`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, true) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, err := d.Parse(`false`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, false) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, err := d.Parse(`null`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, nil) {
		t.Errorf("Unexpected value: %v", x)
	}

	if n, err := d.Parse(``); err == nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, nil) {
		t.Errorf("Unexpected value: %v", x)
	}
}

func TestNode_PrintJSON(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()
	buf := bytes.NewBuffer(nil)
	s := `{"answer":42}`
	if n, err := d.Parse(s); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if N, err := n.PrintJSON(buf); err != nil {
		t.Errorf("Failed to print %v to buffer.", n)
	} else if N != len(s) {
		t.Errorf("Invalid number of written bytes %d != %d", N, len(s))
	} else if actual := buf.String(); actual != s {
		t.Errorf("Unexpected value: %s", actual)
	}
}

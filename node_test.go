package njson

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestNodeToBool(t *testing.T) {
	d := Document{}
	if n, _, err := d.Parse("true"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if v, ok := n.ToBool(); !ok {
		t.Errorf("Unexpected conversion %v", n)
	} else if !v {
		t.Errorf("Unexpected conversion %v", n)
	}
	if n, _, err := d.Parse("false"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if v, ok := n.ToBool(); !ok {
		t.Errorf("Unexpected conversion %v", n)
	} else if v {
		t.Errorf("Unexpected conversion %v", n)
	}
	// if v, ok := ((*Node)(nil)).ToBool(); ok {
	// 	t.Errorf("Unexpected conversion %v", v)
	// } else if v {
	// 	t.Errorf("Unexpected conversion %v", v)
	// }
	if v, ok := new(Node).ToBool(); ok {
		t.Errorf("Unexpected conversion %v", v)
	} else if v {
		t.Errorf("Unexpected conversion %v", v)
	}
}

func TestNodeToFloat(t *testing.T) {
	d := Document{}
	n, _, err := d.Parse("1.2")
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

	// if n, _, err := d.Parse("NaN"); err != nil {
	// 	t.Errorf("Unexpected error: %s", err)
	// } else if f, ok := n.Node().ToFloat(); !ok {
	// 	t.Errorf("Unexpected conversion error")
	// } else if !math.IsNaN(f) {
	// 	t.Errorf("Unexpected conversion %f", f)
	// }

	if n, _, err := d.Parse("0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 0 {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, _, err := d.Parse("-17"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != -17 {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, _, err := d.Parse("-a7"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToFloat(); ok {
		t.Errorf("Expected conversion error")
	}
	if n, _, err := d.Parse("true"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToFloat(); ok {
		t.Errorf("Unexpected conversion error")
	}
	// if v, ok := ((*Node)(nil)).ToFloat(); ok {
	// 	t.Errorf("Unexpected conversion %v", v)
	// } else if v != 0 {
	// 	t.Errorf("Unexpected conversion %v", v)
	// }
}

func TestNodeToInt(t *testing.T) {
	d := Document{}
	if n, _, err := d.Parse("1.2"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToInt(); ok {
		t.Errorf("Unexpected conversion ok")
	}
	if n, _, err := d.Parse("-1.0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error %#v %d", n, f)
	} else if f != -1 {
		t.Errorf("Unexpected conversion %d", f)
	}

	// if n, _, err := d.Parse("NaN"); err != nil {
	// 	t.Errorf("Unexpected error: %s", err)
	// } else if _, ok := n.Node().ToInt(); ok {
	// 	t.Errorf("Unexpected conversion error")
	// }

	if n, _, err := d.Parse("0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 0 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, _, err := d.Parse("-17"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != -17 {
		t.Errorf("Unexpected conversion %d", f)
	}
	if n, _, err := d.Parse("42.0"); err != nil {
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

	if n, _, err := d.Parse("-a7"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if v, ok := n.ToInt(); ok {
		t.Errorf("Unexpected conversion %v", v)
	}
}

func TestNodeToUint(t *testing.T) {
	d := Document{}
	n, _, err := d.Parse("1.2")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if _, ok := n.ToUint(); ok {
		t.Errorf("Unexpected conversion ok")
	}
	if n, _, err := d.Parse("0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToUint(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 0 {
		t.Errorf("Unexpected conversion %d", f)
	}

	// if n, _, err := d.Parse("NaN"); err != nil {
	// 	t.Errorf("Unexpected error: %s", err)
	// } else if _, ok := n.Node().ToUint(); ok {
	// 	t.Errorf("Unexpected conversion error")
	// }

	if n, _, err := d.Parse("-17"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToUint(); ok {
		t.Errorf("Unexpected conversion ok")
	}

	if n, _, err := d.Parse("42.0"); err != nil {
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

	if n, _, err := d.Parse("-a7"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if u, ok := n.ToUint(); ok {
		t.Errorf("Unexpected conversion %d", u)
	}
	// if v, ok := ((*Node)(nil)).ToUint(); ok {
	// 	t.Errorf("Unexpected conversion %v", v)
	// } else if v != 0 {
	// 	t.Errorf("Unexpected conversion %v", v)
	// }
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
	d := Blank()
	defer d.Close()
	{
		c := customJSONUnmarshaler{}
		if n, _, err := d.ParseUnsafe([]byte("[42]")); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if err := n.WrapUnmarshalJSON(&c); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if c.Foo != 42 {
			t.Errorf("Unexpected value: %d", c.Foo)
		}

	}
	{
		c := customJSONUnmarshaler{}
		if n, _, err := d.ParseUnsafe([]byte("[]")); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if err := n.WrapUnmarshalJSON(&c); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if c.Foo != 0 {
			t.Errorf("Unexpected value: %d", c.Foo)
		}

	}
}
func TestNode_Unescaped(t *testing.T) {
	d := Blank()
	defer d.Close()
	if n, _, err := d.Parse(`"foo"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.Unescaped(); s != "foo" {
		t.Errorf("Unexpected value: %s", s)
	}
	if n, _, err := d.Parse(`42`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.Unescaped(); s != "" {
		t.Errorf("Unexpected value: %s", s)
	}
	if n, _, err := d.Parse(`"foo\n"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.Unescaped(); s != "foo\n" {
		t.Errorf("Unexpected value: %s", s)
	} else if s := n.Unescaped(); s != "foo\n" {
		t.Errorf("Unexpected value: %s", s)
	}
	if n, _, err := d.Parse(`"foo\n"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.Unescaped(); string(s) != "foo\n" {
		t.Errorf("Unexpected value: %s", s)
	} else if s := n.Unescaped(); string(s) != "foo\n" {
		t.Errorf("Unexpected value: %s", s)
	}
	if n, _, err := d.Parse(`"foo"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s := n.Unescaped(); string(s) != "foo" {
		t.Errorf("Unexpected value: %s", s)
	}
}

func TestNode_ToInterface(t *testing.T) {
	d := Blank()
	defer d.Close()

	if n, _, err := d.Parse(`"foo"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, "foo") {
		t.Errorf("Unexpected value: %v", x)
	}

	if n, _, err := d.Parse(`42`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, 42.0) {
		t.Errorf("Unexpected value: %v", x)
	}

	if n, _, err := d.Parse(`["foo"]`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, []interface{}{"foo"}) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, _, err := d.Parse(`{}`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, map[string]interface{}{}) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, _, err := d.Parse(`{"answer": 42}`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, map[string]interface{}{"answer": 42.0}) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, _, err := d.Parse(`true`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, true) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, _, err := d.Parse(`false`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, false) {
		t.Errorf("Unexpected value: %v", x)
	}
	if n, _, err := d.Parse(`null`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); !ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, nil) {
		t.Errorf("Unexpected value: %v", x)
	}

	if n, _, err := d.Parse(``); err == nil {
		t.Errorf("Unexpected error: %s", err)
	} else if x, ok := n.ToInterface(); ok {
		t.Errorf("Failed to convert %v to interface.", n)
	} else if !reflect.DeepEqual(x, nil) {
		t.Errorf("Unexpected value: %v", x)
	}
}

func TestNode_PrintJSON(t *testing.T) {
	d := Document{}
	buf := bytes.NewBuffer(nil)
	s := `{"answer":42}`
	if n, _, err := d.Parse(s); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if N, err := PrintJSON(buf, n); err != nil {
		t.Errorf("Failed to print %v to buffer.", n)
	} else if N != len(s) {
		t.Errorf("Invalid number of written bytes %d != %d", N, len(s))
	} else if actual := buf.String(); actual != s {
		t.Errorf("Unexpected value: %s", actual)
	}
}

package njson

import (
	"bytes"
	"encoding/json"
	"github.com/alxarch/njson/numjson"
	"reflect"
	"testing"
)

func TestNode_ToBool(t *testing.T) {
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
	if n, _, err := d.Parse("true"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n.Boolean() != True {
		t.Errorf("Unexpected conversion %v", n)
	}
}

//func TestNode_ToFloat(t *testing.T) {
//	d := Document{}
//	n, _, err := d.Parse("1.2")
//	if err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	}
//	if f, ok := n.ToFloat(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != 1.2 {
//		t.Errorf("Unexpected conversion %f", f)
//	} else if f, ok := n.ToFloat(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != 1.2 {
//		t.Errorf("Unexpected conversion %f", f)
//	}
//
//	// if n, _, err := d.Parse("NaN"); err != nil {
//	// 	t.Errorf("Unexpected error: %s", err)
//	// } else if f, ok := n.Node().ToFloat(); !ok {
//	// 	t.Errorf("Unexpected conversion error")
//	// } else if !math.IsNaN(f) {
//	// 	t.Errorf("Unexpected conversion %f", f)
//	// }
//
//	if n, _, err := d.Parse("0"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if f, ok := n.ToFloat(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != 0 {
//		t.Errorf("Unexpected conversion %f", f)
//	}
//
//	if n, _, err := d.Parse("-17"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if f, ok := n.ToFloat(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != -17 {
//		t.Errorf("Unexpected conversion %f", f)
//	}
//
//	if n, _, err := d.Parse("-a7"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if _, ok := n.ToFloat(); ok {
//		t.Errorf("Expected conversion error")
//	}
//	if n, _, err := d.Parse("true"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if _, ok := n.ToFloat(); ok {
//		t.Errorf("Unexpected conversion error")
//	}
//	// if v, ok := ((*Node)(nil)).ToFloat(); ok {
//	// 	t.Errorf("Unexpected conversion %v", v)
//	// } else if v != 0 {
//	// 	t.Errorf("Unexpected conversion %v", v)
//	// }
//}
//
//func TestNode_ToInt(t *testing.T) {
//	d := Document{}
//	if n, _, err := d.Parse("1.2"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if _, ok := n.ToInt(); ok {
//		t.Errorf("Unexpected conversion ok")
//	}
//	if n, _, err := d.Parse("-1.0"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if f, ok := n.ToInt(); !ok {
//		t.Errorf("Unexpected conversion error %v %d", n, f)
//	} else if f != -1 {
//		t.Errorf("Unexpected conversion %d", f)
//	}
//
//	// if n, _, err := d.Parse("NaN"); err != nil {
//	// 	t.Errorf("Unexpected error: %s", err)
//	// } else if _, ok := n.Node().ToInt(); ok {
//	// 	t.Errorf("Unexpected conversion error")
//	// }
//
//	if n, _, err := d.Parse("0"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if f, ok := n.ToInt(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != 0 {
//		t.Errorf("Unexpected conversion %d", f)
//	}
//
//	if n, _, err := d.Parse("-17"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if f, ok := n.ToInt(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != -17 {
//		t.Errorf("Unexpected conversion %d", f)
//	}
//	if n, _, err := d.Parse("42.0"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if f, ok := n.ToInt(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != 42 {
//		t.Errorf("Unexpected conversion %d", f)
//	} else if f, ok := n.ToInt(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != 42 {
//		t.Errorf("Unexpected conversion %d", f)
//	}
//
//	if n, _, err := d.Parse("-a7"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if v, ok := n.ToInt(); ok {
//		t.Errorf("Unexpected conversion %v", v)
//	}
//}
//
//func TestNode_ToUint(t *testing.T) {
//	d := Document{}
//	n, _, err := d.Parse("1.2")
//	if err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	}
//	if _, ok := n.ToUint(); ok {
//		t.Errorf("Unexpected conversion ok")
//	}
//	if n, _, err := d.Parse("0"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if f, ok := n.ToUint(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != 0 {
//		t.Errorf("Unexpected conversion %d", f)
//	}
//
//	// if n, _, err := d.Parse("NaN"); err != nil {
//	// 	t.Errorf("Unexpected error: %s", err)
//	// } else if _, ok := n.Node().ToUint(); ok {
//	// 	t.Errorf("Unexpected conversion error")
//	// }
//
//	if n, _, err := d.Parse("-17"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if _, ok := n.ToUint(); ok {
//		t.Errorf("Unexpected conversion ok")
//	}
//
//	if n, _, err := d.Parse("42.0"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if f, ok := n.ToUint(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != 42 {
//		t.Errorf("Unexpected conversion %d", f)
//	} else if f, ok := n.ToUint(); !ok {
//		t.Errorf("Unexpected conversion error")
//	} else if f != 42 {
//		t.Errorf("Unexpected conversion %d", f)
//	}
//
//	if n, _, err := d.Parse("-a7"); err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	} else if u, ok := n.ToUint(); ok {
//		t.Errorf("Unexpected conversion %d", u)
//	}
//	// if v, ok := ((*Node)(nil)).ToUint(); ok {
//	// 	t.Errorf("Unexpected conversion %v", v)
//	// } else if v != 0 {
//	// 	t.Errorf("Unexpected conversion %v", v)
//	// }
//}

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
		if n, _, err := d.Parse("[42]"); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if err := n.WrapUnmarshalJSON(&c); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if c.Foo != 42 {
			t.Errorf("Unexpected value: %d", c.Foo)
		}

	}
	{
		c := customJSONUnmarshaler{}
		if n, _, err := d.Parse("[]"); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if err := n.WrapUnmarshalJSON(&c); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if c.Foo != 0 {
			t.Errorf("Unexpected value: %d", c.Foo)
		}

	}
	{
		c := customJSONUnmarshaler{}
		if n, _, err := d.Parse(`{}`); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if err := n.WrapUnmarshalJSON(&c); err == nil {
			t.Errorf("Expected error got nil")
		} else if c.Foo != 0 {
			t.Errorf("Unexpected value: %d", c.Foo)
		}
	}
	{
		c := customJSONUnmarshaler{}
		if n, _, err := d.Parse(`1`); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if err := n.WrapUnmarshalJSON(&c); err == nil {
			t.Errorf("Expected error got nil")
		} else if c.Foo != 0 {
			t.Errorf("Unexpected value: %d", c.Foo)
		}

	}
	{
		c := customJSONUnmarshaler{}
		n := Node{}
		if err := n.WrapUnmarshalJSON(&c); err == nil {
			t.Errorf("Expected error got nil")
		} else if c.Foo != 0 {
			t.Errorf("Unexpected value: %d", c.Foo)
		} else if e, ok := err.(typeError); !ok {
			t.Errorf("Unexpected error: %v", err)
		} else if e.Want != TypeAnyValue {
			t.Errorf("Unexpected type error: %v", e.Want)
		} else if e.Type != TypeInvalid {
			t.Errorf("Unexpected type error: %v", e.Type)
		}

	}
}

type customTextUnmarshaler struct {
	Foo string
}

func (c *customTextUnmarshaler) UnmarshalText(data []byte) error {
	c.Foo = string(data)
	return nil
}

func TestNode_WrapUnmarshalText(t *testing.T) {
	d := Blank()
	defer d.Close()
	{
		c := customTextUnmarshaler{}
		if n, _, err := d.Parse(`"foo"`); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if err := n.WrapUnmarshalText(&c); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if c.Foo != "foo" {
			t.Errorf("Unexpected value: %s", c.Foo)
		}
	}
	{
		c := customTextUnmarshaler{}
		n, _, err := d.Parse(`{}`)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		err = n.WrapUnmarshalText(&c)
		assertEqual(t, err, typeError{TypeObject, TypeString})
	}
}
func TestNode_Unescaped(t *testing.T) {
	d := Blank()
	defer d.Close()
	if n, _, err := d.Parse(`"foo"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s, typ := n.ToString(); typ != TypeString || s != "foo" {
		t.Errorf("Unexpected value: %s(%s)", typ,s)
	}
	if n, _, err := d.Parse(`42`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s, typ := n.ToString(); typ != TypeNumber || s != "42" {
		t.Errorf("Unexpected value: %s(%s)", typ,s)
	}
	if n, _, err := d.Parse(`"foo\n"`); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if s, typ := n.ToString(); typ != TypeString || s != "foo\n" {
		t.Errorf("Unexpected value: %s(%s)", typ,s)
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
	} else if !reflect.DeepEqual(x, int64(42)) {
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
	} else if !reflect.DeepEqual(x, map[string]interface{}{"answer": int64(42)}) {
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

func TestNode_Remove(t *testing.T) {
	d := Document{}
	s := `{"results":[1,2,3]}`
	n, _, err := d.Parse(s)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	n.Object().Get("results").Array().Remove(1)
	data, err := n.AppendJSON(nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if string(data) != `{"results":[1,3]}` {
		t.Errorf("Unexpected JSON value: %s", data)

	}

}

func TestNode_Index(t *testing.T) {
	d := Document{}
	n, _, _ := d.Parse(`[1,2,3]`)
	arr := Array(n)
	{
		n := arr.Get(2)
		assertEqual(t, n.Type(), TypeNumber)
		assertEqual(t, n.Number(), numjson.Int64(3))
		assertEqual(t, n.id, uint(3))
	}
	{
		n := arr.Get(8)
		assertEqual(t, Node{}, n)
	}
}
func TestNode_Del(t *testing.T) {
	d := Document{}
	s := `{"answer":42,"wrong_answer":41}`
	n, _, err := d.Parse(s)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	Object(n).Del("wrong_answer")
	data, err := n.AppendJSON(nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if string(data) != `{"answer":42}` {
		t.Errorf("Unexpected JSON value: %s", data)
	}

}

func TestNode_Strip(t *testing.T) {
	d := Document{}
	n, _, err := d.Parse(`{
		"bar": {
			"foo": "bar",
			"bar":"baz"
		},
		"foo": {}
	}`)
	assertNoError(t, err)
	total := Object(n).Strip("foo")
	data, err := n.AppendJSON(nil)
	assertNoError(t, err)
	assertEqual(t, 2, total)
	assertEqual(t, `{"bar":{"bar":"baz"}}`, string(data))
}

func TestNode_Lookup(t *testing.T) {
	d := Document{}
	n, _, err := d.Parse(`{
		"foo": {},
		"bar": {
			"foo": "bar",
			"bar":"baz",
			"baz": ["foo","bar","baz"]
		}
	}`)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	n = n.Lookup("bar", "baz", "2")
	if n.Type() != TypeString {
		t.Errorf("Invalid type: %s", n.Type())
	}
	if s, _ := n.ToString(); s != "baz" {
		t.Errorf("Invalid value: %s", s)
	}
	data, err := n.AppendJSON(nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if string(data) != `"baz"` {
		t.Errorf("Unexpected JSON value: %s", data)
	}
}

func TestNode_Append(t *testing.T) {
	d := Document{}
	n := d.NewArray()
	n.Push(d.NewString("foo"))
	n.Push(d.NewString("bar"))
	n.Push(d.NewString("baz"))
	data, err := n.Node().AppendJSON(nil)
	assertNoError(t, err)
	assertEqual(t, string(data), `["foo","bar","baz"]`)
	n.Slice(0, 2)
	data, err = n.Node().AppendJSON(nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if string(data) != `["foo","bar"]` {
		t.Errorf("Invalid append result: %s", data)
	}
	n.Replace(0, d.NewInt(42))
	data, err = n.Node().AppendJSON(nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if string(data) != `[42,"bar"]` {
		t.Errorf("Invalid append result: %s", data)
	}

}
func TestNode_Values(t *testing.T) {
	d := Document{}
	n, _, err := d.Parse(`{
		"foo": 1,
		"bar": 2,
		"baz": 3
	}`)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	v := n.Object().Iter()
	if n.Object().Len() != 3 {
		t.Errorf("Invalid iterator len: %d", n.Object().Len())
	}
	iterKeys := []string{}
	iterValues := []int64{}
	iterIndices := []int{}
	for i := 0; v.Next(); i++ {
		iterKeys = append(iterKeys, v.Key())
		n := v.Node().Number()
		iterValues = append(iterValues, n.Int64())
		iterIndices = append(iterIndices, i)
	}
	assertEqual(t, iterKeys, []string{"foo", "bar", "baz"})
	assertEqual(t, iterValues, []int64{1, 2, 3})
	assertEqual(t, iterIndices, []int{0, 1, 2})
	assertEqual(t, v.Next(), false)
	v.Close()
	assertEqual(t, v.Next(), false)
	assertEqual(t, v.node, Node{})
	assertEqual(t, v.iter, iterator{})
}

func TestNode_SetX(t *testing.T) {
	d := Document{}
	//n := d.Text("foo")
	//n.SetString("bar")
	//assertEqual(t, n.Raw(), "bar")
	//assertEqual(t, n.Type(), TypeString)
	//n.SetInt(42)
	//assertEqual(t, n.Raw(), "42")
	//assertEqual(t, n.Type(), TypeNumber)
	//n.SetFloat(1)
	//assertEqual(t, n.Raw(), "1")
	//assertEqual(t, n.Type(), TypeNumber)
	//n.SetUint(2)
	//assertEqual(t, n.Raw(), "2")
	//assertEqual(t, n.Type(), TypeNumber)
	//n.SetNull()
	//assertEqual(t, n.Raw(), "null")
	//assertEqual(t, n.Type(), TypeNull)
	//n.SetFalse()
	//assertEqual(t, n.Raw(), "false")
	//assertEqual(t, n.Type(), TypeBoolean)
	//n.SetTrue()
	//assertEqual(t, n.Raw(), "true")
	//assertEqual(t, n.Type(), TypeBoolean)
	//n.SetStringHTML("<p>foo</p>")
	//assertEqual(t, n.Raw(), `\u003cp\u003efoo\u003c\/p\u003e`)
	//assertEqual(t, n.Type(), TypeString)
	obj := d.NewObject()
	obj.Set("foo", d.NewStringSafe("bar"))
	obj.Set("foo", d.NewString("baz"))
	foo := obj.Get("foo")
	v, ok := foo.ToInterface()
	assert(t, ok, "interface is valid")
	assertEqual(t, "baz", v)
}

func TestNode_Empty(t *testing.T) {
	n := Node{}
	assertEqual(t, n.Type(), TypeInvalid)
	s, typ := n.Inspect()
	assertEqual(t, TypeInvalid, typ)
	assertEqual(t, "", s)
	assertEqual(t, n.value(), (*value)(nil))
	//{
	//	n, ok := n.ToUint()
	//	assertEqual(t, n, uint64(0))
	//	assertEqual(t, ok, false)
	//}
	//{
	//	n, ok := n.ToFloat()
	//	assertEqual(t, n, float64(0))
	//	assertEqual(t, ok, false)
	//}
	//{
	//	n, ok := n.ToInt()
	//	assertEqual(t, n, int64(0))
	//	assertEqual(t, ok, false)
	//}
	{
		b, ok := n.ToBool()
		assertEqual(t, b, false)
		assertEqual(t, ok, false)
	}
	{
		n := Object(n).Get("foo")
		assertEqual(t, n, Node{})
	}
}

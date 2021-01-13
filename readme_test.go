package njson_test

import (
	"fmt"

	"github.com/alxarch/njson"
)

func Example() {
	d := njson.Document{}

	root, _, _ := d.Parse(`{"answer":42, "foo": {"bar": "baz"}}`)

	answer := root.Object().Get("answer").Number().Int64()
	fmt.Println(answer)

	n := root.Lookup("foo", "bar")
	bar, typ := n.ToString()
	fmt.Println(bar, typ)

	obj := root.Lookup("foo").Object()
	obj.Set("bar", d.NewString("Hello, 世界"))

	data := make([]byte, 64)
	data, _ = root.AppendJSON(data[:0])
	fmt.Println(string(data))

	// Output:
	// 42
	// baz String
	// {"answer":42,"foo":{"bar":"Hello, 世界"}}
}

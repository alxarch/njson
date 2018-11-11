package njson_test

import (
	"fmt"

	"github.com/alxarch/njson"
)

func Example() {
	d := njson.Document{}

	root, _, _ := d.Parse(`{"answer":42, "foo": {"bar": "baz"}}`)

	answer, _ := root.Get("answer").ToInt()
	fmt.Println(answer)

	n := root.Lookup("foo", "bar")
	bar := n.Unescaped()
	fmt.Println(bar)

	n.SetString("Hello, 世界")

	data := make([]byte, 64)
	data, _ = root.AppendJSON(data[:0])
	fmt.Println(string(data))

	// Output:
	// 42
	// baz
	// {"answer":42,"foo":{"bar":"Hello, 世界"}}
}

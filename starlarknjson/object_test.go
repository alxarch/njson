package starlarknjson

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

func TestObject(t *testing.T) {
	thread := &starlark.Thread{
		Name: "test",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Println(msg)
		},
	}
	env := starlark.StringDict{
		"njson": &Module,
	}
	code := `#
obj = njson.parse('{"foo":"bar"}')
foo = obj["foo"]
obj["bar"] = 'baz'
d = dict(obj)
`
	globals, err := starlark.ExecFile(thread, "test.star", code, env)
	require.NoError(t, err)
	require.Equal(t, globals["foo"], starlark.String("bar"))
	d := globals["d"].String()
	require.Equal(t, `{"foo": "bar", "bar": "baz"}`, d)
}

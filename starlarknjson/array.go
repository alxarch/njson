package starlarknjson

import (
	"errors"
	"fmt"
	"github.com/alxarch/njson"
	"go.starlark.net/starlark"
)

var _ starlark.Value = (*Array)(nil)
var _ starlark.HasAttrs = (*Array)(nil)
var _ starlark.Indexable = (*Array)(nil)

type Array struct {
	node      njson.Array
	iterators uint32
	len       uint32
	frozen    bool
}

func (a *Array) Index(i int) starlark.Value {
	if v := nodeValue(a.node.Get(i)); v != nil {
		return v
	}
	return starlark.None
}
func (a *Array) isMutable() bool {
	return a.frozen == false && a.iterators == 0
}

// checkMutable reports an error if the list should not be mutated.
// verb+" list" should describe the operation.
func (a *Array) checkMutable(verb string) error {
	if a.frozen {
		return fmt.Errorf("cannot %s frozen array", verb)
	}
	if a.iterators > 0 {
		return fmt.Errorf("cannot %s array during iteration", verb)
	}
	return nil
}

func (a *Array) Len() int {
	return int(a.len)
}

func (a *Array) Attr(name string) (starlark.Value, error) {
	return arrayMethods.Get(name, a)
}

func (a *Array) AttrNames() []string {
	return arrayMethods.Names()
}

func (a *Array) String() string {
	return "[]"
}

func (a *Array) Type() string {
	return "njson_array"
}

func (a *Array) Freeze() {
	a.frozen = true
}

func (a *Array) Truth() starlark.Bool {
	return a.len != 0
}

func (a *Array) Hash() (uint32, error) {
	return 0, errors.New("njson array is not hashable")
}

var arrayMethods = newProto(map[string]*starlark.Builtin{
	"append": starlark.NewBuiltin("append", arrayAppend),
	"clear":  starlark.NewBuiltin("clear", arrayClear),
	"extend": starlark.NewBuiltin("extend", arrayExtend),
	"index":  starlark.NewBuiltin("index", arrayIndex),
	"insert": starlark.NewBuiltin("insert", arrayInsert),
	"pop":    starlark.NewBuiltin("pop", arrayPop),
	"remove": starlark.NewBuiltin("remove", arrayRemove),
})

func arrayRemove(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("implement me")
}

func arrayPop(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("implement me")
}

func arrayInsert(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("implement me")
}

func arrayIndex(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("implement me")
}

func arrayExtend(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("implement me")
}

func arrayClear(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("implement me")
}

func arrayAppend(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	arr := fn.Receiver().(*Array)
	if !arr.isMutable() {
		return nil, errors.New("array is immutable")
	}
	var el starlark.Value
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &el); err != nil {
		return nil, err
	}
	node, err := nodeOf(arr.node.Document(), el)
	if err != nil {
		return nil, err
	}
	arr.node.Push(node)
	arr.len++
	return starlark.None, nil
}

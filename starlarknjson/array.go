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
var _ starlark.HasSetIndex = (*Array)(nil)
var _ starlark.Iterable = (*Array)(nil)

type Array struct {
	node      njson.Array
	iterators uint32
	len       uint32
	frozen    bool
}

func (a *Array) SetIndex(index int, value starlark.Value) error {
	if err := a.checkMutable("insert to"); err != nil {
		return err
	}
	node, err := Node(a.node.Node().Document(), value)
	if err != nil {
		return err
	}
	if a.node.Set(index, node).IsValid() {
		return errors.New("unexpected error while inserting to njson array")
	}
	return nil
}

func (a *Array) Iterate() starlark.Iterator {
	return &arrayIter{
		ArrayIterator: a.node.Iter(),
	}
}

type arrayIter struct {
	njson.ArrayIterator
}

func (i *arrayIter) Next(p *starlark.Value) bool {
	if i.ArrayIterator.Next() {
		*p = Value(i.Node())
		return true
	}
	return false
}

func (i *arrayIter) Done() {
	i.ArrayIterator.Close()
}

func (a *Array) Index(i int) starlark.Value {
	if v := Value(a.node.Get(i)); v != nil {
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

func arrayIndex(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var seek, start_, end_ starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &seek, &start_, &end_); err != nil {
		return nil, err
	}
	arr := b.Receiver().(*Array)
	len := arr.Len()
	start, end, err := indices(len, start_, end_)
	if err != nil {
		return nil, builtinError(b, err)
	}
	iter := arr.node.Iter()
	defer iter.Close()
	i := 0
	for ; i < start; i++ {
		iter.Next()
	}
	for ; iter.Next() && i < end; i++ {
		el := Value(iter.Node())
		if err != nil {
			return nil, builtinError(b, err)
		}
		eq, err := starlark.Equal(el, seek)
		if err != nil {
			return nil, builtinError(b, err)
		}
		if eq {
			return starlark.MakeInt(i), nil
		}
	}
	return nil, builtinError(b, "value not in array")
}

func indices(len int, start_, end_ starlark.Value) (int, int, error) {
	if start_ == starlark.None {
		return 0, len, nil
	}
	start, err := starlark.AsInt32(start_)
	if err != nil {
		return 0, len, err
	}
	start = asIndex(start, len)
	if end_ == starlark.None {
		return start, len, nil
	}
	end, err := starlark.AsInt32(end_)
	if err != nil {
		return start, len, err
	}
	end = asIndex(end, len)
	return start, end, nil
}

func asIndex(i, n int) int {
	if 0 <= i && i < n {
		return i
	}
	if i >= n {
		return n
	}
	i += n
	if i < 0 {
		return 0
	}
	return i
}

func arrayExtend(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("implement me")
}

func arrayClear(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("implement me")
}

func arrayAppend(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	arr := fn.Receiver().(*Array)
	if err := arr.checkMutable("append to"); err != nil {
		return nil, err
	}
	var el starlark.Value
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &el); err != nil {
		return nil, err
	}
	node, err := Node(arr.node.Node().Document(), el)
	if err != nil {
		return nil, builtinError(fn, err)
	}
	arr.node.Push(node)
	arr.len++
	return starlark.None, nil
}

func builtinError(b *starlark.Builtin, msg interface{}) error {
	err, _ := msg.(error)
	return &namedError{
		msg: fmt.Sprintf("%s: %s", b.Name(), msg),
		err: err,
	}
}

type namedError struct {
	msg string
	err error
}

func (e *namedError) Error() string {
	return e.msg
}

func (e *namedError) Unwrap() error {
	return e.err
}
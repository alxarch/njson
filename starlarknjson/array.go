package starlarknjson

import (
	"errors"
	"github.com/alxarch/njson"
	"go.starlark.net/starlark"
)

var _ starlark.Value = (*Array)(nil)
var _ starlark.HasAttrs = (*Array)(nil)
var _ starlark.Iterable = (*Array)(nil)
var _ starlark.Indexable = (*Array)(nil)
var _ starlark.HasSetIndex = (*Array)(nil)

type Array struct {
	arr          njson.Array
	frozen       bool
}

func (a *Array) Attr(name string) (starlark.Value, error) {
	return arrayMethods.Get(name, a)
}

func (a *Array) AttrNames() []string {
	return arrayMethods.Names()
}

func (a *Array) mut() (njson.Array, error) {
	if err := a.checkMutable(); err != nil {
		return njson.Array{}, err
	}
	return a.arr, nil
}

func (a *Array) checkMutable() error {
	if !a.arr.IsMutable() {
		return errors.New("cannot modify an array while iterating")
	}
	if a.frozen {
		return errors.New("cannot modify a frozen array")
	}
	return nil
}

func (a *Array) SetIndex(index int, v starlark.Value) error {
	node, err := Node(njson.Node(a.arr).Document(), v)
	if err != nil {
		return err
	}
	if a.frozen {
		return errors.New("array frozen")
	}
	node = a.arr.Set(index, node)
	if err := node.Err(); err != nil {
		return err
	}
	return nil
}

func (a *Array) Index(i int) starlark.Value {
	if v := Value(a.arr.Get(i)); v != nil {
		return v
	}
	return starlark.None
}

func (a *Array) Len() int {
	return a.arr.Len()
}

func (a *Array) String() string {
	return "[]"
}

func (a *Array) Type() string {
	return "njson_array"
}

func (a *Array) Freeze() {
	a.frozen =  true
}

func (a Array) Truth() starlark.Bool {
	return a.Len() != 0
}

func (a Array) Hash() (uint32, error) {
	return 0, errors.New("njson array is not hashable")
}

func (a Array) Iterate() starlark.Iterator {
	return &arrayIter{
		ArrayIterator: a.arr.Iterate(),
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

var arrayMethods = newProto(map[string]*starlark.Builtin{
	"append": starlark.NewBuiltin("append", arrayAppend),
	"clear":  starlark.NewBuiltin("clear", arrayClear),
	"extend": starlark.NewBuiltin("extend", arrayExtend),
	"index":  starlark.NewBuiltin("index", arrayIndex),
	"insert": starlark.NewBuiltin("insert", arrayInsert),
	"pop":    starlark.NewBuiltin("pop", arrayPop),
	"remove": starlark.NewBuiltin("remove", arrayRemove),
})

func arrayAppend(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var el starlark.Value
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &el); err != nil {
		return nil, err
	}
	arr, err := fn.Receiver().(*Array).mut()
	if err != nil {
		return nil, err
	}
	val, err := Node(arr.Node().Document(), el)
	if err != nil {
		return nil, err
	}
	if err := arr.Push(val).Err(); err != nil {
		return nil, err
	}
	return starlark.None, nil
}

func arrayClear(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 0 {
		return nil, errors.New("clear does not accept any arguments")
	}
	if len(kwargs) != 0 {
		return nil, errors.New("clear does not accept any keyword arguments")
	}
	arr, err := fn.Receiver().(*Array).mut()
	if err != nil {
		return nil, err
	}
	if err := arr.Clear().Node().Err(); err != nil {
		return nil, err
	}
	return starlark.None, nil
}

func arrayExtend(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var iterable starlark.Iterable
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &iterable); err != nil {
		return nil, err
	}
	arr, err := fn.Receiver().(*Array).mut()
	if err != nil {
		return nil, err
	}
	doc := arr.Node().Document()
	var el starlark.Value
	iter := iterable.Iterate()
	defer iter.Done()
	for iter.Next(&el) {
		node, err := Node(doc, el)
		if err != nil {
			return nil, err
		}
		if err := arr.Push(node).Err(); err != nil {
			return nil, err
		}
	}
	return starlark.None, nil
}

func arrayIndex(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	arr := b.Receiver().(*Array)
	var seek, start_, end_ starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &seek, &start_, &end_); err != nil {
		return nil, err
	}
	len := arr.Len()
	start, end, err := indices(len, start_, end_)
	if err != nil {
		return nil, errors.New(methodError(b, err.Error()))
	}
	iter := arr.arr.Iterate()
	defer iter.Close()
	i := 0
	for ; i < start; i++ {
		iter.Next()
	}
	for ; iter.Next() && i < end; i++ {
		el := Value(iter.Node())
		if err != nil {
			return nil, errors.New(methodError(b, err.Error()))
		}
		eq, err := starlark.Equal(el, seek)
		if err != nil {
			return nil, errors.New(methodError(b, err.Error()))
		}
		if eq {
			return starlark.MakeInt(i), nil
		}
	}
	return nil, errors.New(methodError(b, "value not in array"))
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

func arrayInsert(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("implement me")
}

func arrayPop(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	recv := fn.Receiver().(*Array)
	n := recv.Len()
	i := n-1
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0, &i); err != nil {
		return nil, err
	}
	if i < 0 {
		i += n
	}
	arr, err := recv.mut()
	if err != nil {
		return nil, err
	}
	node := arr.Remove(i)
	if err := node.Err(); err != nil {
		return nil, err
	}
	return Value(node), nil
}

func arrayRemove(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var seek starlark.Value
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &seek); err != nil {
		return nil, err
	}
	arr, err := fn.Receiver().(*Array).mut()
	if err != nil {
		return nil, err
	}
	iter := arr.Iterate()
	defer iter.Close()
	for i := 0; iter.Next(); i++ {
		v := Value(iter.Node())
		ok, err := starlark.Equal(seek, v)
		if err != nil {
			return nil, err
		}
		if ok {
			iter.Close()
			if err := arr.Remove(i).Err(); err != nil {
				return nil, err
			}
			return starlark.None, nil
		}
	}
	return nil, errors.New(methodError(fn, "element not found"))
}


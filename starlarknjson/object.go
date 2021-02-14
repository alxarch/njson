package starlarknjson

import (
	"errors"
	"fmt"
	"github.com/alxarch/njson"
	"go.starlark.net/starlark"
)

var _ starlark.Value = (*Object)(nil)
var _ starlark.IterableMapping = (*Object)(nil)
var _ starlark.HasAttrs = (*Object)(nil)
var _ starlark.HasSetKey = (*Object)(nil)

type Object struct {
	obj    njson.Object
	frozen bool
}

func (o *Object) Get(key starlark.Value) (v starlark.Value, found bool, err error) {
	k, ok := key.(starlark.String)
	if !ok {
		// The key is not a string, we won't find it
		return nil, false, TypeError(fmt.Sprintf("cannot use instance of %q as key for %s", key.Type(), o.Type()))
	}
	node := o.obj.Get(string(k))
	if node.IsValid() {
		return Value(node), true, nil
	}
	return nil, false, nil
}

func (o *Object) SetKey(k, v starlark.Value) error {
	key, err := checkKey(k)
	if err != nil {
		return err
	}
	node, err := o.node(v)
	if err != nil {
		return err
	}
	obj, err := o.mut()
	if err != nil {
		return err
	}
	obj.Set(key, node)
	return nil
}

func (o *Object) node(v starlark.Value) (njson.Node, error) {
	doc := o.obj.Node().Document()
	return Node(doc, v)
}

func checkKey(key starlark.Value) (string, error) {
	if key, ok := key.(starlark.String); ok {
		return string(key), nil
	}
	return "", newValueError(fmt.Errorf("instance of %q cannot be used as key for njson_object", key.Type()))
}

func (o *Object) Attr(name string) (starlark.Value, error) {
	return objectMethods.Get(name, o)
}

func (o *Object) AttrNames() []string {
	return objectMethods.names
}

var objectMethods = newProto(map[string]*starlark.Builtin{
	"get":        starlark.NewBuiltin("get", objectGet),
	"clear":      starlark.NewBuiltin("clear", objectClear),
	"items":      starlark.NewBuiltin("items", objectItems),
	"keys":       starlark.NewBuiltin("keys", objectKeys),
	"pop":        starlark.NewBuiltin("pop", objectPop),
	"popitem":    starlark.NewBuiltin("popitem", objectPopItem),
	"setdefault": starlark.NewBuiltin("setdefault", objectSetDefault),
	"update":     starlark.NewBuiltin("update", objectUpdate),
	"values":     starlark.NewBuiltin("values", objectValues),
})

func objectSetDefault(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key, dflt starlark.Value = nil, starlark.None
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &key, &dflt); err != nil {
		return nil, err
	}
	obj := fn.Receiver().(*Object)
	v, ok, err := obj.Get(key)
	if err != nil {
		return nil, err
	}
	if ok {
		return v, nil
	}
	if err := obj.SetKey(key, dflt); err != nil {
		return nil, err
	}
	return dflt, nil
}

func objectUpdate(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	obj, err := fn.Receiver().(*Object).mut()
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		switch arg := args[0].(type) {
		case starlark.IterableMapping:
			err = setIterableMapping(obj, arg)
		case starlark.Iterable:
			err = setIterable(obj, arg)
		default:
			return nil, errors.New(methodError(fn, "invalid first argument"))
		}
		if err != nil {
			return nil, err
		}
	}
	if len(args) != 0 {
		return nil, errors.New(methodError(fn, "receives up to one positional argument"))
	}
	err = setKWArgs(obj, kwargs)
	if err != nil {
		return nil, err
	}
	return starlark.None, nil
}

func iterKV(pair starlark.Value) (k, v starlark.Value, err error) {
	iter := starlark.Iterate(pair)
	if iter == nil {
		return k, v, fmt.Errorf("non iterable pair")
	}
	defer iter.Done()
	n := starlark.Len(pair)
	if n < 0 {
		return k, v, fmt.Errorf("cannot get pair len")
	}
	if n != 2 {
		return k, v, fmt.Errorf("invalid pair")
	}
	iter.Next(&k)
	iter.Next(&v)
	return k, v, nil
}

func setIterableMapping(obj njson.Object, arg starlark.IterableMapping) error {
	doc := obj.Node().Document()
	iter := arg.Iterate()
	defer iter.Done()
	var k starlark.Value
	for iter.Next(&k) {
		key, err := checkKey(k)
		if err != nil {
			return err
		}
		v, ok, err := arg.Get(k)
		if err != nil {
			return err
		}
		if !ok {
			return KeyError(fmt.Sprintf("key %q not found in pair", key))
		}
		val, err := Node(doc, v)
		if err != nil {
			return err
		}
		obj.Set(key, val)
	}
	return nil
}

func setIterable(obj njson.Object, arg starlark.Iterable) error {
	doc := obj.Node().Document()
	iter := arg.Iterate()
	defer iter.Done()
	var pair starlark.Value
	for iter.Next(&pair) {
		k, v, err := iterKV(pair)
		if err != nil {
			return err
		}
		key, err := checkKey(k)
		if err != nil {
			return err
		}
		val, err := Node(doc, v)
		if err != nil {
			return err
		}
		obj.Set(key, val)
	}
	return nil
}

func setKWArgs(obj njson.Object, kwargs []starlark.Tuple) error {
	doc := obj.Node().Document()
	for _, kw := range kwargs {
		if len(kw) == 2 {
			k, v := kw[0], kw[1]
			key, err := checkKey(k)
			if err != nil {
				return err
			}
			val, err := Node(doc, v)
			if err != nil {
				return err
			}
			obj.Set(key, val)
		}
	}
	return nil
}

func objectValues(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	obj := fn.Receiver().(*Object).obj
	iter := obj.Iterate()
	values := make([]starlark.Value, 0, iter.Len())
	for iter.Next() {
		values = append(values, Value(iter.Node()))
	}
	return starlark.NewList(values), nil
}

func objectPopItem(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}

	obj, err := fn.Receiver().(*Object).mut()
	if err != nil {
		return nil, err
	}
	key, node := obj.Pop()
	if node.IsValid() {
		return starlark.Tuple{
			starlark.String(key),
			Value(node),
		}, nil
	}
	return nil, KeyError(methodError(fn, "object is empty"))
}

func objectPop(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key, dflt starlark.Value
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &key, &dflt); err != nil {
		return nil, err
	}
	k, err := checkKey(key)
	if err != nil {
		return nil, err
	}
	obj, err := fn.Receiver().(*Object).mut()
	if err != nil {
		return nil, err
	}
	node := obj.Del(k, true)
	if err := node.Err(); err != nil {
		return nil, err
	}
	if node.IsValid() {
		return Value(node), nil
	}
	if dflt != nil {
		return dflt, nil
	}
	return nil, KeyError(fmt.Sprintf("%q not found in object", k))
}

func objectKeys(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	o := fn.Receiver().(*Object)
	iter := o.obj.Iterate()
	defer iter.Close()
	keys := starlark.NewSet(iter.Len())
	for iter.Next() {
		keys.Insert(starlark.String(iter.Key()))
	}
	return keys, nil
}

func objectItems(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	obj := fn.Receiver().(*Object)
	items := obj.Items()
	out := make([]starlark.Value, len(items))
	for i, pair := range items {
		out[i] = pair
	}
	return starlark.NewList(out), nil
}

func objectClear(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	obj, err := fn.Receiver().(*Object).mut()
	if err != nil {
		return nil, err
	}
	if err := obj.Clear().Node().Err(); err != nil {
		return nil, err
	}
	return starlark.None, nil
}

func (o *Object) Iterate() starlark.Iterator {
	return &objectIter{
		ObjectIterator: o.obj.Iterate(),
	}
}

type objectIter struct {
	njson.ObjectIterator
}

func (o *objectIter) Next(p *starlark.Value) bool {
	if o.ObjectIterator.Next() {
		*p = starlark.String(o.ObjectIterator.Key())
		return true
	}
	return false
}

func (o objectIter) Done() {
	o.ObjectIterator.Close()
}

func (o *Object) Items() []starlark.Tuple {
	iter := o.obj.Iterate()
	defer iter.Close()
	items := make([]starlark.Tuple, 0, iter.Len())
	values := make([]starlark.Value, cap(items)*2)
	var item starlark.Tuple
	for iter.Next() {
		if len(values) >= 2 { // elide bounds checks
			item, values = values[:2], values[2:]
			if len(item) == 2 { // elide bounds checks
				item[0], item[1] = starlark.String(iter.Key()), Value(iter.Node())
				items = append(items, item)
			}
		}
	}
	return items
}

func (o *Object) String() string {
	return "{}"
}

func (o *Object) Type() string {
	return "njson_object"
}

func (o *Object) Freeze() {
	o.frozen = true
}

func (o *Object) Truth() starlark.Bool {
	return o.obj.Len() != 0
}

func (o *Object) Hash() (uint32, error) {
	return 0, errors.New("njson_object is not hashable")
}

func objectGet(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key, fallback starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &key, &fallback); err != nil {
		return nil, err
	}
	obj := b.Receiver().(*Object)

	v, ok, err := obj.Get(key)
	if err != nil {
		return nil, err
	}
	if ok {
		return v, nil
	}
	if fallback != nil {
		return fallback, nil
	}
	return starlark.None, nil
}

func (o *Object) checkMutable() error {
	if !o.obj.IsMutable() {
		return errors.New("cannot modify an object while iterating")
	}
	if o.frozen {
		return errors.New("cannot modify a frozen object")
	}
	return nil
}

func (o *Object) mut() (njson.Object, error) {
	return o.obj, o.checkMutable()
}

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

type Object struct {
	node      njson.Object
	iterators uint32
	len uint32
	frozen    bool
}
var objectMethods = newProto(map[string]*starlark.Builtin{
	"get": starlark.NewBuiltin("get", objectGet),
})

func (o *Object) Attr(name string) (starlark.Value, error) {
	return objectMethods.Get(name, o)
}

func (o *Object) AttrNames() []string {
	return objectMethods.names
}

func (o *Object) Get(key starlark.Value) (v starlark.Value, found bool, err error) {
	if k, ok := key.(starlark.String); ok {
		node := o.node.Get(string(k))
		if node.IsZero() {
			return nil, false, nil
		}
		return nodeValue(node), true, nil
	}
	return nil, false, errors.New("invalid key")
}

func (o *Object) Iterate() starlark.Iterator {
	return &objectIter{}
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
	items := make([]starlark.Tuple, 0, o.len)
	values := make([]starlark.Value, 2*len(items))
	var item starlark.Tuple
	iter := o.node.Iter()
	defer iter.Close()
	for iter.Next() && len(values) >= 2 {
		item, values = values[:2], values[2:]
		item[0] = starlark.String(iter.Key())
		item[1], _ = nodeValue(iter.Node())
		items = append(items, item)
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
	return o.len != 0
}

func (o *Object) Hash() (uint32, error) {
	return 0, errors.New("njson_object is not hashable")
}

func objectGet(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key, fallback starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &key, &fallback); err != nil {
		return nil, err
	}
	if v, ok, err := b.Receiver().(*Object).Get(key); err != nil {
		return nil, fmt.Errorf("%s: %s", b.Name(), err)
	} else if ok {
		return v, nil
	} else if fallback != nil {
		return fallback, nil
	}
	return starlark.None, nil
}

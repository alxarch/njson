package starlarknjson

import (
	"errors"
	"github.com/alxarch/njson"
	"go.starlark.net/starlark"
)

var _ starlark.Value = (*Object)(nil)
var _ starlark.IterableMapping = (*Object)(nil)
type Object struct {
	node      njson.Object
	iterators uint32
	len uint32
	frozen    bool
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
	panic("implement me")
}

func (o *Object) Items() []starlark.Tuple {
	panic("implement me")
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



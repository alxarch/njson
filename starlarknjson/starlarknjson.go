// Package starlarknjson provides low allocation JSON support for Starlark
package starlarknjson

import (
	"errors"
	"github.com/alxarch/njson"
	"go.starlark.net/starlark"
	"strings"
)

const (
	keyThreadLocalDocument = "njson"
)

func Parse(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(kwargs) != 0 {
		return nil, errors.New("parse does not accept keyword arguments")
	}
	if len(args) != 1 {
		return nil, errors.New("parse expects an argument")
	}
	var input string
	if err := starlark.UnpackPositionalArgs("parse", args, kwargs, 1, &input); err != nil {
		return nil, err
	}
	doc := documentFromThread(thread)
	if doc == nil {
		doc = &njson.Document{}
		thread.SetLocal(keyThreadLocalDocument, doc)
	}
	node, tail, err := doc.Parse(input)
	if err != nil {
		return nil, err
	}
	if tail = strings.TrimSpace(tail); tail != "" {
		return nil, errors.New("leftover text after parsing JSON")
	}
	return nodeValue(node)
}

func documentFromThread(thread *starlark.Thread) *njson.Document {
	if doc, ok := thread.Local(keyThreadLocalDocument).(*njson.Document); ok {
		return doc
	}
	return nil
}

type proto struct {
	methods map[string]*starlark.Builtin
	names   []string
}

func newProto(methods map[string]*starlark.Builtin) *proto {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	return &proto{
		methods: methods,
		names:   names,
	}
}

func (p *proto) Get(name string, recv starlark.Value) (starlark.Value, error) {
	if m := p.methods[name]; m != nil {
		return m.BindReceiver(recv), nil
	}
	return nil, nil
}
func (p *proto) Names() []string {
	return p.names
}

func nodeValue(n njson.Node) (starlark.Value, error) {
	switch s, typ := n.Text(); typ {
	case njson.TypeString:
		return starlark.String(s), nil
	case njson.TypeArray:
		arr := njson.Array(n)
		return &Array{
			node: arr,
			len:  uint32(arr.Len()),
		}, nil
	case njson.TypeObject:
		obj := njson.Object(n)
		return &Object{
			node: obj,
			len:  uint32(obj.Len()),
		}, nil
	case njson.TypeNumber:
		v, err := readNumber(s)
		if err != nil {
			return nil, nil
		}
		return v, nil
	case njson.TypeBoolean:
		if s == "true" {
			return starlark.True, nil
		}
		return starlark.False, nil
	case njson.TypeNull:
		return starlark.None, nil
	default:
		return nil, nil
	}
}

var zeroNode njson.Node

func nodeOf(doc *njson.Document, v starlark.Value) (njson.Node, error) {
	switch v := v.(type) {
	case *Array:
		return njson.Node(v.node), nil
	case *Object:
		return njson.Node(v.node), nil
	case starlark.String:
		return doc.Text(string(v)), nil
	case starlark.IterableMapping:
		obj := doc.Object()
		iter := v.Iterate()
		var key starlark.Value
		for iter.Next(&key) {
			key, ok := key.(starlark.String)
			if !ok {
				return zeroNode, errors.New("invalid dict key")
			}
			val, _, _ := v.Get(key)
			node, err := nodeOf(doc, val)
			if err != nil {
				return zeroNode, err
			}
			obj.Set(string(key), node)
		}
		return obj.Node(), nil
	case starlark.Iterable:
		arr := doc.Array()
		iter := v.Iterate()
		var el starlark.Value
		for iter.Next(&el) {
			el, err := nodeOf(doc, el)
			if err != nil {
				return njson.Node{}, err
			}
			arr.Push(el)
		}
		return njson.Node(arr), nil
	case starlark.Bool:
		if v {
			return doc.True(), nil
		}
		return doc.False(), nil
	case starlark.Float:
		return doc.Number(float64(v)), nil
	case starlark.NoneType:
		return doc.Null(), nil
	case starlark.Int:
		return doc.RawNumber(v.String()), nil
	case interface{ NJSON() njson.Node }:
		return v.NJSON(), nil
	default:
		return zeroNode, errors.New("cannot handle type")
	}
}

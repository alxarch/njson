// Package starlarknjson provides low allocation JSON support for Starlark
package starlarknjson

import (
	"errors"
	"github.com/alxarch/njson"
	"go.starlark.net/starlark"
	"math/big"
)

type proto struct {
	methods map[string]*starlark.Builtin
	names  []string
}

func newProto(methods map[string]*starlark.Builtin) *proto {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	return &proto{
		methods: methods,
		names: names,
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

func nodeValue(n njson.Node) starlark.Value {
	switch n.Type() {
	case njson.TypeString:
		s, _ := n.ToString()
		return starlark.String(s)
	case njson.TypeArray:
		arr := njson.Array(n)
		return &Array{
			node: arr,
			len: uint32(arr.Len()),
		}
	case njson.TypeObject:
		obj := njson.Object(n)
		return &Object{
			node: obj,
		}
	case njson.TypeNumber:
		if n, ok := n.ToInt(); ok {
			return starlark.MakeInt64(n)
		}
		if n, ok := n.ToUint(); ok {
			return starlark.MakeUint64(n)
		}
		if f, ok := n.ToFloat(); ok {
			return starlark.Float(f)
		}
		b := big.NewInt(0)
		b.SetString(n.Raw(), 10)
		return starlark.MakeBigInt(b)
	case njson.TypeBoolean:
		if n.Raw() == "true" {
			return starlark.True
		}
		return starlark.False
	case njson.TypeNull:
		return starlark.None
	default:
		return starlark.None
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
	case interface {NJSON() njson.Node}:
		return v.NJSON(), nil
	default:
		return zeroNode, errors.New("cannot handle type")
	}
}
// Package starlarknjson provides low allocation JSON support for Starlark
package starlarknjson

import (
	"errors"
	"github.com/alxarch/njson"
	"github.com/alxarch/njson/numjson"
	"go.starlark.net/starlark"
	"math"
	"math/big"
	"strconv"
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
	return Value(node), nil
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

func Value(n njson.Node) starlark.Value {
	switch s, typ := n.ToString(); typ {
	case njson.TypeString:
		return starlark.String(s)
	case njson.TypeArray:
		arr := njson.Array(n)
		return &Array{
			node: arr,
			len:  uint32(arr.Len()),
		}
	case njson.TypeObject:
		obj := njson.Object(n)
		return &Object{
			node: obj,
			len:  uint32(obj.Len()),
		}
	case njson.TypeNumber:
		num, err := numjson.Parse(s)
		if err != nil {
			if numErr, ok := err.(*strconv.NumError); ok {
				if _, ok := numErr.Err.(*numjson.TooBigError); ok {
					if b, ok := big.NewInt(0).SetString(s, 10); ok {
						return starlark.MakeBigInt(b)
					}
				}
			}
			// We handle failed parse in the default clause below.
			// This way, any changes to the way parse behaves will not break this code.
		}
		switch num.Type() {
		case numjson.Float:
			return starlark.Float(num.Float64())
		case numjson.Int:
			return starlark.MakeInt64(num.Int64())
		case numjson.Uint:
			u := num.Uint64()
			if v := uint(u); uint64(v) == u {
				return starlark.MakeUint(v)
			}
			b := big.NewInt(0).SetUint64(u)
			return starlark.MakeBigInt(b)
		default:
			// An error occurred during parsing, we fallback to strconv.ParseFloat
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return starlark.Float(f)
			}
			// If that failed as well, return NaN
			return starlark.Float(math.NaN())
		}
	case njson.TypeBoolean:
		return starlark.Bool(njson.Const(s).IsTrue())
	case njson.TypeNull:
		return starlark.None
	default:
		return nil
	}
}

var zeroNode njson.Node

func Node(doc *njson.Document, v starlark.Value) (njson.Node, error) {
	switch v := v.(type) {
	case *Array:
		return njson.Node(v.node), nil
	case *Object:
		return njson.Node(v.node), nil
	case starlark.String:
		return doc.NewString(string(v)), nil
	case starlark.IterableMapping:
		obj := doc.NewObject()
		iter := v.Iterate()
		var key starlark.Value
		for iter.Next(&key) {
			key, ok := key.(starlark.String)
			if !ok {
				return zeroNode, errors.New("invalid dict key")
			}
			val, _, _ := v.Get(key)
			node, err := Node(doc, val)
			if err != nil {
				return zeroNode, err
			}
			obj.Set(string(key), node)
		}
		return obj.Node(), nil
	case starlark.Iterable:
		arr := doc.NewArray()
		iter := v.Iterate()
		var el starlark.Value
		for iter.Next(&el) {
			el, err := Node(doc, el)
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
		return doc.NewFloat(float64(v)), nil
	case starlark.NoneType:
		return doc.Null(), nil
	case starlark.Int:
		return doc.NewNumberString(v.String()), nil
	case interface{ NJSON(d *njson.Document) njson.Node }:
		return v.NJSON(doc), nil
	default:
		return zeroNode, errors.New("cannot handle type")
	}
}

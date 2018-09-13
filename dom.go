package njson

import (
	"strconv"

	"github.com/alxarch/njson/numjson"
	"github.com/alxarch/njson/strjson"
)

func (n *Node) SetNull() {
	n.info = vNull
}

func (n *Node) Remove(c *Node) {
	if n != nil && n.info == vArray {
		for i := range n.values {
			v := n.values[i].Value
			if v == c {
				if j := i + 1; 0 <= j && j < len(n.values) {
					copy(n.values[i:], n.values[j:])
				}
				if j := len(n.values) - 1; 0 <= j && j < len(n.values) {
					n.values[j] = KV{}
					n.values = n.values[:j]
				}
				return
			}
		}
	}
}
func (n *Node) remove(i int) {
	if 0 <= i && i < len(n.values) {
		if j := len(n.values) - 1; 0 <= j && j < len(n.values) {
			n.values[i] = n.values[j]
			n.values[j] = KV{}
			n.values = n.values[:j]
		}
	}
}

func (n *Node) Get(key string) *Node {
	if n == nil || !n.info.IsObject() {
		return nil
	}
	for _, k := range n.values {
		if k.Key == key {
			return k.Value
		}
	}
	return nil
}
func (n *Node) Index(i int) *Node {
	if n == nil || !n.info.HasLen() {
		return nil
	}
	if 0 <= i && i < len(n.values) {
		return n.values[i].Value
	}
	return nil
}
func MakeArray(values ...*Node) Node {
	kvs := make([]KV, len(values))
	if len(kvs) >= len(values) {
		kvs = kvs[:len(values)]
		for i := range values {
			kvs[i].Value = values[i]
		}
	}
	return Node{
		info:   vArray,
		values: kvs,
	}
}
func MakeObject(values map[string]*Node) Node {
	kvs := make([]KV, 0, len(values))
	for k, v := range values {
		kvs = append(kvs, KV{k, v})
	}
	return Node{
		info:   vObject,
		values: kvs,
	}
}
func MakeNode(raw string, inf Info) Node {
	return Node{inf, raw, nil}
}

func (n *Node) Lookup(path ...string) *Node {
	for _, key := range path {
		if n == nil {
			return nil
		}
		switch Type(n.info) {
		case TypeObject:
			for _, kv := range n.values {
				if kv.Key == key {
					n = kv.Value
					break
				}
			}
		case TypeArray:
			if i, err := strconv.Atoi(key); err == nil && 0 <= i && i <= len(n.values) {
				n = n.values[i].Value
			}
		default:
			return nil
		}
	}
	return n
}
func (n *Node) Delete(key string) {
	if n == nil || !n.info.IsObject() {
		return
	}
	for i, kv := range n.values {
		if kv.Key == key {
			n.remove(i)
			return
		}
	}
}
func (n *Node) Slice(i, j int) {
	if n == nil || !n.info.IsArray() {
		return
	}
	if 0 <= i && i < len(n.values) {
		if i <= j && j <= len(n.values) {
			n.values = n.values[i:j]
		}
	}
}

func (n *Node) Set(key string, v *Node) {
	if n == nil || !n.info.IsObject() {
		return
	}
	for i := range n.values {
		if n.values[i].Key == key {
			n.values[i].Value = v
			return
		}
	}
	n.values = append(n.values, KV{key, v})
}

func (n *Node) SetBool(t bool) {
	if t {
		n.info = vTrue
	} else {
		n.info = vFalse
	}
}

func (n *Node) SetNumber(f float64) {
	buf := make([]byte, 0, 64)
	buf = numjson.AppendFloat(buf, f, 64)
	n.raw = b2s(buf)
	n.info = vNumber
}

func (n *Node) SetHTML(s string) {
	n.raw = strjson.Escaped(s, true, false)
	n.info = vString
}
func (n *Node) SetString(s string) {
	n.raw = strjson.Escaped(s, false, false)
	n.info = vString
}

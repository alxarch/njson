package foojson

import (
	"github.com/alxarch/njson/strjson"
	"strconv"
)

type Type uint8

const (
	TypeInvalid Type = iota
	TypeString  Type = 1 << iota
	TypeObject
	TypeArray
	TypeNumber
	TypeBoolean
	TypeNull
	TypeAnyValue = TypeString | TypeNumber | TypeBoolean | TypeObject | TypeArray | TypeNull
)

type nodeInfo uint8

const (
	infRoot = 1 << iota
	infEmpty
	infSimpleString
	infUnescaped
)

type Document struct {
	nodes   []value
	rev     uint
	scratch []byte
}

func (d *Document) get(id, rev uint) *value {
	if d != nil && d.rev == rev && id < uint(len(d.nodes)) {
		return &d.nodes[id]
	}
	return nil
}

func (d *Document) unescape(s string) string {
	buf := d.buffer(len(s))
	buf = buf[:strjson.Unescape(buf, s)]
	return string(buf)
}

func (d *Document) buffer(size int) []byte {
	if buf := d.scratch; 0 <= size && size <= cap(buf) {
		return buf[:size]
	}
	d.scratch = make([]byte, size)
	return d.scratch
}

type value struct {
	typ  Type
	info nodeInfo
	raw  string
	next uint
}

func (n value) IsEmptyObject() bool {
	return n.Type() == TypeObject && n.raw == "" && n.next == 0
}

func (n value) Type() Type {
	return n.typ
}

func (n value) IsEmptyArray() bool {
	return n.Type() == TypeArray && n.next == 0
}

func (n *value) KeyRaw() string {
	if n.Type() == TypeObject {
		return n.raw
	}
	return ""
}

func (n *value) Key() (string, bool) {
	if n.Type() == TypeObject {
		if n.info.needsUnescape() {
			n.unescape()
		}
		return n.raw, true
	}
	return "", false
}

func (i nodeInfo) isRoot() bool {
	return i&infRoot == infRoot
}
func (i nodeInfo) isEmpty() bool {
	return i&infEmpty == infEmpty
}
func (i nodeInfo) needsUnescape() bool {
	return i&(infSimpleString|infUnescaped) == 0
}

// Node is a reference to a JSON value in a Document
type Node struct {
	id  uint
	rev uint
	doc *Document
}

// Document returns the document this Node references.
func (n *Node) Document() *Document {
	if n.doc != nil && n.doc.rev == n.rev {
		return n.doc
	}
	return nil
}

func (n *Node) Type() Type {
	if n := n.value(); n != nil {
		return n.Type()
	}
	return TypeInvalid
}

func (n *Node) value() *value {
	return n.doc.get(n.id, n.rev)
}

func (n Node) jump(id uint) Node {
	n.id = id
	return n
}

func (n Node) ToString() (string, bool) {
	if v := n.value(); v != nil && v.Type() == TypeString {
		if v.info.needsUnescape() {
			return n.doc.unescape(v.raw), true
		}
		return v.raw, true
	}
	return "", false
}

func (n Node) Iter() (it Iter) {
	if d := n.doc; d != nil && n.rev == d.rev && n.id < uint(len(d.nodes)) {
		copy(it.pair[:], d.nodes[n.id:])
	}
	return
}

type Iter struct {
	pair [2]value
	index  int
}

func (a *Iter) Index() int {
	return a.index
}

func (a *Iter) Next() bool {
	switch v := a.pair[0]; v.typ{
	case TypeObject, TypeArray:
		return !v.info.isEmpty()
	default:
		return false
	}
}

func (a *Iter) Key() string {
	v := a.cursor.value()
	if v == nil {
		return ""
	}
	switch v.Type() {
	case TypeObject:
		if v.info.needsUnescape() {
			return a.cursor.doc.unescape(v.raw)
		}
		return v.raw
	case TypeArray:
		return strconv.Itoa(a.index)
	default:
		return ""
	}
}

func (a *Iter) Value() Node {
	return a.cursor.jump(a.cursor.id + 1)
}

func (d *Document) copyNode(other *Document, id uint) uint {
	src := other.get(id, other.rev)
	dst := uint(len(d.nodes))
	d.nodes = append(d.nodes, value{
		typ:  src.typ,
		info: src.info,
		raw:  src.raw,
	})
	switch src.typ {
	case TypeArray, TypeObject:
		if src.next != 0 {
			_ = d.copyNode(other, id+1)
			d.nodes[dst].next = d.copyNode(other, src.next)
		}
		return dst
	default:
		return dst
	}
}

func (d *Document) copyOrAdopt(other *Document, id, parent uint) uint {
	n := other.get(id, other.rev)
	if n == nil {
		return 0
	}
	if other == d && id != parent && n.info.isRoot() && id != 0 {
		n.info &^= infRoot
		return id
	}
	return d.copyNode(other, id)
}

func (n Node) Set(key string, value Node) {
	v := n.value()
	if v == nil {
		return
	}
	if v.Type() != TypeObject {
		return
	}
	id := n.doc.copyOrAdopt(value.Document(), value.id, n.id)
	if id == 0 {
		return
	}
	v :=

}

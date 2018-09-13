package njson

import (
	"encoding/json"
	"io"
	"math"
	"sync"

	"github.com/alxarch/njson/numjson"
	"github.com/alxarch/njson/strjson"
)

type Node struct {
	info Info
	raw  string // json string
	// key    string
	values []KV
}
type KV struct {
	Key   string
	Value *Node
}

// func (n *Node) Key() string {
// 	return n.key
// }

func (n *Node) Values() []KV {
	if n != nil && n.info.HasLen() {
		return n.values
	}
	return nil

}

const minNumValues = 8

// func (n *Node) append(v *Node, i uint) {
// 	if i < uint(len(n.values)) {
// 		n.values[i] = v
// 		return
// 	}
// 	if tmp := make([]*Node, 2*len(n.values)+minNumValues); len(tmp) >= len(n.values) {
// 		copy(tmp[:len(n.values)], n.values)
// 		if i < uint(len(tmp)) {
// 			tmp[i] = v
// 		}
// 		n.values = tmp
// 	}
// 	return
// }

func (n *Node) Raw() string {
	return n.raw
}
func (n *Node) Bytes() []byte {
	return s2b(n.raw)
}
func (n *Node) Info() Info {
	return n.info
}

// Appender is a Marshaler interface for buffer append workflows.
type Appender interface {
	AppendJSON([]byte) ([]byte, error)
}

func (n *Node) AppendJSON(dst []byte) ([]byte, error) {
	switch Type(n.info) {
	case TypeObject:
		dst = append(dst, delimBeginObject)
		for i, kv := range n.values {
			if i > 0 {
				dst = append(dst, delimValueSeparator)
			}
			dst = append(dst, delimString)
			dst = append(dst, kv.Key...)
			dst = append(dst, delimString, delimNameSeparator)
			dst, _ = kv.Value.AppendJSON(dst)
		}
		dst = append(dst, delimEndObject)
	case TypeArray:
		dst = append(dst, delimBeginArray)
		for i, kv := range n.values {
			if i > 0 {
				dst = append(dst, delimValueSeparator)
			}
			dst, _ = kv.Value.AppendJSON(dst)
		}
		dst = append(dst, delimEndArray)
	case TypeString:
		dst = append(dst, delimString)
		dst = append(dst, n.raw...)
		dst = append(dst, delimString)
	case TypeNumber:
		dst = append(dst, n.raw...)
	case TypeNull:
		dst = append(dst, strNull...)
	case TypeBoolean:
		if n.info.IsTrue() {
			dst = append(dst, strTrue...)
		} else {
			dst = append(dst, strFalse...)
		}
		// default:
	}
	return dst, nil

}

// Len gets the number of children a node has.
func (n *Node) Len() (i int) {
	if n != nil && n.info.HasLen() {
		return len(n.values)
	}
	return
}

// String returns the unescaped string form of the Node.
// The returned string is safe to use even if ParseUnsafe was used.
func (n *Node) String() string {
	if n == nil {
		return ""
	}
	if n.info.IsString() {
		s := strjson.Unescaped(n.raw)
		if s == n.raw && !n.info.Safe() {
			// When input is unsafe we need to copy the string so
			// any calls to Unescaped() return a safe string to use.
			return scopy(s)
		}
		return s
	}
	if n.info.IsNumber() {
		if n.info.Safe() {
			return n.raw
		}
		return scopy(n.raw)
	}
	switch n.info {
	case vNull:
		return strNull
	case vTrue:
		return strTrue
	case vFalse:
		return strFalse
	default:
		return ""
	}
}

// func (n *Node) unescape() {
// }

// func (n *Node) UnescapedBytes() []byte {
// 	if n == nil {
// 		return nil
// 	}
// 	if n.info.Unescaped() {
// 		return s2b(n.unescaped)
// 	}
// 	if n.info.Quoted() {
// 		n.unescape()
// 		return s2b(n.unescaped)
// 	}
// 	return nil
// }

// // AppendUnescaped appends the unescaped string form of the Node to dst.
// func (n *Node) AppendUnescaped(dst []byte) []byte {
// 	if n == nil {
// 		return dst
// 	}
// 	if n.info.Unescaped() {
// 		return append(dst, n.unescaped...)
// 	}
// 	if n.info.Quoted() {
// 		n.unescape()
// 		return append(dst, n.unescaped...)
// 	}
// 	return dst
// }

// Type returns the type of the node
func (n *Node) Type() Type {
	if n == nil {
		return TypeInvalid
	}
	return n.info.Type()
}

var (
	emptyArrayBytes  = []byte{delimBeginArray, delimEndArray}
	emptyObjectBytes = []byte{delimBeginObject, delimEndObject}
)

// WrapUnmarshalJSON wraps a call to the json.Unmarshaler interface
func (n *Node) WrapUnmarshalJSON(u json.Unmarshaler) (err error) {
	switch n.Type() {
	case TypeArray:
		if len(n.values) == 0 {
			return u.UnmarshalJSON(emptyArrayBytes)
		}
	case TypeObject:
		if len(n.values) == 0 {
			return u.UnmarshalJSON(emptyObjectBytes)
		}
	case TypeInvalid:
		return n.TypeError(TypeAnyValue)
	default:
		return u.UnmarshalJSON(s2b(n.raw))
	}
	data := bufferpool.Get().([]byte)
	data, _ = n.AppendJSON(data[:0])
	err = u.UnmarshalJSON(data)
	bufferpool.Put(data)
	return
}

const minBufferSize = 512

var bufferpool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, minBufferSize)
	},
}

func blankBuffer(size int) []byte {
	if b := bufferpool.Get().([]byte); 0 <= size && size <= cap(b) {
		return b[:size]
	}
	if size < minBufferSize {
		size = minBufferSize
	}
	return make([]byte, size)
}

func putBuffer(b []byte) {
	if b != nil && cap(b) >= minBufferSize {
		bufferpool.Put(b)
	}
}

// Printer is a Marshaler interface for io.Writer workflows.
type Printer interface {
	PrintJSON(w io.Writer) (int, error)
}

// PrintJSON implements the Printer interface
func (n *Node) PrintJSON(w io.Writer) (int, error) {
	return PrintJSON(w, n)
}

// PrintJSON is a helper to write an Appender to an io.Writer
func PrintJSON(w io.Writer, a Appender) (n int, err error) {
	b := bufferpool.Get().([]byte)
	if b, err = a.AppendJSON(b[:0]); err == nil {
		n, err = w.Write(b)
	}
	bufferpool.Put(b)
	return
}

// Unmarshaler unmarshals from a Node
type Unmarshaler interface {
	UnmarshalNodeJSON(*Node) error
}

// ToInterface converts a node to any combatible go value (many allocations on large trees).
func (n *Node) ToInterface() (interface{}, bool) {
	switch n.Type() {
	case TypeObject:
		m := make(map[string]interface{}, n.Len())
		ok := false
		for _, kv := range n.values {
			if m[kv.Key], ok = kv.Value.ToInterface(); !ok {
				return nil, false
			}
		}
		return m, true
	case TypeArray:
		s := make([]interface{}, len(n.values))
		if len(n.values) == len(s) {
			ok := false
			// Avoid bounds check
			s = s[:len(n.values)]
			for i, kv := range n.values {
				if s[i], ok = kv.Value.ToInterface(); !ok {
					return nil, false

				}
			}

		}
		return s, true
	case TypeString:
		return n.String(), true
	case TypeBoolean:
		switch n.info {
		case vTrue:
			return true, true
		case vFalse:
			return false, true
		default:
			return nil, false
		}
	case TypeNull:
		return nil, true
	case TypeNumber:
		return n.ToFloat()
	default:
		return nil, false
	}

}

// TypeError creates an error for the Node's type.
func (n *Node) TypeError(want Type) error {
	return newTypeError(n.Type(), want)
}

func (n *Node) ToUint() (uint64, bool) {
	f := numjson.ParseFloat(n.raw)
	return uint64(f), 0 <= f && f < math.MaxUint64 && math.Trunc(f) == f
}
func (n *Node) ToInt() (int64, bool) {
	f := numjson.ParseFloat(n.raw)
	return int64(f), math.MinInt64 <= f && f < math.MaxInt64 && math.Trunc(f) == f
}
func (n *Node) ToFloat() (float64, bool) {
	f := numjson.ParseFloat(n.raw)
	return f, f == f
}

func (n *Node) ToBool() (bool, bool) {
	switch n.info {
	case vTrue:
		return true, true
	case vFalse:
		return false, true
	default:
		return false, false
	}

}

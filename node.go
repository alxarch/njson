package njson

import (
	"encoding/json"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/alxarch/njson/strjson"
)

type Node struct {
	info   Info
	unsafe bool
	raw    string // json string
	str    string // unescaped string
	num    float64
	key    string
	values []*Node
}

func (n *Node) Key() string {
	return n.key
}

func (n *Node) Values() []*Node {
	if n == nil || n.info&(vObject|vArray) == 0 {
		return nil
	}
	return n.values

}

const minNumValues = 8

func (n *Node) append(v *Node, i int) {
	if 0 <= i && i < len(n.values) {
		n.values[i] = v
		return
	}
	if tmp := make([]*Node, 2*len(n.values)+minNumValues); len(tmp) >= len(n.values) {
		if i = copy(tmp[:len(n.values)], n.values); 0 <= i && i < len(tmp) {
			tmp[i] = v
		}
		n.values = tmp
	}
	return
}

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
		for i, n := range n.values {
			if i > 0 {
				dst = append(dst, delimValueSeparator)
			}
			dst = append(dst, delimString)
			dst = append(dst, n.key...)
			dst = append(dst, delimString, delimNameSeparator)
			dst, _ = n.AppendJSON(dst)
		}
		dst = append(dst, delimEndObject)
	case TypeArray:
		dst = append(dst, delimBeginArray)
		for i, n := range n.values {
			if i > 0 {
				dst = append(dst, delimValueSeparator)
			}
			dst, _ = n.AppendJSON(dst)
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
	if n.info.Unescaped() {
		return n.str
	}
	if n.info.IsString() {
		n.str = strjson.Unescaped(n.raw)
		if n.str == n.raw && !n.info.Safe() {
			// When input is unsafe we need to copy the string so
			// any calls to Unescaped() return a safe string to use.
			b := strings.Builder{}
			b.WriteString(n.str)
			n.str = n.String()
		}
		n.info |= Unescaped
		return n.str
	}
	if n.info.IsNumber() {
		if n.info.Safe() {
			n.str = n.raw
		} else {
			n.str = scopy(n.raw)
		}
		n.info |= Unescaped
		return n.str
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
		for _, k := range n.values {
			if m[k.key], ok = k.ToInterface(); !ok {
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
			for i, n := range n.values {
				if s[i], ok = n.ToInterface(); !ok {
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

func (n *Node) parseFloat() (f float64, ok bool) {
	f, err := strconv.ParseFloat(n.raw, 10)
	if ok = err == nil; ok {
		n.num = f
		n.info |= NumberParsed
		if math.Trunc(f) == f {
			n.info |= NumberZeroDecimal
		}
		// if f < 0 {
		// 	n.info |= NumberSigned
		// }
	}
	return

}
func (n *Node) ToUint() (uint64, bool) {
	if n.info.ToUint() {
		return uint64(n.num), true
	}
	if n.info.IsNumber() {
		n.parseFloat()
		return uint64(n.num), n.info&vNumberInt == vNumberUint
	}
	return 0, false
}
func (n *Node) ToFloat() (float64, bool) {
	if n.info.NumberParsed() {
		return n.num, true
	}
	if n.info.IsNumber() {
		return n.parseFloat()
	}
	return 0, false
}

func (n *Node) ToInt() (int64, bool) {
	// vNumberInt is a superset of vNumberUint
	if n.info.ToInt() {
		return int64(n.num), true
	}
	if n.info.IsNumber() {
		n.parseFloat()
		return int64(n.num), n.info&vNumberUint == vNumberUint
	}
	return 0, false
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

func (n *Node) SetNumber(f float64) {
	// n.raw = FormatFloat()
	// n.num = f
	// n.info = TypeNumber
}

func (n *Node) SetString(s string) {
	n.raw = strjson.Unescaped(s)
	n.str = s
	n.info = vString
}

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
	safe   bool
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
	// if cap(n.values) == 0 {
	// 	n.values = []*Node{v, nil, nil, nil, nil, nil, nil, nil}
	// 	return
	// }
	tmp := make([]*Node, 2*len(n.values)+minNumValues)
	if i = copy(tmp, n.values); 0 <= i && i < len(tmp) {
		tmp[i] = v
	}
	n.values = tmp
	return
}

func (n *Node) Raw() string {
	// if n != nil && n.info.HasRaw() {
	return n.raw
	// }
	// return ""
}
func (n *Node) Bytes() []byte {
	return s2b(n.raw)
}

// Appender is a Marshaler interface for buffer append workflows.
type Appender interface {
	AppendJSON([]byte) ([]byte, error)
}

func (n *Node) AppendJSON(dst []byte) ([]byte, error) {
	if n == nil {
		return dst, nil
	}
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
	default:
		dst = append(dst, n.raw...)
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

// Unescaped returns the unescaped string form of the Node
// The returned string is safe to use as a value even if ParseUnsafe was used
func (n *Node) Unescaped() string {
	if n == nil {
		return ""
	}
	if n.info.Unescaped() {
		return n.str
	}
	if n.info == vString {
		if strings.IndexByte(n.raw, delimEscape) == -1 {
			if n.safe {
				n.str = n.raw
			} else {
				// When input is unsafe we need to copy the string so
				// any calls to Unescaped() return a safe string to use.
				n.str = scopy(n.raw)
			}
			return n.str
		}
		b := blankBuffer(strjson.MaxUnescapedLen(n.raw))
		b = strjson.Unescape(b[:0], n.raw)
		n.str = string(b)
		putBuffer(b)
		n.info |= Unescaped

		return n.str
	}
	return ""
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
		return n.Unescaped(), true
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

const (
	vNumberParsed = vNumber | NumberParsed
)

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
	switch n.info {
	case vNumberUint:
		return uint64(n.num), true
	case vNumber:
		n.parseFloat()
		return uint64(n.num), n.info == vNumberUint
	}
	return 0, false
}
func (n *Node) ToFloat() (float64, bool) {
	if n.info&vNumberParsed == vNumberParsed {
		return n.num, true
	}
	if n.info&vNumber == vNumber {
		return n.parseFloat()
	}
	return 0, false
}

func (n *Node) ToInt() (int64, bool) {
	if n.info&vNumberUint == vNumberUint {
		return int64(n.num), true
	}
	if Type(n.info) == TypeNumber {
		n.parseFloat()
		return int64(n.num), n.info&vNumberUint == vNumberUint
	}
	return 0, false
}

func (n *Node) ToString() (string, bool) {
	return n.Unescaped(), n.info.Unescaped()
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

func (n *Node) IsNull() bool {
	return n.info == vNull
}
func (n *Node) IsArray() bool {
	return n.info == vArray
}
func (n *Node) IsValue() bool {
	const vAnyValue = Info(TypeAnyValue)
	return n.info&vAnyValue != 0
}
func (n *Node) IsString() bool {
	return n.info&vString == vString
}
func (n *Node) IsObject() bool {
	return n.info == vObject
}

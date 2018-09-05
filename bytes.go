//+build !appengine

package njson

import (
	"reflect"
	"unsafe"
)

// Flag to indicate unsafe byte operations
const safebytes = false

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) []byte {
	h := (*reflect.StringHeader)(unsafe.Pointer(&s))
	b := reflect.SliceHeader{
		Data: h.Data,
		Len:  h.Len,
		Cap:  h.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&b))
}

// func b2sEqual(b []byte, s string) bool {
// 	return b2s(b) == s
// }

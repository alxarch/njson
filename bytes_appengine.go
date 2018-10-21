//+build appengine

package njson

const safebytes = true

func s2b(s string) []byte {
	return []byte(s)
}
func b2s(b []byte) string {
	return string(b)
}

func scopy(s string) string {
	b := make([]byte, len(s))
	copy(b, s)
	return string(b)
}

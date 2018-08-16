//+build appengine

package njson

func s2b(s string) []byte {
	return []byte(s)
}
func b2s(b []byte) string {
	return string(b)
}

//+build appengine

package njson

func s2b(s string) []byte {
	return []byte(s)
}
func b2s(b []byte) string {
	return string(b)
}

func b2sEqual(b []byte, s string) bool {
	if len(b) == len(s) {
		for i := range b {
			if b[i] != s[i] {
				return false
			}
		}
		return true
	}
	return false
}

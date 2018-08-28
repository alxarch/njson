//+build !appengine

package njson

import "testing"

func TestS2B(t *testing.T) {
	var s string
	b := s2b(s)
	if b != nil {
		t.Errorf("Invalid b %#v", b)
		return

	}
	s = "foo"
	b = s2b(s)
	if len(b) != 3 {
		t.Errorf("Invalid b %#v", b)
		return
	}
}

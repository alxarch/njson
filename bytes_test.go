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

func Test_scopy(t *testing.T) {
	b := []byte("Lorem ipsum dolor")
	s := b2s(b)
	ss := scopy(s)
	b[0] = 'F'
	if s == ss {
		t.Errorf("Invalid byte string conversion: %s == %s", s, ss)
	}

}
func Benchmark_scopy(b *testing.B) {
	s := "Lorem ipsum dolor"
	var ss string
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ss = scopy(s)
	}
	_ = ss
}

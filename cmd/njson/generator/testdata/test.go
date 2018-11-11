package testdata

import "time"
import "github.com/alxarch/njson/generator/testdata/internal/foo"

type Bar struct {
	Baz string
}
type Foo struct {
	Keywords []string
	Time     time.Duration
	Map      map[string]Baz
	N        int64
	Bar      *Bar
	*Baz
	foo foo.InternalFoo
}

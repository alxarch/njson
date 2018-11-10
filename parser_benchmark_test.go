package njson

import (
	"io/ioutil"
	"strings"
	"testing"
)

func benchmark(src string) func(b *testing.B) {
	p := Document{}
	return func(b *testing.B) {
		n, tail, err := p.Parse(src)
		if err != nil {
			b.Errorf("Parse error: %s", err)
			return
		}
		if strings.TrimSpace(tail) != "" {
			b.Errorf("Non empty tail: %q", tail)
			return
		}
		if n.get() == nil {
			b.Errorf("Nil root")
			return
		}
		b.ReportAllocs()
		b.SetBytes(int64(len(src)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			p.Reset()
			p.Parse(src)
		}
	}
}
func BenchmarkParse(b *testing.B) {
	b.Run("small.json", benchmark(smallJSON))
	b.Run("medium.min.json", benchmark(mediumJSON))
	b.Run("medium.json", benchmark(mediumJSONFormatted))
	b.Run("large.json", benchmark(largeJSON))
	b.Run("twitter.json", benchmark(twitterJSON))
	b.Run("canada.json", benchmark(canadaJSON))

}

var (
	largeJSON           string
	mediumJSON          string
	mediumJSONFormatted string
	smallJSON           string
	twitterJSON         string
	canadaJSON          string
)

func init() {
	if data, err := ioutil.ReadFile("./testdata/large.min.json"); err != nil {
		panic(err)
	} else {
		largeJSON = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/medium.min.json"); err != nil {
		panic(err)
	} else {
		mediumJSON = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/small.json"); err != nil {
		panic(err)
	} else {
		smallJSON = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/medium.json"); err != nil {
		panic(err)
	} else {
		mediumJSONFormatted = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/twitter.json"); err != nil {
		panic(err)
	} else {
		twitterJSON = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/canada.json"); err != nil {
		panic(err)
	} else {
		canadaJSON = string(data)
	}
}

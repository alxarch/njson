[![Build Status](https://travis-ci.org/alxarch/njson.svg)](https://travis-ci.org/alxarch/njson)
[![GoDoc](https://godoc.org/github.com/alxarch/njson?status.svg)](http://godoc.org/github.com/alxarch/njson)
[![Go Report](https://goreportcard.com/badge/github.com/alxarch/njson)](https://goreportcard.com/report/github.com/alxarch/njson)
[![codecov](https://codecov.io/gh/alxarch/njson/branch/master/graph/badge.svg)](https://codecov.io/gh/alxarch/njson)
# njson

High performance parsing and manipulation of JSON documents for Go.

Inspired by [`github.com/valyala/fastjson`](https://github.com/valyala/fastjson)

## Features

  - Does *not* try to be a 'drop-in' replacement for `encoding/json`
  - Deserialize arbitrary JSON input to a DOM tree
  - Manipulate DOM tree
  - Path lookups
  - Lazy unescape and number conversions for faster parsing
  - Reserialze to JSON data
  - Iterate over tree
  - Documents can be reused to avoid allocations
  - Fast, fast, fast
  - [WIP] Support for `reflect` based struct Marshal/Unmarshal via `github.com/alxarch/njson/unjson` package
  - [WIP] CLI tool for Marshal/Unmarshal generated code via `github.com/alxarch/njson/cmd/njson` package

## Usage

```go

	d := njson.Document{}

	root, _, _ := d.Parse(`{"answer":42, "foo": {"bar": "baz"}}`)

	answer, _ := root.Get("answer").ToInt()
	fmt.Println(answer)

	n := root.Lookup("foo", "bar")
	bar := n.Unescaped()
	fmt.Println(bar)

	n.SetString("Hello, 世界")

	data := make([]byte, 64)
	data, _ = root.AppendJSON(data[:0])
	fmt.Println(string(data))

	// Output:
	// 42
	// baz
	// {"answer":42,"foo":{"bar":"Hello, 世界"}}

```




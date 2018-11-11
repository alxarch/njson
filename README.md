# njson

High performance parsing and manipulation of JSON documents for Go.

Inspired by [`github.com/valyala/fastjson`](https://github.com/valyala/fastjson)

## Features

    * Does *not* try to be a 'drop-in' replacement for `encoding/json`
    * Deserialize arbitrary JSON input to a DOM tree
    * Manipulate DOM tree
    * Path lookups
    * Lazy unescape and number conversions for faster parsing
    * Reserialze to JSON data
    * Iterate over tree
    * Documents can be reused to avoid allocations
    * Fast, fast, fast
    * [WIP] Support for `reflect` based struct Marshal/Unmarshal via `github.com/alxarch/njson/unjson` package
    * [WIP] CLI tool for Marshal/Unmarshal generated code via `github.com/alxarch/njson/cmd/njson` package

## Usage

```go
    import "github.com/alxarch/njson"

    d := njson.Document{}

    root, tail, err := d.Parse(`{"answer":42, {"foo": {"bar": "baz"}}}`)
    answer, ok := root.Get("answer").ToInt()
    if ok {
        fmt.Println("answer is", answer)
    }
    n := root.Lookup("foo", "bar").
    bar, ok := n.ToString()
    if ok {
        fmt.Println("bar is", bar)
    }
    bar.SetString("foo")

    data := root.AppendJSON(nil)
    fmt.Println(string(data)) // {"answer":42, {"foo": {"bar": "foo"}}}


```




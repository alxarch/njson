package unjson_test

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/alxarch/njson"
	"github.com/alxarch/njson/unjson"
)

func BenchmarkJSON_Unmarshal(b *testing.B) {
	var err error
	b.SetBytes(int64(len(mediumJSON)))
	b.ResetTimer()
	m := medium{}
	for i := 0; i < b.N; i++ {
		if err = json.Unmarshal(mediumJSONBytes, &m); err != nil {
			b.Errorf("UnexpectedError: %s", err)
		}
	}
}
func BenchmarkUnmarshalFromString(b *testing.B) {
	var err error
	b.SetBytes(int64(len(mediumJSON)))
	b.ResetTimer()
	m := medium{}
	for i := 0; i < b.N; i++ {
		if err = unjson.UnmarshalFromString(mediumJSON, &m); err != nil {
			b.Errorf("UnexpectedError: %s", err)
		}
	}
}
func BenchmarkParseAndUnmarshalFromNode(b *testing.B) {
	p := njson.Get()
	defer p.Close()
	b.SetBytes(int64(len(mediumJSON)))
	b.ResetTimer()
	m := medium{}
	var err error
	var n *njson.Node
	for i := 0; i < b.N; i++ {
		if n, err = p.Parse(mediumJSON); err != nil {
			b.Errorf("UnexpectedError: %s", err)
		}
		if err = unjson.UnmarshalFromNode(n, &m); err != nil {
			b.Errorf("UnexpectedError: %s", err)
		}
	}
}
func BenchmarkUnmarshaler_Unmarshal(b *testing.B) {
	dec, err := unjson.TypeUnmarshaler(reflect.TypeOf(&medium{}), unjson.DefaultOptions())
	if err != nil {
		b.Errorf("UnexpectedError: %s", err)
		return
	}
	if dec == nil {
		b.Errorf("Nil decoder")
		return
	}
	doc := njson.Get()
	root, err := doc.Parse(mediumJSON)
	doc.Close()
	if err != nil {
		b.Errorf("UnexpectedError: %s", err)
		return
	}
	if root == nil {
		b.Errorf("Nil root")
		return
	}
	b.ResetTimer()
	b.SetBytes(int64(len(mediumJSON)))
	m := medium{}
	for i := 0; i < b.N; i++ {
		if err = dec.Unmarshal(&m, root); err != nil {
			b.Errorf("UnexpectedError: %s", err)
		}
	}
}

var (
	mediumJSON      string
	mediumJSONBytes []byte
)

func init() {
	data, err := ioutil.ReadFile("../testdata/medium.min.json")
	if err != nil {
		panic(err)
	}
	mediumJSON = string(data)
	mediumJSONBytes = []byte(data)
}

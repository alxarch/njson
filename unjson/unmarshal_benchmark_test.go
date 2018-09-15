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
	m := medium{}
	if err = unjson.UnmarshalFromString(mediumJSON, &m); err != nil {
		b.Errorf("UnexpectedError: %s", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unjson.UnmarshalFromString(mediumJSON, &m)
	}
}
func BenchmarkParseAndUnmarshal(b *testing.B) {
	d := njson.Blank()
	defer d.Close()
	b.SetBytes(int64(len(mediumJSON)))
	b.ResetTimer()
	m := medium{}
	var err error
	var n njson.Node
	for i := 0; i < b.N; i++ {
		d.Reset()
		if n, _, err = d.Parse(mediumJSON); err != nil {
			b.Errorf("UnexpectedError: %s", err)
		}
		if err = unjson.UnmarshalFromNode(n, &m); err != nil {
			b.Errorf("UnexpectedError: %s", err)
		}
	}
}
func BenchmarkUnmarshaler_Unmarshal(b *testing.B) {
	dec, err := unjson.TypeDecoder(reflect.TypeOf(&medium{}), "")
	if err != nil {
		b.Errorf("UnexpectedError: %s", err)
		return
	}
	if dec == nil {
		b.Errorf("Nil decoder")
		return
	}
	d := njson.Blank()
	defer d.Close()
	b.ResetTimer()
	b.SetBytes(int64(len(mediumJSON)))
	m := medium{}
	var n njson.Node
	for i := 0; i < b.N; i++ {
		d.Reset()
		if n, _, err = d.Parse(mediumJSON); err != nil {
			b.Errorf("UnexpectedError: %s", err)
			return
		}
		if err = dec.Decode(&m, n); err != nil {
			b.Errorf("UnexpectedError: %s", err)
			return
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

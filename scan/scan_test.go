package scan_test

import (
	"bytes"
	"testing"

	"github.com/alxarch/jsonv/scan"
)

func TestNumber(t *testing.T) {
	data := []byte(`  42.42e+42,`)
	num, err := scan.Number(data)
	if err != nil {
		t.Error(err)
	} else if !bytes.Equal(num, data[2:len(data)-1]) {
		t.Errorf("Invalid scan result: %s", num)
	}
}

func TestArray(t *testing.T) {
	data := []byte(`["foo", "bar"]`)
	arr, err := scan.Array(data)
	if err != nil {
		t.Error(err)
	} else if !bytes.Equal(arr, data) {
		t.Errorf("Invalid scan result:\nactual: %s\nexpect: %s", arr, data)
	}

}
func TestString(t *testing.T) {
	data := []byte(`  "foo\\\nbar" `)
	num, err := scan.String(data)
	if err != nil {
		t.Error(err)
	} else if !bytes.Equal(num, data[2:len(data)-1]) {
		t.Errorf("Invalid scan result:\nactual: %s\nexpect: %s", num, data[2:len(data)-1])
	}
}

func TestScanToken(t *testing.T) {
	data := []byte(`["foo","bar"]`)
	tokens, next, err := scan.Tokenize(nil, data)
	if err != nil {
		t.Errorf("Unexpected scan error: %s", err)
	}
	if next != len(data) {
		t.Errorf("Invalid next: %v", next)
	}
	if len(tokens) != 4 {
		t.Errorf("Invalid tokens: %d, %v", len(tokens), tokens)
	}

	out := tokens.AppendTo(nil)
	if string(out) != string(data) {
		t.Errorf("Invalid out: %s", out)

	}

}

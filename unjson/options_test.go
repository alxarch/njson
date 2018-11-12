package unjson

import (
	"reflect"
	"testing"
)

func TestParseField(t *testing.T) {
	tag := reflect.StructTag(`json:",omitempty"`)
	name, h, ok := parseTag(tag, "json")
	assert(t, ok, "Parse OK")
	assertEqual(t, name, "")
	assertEqual(t, h, hintOmitempty)
}

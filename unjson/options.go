package unjson

import (
	"reflect"
	"strings"
)

// Options holds options for an Encoder/Decoder
type Options struct {
	Tag        string // Tag name to use for hints
	OmitEmpty  bool   // Force omitempty on all fields
	OmitMethod string // Method name for checking if a value is empty defaults to 'Omit'
	AllowNaN   bool   // Allow NaN values for numbers
	AllowInf   bool   // Allow ±Inf values for numbers
}

func (o *Options) tagKey() string {
	if o != nil && o.Tag != "" {
		return o.Tag
	}
	return defaultTag
}

func (o *Options) parseField(f reflect.StructField) (name string, hints hint, ok bool) {
	key := o.tagKey()
	name, hints, ok = parseTag(f.Tag, key)
	if name == "" {
		// TODO: add Options.DefaultFieldCase
		name = f.Name
	}
	if o.OmitEmpty {
		hints |= hintOmitempty
	}
	return
}

func (o Options) normalize() Options {
	if o.Tag == "" {
		o.Tag = defaultTag
	}
	if o.OmitMethod == "" {
		o.OmitMethod = defaultOmitMethod
	}
	return o
}

const (
	defaultTag        = "json"
	defaultOmitMethod = "Omit"
)

var (
	defaultOptions = Options{
		Tag:        defaultTag,
		OmitEmpty:  false,
		OmitMethod: defaultOmitMethod,
		AllowInf:   false,
		AllowNaN:   false,
	}
)

// DefaultOptions returns the default options for an Encoder/Decoder
func DefaultOptions() Options {
	return defaultOptions
}

func parseHints(tag string) (hints hint) {
	var hint string
	for len(tag) > 0 {
		tag = tag[1:]
		i := strings.IndexByte(tag, ',')
		if 0 <= i && i < len(tag) {
			hint = tag[:i]
			tag = tag[i:]
		} else {
			hint = tag
			tag = ""
		}
		switch hint {
		case "omitempty":
			hints |= hintOmitempty
		case "8bit":
			hints |= hintRaw
		case "html":
			hints |= hintHTML
		}
	}
	return
}

func parseTag(tag reflect.StructTag, key string) (name string, hints hint, ok bool) {
	if name, ok = tag.Lookup(key); ok {
		if i := strings.IndexByte(name, ','); 0 <= i && i < len(tag) {
			hints = parseHints(name[i:])
			name = name[:i]
		}
	}
	return
}

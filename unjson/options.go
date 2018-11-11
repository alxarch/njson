package unjson

import (
	"reflect"
	"strings"
)

// Options holds options for an Encoder/Decoder
type Options struct {
	Tag        string
	OmitEmpty  bool   // Force omitempty on all fields
	OmitMethod string // Method name for checking if a value is empty
	HTML       bool   // Escape HTML-safe
	AllowNaN   bool   // Allow NaN values for numbers
	AllowInf   bool   // Allow ±Inf values for numbers
}

func (o *Options) parseField(f reflect.StructField) (name string, omiempty, ok bool) {
	p := fieldParser{}
	if o != nil {
		p.Key = o.Tag
		p.OmitEmpty = o.OmitEmpty
	}
	if p.Key == "" {
		p.Key = defaultTag
	}
	return p.parseField(f)
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
		HTML:       false,
		AllowInf:   false,
		AllowNaN:   false,
	}
)

// DefaultOptions returns the default options for an Encoder/Decoder
func DefaultOptions() Options {
	return defaultOptions
}

type fieldParser struct {
	Key       string // Tag key to use for encoder/decoder
	OmitEmpty bool   // Force omitempty on all fields
}

func (o fieldParser) parseField(field reflect.StructField) (tag string, omitempty bool, ok bool) {
	omitempty = o.OmitEmpty
	if tag, ok = field.Tag.Lookup(o.Key); ok {
		if i := strings.IndexByte(tag, ','); 0 <= i && i < len(tag) {
			if !omitempty {
				omitempty = strings.Index(tag[i:], "omitempty") > 0
			}
			tag = tag[:i]
		}
	} else {
		tag = field.Name
	}
	return
}

package unjson

import (
	"reflect"
	"strings"
)

type Options struct {
	Tag       string
	OmitEmpty bool
	// FieldParser           // If nil DefaultFieldParser is used
	OmitMethod string // Method name for checking if a value is empty
	HTML       bool   // Escape HTML-safe
	AllowNaN   bool   // Allow NaN values for numbers
	AllowInf   bool   // Allow ±Inf values for numbers
}

const (
	offset64 = 14695981039346656037
	prime64  = 1099511628211
)

var (
	defaultOptionsHash = defaultOptions.hash()
)

// hashNew initializies a new fnv64a hash value.
func hashNew() uint64 {
	return offset64
}

// hashAddByte adds a byte to a fnv64a hash value, returning the updated hash.
func hashAddByte(h uint64, b byte) uint64 {
	h ^= uint64(b)
	h *= prime64
	return h
}
func hashAddUint64(h, n uint64) uint64 {
	h ^= n
	h *= prime64
	return h
}
func (o *Options) hash() uint64 {
	h := hashNew()
	for _, c := range []byte(o.Tag) {
		h = hashAddByte(h, c)
	}
	for _, c := range []byte(o.OmitMethod) {
		h = hashAddByte(h, c)
	}
	if o.OmitEmpty {
		h = hashAddByte(h, 'O')
	}
	if o.HTML {
		h = hashAddByte(h, 'H')
	}
	if o.AllowInf {
		h = hashAddByte(h, 'I')
	}
	if o.AllowNaN {
		h = hashAddByte(h, 'N')
	}
	return h
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
	// if o.FloatPrecision <= 0 {
	// 	o.FloatPrecision = defaultOptions.FloatPrecision
	// }
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
	defaultFieldParser = NewFieldParser(defaultTag, false)
	defaultOptions     = Options{
		Tag:       defaultTag,
		OmitEmpty: false,
		// FloatPrecision: -1,
		OmitMethod: defaultOmitMethod,
		HTML:       false,
		AllowInf:   false,
		AllowNaN:   false,
	}
)

func DefaultOptions() Options {
	return defaultOptions
}

type FieldParser interface {
	parseField(f reflect.StructField) (name string, omitempty, ok bool)
}

type fieldParser struct {
	Key       string // Tag key to use for encoder/decoder
	OmitEmpty bool   // Force omitempty on all fields
}

func NewFieldParser(key string, omitempty bool) FieldParser {
	if key == "" {
		key = defaultTag
	}
	return fieldParser{key, omitempty}
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

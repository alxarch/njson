package unjson

import (
	"reflect"

	"github.com/alxarch/njson"
)

type structCodec struct {
	fields    []codec
	zeroValue reflect.Value
	typ       reflect.Type
}

func (c *structCodec) Add(f codec) {
	c.fields = append(c.fields, f)
}

func (c *structCodec) Get(key string) *codec {
	for i := range c.fields {
		f := &c.fields[i]
		if f.key == key {
			return f
		}
	}
	return nil
}

// codec is a field encoder/decoder
type codec struct {
	key   string
	index []int // embedded struct index
	decoder
	encoder
	omit omiter
}

// omit checks if a value should be omitted
func (c *structCodec) omit(v reflect.Value) bool {
	for i := range c.fields {
		field := &c.fields[i]
		if f := fieldByIndex(v, field.index); f.IsValid() && !field.omit(f) {
			return false
		}
	}
	return true
}

func (c *structCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	const (
		start = `{,`
		end   = `{}`
	)
	var (
		err  error
		more uint
		fv   reflect.Value
		fc   *codec
	)
	for i := range c.fields {
		fc = &c.fields[i]
		fv = fieldByIndex(v, fc.index)
		if !fv.IsValid() || fc.omit(fv) {
			continue
		}
		b = append(b, start[more])
		more = 1
		b = append(b, delimString)
		b = append(b, fc.key...)
		b = append(b, delimString)
		b = append(b, delimNameSeparator)
		b, err = fc.encode(b, fv)
		if err != nil {
			return b, err
		}
	}
	b = append(b, end[more:]...)
	return b, nil
}

func (c *structCodec) merge(typ reflect.Type, options *Options, index []int, codecs cache) error {
	if typ == nil {
		return nil
	}
	n := typ.NumField()
	v := reflect.New(typ).Elem()
	depth := len(index)
	for i := 0; i < n; i++ {
		// Check field is exported and settable
		if !v.Field(i).CanSet() {
			continue
		}
		field := typ.Field(i)
		tag, hints, tagged := options.parseField(field)
		if tag == "-" {
			continue
		}
		index = append(index[:depth], field.Index...)
		if !tagged && field.Anonymous {
			t := field.Type
			if t.Kind() == reflect.Ptr {
				// Flag for fieldByIndex
				index = append(index, -1)
				t = t.Elem()
			}
			if t.Kind() == reflect.Struct {
				// embedded struct
				if err := c.merge(t, options, index, codecs); err != nil {
					return err
				}
				continue
			}
		}
		// tag = string(strjson.Escape(nil, tag))
		if ff := c.Get(tag); ff != nil && cmpIndex(ff.index, index) != -1 {
			continue
		}

		u, err := codecs.decoder(field.Type, options)
		if err != nil {
			return err
		}
		enc, err := codecs.encoder(field.Type, options, hints)
		if err != nil {
			return err
		}
		omit := omitNever
		if hints&hintOmitempty == hintOmitempty {
			if m, ok := enc.(*structCodec); ok {
				omit = m.omit
			} else {
				omit = newOmiter(field.Type, options.OmitMethod)
			}
		}
		c.Add(codec{
			key:     tag,
			index:   copyIndex(index),
			decoder: u,
			encoder: enc,
			omit:    omit,
		})
	}
	return nil
}

func newStructCodec(typ reflect.Type, options *Options, codecs cache) (*structCodec, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errInvalidType
	}

	if c := codecs.codec(typ); c != nil {
		return c, nil
	}
	c := structCodec{
		typ:       typ,
		fields:    make([]codec, 0, typ.NumField()),
		zeroValue: reflect.Zero(typ),
	}
	codecs[cacheKey{typ, 0}] = &c
	if err := c.merge(typ, options, nil, codecs); err != nil {
		return nil, err
	}
	return &c, nil
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if i == -1 {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		} else {
			v = v.Field(i)
		}
	}
	return v
}

func (c *structCodec) decode(v reflect.Value, n njson.Node) (err error) {
	switch n.Type() {
	case njson.TypeNull:
		v.Set(c.zeroValue)
		return nil
	case njson.TypeObject:
		var (
			field  reflect.Value
			fc     *codec
			values = n.Values()
		)
		for values.Next() {
			fc = c.Get(values.Key())
			if fc == nil {
				continue
			}
			switch len(fc.index) {
			case 1:
				field = v.Field(fc.index[0])
			case 0:
				continue
			default:
				field = v.Field(fc.index[0])
				for _, i := range fc.index[1:] {
					if i == -1 {
						if field.IsNil() {
							field = reflect.New(field.Type().Elem())
						}
						field = field.Elem()
					} else {
						field = field.Field(i)
					}
				}
			}
			if err = fc.decode(field, n.With(values.ID())); err != nil {
				return
			}
		}
		return
	default:
		return n.TypeError(njson.TypeObject | njson.TypeNull)
	}
}

func copyIndex(a []int) (b []int) {
	b = make([]int, len(a))
	copy(b, a)
	return
}

func cmpIndex(a, b []int) int {
	if len(a) > len(b) {
		return -1
	}
	if len(a) == len(b) {
		// Avoid bounds check
		b = b[:len(a)]
		for i, j := range a {
			if jj := b[i]; j > jj {
				return -1
			} else if jj > j {
				return 1
			}
		}
		return 0
	}
	return 1
}

type hint uint8

const (
	hintRaw hint = 1 << iota
	hintHTML
	hintOmitempty
)

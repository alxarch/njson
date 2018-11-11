package unjson

import (
	"reflect"

	"github.com/alxarch/njson"
)

type structCodec struct {
	fields    []codec
	zeroValue reflect.Value
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
	// // Binary search in sorted fields
	// var (
	// 	i, j, h = 0, len(c.fields), 0
	// 	f       *fieldCodec
	// )
	// for i < j {
	// 	h = int(uint(i+j) >> 1) // avoid overflow when computing h
	// 	// i ≤ h < j
	// 	if 0 <= h && h < len(c.fields) {
	// 		f = &c.fields[h]
	// 		if f.Key < key {
	// 			i = h + 1
	// 		} else if f.Key > key {
	// 			j = h
	// 		} else {
	// 			return f.codec
	// 		}
	// 	}
	// }
	// return nil
}

type codec struct {
	key   string
	index []int
	n     int
	decoder
	encoder
	omit omiter
}

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

func (d *structCodec) merge(typ reflect.Type, options *Options, index []int) error {
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
		tag, omitempty, tagged := options.parseField(field)
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
				if err := d.merge(t, options, index); err != nil {
					return err
				}
				continue
			}
		}
		// tag = string(strjson.Escape(nil, tag))
		if ff := d.Get(tag); ff != nil && cmpIndex(ff.index, index) != -1 {
			continue
		}
		u, err := newDecoder(field.Type, options)
		if err != nil {
			return err
		}
		m, err := newEncoder(field.Type, options)
		if err != nil {
			return err
		}
		omit := omitNever
		if omitempty {
			if m, ok := m.(*structCodec); ok {
				omit = m.omit
			} else {
				omit = newOmiter(field.Type, options.OmitMethod)
			}
		}
		d.Add(codec{
			key:     tag,
			index:   copyIndex(index),
			n:       len(index),
			decoder: u,
			encoder: m,
			omit:    omit,
		})
	}
	return nil

}

func newStructCodec(typ reflect.Type, options *Options) (*structCodec, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errInvalidType
	}
	d := &structCodec{
		fields:    make([]codec, 0, typ.NumField()),
		zeroValue: reflect.Zero(typ),
	}
	if err := d.merge(typ, options, nil); err != nil {
		return nil, err
	}
	// // Sort fields for binary search
	// sort.Slice(d.fields, func(i, j int) bool {
	// 	return d.fields[i].Key < d.fields[j].Key
	// })
	return d, nil
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

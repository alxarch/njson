package unjson

import (
	"reflect"

	"github.com/alxarch/njson"
	"github.com/alxarch/njson/strjson"
)

type structCodec struct {
	fields map[string]fieldCodec
	zero   reflect.Value
}

type fieldCodec struct {
	index []int
	n     int
	unmarshaler
	marshaler
	omit omiter
}

func (d *structCodec) omit(v reflect.Value) bool {
	for _, field := range d.fields {
		if f := fieldByIndex(v, field.index); f.IsValid() && !field.omit(f) {
			return false
		}
	}
	return true
}

func (d *structCodec) marshal(b []byte, v reflect.Value) ([]byte, error) {
	var (
		i   = 0
		err error
		fv  reflect.Value
	)
	b = append(b, delimBeginObject)
	for name, field := range d.fields {
		fv = fieldByIndex(v, field.index)
		if !fv.IsValid() || field.omit(fv) {
			continue
		}
		if i > 0 {
			b = append(b, delimValueSeparator)
		}
		b = append(b, delimString)
		b = append(b, name...)
		b = append(b, delimString)
		b = append(b, delimNameSeparator)
		b, err = field.marshal(b, fv)
		if err != nil {
			return b, err
		}
		i++
	}
	b = append(b, delimEndObject)
	return b, nil
}

func (d *structCodec) merge(typ reflect.Type, options Options, index []int) error {
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
		tag = string(strjson.Escape(nil, tag))
		if ff, duplicate := d.fields[tag]; duplicate && cmpIndex(ff.index, index) != -1 {
			continue
		}
		u, err := newUnmarshaler(field.Type, options)
		if err != nil {
			return err
		}
		m, err := newMarshaler(field.Type, options)
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
		d.fields[tag] = fieldCodec{
			index:       copyIndex(index),
			n:           len(index),
			unmarshaler: u,
			marshaler:   m,
			omit:        omit,
		}
	}
	return nil

}

func newStructCodec(typ reflect.Type, options Options) (*structCodec, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errInvalidType
	}
	d := structCodec{
		fields: make(map[string]fieldCodec, typ.NumField()),
		zero:   reflect.Zero(typ),
	}
	if err := d.merge(typ, options, nil); err != nil {
		return nil, err
	}
	return &d, nil
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

func (d *structCodec) unmarshal(v reflect.Value, n *njson.Node) (err error) {
	switch n.Type() {
	case njson.TypeNull:
		v.Set(d.zero)
		return nil
	case njson.TypeObject:
		var (
			field reflect.Value
			fc    fieldCodec
			i, j  int
		)
		for _, n := range n.Values() {
			switch fc = d.fields[n.Key()]; len(fc.index) {
			case 0:
				continue
			case 1:
				field = v.Field(fc.index[0])
			default:
				field = v.Field(fc.index[0])
				for i = 1; 0 <= i && i < len(fc.index); i++ {
					switch j = fc.index[i]; j {
					case -1:
						if field.IsNil() {
							field = reflect.New(field.Type().Elem())
						}
						field = field.Elem()
					default:
						field = field.Field(j)
					}
				}
			}
			if err = fc.unmarshal(field, n); err != nil {
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

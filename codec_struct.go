package njson

import (
	"reflect"
	"unicode"
)

type structCodec struct {
	fields map[string]fieldCodec
	zero   reflect.Value
}

type fieldCodec struct {
	index []int
	decoder
	encoder
	omit omiter
}

type omiter func(v reflect.Value) bool

func (d *structCodec) omit(v reflect.Value) bool {
	for _, field := range d.fields {
		if !field.omit(v.FieldByIndex(field.index)) {
			return false
		}
	}
	return true
}

func (d *structCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
	var (
		i   int
		err error
		fv  reflect.Value
	)
	b = append(b, delimBeginObject)
	for name, field := range d.fields {
		fv = v.FieldByIndex(field.index)
		if field.omit(fv) {
			continue
		}
		if i > 0 {
			b = append(b, delimValueSeparator)
		}
		b = append(b, name...)
		b = append(b, delimNameSeparator)
		b, err = field.encode(b, fv)
		if err != nil {
			return b, err
		}
		i++
	}
	b = append(b, delimEndObject)
	return b, nil
}

func (d *structCodec) merge(typ reflect.Type, options CodecOptions, depth []int) error {
	if typ == nil {
		return nil
	}
	n := typ.NumField()
	for i := 0; i < n; i++ {
		field := typ.Field(i)
		if !isExported(field.Name) {
			continue
		}
		tag, omitempty, tagged := options.tag(field)
		if tag == "-" {
			continue
		}
		var index []int
		if len(depth) > 0 {
			index = append(index, depth...)
		}
		index = append(index, field.Index...)
		if !tagged && field.Anonymous {
			// embedded struct
			if err := d.merge(resolveStruct(field.Type), options, index); err != nil {
				return err
			}
			continue
		}
		tag = QuoteString(tag)
		if ff, duplicate := d.fields[tag]; duplicate && cmpIndex(ff.index, index) != -1 {
			continue
		}
		dec, err := newDecoder(field.Type, options)
		if err != nil {
			return err
		}
		enc, err := newEncoder(field.Type, options)
		if err != nil {
			return err
		}
		omit := omitNever
		if omitempty || options.OmitEmpty {
			if enc, ok := enc.(*structCodec); ok {
				omit = enc.omit
			} else {
				omit = newOmiter(field.Type)
			}
		}
		d.fields[tag] = fieldCodec{
			index:   index,
			decoder: dec,
			encoder: enc,
			omit:    omit,
		}
	}
	return nil

}

func newStructCodec(typ reflect.Type, options CodecOptions) (*structCodec, error) {
	if typ = resolveStruct(typ); typ == nil {
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

func (d *structCodec) decode(v reflect.Value, n *Node) (err error) {
	switch n.Type() {
	case TypeNull:
		v.Set(d.zero)
		return nil
	case TypeObject:
		for n = n.Value(); n != nil; n = n.Next() {
			if f, ok := d.fields[n.src]; ok {
				vv := v.FieldByIndex(f.index)
				if !vv.IsValid() {
					panic(f.index)
				}
				err = f.decode(vv, n.Value())
				if err != nil {
					return
				}
			}
		}
		return
	default:
		return errInvalidNodeType
	}
}

func isExported(name string) bool {
	for _, c := range name {
		if unicode.IsUpper(c) {
			return true
		}
		break
	}
	return false
}

func resolveStruct(typ reflect.Type) reflect.Type {
	switch typ.Kind() {
	case reflect.Ptr:
		return resolveStruct(typ.Elem())
	case reflect.Struct:
		return typ
	default:
		return nil
	}
}

func cmpIndex(a, b []int) int {
	if len(a) > len(b) {
		return -1
	}
	if len(a) < len(b) {
		return 1
	}
	for i, j := range a {
		if jj := b[i]; j > jj {
			return -1
		} else if jj > j {
			return 1
		}
	}
	return 0
}

func omitNil(v reflect.Value) bool {
	return v.IsNil()
}
func omitFloat(v reflect.Value) bool {
	return v.Float() == 0
}
func omitInt(v reflect.Value) bool {
	return v.Int() == 0
}
func omitUint(v reflect.Value) bool {
	return v.Uint() == 0
}
func omitBool(v reflect.Value) bool {
	return !v.Bool()
}
func omitZeroLen(v reflect.Value) bool {
	return v.Len() == 0
}
func omitNever(reflect.Value) bool {
	return false
}
func omitAlways(reflect.Value) bool {
	return true
}
func omitCustom(v reflect.Value) bool {
	return v.Interface().(Omiter).Omit()
}

func newOmiter(typ reflect.Type) omiter {
	if typ == nil {
		return omitAlways
	}
	if typ.Implements(typOmiter) {
		return omitCustom
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return omitNil
	case reflect.Struct:
		return omitNever
	case reflect.Slice, reflect.Map, reflect.String:
		return omitZeroLen
	case reflect.Interface:
		if typ.NumMethod() == 0 {
			return omitNil
		}
		return omitAlways
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return omitInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return omitUint
	case reflect.Float32, reflect.Float64:
		return omitFloat
	case reflect.Bool:
		return omitBool
	default:
		return omitAlways
	}
}

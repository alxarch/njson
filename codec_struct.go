package njson

import (
	"reflect"
)

type structCodec struct {
	fields map[string]fieldCodec
	zero   reflect.Value
}

type fieldCodec struct {
	index []int
	n     int
	decoder
	encoder
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

func (d *structCodec) encode(b []byte, v reflect.Value) ([]byte, error) {
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
	v := reflect.New(typ).Elem()
	for i := 0; i < n; i++ {
		// Check field is exported and settable
		if !v.Field(i).CanSet() {
			continue
		}
		field := typ.Field(i)
		tag, omitempty, tagged := options.ParseField(field)
		if tag == "-" {
			continue
		}
		var index []int
		if len(depth) > 0 {
			index = append(index, depth...)
		}
		index = append(index, field.Index...)
		if !tagged && field.Anonymous {
			t := field.Type
			if t.Kind() == reflect.Ptr {
				// Flag for fieldByIndex
				index = append(index, -1)
				t = field.Type.Elem()
			}
			if t.Kind() == reflect.Struct {
				// embedded struct
				err := d.merge(t, options, index)
				if err != nil {
					return err
				}
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
		if omitempty {
			if enc, ok := enc.(*structCodec); ok {
				omit = enc.omit
			} else {
				omit = newOmiter(field.Type, options.OmitMethod)
			}
		}
		d.fields[tag] = fieldCodec{
			index:   index,
			n:       len(index),
			decoder: dec,
			encoder: enc,
			omit:    omit,
		}
	}
	return nil

}

func newStructCodec(typ reflect.Type, options CodecOptions) (*structCodec, error) {
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

func (d *structCodec) decode(v reflect.Value, n *Node) (err error) {
	switch n.Type() {
	case TypeNull:
		v.Set(d.zero)
		return nil
	case TypeObject:
		var (
			field reflect.Value
			fc    fieldCodec
			i, j  int
		)
		for n = n.Value(); n != nil; n = n.Next() {
			switch fc = d.fields[n.src]; fc.n {
			case 0:
				continue
			case 1:
				field = v.Field(fc.index[0])
			default:
				field = v.Field(fc.index[0])
				for i = 1; i < fc.n; i++ {
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
			if err = fc.decode(field, n.Value()); err != nil {
				return
			}
		}
		return
	default:
		return errInvalidNodeType
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

type omiter func(v reflect.Value) bool

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

func customOmiter(typ reflect.Type, methodName string) omiter {
	if method, ok := typ.MethodByName(methodName); ok {
		f := method.Func.Type()
		if f.NumIn() == 1 && f.NumOut() == 1 && f.Out(0).Kind() == reflect.Bool {
			return omiter(func(v reflect.Value) bool {
				return v.Method(method.Index).Call(nil)[0].Bool()
			})
		}
	}
	return nil
}

// newCustomOmiter creates an omiter func for a type.
//
// It resolves an omiter event if the method is defined on
// the type's pointer or the type's element (if it's a pointer)
func newCustomOmiter(typ reflect.Type, methodName string) omiter {
	if typ == nil {
		return nil
	}
	if methodName == "" {
		methodName = defaultOmitMethod
	}
	if om := customOmiter(typ, methodName); om != nil {
		return om
	}

	switch typ.Kind() {
	case reflect.Ptr:
		// If pointer element implements omiter wrap it
		if om := customOmiter(typ.Elem(), methodName); om != nil {
			return omiter(func(v reflect.Value) bool {
				return v.IsNil() || om(v.Elem())
			})
		}
	default:
		// If pointer to type implements omiter wrap it
		if om := customOmiter(reflect.PtrTo(typ), methodName); om != nil {
			return omiter(func(v reflect.Value) bool {
				return v.CanAddr() && om(v.Addr())
			})
		}
	}
	return nil

}
func omitCustom(v reflect.Value) bool {
	return v.Interface().(Omiter).Omit()
}

func newOmiter(typ reflect.Type, methodName string) omiter {
	if typ == nil {
		return omitAlways
	}
	if om := newCustomOmiter(typ, methodName); om != nil {
		return om
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

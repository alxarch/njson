package unjson

import "reflect"

type Omiter interface {
	Omit() bool
}

var (
	typOmiter = reflect.TypeOf((*Omiter)(nil)).Elem()
)

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
// It resolves an omiter even if the method is defined on
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

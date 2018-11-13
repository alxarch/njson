package unjson

import "reflect"

// Omiter implements custom omitempty logic
type Omiter interface {
	Omit() bool
}

var (
	typOmiter = reflect.TypeOf((*Omiter)(nil)).Elem()
)

type omiter interface {
	omit(v reflect.Value) bool
}
type omiterer interface {
	omiter(methodName string) omiter
}

type omitNil struct{}

func (omitNil) omit(v reflect.Value) bool {
	return v.IsNil()
}

type omitFloat struct{}

func (omitFloat) omit(v reflect.Value) bool {
	return v.Float() == 0
}

type omitInt struct{}

func (omitInt) omit(v reflect.Value) bool {
	return v.Int() == 0
}

type omitUint struct{}

func (omitUint) omit(v reflect.Value) bool {
	return v.Uint() == 0
}

type omitBool struct{}

func (omitBool) omit(v reflect.Value) bool {
	return !v.Bool()
}

type omitZeroLen struct{}

func (omitZeroLen) omit(v reflect.Value) bool {
	return v.Len() == 0
}

type omitNever struct{}

func (omitNever) omit(reflect.Value) bool {
	return false
}

type omitAlways struct{}

func (omitAlways) omit(reflect.Value) bool {
	return true
}

type omitFunc func(reflect.Value) bool

func (f omitFunc) omit(v reflect.Value) bool {
	return f(v)
}

type omitMethod int

func (m omitMethod) omit(v reflect.Value) bool {
	results := v.Method(int(m)).Call(nil)
	if len(results) > 0 {
		return results[0].Bool()
	}
	return false
}

func customOmiter(typ reflect.Type, methodName string) omiter {
	if method, ok := typ.MethodByName(methodName); ok {
		f := method.Func.Type()
		if f.NumIn() == 1 && f.NumOut() == 1 && f.Out(0).Kind() == reflect.Bool {
			return omitMethod(method.Index)
		}
	}
	return nil
}

type ptrOmiter struct {
	omiter
}

func (o *ptrOmiter) omit(v reflect.Value) bool {
	return v.IsNil() || o.omiter.omit(v.Elem())
}

type elemOmiter struct {
	omiter
}

func (o *elemOmiter) omit(v reflect.Value) bool {
	return v.CanAddr() && o.omiter.omit(v.Addr())
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
			return &ptrOmiter{om}
		}
	default:
		// If pointer to type implements omiter wrap it
		if om := customOmiter(reflect.PtrTo(typ), methodName); om != nil {
			return &elemOmiter{om}
		}
	}
	return nil

}
func omitCustom(v reflect.Value) bool {
	return v.Interface().(Omiter).Omit()
}

type arrayOmiter struct {
	size int
	omiter
}

func (o *arrayOmiter) omit(v reflect.Value) bool {
	for i := 0; i < o.size; i++ {
		if !o.omiter.omit(v.Index(i)) {
			return false
		}
	}
	return true
}

func newArrayOmiter(typ reflect.Type, methodName string) omiter {
	return &arrayOmiter{
		omiter: newOmiter(typ.Elem(), methodName),
		size:   typ.Len(),
	}
}

func newOmiter(typ reflect.Type, methodName string) omiter {
	if typ == nil {
		return omitAlways{}
	}
	if om := newCustomOmiter(typ, methodName); om != nil {
		return om
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return omitNil{}
	case reflect.Struct:
		return omitNever{}
	case reflect.Slice, reflect.Map, reflect.String:
		return omitZeroLen{}
	case reflect.Array:
		return newArrayOmiter(typ, methodName)
	case reflect.Interface:
		if typ.NumMethod() == 0 {
			return omitNil{}
		}
		return omitAlways{}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return omitInt{}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return omitUint{}
	case reflect.Float32, reflect.Float64:
		return omitFloat{}
	case reflect.Bool:
		return omitBool{}
	default:
		return omitAlways{}
	}
}

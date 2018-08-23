package njsonutil

import (
	"fmt"
	"reflect"
)

func CheckImplementsTagged(typ reflect.Type, method reflect.Method, tag string) (methodName string, err error) {
	methodName = TaggedMethodName(method.Name, tag)
	if method, ok := typ.MethodByName(methodName); !ok {
		err = fmt.Errorf("Type %s does not have a %s method", typ, methodName)
	} else if fn := method.Type; fn == nil || !fn.ConvertibleTo(method.Type) {
		err = fmt.Errorf("Type %s does not have a valid %s method", typ, method.Type)
	}
	return

}

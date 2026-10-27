package registry

import "reflect"

var METHODS = NewMethods(512)

func For(rv reflect.Value, methodNames []string, validChk ...func(reflect.Value, *reflect.Method) error) (rec reflect.Value, rm *reflect.Method, err error) {
	return METHODS.For(rv, methodNames, validChk...)
}

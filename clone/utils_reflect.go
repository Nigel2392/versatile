package clone

import (
	"reflect"
	"unsafe"
)

type value struct {
	abiType unsafe.Pointer
	ptr     unsafe.Pointer
	flag    reflect.Kind
}

func getCacheKey(v reflect.Value) uintptr {
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.UnsafePointer:
		return v.Pointer()
	default:
		op := *(*value)(unsafe.Pointer(&v))
		return uintptr(op.ptr)
	}
}

func newMapFromPtr(typ reflect.Type, ptr unsafe.Pointer) reflect.Value {
	// creates a valid nil map.
	v := reflect.New(typ)
	v = v.Elem()
	// cast with our cached *hmap pointer.
	(*value)(unsafe.Pointer(&v)).ptr = ptr

	return v
}

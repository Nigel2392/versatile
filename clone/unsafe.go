package clone

import (
	"reflect"
	"unsafe"
)

//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

//go:linkname isSafeConversion github.com/Nigel2392/versatile.isSafeConversion
func isSafeConversion(from, to reflect.Type) bool

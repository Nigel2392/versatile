package danger

import (
	"reflect"
	"unsafe"
)

func UnsafeRunes(rv reflect.Value) []rune {
	return unsafe.Slice((*rune)(rv.UnsafePointer()), rv.Len())
}

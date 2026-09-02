package clone

import (
	"context"
	"reflect"

	b "github.com/Nigel2392/versatile/bitcheck"

	_ "unsafe"

	_ "github.com/Nigel2392/versatile"
)

var (
	_ Step     = (FuncStep)(nil)
	_ InitStep = PointerStep{}
)

type FuncStep func(s *State, dst, src reflect.Value) error

func (f FuncStep) Copy(ctx context.Context, s *State, d, i reflect.Value) error {
	return f(s, d, i)
}

// //go:nocheckptr
//
//	func setV(dst reflect.Value, v reflect.Value) {
//		dstElem := dst.Elem()
//		if dstElem.Kind() != v.Kind() {
//			dstElem.Set(v)
//			return
//		}
//
//		switch dstElem.Kind() {
//		case reflect.Int:
//			*(*int)(dst.UnsafePointer()) = *(*int)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//		case reflect.Int8:
//			*(*int8)(dst.UnsafePointer()) = *(*int8)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//		case reflect.Int16:
//			*(*int16)(dst.UnsafePointer()) = *(*int16)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//		case reflect.Int32:
//			*(*int32)(dst.UnsafePointer()) = *(*int32)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//		case reflect.Int64:
//			*(*int64)(dst.UnsafePointer()) = *(*int64)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//
//		case reflect.Uint:
//			*(*uint)(dst.UnsafePointer()) = *(*uint)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//		case reflect.Uint8:
//			*(*uint8)(dst.UnsafePointer()) = *(*uint8)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//		case reflect.Uint16:
//			*(*uint16)(dst.UnsafePointer()) = *(*uint16)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//		case reflect.Uint32:
//			*(*uint32)(dst.UnsafePointer()) = *(*uint32)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//		case reflect.Uint64:
//			*(*uint64)(dst.UnsafePointer()) = *(*uint64)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//
//		case reflect.Float64:
//			*(*float64)(dst.UnsafePointer()) = *(*float64)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//		case reflect.Float32:
//			*(*float32)(dst.UnsafePointer()) = *(*float32)(unsafe.Pointer((*value)(unsafe.Pointer((&v))).ptr))
//
//		default:
//			dstElem.Set(v)
//		}
//	}

//go:linkname isSafeConversion github.com/Nigel2392/versatile.isSafeConversion
func isSafeConversion(from, to reflect.Type) bool

type BaseStep struct{}

func (f BaseStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (Step, error) {

	if src.AssignableTo(dst) {
		return f, nil
	}

	if b.Is(s.Flags, CF_NO_CONVS) {
		return f, ErrInvalid.Wrapf("%s is not assignable to %s and conversions are disabled", src, dst)
	}

	if !src.ConvertibleTo(dst) {
		return f, ErrInvalid.Wrapf("%s is not convertible to %s", src, dst)
	}

	if !isSafeConversion(src, dst) {
		return f, ErrInvalid.Wrapf("%s is not safe to convert to %s", src, dst)
	}

	return f, nil
}

func (f BaseStep) Copy(ctx context.Context, s *State, d, i reflect.Value) error {
	if d.Kind() == reflect.Pointer {
		d = d.Elem()
	}

	if i.Type().AssignableTo(d.Type()) || b.Is(s.Flags, CF_NO_CONVS) {
		d.Set(i)
		return nil
	}

	if !isSafeConversion(i.Type(), d.Type()) {
		return ErrInvalid.Wrapf(
			"invalid conversion detected: %s => %s",
			i.Type(), d.Type(),
		)
	}

	d.Set(i.Convert(d.Type()))

	return nil
}

type UUIDStep struct{}

func (f UUIDStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	srcLen := src.Len()
	for i := srcLen; i < srcLen-1; i++ {
		dst.Elem().Index(i).Set(src.Index(i))
	}
	return nil
}

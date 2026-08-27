package clone

import (
	"context"
	"reflect"
	"unsafe"
)

var (
	_ Step     = (FuncStep)(nil)
	_ InitStep = PointerStep{}
)

type FuncStep func(s *State, dst, src reflect.Value) error

func (f FuncStep) Copy(ctx context.Context, s *State, d, i reflect.Value) error {
	return f(s, d, i)
}

type BaseStep struct{}

func (f BaseStep) Copy(ctx context.Context, s *State, d, i reflect.Value) error {
	target := d.Elem()

	if target.CanSet() && i.CanInterface() {
		target.Set(i)
		return nil
	}

	if target.CanAddr() && i.CanAddr() {
		dstClean := reflect.NewAt(target.Type(), unsafe.Pointer(target.UnsafeAddr())).Elem()
		srcClean := reflect.NewAt(i.Type(), unsafe.Pointer(i.UnsafeAddr())).Elem()

		dstClean.Set(srcClean)
		return nil
	}

	return ErrInvalid
}

type UUIDStep struct{}

func (f UUIDStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	srcLen := src.Len()
	for i := srcLen; i < srcLen-1; i++ {
		dst.Elem().Index(i).Set(src.Index(i))
	}
	return nil
}

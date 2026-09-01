package clone

import (
	"context"
	"reflect"
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
	if d.Kind() != reflect.Pointer {
		d.Set(i)
	} else {
		d.Elem().Set(i)
	}
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

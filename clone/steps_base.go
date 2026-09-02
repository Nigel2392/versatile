package clone

import (
	"context"
	"reflect"

	b "github.com/Nigel2392/versatile/bitcheck"

	_ "unsafe"

	_ "github.com/Nigel2392/versatile"
)

type FuncStep func(s *State, dst, src reflect.Value) error

func (f FuncStep) Copy(ctx context.Context, s *State, d, i reflect.Value) error {
	return f(s, d, i)
}

type BaseStep struct{}

func (f BaseStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (Step, error) {
	if src == dst || src.AssignableTo(dst) {
		return f, nil
	}

	if b.Is(s.Flags, CF_NO_CONVS) {
		return f, ErrInvalid.Wrapf("%s is not assignable to %s and conversions are disabled", src, dst)
	}

	if !isSafeConversion(src, dst) {
		return f, ErrInvalid.Wrapf("%s is not safe to convert to %s", src, dst)
	}

	if !src.ConvertibleTo(dst) {
		return f, ErrInvalid.Wrapf("%s is not convertible to %s", src, dst)
	}

	return f, nil
}

func (f BaseStep) Copy(ctx context.Context, s *State, d, i reflect.Value) error {
	if d.Kind() == reflect.Pointer {
		d = d.Elem()
	}

	srcTyp := i.Type()
	dstTyp := d.Type()

	if b.Is(s.Flags, CF_NO_CONVS) || srcTyp == dstTyp || srcTyp.AssignableTo(dstTyp) {
		d.Set(i)
		return nil
	}

	if !isSafeConversion(srcTyp, dstTyp) {
		return ErrInvalid.Wrapf(
			"invalid conversion detected: %s => %s",
			srcTyp, dstTyp,
		)
	}

	d.Set(i.Convert(dstTyp))

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

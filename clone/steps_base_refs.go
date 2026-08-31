package clone

import (
	"context"
	"reflect"

	"github.com/Nigel2392/errors"
)

type PointerStep struct {
	step Step
}

func (f PointerStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (_ Step, err error) {
	f.step, err = s.StepInit(ctx, dst, src.Elem())
	return f, err
}

func (f PointerStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	target := dst
	if dst.Kind() == reflect.Pointer && dst.Elem().Kind() == reflect.Pointer {
		target = dst.Elem()
	}

	if src.IsNil() {
		if target.CanSet() {
			target.Set(reflect.Zero(target.Type()))
		}
		return nil
	}

	newPtr, cached := s.New(src, target.Type().Elem())
	if target.CanSet() {
		target.Set(newPtr)
	}

	if cached {
		return nil
	}

	return f.step.Copy(ctx, s, newPtr, src.Elem())
}

type FromInterfaceStep struct{}

func (f FromInterfaceStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	for src.Kind() == reflect.Interface {
		src = src.Elem()
		if !src.IsValid() {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
	}

	step, err := s.StepInit(ctx, dst.Type(), src.Type())
	if err != nil {
		return errors.Wrap(err, "FromInterfaceStep.Copy")
	}

	err = step.Copy(ctx, s, dst, src)
	if err != nil {
		err = errors.Wrap(err, "FromInterfaceStep.Copy")
	}
	return err
}

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
	// Bepaal het onderliggende type van de bestemming, tenzij het een interface is.
	dstTyp := dst
	if dst.Kind() == reflect.Pointer {
		dstTyp = dst.Elem()
	}
	f.step, err = s.StepInit(ctx, dstTyp, src.Elem())
	return f, err
}

func (f PointerStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	target := dst.Elem()

	if src.IsNil() {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	var allocTyp reflect.Type
	if target.Kind() == reflect.Interface {
		allocTyp = src.Type().Elem()
	} else {
		allocTyp = target.Type().Elem()
	}

	newPtr, cached := s.New(src, allocTyp)

	target.Set(newPtr)
	if cached {
		return nil
	}

	return f.step.Copy(ctx, s, newPtr, src.Elem())
}

type InterfaceStep struct{}

func (f InterfaceStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	for src.Kind() == reflect.Interface {
		src = src.Elem()
		if !src.IsValid() {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
	}

	step, err := s.StepInit(ctx, dst.Type(), src.Type())
	if err != nil {
		return errors.Wrap(err, "InterfaceStep.Copy")
	}

	err = step.Copy(ctx, s, dst, src)
	if err != nil {
		err = errors.Wrap(err, "InterfaceStep.Copy")
	}

	return err
}

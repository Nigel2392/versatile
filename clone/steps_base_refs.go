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
	if step, ok := s.Cache().Step(dst, src); ok {
		return step, nil
	}

	var dstElem reflect.Type
	switch dst.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
		dstElem = dst.Elem()
	}

	f.step, err = s.StepInit(ctx, dstElem, src.Elem())
	if err != nil {
		return f, errors.Wrap(err, "PointerStep.Init")
	}

	s.Cache().AddStepType(dst, src, f)
	return f, nil
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

	newPtr, cached := s.New(src, target, allocTyp)

	target.Set(newPtr)
	if cached {
		return nil
	}

	return s.StepCopy(ctx, f.step, newPtr, src.Elem())
}

type InterfaceStep struct {
	step Step
}

func (f InterfaceStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (_ Step, err error) {

	cDst := dst
	cSrc := src

	if dst.Kind() == reflect.Interface {
		cDst = src
	}

	if src.Kind() == reflect.Interface {
		cSrc = cDst
	}

	if cDst.Kind() == reflect.Interface && cSrc.Kind() == reflect.Interface {
		return f, nil
	}

	if step, ok := s.Cache().Step(cDst, cSrc); ok {
		return step, nil
	}

	f.step, err = s.StepInit(ctx, cDst, cSrc)
	if err != nil && !errors.Is(err, ErrNoSteps) {
		return f, errors.Wrap(err, "InterfaceStep.Init")
	}

	s.Cache().AddStepType(cDst, cSrc, f)
	return f, nil
}

func (f InterfaceStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) (err error) {
	for src.Kind() == reflect.Interface {
		src = src.Elem()
		if !src.IsValid() {
			dst.Elem().Set(reflect.Zero(dst.Elem().Type()))
			return nil
		}
	}

	if f.step == nil {
		f.step, err = s.StepInit(ctx, src.Type(), src.Type())
		if err != nil {
			return errors.Wrap(err, "InterfaceStep.Copy")
		}
	}

	newVal := reflect.New(src.Type())
	err = s.StepCopy(ctx, f.step, newVal, src)
	if err != nil {
		return errors.Wrap(err, "InterfaceStep.Copy")
	}

	dst.Elem().Set(newVal.Elem())
	return nil
}

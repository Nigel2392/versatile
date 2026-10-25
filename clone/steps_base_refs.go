package clone

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Nigel2392/errors"
	"github.com/Nigel2392/versatile/internal/develop"
)

type PointerStep struct {
	Step Step
}

func (f PointerStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (_ Step, err error) {
	if step, ok := s.Cache().Step(dst, src); ok {
		return step, nil
	}

	var dstElem = dst
	switch dst.Kind() {
	case reflect.Pointer:
		dstElem = dst.Elem()
	}

	f.Step, err = s.StepInit(ctx, dstElem, src.Elem())
	if err != nil {
		return f, errors.Wrap(err, "PointerStep.Init")
	}

	s.Cache().AddStepType(dst, src, f)
	return f, nil
}

func NewDstForSrc(s *State, dst, src reflect.Value) (newPtr reflect.Value, zero, cached bool) {
	for src.Kind() == reflect.Interface {
		src = src.Elem()
		if !src.IsValid() {
			dst.Set(reflect.Zero(dst.Type()))
			return dst, true, false
		}
	}

	if src.Kind() == reflect.Pointer && src.IsNil() {
		if develop.DEV {
			develop.Logf("src is nil, setting dst to zero of %s", src.Type())
		}

		dst.Set(reflect.Zero(src.Type()))
		return dst, true, false
	}

	var allocTyp reflect.Type
	switch dst.Kind() {
	case reflect.Pointer:
		allocTyp = dst.Type().Elem()

	case reflect.Interface:
		if src.Kind() == reflect.Pointer {
			allocTyp = src.Type().Elem()
		} else {
			allocTyp = src.Type()
		}

	default:
		allocTyp = dst.Type()
	}

	if develop.DEV {
		defer func() {
			var newPtrStr any
			if newPtr.Kind() == reflect.Invalid {
				newPtrStr = newPtr.Kind()
			} else {
				newPtrStr = newPtr.Type()
			}

			if r := recover(); r != nil {
				panic(fmt.Sprintf(
					"NewDstValue(\n\tdst: %s\n\tsrc: %s\n\tallocTyp: %s\n\tnewPtrTyp: %s\n)",
					dst.Type(), src.Type(), allocTyp, newPtrStr,
				))
			} else {
				develop.Logf(
					"NewDstValue(\n\tdst: %s\n\tsrc: %s\n\tallocTyp: %s\n\tnewPtrTyp: %s\n)",
					dst.Type(), src.Type(), allocTyp, newPtrStr,
				)
			}
		}()
	}

	newPtr, cached = s.New(src, dst, allocTyp)

	dst.Set(newPtr)

	return newPtr, zero, cached
}

func (f PointerStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	target := dst.Elem()
	newPtr, zero, cached := NewDstForSrc(s, target, src)
	if cached || zero {
		return nil
	}

	return s.StepCopy(ctx, f.Step, newPtr, src.Elem(), ReferenceCaller{Ref: newPtr})
}

type InterfaceStep struct {
	Step Step
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

	f.Step, err = s.StepInit(ctx, cDst, cSrc)
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

	if f.Step == nil {
		f.Step, err = s.StepInit(ctx, src.Type(), src.Type())
		if err != nil {
			return errors.Wrap(err, "InterfaceStep.Copy")
		}
	}

	newVal := reflect.New(src.Type())
	err = s.StepCopy(ctx, f.Step, newVal, src, ReferenceCaller{newVal})
	if err != nil {
		return errors.Wrap(err, "InterfaceStep.Copy")
	}

	dst.Elem().Set(newVal.Elem())
	return nil
}

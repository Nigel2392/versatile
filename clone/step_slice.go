package clone

import (
	"context"
	"reflect"

	"github.com/Nigel2392/errors"
)

type SliceStep struct {
	step Step
}

func (f SliceStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {
	if step, ok := s.Cache().Step(dst, src); ok {
		return step, nil
	}

	dstElem := dst
	if dstElem.Kind() == reflect.Slice || dstElem.Kind() == reflect.Array {
		dstElem = dstElem.Elem()
	}

	f.step, err = s.StepInit(ctx, dstElem, src.Elem())
	if err != nil {
		err = errors.Wrap(err, "SliceStep.Init")
	}

	s.Cache().AddStepType(dst, src, f)
	return f, err
}

func (f SliceStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	srcLen := src.Len()
	if src.Kind() == reflect.Slice && src.IsNil() {
		dst.Elem().Set(reflect.Zero(dst.Type().Elem()))
		return nil
	}

	newSlice, cached := s.MakeSlice(src, dst, dst.Type().Elem(), srcLen)

	if cached {
		dst.Elem().Set(newSlice)
		return nil
	}

	for i := range srcLen {
		if err := s.StepCopy(ctx, f.step, newSlice.Index(i).Addr(), src.Index(i)); err != nil {
			return errors.Wrap(err, "SliceStep.Copy")
		}
	}

	dst.Elem().Set(newSlice)

	return nil
}

type ToArrayStep struct {
	SliceStep
}

func (f ToArrayStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {
	step, err = f.SliceStep.Init(ctx, s, dst, src)
	if err != nil {
		err = errors.Wrap(err, "ArrayStep.Init")
	}
	f.SliceStep = step.(SliceStep)
	return f, err
}

func (f ToArrayStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	dst = dst.Elem() // deref ptr to dst

	srcLen := src.Len()
	dstLen := dst.Len()
	if dstLen < srcLen {
		srcLen = dstLen
	}

	var i int
	for i = 0; i < srcLen; i++ {
		if err := s.StepCopy(ctx, f.step, dst.Index(i).Addr(), src.Index(i)); err != nil {
			return errors.Wrap(err, "ArrayStep.Copy")
		}
	}

	for ; i < dstLen; i++ {
		item := dst.Index(i)
		item.Set(reflect.Zero(item.Type()))
	}

	return nil
}

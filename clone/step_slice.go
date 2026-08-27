package clone

import (
	"context"
	"reflect"
)

type SliceStep struct {
	step Step
}

func (f SliceStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {
	var ok bool
	f.step, ok = s.Step(dst.Elem(), src.Elem())
	if !ok {
		return nil, ErrNoSteps.Wrapf("no steps found for %s", src)
	}
	f.step, err = initStep(ctx, s, f.step, dst, src)
	return f, err
}

func (f SliceStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	srcLen := src.Len()
	if src.Kind() == reflect.Slice && src.IsNil() {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	newSlice, cached := s.MakeSlice(src, dst.Type().Elem(), srcLen)
	if cached {
		return nil
	}

	for i := range srcLen {
		if err := f.step.Copy(ctx, s, newSlice.Index(i).Addr(), src.Index(i)); err != nil {
			return err
		}
	}

	dst.Elem().Set(newSlice)
	return nil
}

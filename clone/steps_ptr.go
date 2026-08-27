package clone

import (
	"context"
	"reflect"
)

type PointerStep struct {
	step Step
}

func (f PointerStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (_ Step, err error) {
	f.step, err = s.StepInit(ctx, dst, src.Elem())
	return f, err
}

func (f PointerStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	if dst.IsNil() {
		dst.Set(reflect.New(dst.Type().Elem()))
	}
	return f.step.Copy(ctx, s, dst, src)
}

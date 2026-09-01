package clone

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Nigel2392/errors"
)

type SliceStep struct {
	step Step
}

func (f SliceStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {
	f.step, err = s.StepInit(ctx, dst.Elem(), src.Elem())
	if err != nil {
		err = errors.Wrap(err, "SliceStep.Init")
	}
	return f, err
}

func (f SliceStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	srcLen := src.Len()
	if src.Kind() == reflect.Slice && src.IsNil() {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	newSlice, cached := s.MakeSlice(src, dst.Type().Elem(), srcLen)

	dst.Elem().Set(newSlice)
	fmt.Println(dst.Type(), dst.Elem().Type(), src.Type(), newSlice.Type(), srcLen)
	if cached {
		return nil
	}

	fmt.Println(newSlice.Interface(), src.Interface(), cached, srcLen)
	for i := range srcLen {
		fmt.Println(newSlice.Index(i), src.Index(i))
		if err := f.step.Copy(ctx, s, newSlice.Index(i).Addr(), src.Index(i)); err != nil {
			return errors.Wrap(err, "SliceStep.Copy")
		}
	}
	fmt.Println(newSlice.Interface(), src.Interface(), cached, srcLen)

	return nil
}

type ArrayStep struct {
	SliceStep
}

func (f ArrayStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {
	step, err = f.SliceStep.Init(ctx, s, dst, src)
	if err != nil {
		err = errors.Wrap(err, "ArrayStep.Init")
	}
	f.SliceStep = step.(SliceStep)
	return f, err
}

func (f ArrayStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	srcLen := src.Len()
	dstLen := dst.Len()
	if dstLen < srcLen {
		srcLen = dstLen
	}

	var i int
	for i = 0; i < srcLen; i++ {
		if err := f.step.Copy(ctx, s, dst.Index(i), src.Index(i)); err != nil {
			return errors.Wrap(err, "ArrayStep.Copy")
		}
	}

	for ; i < dstLen; i++ {
		item := dst.Index(i)
		item.Set(reflect.Zero(item.Type()))
	}

	return nil
}

package clone

import (
	"context"
	"reflect"
	"unsafe"

	"github.com/Nigel2392/versatile/bitcheck"
)

type (
	CloneFlag = bitcheck.Flag
	oldPtr    = uintptr
	newPtr    = unsafe.Pointer
)

type State struct {
	Flags CloneFlag

	pointers map[oldPtr]newPtr
	reg      *stepRegistry
}

func (s *State) New(oldPtr reflect.Value, newTyp reflect.Type) (newOrCached reflect.Value, wasCached bool) {
	oldP := oldPtr.Pointer()
	if v, ok := s.pointers[oldP]; ok {
		return reflect.NewAt(newTyp, v), true
	}

	n := reflect.New(newTyp)
	s.pointers[oldP] = n.UnsafePointer()
	return n, false
}

func (s *State) MakeSlice(oldPtr reflect.Value, newTyp reflect.Type, l int) (newOrCached reflect.Value, wasCached bool) {
	oldP := oldPtr.Pointer()
	if v, ok := s.pointers[oldP]; ok {
		return reflect.SliceAt(newTyp, v, l), true
	}

	n := reflect.MakeSlice(newTyp, l, l)
	s.pointers[oldP] = n.UnsafePointer()
	return n, false
}

func (s *State) MakeMap(oldPtr reflect.Value, newTyp reflect.Type, _len int) (newOrCached reflect.Value, wasCached bool) {
	oldP := oldPtr.Pointer()
	if v, ok := s.pointers[oldP]; ok {
		return reflect.NewAt(newTyp, v).Elem(), true
	}

	n := reflect.MakeMapWithSize(newTyp, _len)
	s.pointers[oldP] = n.UnsafePointer()
	return n, false
}

func (s *State) Step(dstIfSrcElseSrc reflect.Type, src ...reflect.Type) (Step, bool) {
	step, ok := s.__get_step(dstIfSrcElseSrc, src)
	return stepForState(s, step), ok
}

func (s *State) __get_step(dstIfSrcElseSrc reflect.Type, src []reflect.Type) (Step, bool) {
	if s.reg != nil {
		step, ok := s.reg.step(dstIfSrcElseSrc, src)
		if ok {
			return step, true
		}
	}

	return stepReg.step(dstIfSrcElseSrc, src)
}

func (s *State) StepInit(ctx context.Context, dst, src reflect.Type) (step Step, err error) {
	step, ok := s.Step(dst, src)
	if !ok {
		return nil, ErrNoSteps
	}

	return initStep(ctx, s, step, dst, src)
}

func stepForState(s *State, st Step) Step {
	if st == nil {
		return nil
	}
	if i, ok := st.(InitStep); ok {
		return istateStep{s: s, w: i}
	}
	return stateStep[Step]{s: s, w: st}
}

type stateStep[T Step] struct {
	s *State
	w T
}

func (t stateStep[T]) Unwrap() Step {
	return t.w
}

func (t stateStep[T]) Copy(ctx context.Context, s *State, dst, src reflect.Value) (err error) {
	if dst.Kind() != reflect.Pointer && !dst.CanSet() {
		if dst.CanInterface() {
			dstf := dst.Interface()
			return ErrInvalid.Wrapf("dst %T(%v) is not settable, cannot copy src %v", dstf, dstf, src.Interface())
		}
		if !dst.IsValid() {
			return ErrInvalid.Wrapf("dst is invalid, cannot copy src %v", src.Interface())
		}
	}

	return t.w.Copy(ctx, s, dst, src)
}

type istateStep struct {
	stateStep[InitStep]
}

func (t istateStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {
	step, err = t.w.Init(ctx, s, dst, src)
	if step != nil {
		step = stateStep[Step]{s: s, w: step}
	}
	return step, err
}

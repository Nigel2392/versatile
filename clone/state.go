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

const (
	CF_INVALID CloneFlag = iota
	CF_NOWRAP  CloneFlag = 1 << iota
	CF_NOVALIDATE
	CF_KEEP_POINTERS
	CF_NO_CONVS
)

type State struct {
	Flags CloneFlag

	wrapStep func(*State, Step) Step
	pointers map[oldPtr]newPtr
	cache    ResettableRegistry
}

func Flag(flag CloneFlag) func(s *State) {
	return func(s *State) {
		s.Flags = flag
	}
}

func WrapStep(fn func(*State, Step) Step) func(s *State) {
	return func(s *State) {
		s.wrapStep = fn
	}
}

type stateContextKey struct{}

func SharedStateContext(ctx context.Context, opts ...func(s *State)) context.Context {
	var (
		state *State
		ok    bool
	)

	if state, ok = StateFromContext(ctx); !ok {
		state = new(State{
			pointers: make(map[oldPtr]newPtr),
			cache:    &cacheRegistry{steps: make(map[any]Step)},
		})

		// only need to override context when state is not present already
		ctx = context.WithValue(ctx, stateContextKey{}, state)
	}

	for _, opt := range opts {
		opt(state)
	}

	return ctx
}

func StateFromContext(ctx context.Context) (*State, bool) {
	s, ok := ctx.Value(stateContextKey{}).(*State)
	return s, ok
}

type value struct {
	_   uintptr
	ptr uintptr
	_   uintptr
}

func getCacheKey(v reflect.Value) uintptr {
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.UnsafePointer:
		return v.Pointer()
	default:
		op := *(*value)(unsafe.Pointer(&v))
		return op.ptr
	}
}

func (s *State) New(keyPtr reflect.Value, dstVal reflect.Value, newTyp reflect.Type) (newOrCached reflect.Value, ok bool) {
	op := getCacheKey(keyPtr)
	if v, ok := s.pointers[op]; ok {
		return reflect.NewAt(newTyp, v), true
	}

	// safely reuse the existing pointer
	if dstVal.Kind() == reflect.Pointer && !dstVal.IsNil() && dstVal.Type().Elem() == newTyp {
		s.pointers[op] = dstVal.UnsafePointer()
		return dstVal, false
	}

	// fallback to allocation
	n := reflect.New(newTyp)
	s.pointers[op] = n.UnsafePointer()
	return n, false
}

func (s *State) MakeSlice(oldPtr reflect.Value, dstVal reflect.Value, newTyp reflect.Type, l int) (n reflect.Value, wasCached bool) {
	if newTyp.Kind() == reflect.Interface {
		newTyp = oldPtr.Type()
	}

	op := getCacheKey(oldPtr)
	if v, ok := s.pointers[op]; ok {
		return reflect.SliceAt(newTyp.Elem(), v, l), true
	}

	if dstVal.Kind() == reflect.Pointer && !dstVal.IsNil() && dstVal.Type().Elem() == newTyp && (newTyp.Kind() == reflect.Array || newTyp.Kind() == reflect.Slice) {
		el := dstVal.Elem()
		switch newTyp.Kind() {
		case reflect.Array:
			s.pointers[op] = dstVal.UnsafePointer()
		case reflect.Slice:
			el.Grow(l - el.Len())
			el.SetLen(l)
			el.SetCap(l)
			s.pointers[op] = el.UnsafePointer()
		}
		return el, false
	}

	if oldPtr.Kind() == reflect.Array && (newTyp.Kind() == reflect.Array) {
		n = reflect.New(oldPtr.Type())
		s.pointers[op] = n.UnsafePointer()
		n = n.Elem()
	} else {
		n = reflect.MakeSlice(newTyp, l, oldPtr.Cap())
		s.pointers[op] = n.UnsafePointer()
	}

	return n, false
}

func (s *State) MakeMap(oldPtr reflect.Value, dstVal reflect.Value, newTyp reflect.Type, _len int) (newOrCached reflect.Value, wasCached bool) {
	op := getCacheKey(oldPtr)
	if v, ok := s.pointers[op]; ok {
		return reflect.NewAt(newTyp, unsafe.Pointer(v)).Elem(), true
	}

	if dstVal.Kind() == reflect.Pointer && !dstVal.IsNil() && dstVal.Type().Elem() == newTyp {
		s.pointers[op] = dstVal.UnsafePointer()
		return dstVal, false
	}

	n := reflect.MakeMapWithSize(newTyp, _len)
	s.pointers[op] = n.UnsafePointer()
	return n, false
}

func (s *State) Cache() ResettableRegistry {
	if s.cache == nil {
		return NopRegistry
	}
	return s.cache
}

func (s *State) Step(dstIfSrcElseSrc reflect.Type, src reflect.Type) (Step, bool) {
	step, ok := s.__get_step(dstIfSrcElseSrc, src)

	if !ok || step == nil || s.Flags.Is(CF_NOWRAP) || s.wrapStep == nil {
		return step, ok
	}

	return s.wrapStep(s, step), true
}

func (s *State) __get_step(dstIfSrcElseSrc reflect.Type, src reflect.Type) (Step, bool) {
	if s.cache != nil {
		step, ok := s.cache.Step(dstIfSrcElseSrc, src)
		if ok {
			return step, true
		}
	}

	return stepReg.Step(dstIfSrcElseSrc, src)
}

func (s *State) StepInit(ctx context.Context, dst, src reflect.Type) (step Step, err error) {
	step, ok := s.Step(dst, src)
	if !ok {
		return nil, ErrNoSteps.Wrapf("No steps found for %v and %v", dst, src)
	}

	return initStep(ctx, s, step, dst, src)
}

func (s *State) StepCopy(ctx context.Context, step Step, dst, src reflect.Value) error {

	if !bitcheck.Is(s.Flags, CF_NOVALIDATE) && dst.Kind() != reflect.Pointer && !dst.CanSet() {
		if dst.CanInterface() {
			dstf := dst.Interface()
			return ErrInvalid.Wrapf("dst %T(%v) is not settable, cannot copy src %v", dstf, dstf, src.Interface())
		}
		if !dst.IsValid() {
			return ErrInvalid.Wrapf("dst is invalid, cannot copy src %v", src.Interface())
		}
	}

	return step.Copy(ctx, s, dst, src)
}

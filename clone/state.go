package clone

import (
	"context"
	"reflect"
	"unsafe"

	"github.com/Nigel2392/versatile/bitcheck"
)

type (
	CloneFlag = bitcheck.Flag

	// type hints for the shared references object cache
	oldPtr = uintptr
	newPtr = unsafe.Pointer
)

const (
	CF_INVALID       CloneFlag = iota
	CF_NOWRAP        CloneFlag = 1 << iota // don't allow wrapping, even if a wrap function is provided
	CF_NOVALIDATE                          // don't validate if dst is actually settable
	CF_KEEP_POINTERS                       // don't reset the pointer cache after before copy with a shared state
	CF_ONLY_PUB_FLDS                       // only public (non exported) struct fields can be cloned.
	CF_NO_CONVS                            // don't allow conversions from, int32 to int64 (singular example)
)

type State struct {
	Flags CloneFlag

	wrapStep func(*State, Step) Step
	pointers map[oldPtr]newPtr
	cache    ResettableRegistry
}

// Allow for enabling and disabling certain features
//
// Setting some of these flags can have an impact on performance and/or
// functionality of the [Clone] function.
func Flag(flag CloneFlag, reset ...bool) func(s *State) {
	var _reset bool
	if len(reset) > 0 {
		_reset = reset[0]
	}

	if _reset {
		return func(s *State) {
			s.Flags = flag
		}
	}

	return func(s *State) {
		s.Flags |= flag
	}
}

// Allow for wrapping of steps to perform actions pre/post init
// and pre/post copy.
func WrapStep(fn func(*State, Step) Step) func(s *State) {
	return func(s *State) {
		s.wrapStep = fn
	}
}

// Initialize a new reflect.Value of type [newTyp]
//
// [keyPtr] is used to generate a cache key to properly clone any references that share the same pointer.
//
// If dstVal is a pointer, and it's [Elem] method returns a value of type [newTyp]
// then [State.New] assumes it is safe to return (and cache) [dstVal].
func (s *State) New(keyPtr reflect.Value, dstVal reflect.Value, newTyp reflect.Type) (newOrCached reflect.Value, cached bool) {
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

// Initialize a new [reflect.Value] of kind [reflect.Array] or [reflect.Slice] with type [newTyp]
//
// [keyPtr] is used to generate a cache key to properly clone any references that share the same pointer.
//
// If dstVal is a pointer, and it's [Elem] method returns a value of type [newTyp]
// then [State.New] assumes it is safe to return (and cache) [dstVal].
//
// If dstVal can be used, the slice is grown to length [l], with it's capacity and len set to [l].
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

// Initialize a new [reflect.Value] of kind [reflect.Map] with type [newTyp]
//
// [keyPtr] is used to generate a cache key to properly clone any references that share the same pointer.
//
// If dstVal is a pointer, and it's [Elem] method returns a value of type [newTyp]
// then [State.New] assumes it is safe to return (and cache) [dstVal].
func (s *State) MakeMap(oldPtr reflect.Value, dstVal reflect.Value, newTyp reflect.Type, _len int) (newOrCached reflect.Value, wasCached bool) {
	op := getCacheKey(oldPtr)
	if v, ok := s.pointers[op]; ok {
		return newMapFromPtr(newTyp, unsafe.Pointer(v)), true
	}

	if dstVal.Kind() == reflect.Pointer && !dstVal.IsNil() && !dstVal.Elem().IsNil() && dstVal.Type().Elem() == newTyp {
		el := dstVal.Elem()
		s.pointers[op] = el.UnsafePointer()
		return el, false
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
	if !bitcheck.Is(s.Flags, CF_NOVALIDATE) && !IsAllowedType(src) {
		return nil, ErrInvalid.Wrapf("%s is specified as non-clonable", src)
	}

	step, ok := s.Step(dst, src)
	if !ok {
		return nil, ErrNoSteps.Wrapf("No steps found for %v and %v", dst, src)
	}

	return initStep(ctx, s, step, dst, src)
}

func (s *State) StepCopy(ctx context.Context, step Step, dst, src reflect.Value) error {
	if !bitcheck.Is(s.Flags, CF_NOVALIDATE) {

		if dst.Kind() != reflect.Pointer && !dst.CanSet() {
			if dst.CanInterface() {
				dstf := dst.Interface()
				return ErrInvalid.Wrapf("dst %T(%v) is not settable, cannot copy src %v", dstf, dstf, src.Interface())
			}
			if !dst.IsValid() {
				return ErrInvalid.Wrapf("dst is invalid, cannot copy src %v", src.Interface())
			}
		}

		if !IsAllowedValue(src) {
			return ErrInvalid.Wrapf(
				"%s is specified as non-clonable", src.Type(),
			)
		}
	}

	return step.Copy(ctx, s, dst, src)
}

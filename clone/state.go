package clone

import (
	"context"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/Nigel2392/errors"
	"github.com/Nigel2392/versatile"
	"github.com/Nigel2392/versatile/bitcheck"
	"github.com/Nigel2392/versatile/internal/develop"
	"github.com/Nigel2392/versatile/registry"
)

type (
	CloneFlag = bitcheck.Flag

	// type hints for the shared references object cache
	oldPtr = uintptr
	newPtr = unsafe.Pointer
)

type Caller interface {
	Value() reflect.Value
}

type SliceCaller struct {
	Val   reflect.Value
	Index int
}

type ReferenceCaller struct {
	Ref reflect.Value
}

type TopLevelCaller struct {
	Val reflect.Value
}

type StructCaller struct {
	Val   reflect.Value
	Field []int
}

type MapCaller struct {
	Val reflect.Value
	Key reflect.Value
}

func (t SliceCaller) Value() reflect.Value {
	return t.Val
}

func (t ReferenceCaller) Value() reflect.Value {
	return t.Ref
}

func (t TopLevelCaller) Value() reflect.Value {
	return t.Val
}

func (t StructCaller) Value() reflect.Value {
	return t.Val
}

func (t MapCaller) Value() reflect.Value {
	return t.Val
}

const (
	CF_INVALID       CloneFlag = iota
	CF_NOWRAP        CloneFlag = 1 << iota // don't allow wrapping, even if a wrap function is provided
	CF_NOVALIDATE                          // don't validate if dst is actually settable
	CF_KEEP_POINTERS                       // don't reset the pointer cache after before copy with a shared state
	CF_NO_CACHE                            // do not use an internal state cache
	CF_ONLY_PUB_FLDS                       // only public (non exported) struct fields can be cloned.
	CF_NO_CONVS                            // don't allow conversions from, int32 to int64 (singular example)
	CF_COPY_METHODS                        // allow for methods on src or dst to be used to copy src to dst
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
	if develop.DEV {
		defer func() {
			if r := recover(); r != nil {
				panic(fmt.Sprintf("New(%s, %s, %s)", keyPtr.Type(), dstVal.Type(), newTyp))
			}
		}()
	}

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
	if develop.DEV {
		defer func() {
			if r := recover(); r != nil {
				panic(fmt.Sprintf("MakeSlice(%s, %s, %s, %d)", oldPtr.Type(), dstVal.Type(), newTyp, l))
			}
		}()
	}

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
	if develop.DEV {
		defer func() {
			if r := recover(); r != nil {
				panic(fmt.Sprintf("MakeMap(%s, %s, %s, %d)", oldPtr.Type(), dstVal.Type(), newTyp, _len))
			}
		}()
	}

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
	if s.cache != nil && !s.Flags.Is(CF_NO_CACHE) {
		step, ok := s.cache.Step(dstIfSrcElseSrc, src)
		if ok {
			return step, true
		}
	}

	return stepReg.Step(dstIfSrcElseSrc, src)
}

func (s *State) StepInit(ctx context.Context, dst, src reflect.Type) (step Step, err error) {
	if !bitcheck.Is(s.Flags, CF_NOVALIDATE) && !OK.Type(ctx, src) {
		return nil, ErrInvalid.Wrapf("State.StepInit: %s is specified as non-clonable", src)
	}

	step, ok := s.Step(dst, src)
	if !ok {
		return nil, ErrNoSteps.Wrapf("No steps found for %v and %v", dst, src)
	}

	if i, ok := step.(InitStep); ok {
		step, err = i.Init(ctx, s, dst, src)
	}

	return step, err
}

var (
	_ctxTyp            = reflect.TypeFor[context.Context]()
	_errTyp            = reflect.TypeFor[error]()
	_callerTyp         = reflect.TypeFor[Caller]()
	_stateTyp          = reflect.TypeFor[*State]()
	_methodCCpyMethods = []string{"CCopy", "CCopyTo"}
)

func _checkCCpyMethod(src reflect.Value, m *reflect.Method) error {
	if m.Type.NumIn() == 1 {
		return versatile.ErrArgCount.Wrapf(
			"%s.%s (%s) must have at least dst as argument",
			src.Type(), m.Name, m.Type).
			// wrap
			WithCause(ErrInvalid).
			WithCause(registry.ErrNotFound)
	}

	if m.Type.NumOut() > 1 {
		return versatile.ErrArgCount.Wrapf(
			"%s.%s (%s) must return a single value of type error or nothing at all",
			src.Type(), m.Name, m.Type).
			// wrap
			WithCause(ErrInvalid).
			WithCause(registry.ErrNotFound)
	}

	if m.Type.NumOut() == 1 && m.Type.Out(1) != _errTyp {
		return versatile.ErrArgCount.Wrapf(
			"%s.%s (%s) must return a single value of type error or nothing at all",
			src.Type(), m.Name, m.Type).
			// wrap
			WithCause(ErrInvalid).
			WithCause(registry.ErrNotFound)
	}

	return nil
}

// StepCopy is to be used inside of steps to call [Step.Copy] method
//
// without calling [StepCopy] certain features and functionality are not honored
func (s *State) StepCopy[CALLER Caller](ctx context.Context, step Step, dst, src reflect.Value, caller CALLER) error {
	if err := s.validate(ctx, dst, src); err != nil {
		return err
	}

	if s.Flags.Is(CF_COPY_METHODS) {
		method, err := stepReg.MethodFor(src, _methodCCpyMethods, _checkCCpyMethod)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				// no custom copy method on src
				// proceed with default copy
				goto baseCpy
			} else {
				return err
			}
		}

		var (
			start        = 1
			dstUsed      = false
			curCallerTyp = reflect.TypeOf(caller)
			numIn        = method.Type.NumIn()
			argList      = make([]reflect.Value, numIn)
			dst          = dst.Elem()
			dstTyp       = dst.Type()
		)

		argList[0] = src

		// add context argument if requested
		if numIn > 1 {
			if argTyp := method.Type.In(1); argTyp == _ctxTyp || _ctxTyp.AssignableTo(argTyp) {
				start++
				argList[1] = reflect.ValueOf(ctx)
			}
		}

		for i := start; i < numIn; i++ {
			argTyp := method.Type.In(i)
			switch {
			case argTyp == _stateTyp || _stateTyp.AssignableTo(argTyp):
				// state argument
				argList[i] = reflect.ValueOf(s)

			case argTyp == dstTyp || dstTyp.AssignableTo(argTyp):
				// dst direct assignment
				argList[i] = dst
				dstUsed = true

			case argTyp == curCallerTyp || curCallerTyp.AssignableTo(argTyp),
				argTyp == _callerTyp || _callerTyp.AssignableTo(argTyp):
				// caller argument
				argList[i] = reflect.ValueOf(caller)

			case argTyp.Kind() == reflect.Interface && curCallerTyp.Implements(argTyp):
				// convert the caller argument to expected iface
				argList[i] = reflect.ValueOf(caller).Convert(argTyp)

			case argTyp.Kind() == reflect.Interface && dstTyp.Implements(argTyp),
				dstTyp.ConvertibleTo(argTyp):
				// convert the dst argument to expected type
				argList[i] = dst.Convert(argTyp)
				dstUsed = true

			default:
				panic(fmt.Sprintf(
					"invalid copy method signature, unknown arg type: %s for function %s with dst %s",
					argTyp, method.Type, dst.Type(),
				))
			}
		}

		if !dstUsed {
			panic(fmt.Sprintf(
				"invalid copy method signature, dst is required for copy method %s.%s",
				src.Type(), method.Func,
			))
		}

		out := method.Func.Call(argList)
		if len(out) == 0 {
			return nil
		}

		if rErr := out[1]; rErr.IsValid() && !rErr.IsNil() {
			return rErr.Interface().(error)
		}

		return nil
	}

baseCpy:
	return s.copy(ctx, step, dst, src, caller)
}

// [State.Copy] is to be used inside of the [Copy] function stack
func (s *State) Copy[DST, SRC any, CALLER Caller](ctx context.Context, dst DST, src SRC, caller CALLER) (err error) {
	var (
		rvDst = versatile.ReflectValue(dst)
		rvSrc = versatile.ReflectValue(src)
	)

	if rvDst.Kind() != reflect.Pointer || rvDst.IsNil() {
		return ErrNotPointer.Wrapf("%s is not a pointer or is nil, cannot copy src %s to dst", rvDst.Type(), rvSrc.Type())
	}

	if rvSrc.Kind() == reflect.Invalid {
		rvDst.Elem().Set(reflect.Zero(rvDst.Elem().Type()))
		return nil
	}

	return s._Copy(ctx, rvDst, rvSrc, caller)
}

// [State.Copy] is to be used inside of the [Copy] function stack
func (s *State) _Copy[CALLER Caller](ctx context.Context, rvDst, rvSrc reflect.Value, caller CALLER) (err error) {

	var (
		step   Step
		srcTyp = rvSrc.Type()
		dstTyp = rvDst.Type()
	)

	// if dst is pointer to interface, see if step registered
	// for said interface
	if dstTyp.Kind() == reflect.Pointer {

		// handle copy(****int, ***int)
		var dptrs int
		for dstTyp.Kind() == reflect.Pointer {
			dstTyp = dstTyp.Elem()
			dptrs++
		}

		var sptrs int
		var sdstTyp = srcTyp
		for sdstTyp.Kind() == reflect.Pointer {
			sdstTyp = sdstTyp.Elem()
			sptrs++
		}

		// src may be passed as ******T, but should always end up as *T
		if sptrs > 1 {
			for sptrs > 1 {
				rvSrc = rvSrc.Elem()
				sptrs--
			}

			srcTyp = rvSrc.Type()
		}

		// -1 because dst should always be ptr for src
		dptrs = (dptrs - 1) - sptrs

		var d = rvDst
		for i := range dptrs {
			if rvDst.IsNil() && i > 0 {
				d.Set(reflect.Zero(rvDst.Type()))
			}
			d = rvDst
			rvDst = rvDst.Elem()
		}

		dstTyp = rvDst.Type().Elem()
	}

	// retrieve copy steps
	step, err = s.StepInit(ctx, dstTyp, srcTyp)
	if err != nil {
		return errors.Wrapf(
			err, "%s => %s",
			srcTyp, dstTyp,
		)
	}

	// copy to dst
	return s.StepCopy(ctx, step, rvDst, rvSrc, caller)
}

func (s *State) validate(ctx context.Context, dst, src reflect.Value) error {
	if bitcheck.Is(s.Flags, CF_NOVALIDATE) {
		return nil
	}

	if dst.Kind() != reflect.Pointer && !dst.CanSet() {
		if dst.CanInterface() {
			dstf := dst.Interface()
			return ErrInvalid.Wrapf("dst %T(%v) is not settable, cannot copy src %v", dstf, dstf, src.Interface())
		}
		if !dst.IsValid() {
			return ErrInvalid.Wrapf("dst is invalid, cannot copy src %v", src.Interface())
		}
	}

	if !OK.Value(ctx, src) {
		return ErrInvalid.Wrapf(
			"%s is specified as non-clonable", src.Type(),
		)
	}

	return nil
}

func (s *State) copy[CALLER Caller](ctx context.Context, step Step, dst, src reflect.Value, caller CALLER) (err error) {
	switch c := step.(type) {
	case CallerStep[CALLER]:
		// wont allocate on heap by default
		err = c.CopyWithCaller(ctx, s, dst, src, caller)
	case CallerStep[Caller]:
		// will allocate on heap (interface boxing)
		err = c.CopyWithCaller(ctx, s, dst, src, caller)
	default:
		err = step.Copy(ctx, s, dst, src)
	}

	if err != nil {
		return err
	}

	return nil

}

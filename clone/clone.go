package clone

import (
	"context"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/Nigel2392/errors"
	"github.com/Nigel2392/versatile"
	"github.com/Nigel2392/versatile/bitcheck"
	"github.com/Nigel2392/versatile/internal/danger"
)

var (
	_ Step     = (*BaseStep)(nil)
	_ Step     = (*UUIDStep)(nil)
	_ InitStep = (*PointerStep)(nil)
	_ InitStep = (*InterfaceStep)(nil)
	_ InitStep = (*StructStep)(nil)
	_ InitStep = (*MapStep)(nil)
	_ InitStep = (*StructToMapStep)(nil)
	// _ InitStep = (*MapToStructStep)(nil)
	_ InitStep = (*SliceStep)(nil)
	_ InitStep = (*ToArrayStep)(nil)
)

const STRUCT_TAG = "clone"

type Step interface {
	Copy(ctx context.Context, s *State, dst, src reflect.Value) error
}

type CallerStep[CALLER Caller] interface {
	CopyWithCaller(ctx context.Context, s *State, dst, src reflect.Value, caller CALLER) error
}

type InitStep interface {
	Step
	Init(ctx context.Context, s *State, dst, src reflect.Type) (Step, error)
}

// Copy value src into pointer dst
//
// Options can be provided to change the state and behaviour
func CopyT[TYP any, PTR any](ctx context.Context, dst *PTR, src TYP, opts ...func(*State)) (err error) {
	return Copy(ctx, dst, src, opts...)
}

// Clone value src
//
// Options can be provided to change the state and behaviour
func Clone[T any](ctx context.Context, src any, opts ...func(*State)) (val T, err error) {
	var (
		rvSrc = versatile.ReflectValue(src)
		rvDst = reflect.New(rvSrc.Type())
	)

	if rvSrc.Kind() == reflect.Invalid {
		return val, ErrInvalid.Wrap("src is invalid")
	}

	err = rcopy(ctx, rvDst, rvSrc, opts)

	return rvDst.Elem().Interface().(T), nil
}

// Copy value src into pointer dst
//
// Options can be provided to change the state and behaviour
func Copy(ctx context.Context, dst any, src any, opts ...func(*State)) (err error) {
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

	return rcopy(ctx, rvDst, rvSrc, opts)
}

func rcopy(ctx context.Context, dst reflect.Value, src reflect.Value, opts []func(*State)) (err error) {

	var state *State
	var _state, ok = StateFromContext(ctx)
	if !ok {
		_state = new(State{
			pointers: make(map[oldPtr]newPtr),
			cache:    &cacheRegistry{steps: make(map[any]Step)},
		})
	} else {
		s := *_state
		_state = &s
	}

	// prevents extra alloc (prevent state escapes to heap)
	state = (*State)(danger.Noescape(unsafe.Pointer(_state)))

	// apply customisations to state
	for _, opt := range opts {
		opt(state)
	}

	if !bitcheck.Is(state.Flags, CF_KEEP_POINTERS) {
		clear(state.pointers)
	}

	var (
		step   Step
		srcTyp = src.Type()
		dstTyp = dst.Type()
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

		// -1 because dst should always be ptr for src
		dptrs = (dptrs - 1) - sptrs

		var d = dst
		for i := range dptrs {
			if dst.IsNil() && i > 0 {
				d.Set(reflect.Zero(dst.Type()))
			}
			d = dst
			dst = dst.Elem()
		}

		dstTyp = dst.Type().Elem()
	}

	// retrieve copy steps
	step, err = state.StepInit(ctx, dstTyp, srcTyp)
	if err != nil {
		return errors.Wrapf(
			err, "%s => %s",
			srcTyp, dstTyp,
		)
	}

	// copy to dst
	err = state.StepCopy(ctx, step, dst, src, TopLevelCaller{Val: dst})

	// ensure state lives through all steps
	runtime.KeepAlive(_state)
	return err

}

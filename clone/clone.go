package clone

import (
	"context"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/Nigel2392/errors"
	"github.com/Nigel2392/versatile/bitcheck"
)

var (
	_ Step     = (*BaseStep)(nil)
	_ Step     = (*UUIDStep)(nil)
	_ InitStep = (*PointerStep)(nil)
	_ InitStep = (*InterfaceStep)(nil)
	_ InitStep = (*StructStep)(nil)
	_ InitStep = (*MapStep)(nil)
	_ InitStep = (*StructToMapStep)(nil)
	_ InitStep = (*MapToStructStep)(nil)
	_ InitStep = (*SliceStep)(nil)
	_ InitStep = (*ToArrayStep)(nil)
)

const STRUCT_TAG = "clone"

type Step interface {
	Copy(ctx context.Context, s *State, dst, src reflect.Value) error
}

type InitStep interface {
	Step
	Init(ctx context.Context, s *State, dst, src reflect.Type) (Step, error)
}

func CopyT[TYP any, PTR any](ctx context.Context, dst *PTR, src TYP, opts ...func(*State)) (err error) {
	return Copy(ctx, dst, src, opts...)
}

func Copy(ctx context.Context, dst any, src any, opts ...func(*State)) (err error) {
	var (
		rvDst = reflect.ValueOf(dst)
		rvSrc = reflect.ValueOf(src)
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
	state = (*State)(noescape(unsafe.Pointer(_state)))

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
	err = step.Copy(ctx, state, dst, src)

	// ensure state lives through all steps
	runtime.KeepAlive(_state)
	return err

}

func initStep(ctx context.Context, state *State, step Step, dst, src reflect.Type) (_ Step, err error) {
	if i, ok := step.(InitStep); ok {
		step, err = i.Init(ctx, state, dst, src)
	}
	return step, err
}

var setDirectKinds = func() reflect.Kind {
	var n reflect.Kind
	n |= 1 << reflect.Bool
	n |= 1 << reflect.String
	n |= 1 << reflect.Int
	n |= 1 << reflect.Int8
	n |= 1 << reflect.Int16
	n |= 1 << reflect.Int32
	n |= 1 << reflect.Int64
	n |= 1 << reflect.Uint
	n |= 1 << reflect.Uint8
	n |= 1 << reflect.Uint16
	n |= 1 << reflect.Uint32
	n |= 1 << reflect.Uint64
	n |= 1 << reflect.Float32
	n |= 1 << reflect.Float64
	n |= 1 << reflect.Complex64
	n |= 1 << reflect.Complex128
	n |= 1 << reflect.Uintptr
	return n
}()

// isValueType reports wether [reflect.Value.Set] is deemed enough of a clone
// this reports false for reference types.
func isValueType(k reflect.Kind) bool {
	return setDirectKinds&(1<<k) > 0
}

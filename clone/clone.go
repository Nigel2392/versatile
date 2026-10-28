package clone

import (
	"context"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/Nigel2392/versatile"
	"github.com/Nigel2392/versatile/internal/danger"
	"github.com/Nigel2392/versatile/registry"
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
		rvDst = new(T)
	)

	if rvSrc.Kind() == reflect.Invalid {
		return val, ErrInvalid.Wrap("src is invalid")
	}

	_rvDst := (*T)(danger.Noescape(unsafe.Pointer(rvDst)))
	err = rcopy(ctx, reflect.ValueOf(_rvDst), rvSrc, opts)

	return *rvDst, nil
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

var _exactTypeGet = []registry.Getter[Step]{registry.GetType[Step]}

func _getterExactType(dst reflect.Type, src reflect.Type) []registry.Getter[Step] {
	return _exactTypeGet
}

func rcopy(ctx context.Context, dst reflect.Value, src reflect.Value, opts []func(*State)) (err error) {

	sCtx, ok := ctx.Value(contextKey{}).(*stateContext)
	if !ok {
		sCtx = &stateContext{
			Context: ctx,
			state: &State{
				pointers: make(map[oldPtr]newPtr),
			},
		}
	} else {
		s := *sCtx.state
		sCtx = &stateContext{
			Context: ctx,
			state:   &s,
		}
	}

	// prevents extra alloc (prevent state escapes to heap)
	state := (*State)(danger.Noescape(unsafe.Pointer(sCtx.state)))
	ctx = (*stateContext)(danger.Noescape(unsafe.Pointer(sCtx)))

	// apply customisations to state
	for _, opt := range opts {
		opt(state)
	}

	if state.cache == nil && !state.Flags.Is(CF_NO_CACHE) {
		state.cache = &cacheRegistry{cache: make(map[registry.RegKey[reflect.Type]]Step)}
	}

	// only clear when context is shared context
	if state.Flags&cf_shared_state == cf_shared_state {
		if state.Flags&CF_KEEP_POINTERS != CF_KEEP_POINTERS {
			clear(state.pointers)
		}

		// set shared to false, if clone or copy get called from
		// within the current stack we want to keep the previous
		// state pointers.
		state.Flags = state.Flags &^ cf_shared_state
	}

	// copy to dst
	err = state.RCopy(ctx, dst, src, TopLevelCaller{Val: dst})

	// ensure state lives through all steps
	runtime.KeepAlive(state)
	runtime.KeepAlive(sCtx)
	return err

}

package clone

import (
	"context"
	"reflect"
	"runtime"
	"unsafe"
)

type Step interface {
	Copy(s *State, dst, src reflect.Value) error
}

type InitStep interface {
	Step
	Init(s *State, dst, src reflect.Type) (Step, error)
}

func Copy[T any](ctx context.Context, dst *T, src T, opts ...func(*State)) (err error) {
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
	_state := new(State{
		Ctx:      ctx,
		pointers: make(map[oldPtr]newPtr),
	})

	// prevents extra alloc
	state := (*State)(noescape(unsafe.Pointer(_state)))

	// apply customisations to state
	for _, opt := range opts {
		opt(state)
	}

	var (
		step Step
		ok   bool

		srcTyp = src.Type()
		dstTyp = dst.Type()
	)

	// if dst is pointer to interface, see if step registered
	// for said interface
	if dst.Elem().Kind() == reflect.Interface {
		// retrieve copy steps
		// not registered is OK -> check registry for src type
		step, ok = state.Step(dstTyp.Elem())
		if ok {
			goto copyStep
		}
	}

	// retrieve copy steps
	step, ok = state.Step(srcTyp)
	if !ok {
		return ErrNoSteps.Wrapf("no steps found for %s", srcTyp)
	}

copyStep:
	// initialize step if needed (complex types)
	if i, ok := step.(InitStep); ok {
		step, err = i.Init(state, dstTyp, srcTyp)
		if err != nil {
			return err
		}
	}

	// copy to dst
	err = step.Copy(state, dst, src)

	// ensure state lives through all steps
	runtime.KeepAlive(_state)
	return err

}

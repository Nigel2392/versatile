package clone

import (
	"fmt"
	"reflect"

	"github.com/Nigel2392/versatile/internal/registry"
)

var (
	_checksMethodsInterface = []registry.Getter[Step]{
		registry.GetTypeDst[Step],
		registry.GetType[Step],
		registry.GetTypeSrc[Step],
		registry.GetKind[Step],
		registry.GetKindSrc[Step],
	}
	_checksPlainInterface = []registry.Getter[Step]{
		_checkIface,
		registry.GetType[Step],
		registry.GetTypeSrc[Step],
		registry.GetKind[Step],
		registry.GetKindSrc[Step],
	}
	_checksDefault = []registry.Getter[Step]{
		registry.GetType[Step],
		registry.GetTypeSrc[Step],
		registry.GetKind[Step],
		registry.GetKindSrc[Step],
	}
)

type Registry interface {
	AddStepType(dstIfSrcElseSrc reflect.Type, src ...any)
	AddStepKind(dstIfSrcElseSrc reflect.Kind, src ...any)
	Step(dst reflect.Type, src reflect.Type) (Step, bool)
}

type ResettableRegistry interface {
	Registry
	Reset() int
}

type stepRegistry struct {
	_steps *registry.TypeRegistry[Step]
}

func _stepFromReg(dst, src reflect.Type) []registry.Getter[Step] {
	switch {
	case dst.Kind() == reflect.Interface:
		if dst.NumMethod() > 0 {
			return _checksMethodsInterface
		} else {
			return _checksPlainInterface
		}
	default:
		return _checksDefault
	}
}

var stepReg = new(stepRegistry{
	_steps: registry.NewTypeRegistry(_stepFromReg),
})

func _checkIface(types map[any]Step, key registry.RegKey[reflect.Type], kindKey registry.RegKey[reflect.Kind]) (Step, bool) {
	step, ok := types[reflect.Interface]
	return step, ok
}

func AddStepType(dstIfSrcElseSrc reflect.Type, args ...any) {
	stepReg.AddStepType(dstIfSrcElseSrc, args...)
}

func AddStepKind(dstIfSrcElseSrc reflect.Kind, args ...any) {
	stepReg.AddStepKind(dstIfSrcElseSrc, args...)
}

func (r *stepRegistry) AddStepType(dst reflect.Type, src ...any) {
	switch len(src) {
	case 1:
		_src := dst
		step := src[0].(Step)

		if _src.Kind() == reflect.Interface {
			r._steps.InterfaceRegister(_src, step)
		} else {
			r._steps.Register(_src, step)
		}

	case 2:
		dstTyp := dst
		_src := src[0].(reflect.Type)
		step := src[1].(Step)

		if dstTyp.Kind() == reflect.Interface {
			r._steps.InterfaceRegister(dstTyp, step)
		} else {
			r._steps.Register(dstTyp, _src, step)
		}

	default:
		panic(fmt.Sprintf("arguments invalid, expected 1 or 2, got %d", len(src)))
	}
}

func (r *stepRegistry) AddStepKind(dst reflect.Kind, src ...any) {
	switch len(src) {
	case 1:
		_src := dst
		step := src[0].(Step)
		r._steps.RegisterKind(0, _src, step)
	case 2:
		dstKnd := dst
		_src := src[0].(reflect.Kind)
		step := src[1].(Step)
		r._steps.RegisterKind(dstKnd, _src, step)
	default:
		panic(fmt.Sprintf("arguments invalid, expected 1 or 2, got %d", len(src)))
	}
}

func (r *stepRegistry) Step(dst reflect.Type, src reflect.Type) (step Step, ok bool) {
	if dst == nil && src == nil {
		panic("nil types provided")
	}

	return r._steps.Get(dst, src)
}

type cacheRegistry struct {
	disabled bool
	stepRegistry
}

func (r *cacheRegistry) SetDisabled(b bool) (old bool) {
	old = r.disabled
	r.disabled = b
	return old
}

func (r cacheRegistry) Reset() int {
	return r._steps.Clear()
}

func (r *cacheRegistry) AddStepType(dstIfSrcElseSrc reflect.Type, src ...any) {
	if r.disabled {
		return
	}
	r.stepRegistry.AddStepType(dstIfSrcElseSrc, src...)
}

func (r *cacheRegistry) AddStepKind(dstIfSrcElseSrc reflect.Kind, src ...any) {
	if r.disabled {
		return
	}
	r.stepRegistry.AddStepKind(dstIfSrcElseSrc, src...)
}

func (r *cacheRegistry) Step(dst reflect.Type, src reflect.Type) (Step, bool) {
	if r.disabled {
		return nil, false
	}

	return r.stepRegistry.Step(dst, src)
}

var NopRegistry ResettableRegistry = nopRegistry{}

type nopRegistry struct{}

func (r nopRegistry) AddStepType(dstIfSrcElseSrc reflect.Type, src ...any)      {}
func (r nopRegistry) AddStepKind(dstIfSrcElseSrc reflect.Kind, src ...any)      {}
func (r nopRegistry) Step(dst reflect.Type, src reflect.Type) (s Step, ok bool) { return }
func (r nopRegistry) Reset() (i int)                                            { return }

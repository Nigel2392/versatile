package clone

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/Nigel2392/versatile/registry"
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

type CacheRegistry interface {
	Add(dstTyp, srcTyp reflect.Type, step Step)
	Step(dst reflect.Type, src reflect.Type) (Step, bool)
	Reset() int
}

type stepRegistry struct {
	_steps   *registry.TypeRegistry[Step]
	_methods *registry.MethodsRegistry
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
	_steps:   registry.NewTypeRegistry(&sync.Mutex{}, registry.NONE, _stepFromReg),
	_methods: registry.NewMethods(1024),
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
		r._steps.Register(dstTyp, _src, step)

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

func (r *stepRegistry) MethodFor(val reflect.Value, names []string, valid ...func(reflect.Value, *reflect.Method) error) (reflect.Value, *reflect.Method, error) {
	return r._methods.For(val, names, valid...)
}

type cacheRegistry struct {
	disabled bool
	cache    map[registry.RegKey[reflect.Type]]Step
}

func (r *cacheRegistry) SetDisabled(b bool) (old bool) {
	old = r.disabled
	r.disabled = b
	return old
}

func (r cacheRegistry) Reset() int {
	l := len(r.cache)
	clear(r.cache)
	return l
}

func (r *cacheRegistry) Add(dstTyp, srcTyp reflect.Type, step Step) {
	if r.disabled {
		return
	}

	r.cache[registry.RegKey[reflect.Type]{Dst: dstTyp, Src: srcTyp}] = step
}

func (r *cacheRegistry) Step(dst reflect.Type, src reflect.Type) (Step, bool) {
	if r.disabled {
		return nil, false
	}

	step, ok := r.cache[registry.RegKey[reflect.Type]{Dst: dst, Src: src}]
	return step, ok
}

var NopRegistry CacheRegistry = nopRegistry{}

type nopRegistry struct{}

func (r nopRegistry) Add(reflect.Type, reflect.Type, Step)                      {}
func (r nopRegistry) Step(dst reflect.Type, src reflect.Type) (s Step, ok bool) { return }
func (r nopRegistry) Reset() (i int)                                            { return }

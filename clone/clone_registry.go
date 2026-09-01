package clone

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

func init() {
	CACHE.SetEnabled(true)
}

type stepRegistry struct {
	steps map[any]Step
}

var stepReg = new(stepRegistry{
	steps: make(map[any]Step),
})

func AddStepType(dstIfSrcElseSrc reflect.Type, args ...any) {
	stepReg.AddStepType(dstIfSrcElseSrc, args...)
}

func AddStepKind(dstIfSrcElseSrc reflect.Kind, args ...any) {
	stepReg.AddStepKind(dstIfSrcElseSrc, args...)
}

func (r *stepRegistry) AddStepType(dstIfSrcElseSrc reflect.Type, src ...any) {
	switch len(src) {
	case 1:
		step := src[0].(Step)
		r.steps[dstIfSrcElseSrc] = step
	case 2:
		dstTyp := dstIfSrcElseSrc
		dstIfSrcElseSrc, _ = src[0].(reflect.Type)
		step := src[1].(Step)
		r.steps[duo[reflect.Type]{dstTyp, dstIfSrcElseSrc}] = step
	default:
		panic(fmt.Sprintf("arguments invalid, expected 1 or 2, got %d", len(src)))
	}
}

func (r *stepRegistry) AddStepKind(dstIfSrcElseSrc reflect.Kind, src ...any) {
	switch len(src) {
	case 1:
		step := src[0].(Step)
		r.steps[dstIfSrcElseSrc] = step
	case 2:
		dstKnd := dstIfSrcElseSrc
		dstIfSrcElseSrc = src[0].(reflect.Kind)
		step := src[1].(Step)
		r.steps[duo[reflect.Kind]{dstKnd, dstIfSrcElseSrc}] = step
	default:
		panic(fmt.Sprintf("arguments invalid, expected 1 or 2, got %d", len(src)))
	}
}

func (r *stepRegistry) Step(dst reflect.Type, src reflect.Type) (Step, bool) {
	if dst == nil && src == nil {
		panic("nil types provided")
	}

	if step, ok := r.steps[duo[reflect.Type]{dst, src}]; ok {
		return step, true
	}

	if dst != nil {
		if dst.Kind() == reflect.Interface && dst.NumMethod() > 0 && (src.Kind() == reflect.Interface || src.AssignableTo(dst) || src.ConvertibleTo(dst)) {
			if step, ok := r.steps[reflect.Interface]; ok {
				return step, true
			}
		}
	}

	if src == nil {
		return nil, false
	}

	if step, ok := r.steps[src]; ok {
		return step, true
	}

	if dst != nil {
		if step, ok := r.steps[duo[reflect.Kind]{dst.Kind(), src.Kind()}]; ok {
			return step, true
		}
	}

	if step, ok := r.steps[src.Kind()]; ok {
		return step, true
	}

	return nil, false
}

type cacheRegistry struct {
	enabled atomic.Bool
	stepRegistry
}

var CACHE = cacheRegistry{
	steps: make(map[any]Step),
}

func (r *cacheRegistry) SetEnabled(b bool) (old bool) {
	return r.enabled.Swap(b)
}

func (r *cacheRegistry) AddStepType(dstIfSrcElseSrc reflect.Type, src ...any) {
	if !r.enabled.Load() {
		return
	}
	r.stepRegistry.AddStepType(dstIfSrcElseSrc, src...)
}

func (r *cacheRegistry) AddStepKind(dstIfSrcElseSrc reflect.Kind, src ...any) {
	if !r.enabled.Load() {
		return
	}
	r.stepRegistry.AddStepKind(dstIfSrcElseSrc, src...)
}

func (r *cacheRegistry) Step(dst reflect.Type, src reflect.Type) (Step, bool) {
	if !r.enabled.Load() {
		return nil, false
	}

	return r.stepRegistry.Step(dst, src)
}

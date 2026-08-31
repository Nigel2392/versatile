package clone

import (
	"fmt"
	"reflect"
)

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
		stepReg.steps[dstIfSrcElseSrc] = step
	case 2:
		dstTyp := dstIfSrcElseSrc
		dstIfSrcElseSrc = src[0].(reflect.Type)
		step := src[1].(Step)
		stepReg.steps[duo[reflect.Type]{dstTyp, dstIfSrcElseSrc}] = step
	default:
		panic(fmt.Sprintf("arguments invalid, expected 1 or 2, got %d", len(src)))
	}
}

func (r *stepRegistry) AddStepKind(dstIfSrcElseSrc reflect.Kind, src ...any) {
	switch len(src) {
	case 1:
		step := src[0].(Step)
		stepReg.steps[dstIfSrcElseSrc] = step
	case 2:
		dstKnd := dstIfSrcElseSrc
		dstIfSrcElseSrc = src[0].(reflect.Kind)
		step := src[1].(Step)
		stepReg.steps[duo[reflect.Kind]{dstKnd, dstIfSrcElseSrc}] = step
	default:
		panic(fmt.Sprintf("arguments invalid, expected 1 or 2, got %d", len(src)))
	}
}

func (r *stepRegistry) Step(dstIfSrcElseSrc reflect.Type, _src ...reflect.Type) (Step, bool) {
	return r.step(dstIfSrcElseSrc, _src)
}

func (r *stepRegistry) step(dstIfSrcElseSrc reflect.Type, _src []reflect.Type) (Step, bool) {
	if len(_src) > 1 {
		panic(fmt.Sprintf("arguments invalid, expected 1 or 0, got %d", len(_src)))
	}

	var (
		dst reflect.Type
		src = dstIfSrcElseSrc
	)

	if len(_src) == 1 {
		dst = dstIfSrcElseSrc
		src = _src[0]
	}

	if dst == nil && src == nil {
		panic("nil types provided")
	}

	if dst != nil {

		if step, ok := r.steps[duo[reflect.Type]{dst, src}]; ok {
			return step, true
		}

		if dst.Kind() == reflect.Interface {
			if step, ok := r.steps[dst]; ok {
				return step, true
			}
		}

		if step, ok := r.steps[duo[reflect.Kind]{dst.Kind(), src.Kind()}]; ok {
			return step, true
		}
	}

	if src == nil {
		return nil, false
	}

	if step, ok := r.steps[src]; ok {
		return step, true
	}

	if step, ok := r.steps[src.Kind()]; ok {
		return step, true
	}

	if src.Kind() == reflect.Pointer {
		return r.step(dst, []reflect.Type{src.Elem()})
	}

	if dst.Kind() == reflect.Pointer {
		return r.step(dst.Elem(), []reflect.Type{src})
	}

	return nil, false
}

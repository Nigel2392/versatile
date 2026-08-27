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

func AddStepType(typ reflect.Type, args ...any) {
	stepReg.AddStepType(typ, args...)
}

func AddStepKind(srcKind reflect.Kind, step Step) {
	stepReg.AddStepKind(srcKind, step)
}

func (r *stepRegistry) AddStepType(dstIfSrcElseSrc reflect.Type, args ...any) {
	switch len(args) {
	case 1:
		step := args[0].(Step)
		stepReg.steps[dstIfSrcElseSrc] = step
	case 2:
		dstTyp := dstIfSrcElseSrc
		dstIfSrcElseSrc = args[0].(reflect.Type)
		step := args[1].(Step)
		stepReg.steps[typeKey{dstTyp, dstIfSrcElseSrc}] = step
	default:
		panic(fmt.Sprintf("arguments invalid, expected 1 or 2, got %d", len(args)))
	}
}

func (r *stepRegistry) AddStepKind(srcKind reflect.Kind, step Step) {
	stepReg.steps[srcKind] = step
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

	if dst != nil && src == nil {
		src = dst
		dst = nil
	}

	if dst != nil {
		if step, ok := r.steps[typeKey{dst, src}]; ok {
			return step, true
		}
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

	return nil, false
}

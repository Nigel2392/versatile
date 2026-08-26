package clone

import "reflect"

type stepRegistry struct {
	byType map[reflect.Type]Step
	byKind map[reflect.Kind]Step
}

var stepReg = new(stepRegistry{
	byType: make(map[reflect.Type]Step),
	byKind: make(map[reflect.Kind]Step),
})

func AddStepType(typ reflect.Type, step Step) {

	if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Interface {
		typ = typ.Elem()
	}

	stepReg.byType[typ] = step
}

func AddStepKind(knd reflect.Kind, step Step) {
	stepReg.byKind[knd] = step
}

func (r *stepRegistry) getStep(src reflect.Type) (Step, bool) {
	if step, ok := r.byType[src]; ok {
		return step, true
	}

	if step, ok := r.byKind[src.Kind()]; ok {
		return step, true
	}

	return nil, false
}

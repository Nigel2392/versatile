package clone

import (
	"context"
	"reflect"
)

const TAG_NAME = "versatile"

type structFieldStep struct {
	idx  []int
	step Step
}

type StructStep struct {
	typ   reflect.Type
	steps []structFieldStep
}

func (f StructStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {
	f.typ = src
	f.steps = make([]structFieldStep, 0, src.NumField())
	for sf := range src.Fields() {
		// if !sf.IsExported() {
		// continue
		// }

		if sf.Tag != "" {
			tag, ok := sf.Tag.Lookup(TAG_NAME)
			if ok && tag == "-" {
				continue
			}
		}

		step, err = s.StepInit(ctx, sf.Type, sf.Type)
		if err != nil {
			return nil, err
		}

		f.steps = append(f.steps, structFieldStep{
			idx:  sf.Index,
			step: step,
		})
	}
	return f, nil
}

func (f StructStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	var ntEl reflect.Value
	var srcIsPtr = src.Kind() == reflect.Pointer
	var srcElem = reflect.Indirect(src)
	var nt, cached = s.New(src, srcElem.Type())
	ntEl = nt.Elem()
	if cached {
		goto setValue
	}

	for _, fld := range f.steps {

		targetFld := ntEl.FieldByIndex(fld.idx)
		srcFld := srcElem.FieldByIndex(fld.idx)

		err := fld.step.Copy(ctx, s,
			targetFld.Addr(),
			srcFld,
		)
		if err != nil {
			return err
		}
	}

setValue:
	if !srcIsPtr {
		nt = ntEl
	}

	//	dst = dst.Elem()
	//
	//	for dst.Kind() == reflect.Interface {
	//		dst = dst.Elem()
	//	}
	//
	//	dst.Set(nt)

	dst.Elem().Set(nt)
	return nil
}

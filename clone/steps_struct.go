package clone

import (
	"context"
	"reflect"
	"unsafe"

	"github.com/Nigel2392/errors"
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
	if step, ok := CACHE.Step(dst, src); ok {
		return step, nil
	}

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
			return f, errors.Wrapf(err, "StructStep.Init(%v)", sf.Index)
		}

		f.steps = append(f.steps, structFieldStep{
			idx:  sf.Index,
			step: step,
		})
	}

	CACHE.AddStepType(dst, src, f)
	return f, nil
}

func (f StructStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	if !src.CanAddr() {
		copyVal := reflect.New(src.Type()).Elem()
		copyVal.Set(src)
		src = copyVal
	}

	dstElem := dst.Elem()

	for _, fld := range f.steps {
		targetFld := dstElem.FieldByIndex(fld.idx)
		srcFld := src.FieldByIndex(fld.idx)

		addrDst := reflect.NewAt(targetFld.Type(), unsafe.Pointer(targetFld.UnsafeAddr()))
		addrSrc := reflect.NewAt(srcFld.Type(), unsafe.Pointer(srcFld.UnsafeAddr())).Elem()

		if err := fld.step.Copy(ctx, s, addrDst, addrSrc); err != nil {
			return errors.Wrapf(err, "StructStep.Copy(%v)", fld.idx)
		}
	}

	return nil
}

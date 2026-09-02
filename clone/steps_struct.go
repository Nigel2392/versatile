package clone

import (
	"context"
	"reflect"
	"unsafe"

	"github.com/Nigel2392/errors"
)

type structFieldStep struct {
	idx  []int
	step Step
}

type StructStep struct {
	steps []structFieldStep
}

func (f StructStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (Step, error) {
	if step, ok := s.Cache().Step(dst, src); ok {
		return step, nil
	}

	var dstStrct reflect.Type
	if dst.Kind() == reflect.Struct {
		dstStrct = dst
	}

	s.Cache().AddStepType(dst, src, Step(&f)) // cache reference to [StructStep]

	f.steps = make([]structFieldStep, 0, src.NumField())
	for sf := range src.Fields() {
		// if !sf.IsExported() {
		// continue
		// }

		if sf.Tag != "" {
			tag, ok := sf.Tag.Lookup(STRUCT_TAG)
			if ok && tag == "-" {
				continue
			}
		}

		if !s.Flags.Is(CF_NOVALIDATE) && !IsAllowedType(sf.Type) {
			continue
		}

		dstSfTyp := sf.Type
		if dstStrct != nil {
			dstSf := dstStrct.FieldByIndex(sf.Index)
			dstSfTyp = dstSf.Type
		}

		step, err := s.StepInit(ctx, dstSfTyp, sf.Type)
		if err != nil {
			return f, errors.Wrapf(err, "StructStep.Init(%v)", sf.Index)
		}

		f.steps = append(f.steps, structFieldStep{
			idx:  sf.Index,
			step: step,
		})
	}

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
		srcFldVal := reflect.NewAt(srcFld.Type(), unsafe.Pointer(srcFld.UnsafeAddr())).Elem()

		if err := s.StepCopy(ctx, fld.step, addrDst, srcFldVal); err != nil {
			return errors.Wrapf(err, "StructStep.Copy(%v)", fld.idx)
		}
	}

	return nil
}

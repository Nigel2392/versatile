package clone

import (
	"context"
	"reflect"
	"unsafe"
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
	var srcIsPtr = src.Kind() == reflect.Pointer
	var srcElem = reflect.Indirect(src)

	// Ensure srcElem is addressable so UnsafeAddr() doesn't panic on unaddressable sources (e.g. passed by value)
	if !srcElem.CanAddr() {
		copyVal := reflect.New(srcElem.Type()).Elem()
		copyVal.Set(srcElem)
		srcElem = copyVal
	}

	var nt = reflect.New(srcElem.Type())
	var ntEl = nt.Elem()

	for _, fld := range f.steps {
		targetFld := ntEl.FieldByIndex(fld.idx)
		srcFld := srcElem.FieldByIndex(fld.idx)

		addrDst := reflect.NewAt(targetFld.Type(), unsafe.Pointer(targetFld.UnsafeAddr()))
		addrSrc := reflect.NewAt(srcFld.Type(), unsafe.Pointer(srcFld.UnsafeAddr())).Elem()

		err := fld.step.Copy(ctx, s, addrDst, addrSrc)
		if err != nil {
			return err
		}
	}

	if !srcIsPtr {
		nt = ntEl
	}

	dst.Elem().Set(nt)
	return nil
}

package clone

import (
	"context"
	"reflect"
	"unsafe"

	"github.com/Nigel2392/tags"
)

// struct => map
type StructToMapStep struct {
	fields []structToMapField
}

func (m StructToMapStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {
	m.fields = make([]structToMapField, 0, src.NumField())
	for sf := range StructFieldsForClone(s, src) {
		structField := structToMapField{
			StructField: sf,
		}

		if sf.Tag != "" {
			tag, ok := sf.Tag.Lookup(STRUCT_TAG)
			if ok && tag == "-" {
				continue
			}

			structField.tags = tags.ParseTags(tag)
		}

		if !isValueType(sf.Type.Kind()) {
			step, err = s.StepInit(ctx, sf.Type, sf.Type)
			if err != nil {
				return m, err
			}

			structField.step = step
		}

		m.fields = append(m.fields, structField)
	}
	return m, nil
}

func (f StructToMapStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {
	var mapVal, cached = s.MakeMap(src, dst, reflect.TypeFor[map[string]any](), len(f.fields))
	if cached {
		dst.Elem().Set(mapVal)
		return nil
	}

	if !src.CanAddr() {
		copyVal := reflect.New(src.Type()).Elem()
		copyVal.Set(src)
		src = copyVal
	}

	for _, fld := range f.fields {
		key := fld.Name
		srcFld := src.FieldByIndex(fld.StructField.Index)
		srcFldVal := reflect.NewAt(srcFld.Type(), unsafe.Pointer(srcFld.UnsafeAddr())).Elem()
		val, err := fld.Copy(ctx, s, srcFldVal)

		if err != nil {
			return err
		}

		mapVal.SetMapIndex(reflect.ValueOf(key), val)
	}

	dst.Elem().Set(mapVal)

	return nil
}

type structToMapField struct {
	step Step
	tags tags.TagMap
	reflect.StructField
}

func (f structToMapField) Copy(ctx context.Context, s *State, src reflect.Value) (reflect.Value, error) {
	if f.step == nil {
		return src, nil
	}

	dst, _ := s.New(src, reflect.Value{}, f.StructField.Type)
	err := f.step.Copy(ctx, s, dst, src)
	return dst, err
}

package clone

import (
	"context"
	"reflect"

	"github.com/Nigel2392/tags"
)

type structToMapField struct {
	tags tags.TagMap
	reflect.StructField
}

// struct => map
type StructToMapStep struct {
	fields []structToMapField
}

func (m StructToMapStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {
	m.fields = make([]structToMapField, 0, src.NumField())
	for sf := range src.Fields() {
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

		m.fields = append(m.fields, structField)
	}
	return m, nil
}

func (f StructToMapStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {

	return nil
}

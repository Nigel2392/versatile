package clone

import (
	"iter"
	"reflect"
)

func StructFieldsForClone(s *State, typ reflect.Type) iter.Seq[reflect.StructField] {
	return func(yield func(reflect.StructField) bool) {
		for i := range typ.NumField() { // stditerators linter error, but why the fuck would I want more closures?
			sf := typ.Field(i)

			if s.Flags.Is(CF_ONLY_PUB_FLDS) && !sf.IsExported() {
				continue
			}

			if !yield(sf) {
				return
			}
		}
	}
}

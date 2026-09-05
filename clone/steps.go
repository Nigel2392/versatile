package clone

import (
	"reflect"
	"uuid"
)

type duo[T comparable] struct {
	dst T
	src T
}

func init() {
	AddStepKind(reflect.Int, BaseStep{})
	AddStepKind(reflect.Int8, BaseStep{})
	AddStepKind(reflect.Int16, BaseStep{})
	AddStepKind(reflect.Int32, BaseStep{})
	AddStepKind(reflect.Int64, BaseStep{})

	AddStepKind(reflect.Uint, BaseStep{})
	AddStepKind(reflect.Uint8, BaseStep{})
	AddStepKind(reflect.Uint16, BaseStep{})
	AddStepKind(reflect.Uint32, BaseStep{})
	AddStepKind(reflect.Uint64, BaseStep{})
	AddStepKind(reflect.Uintptr, BaseStep{})

	AddStepKind(reflect.Float32, BaseStep{})
	AddStepKind(reflect.Float64, BaseStep{})

	AddStepKind(reflect.Complex64, BaseStep{})
	AddStepKind(reflect.Complex128, BaseStep{})

	AddStepKind(reflect.Bool, BaseStep{})
	AddStepKind(reflect.String, BaseStep{})

	AddStepKind(reflect.Pointer, PointerStep{})
	AddStepKind(reflect.Interface, InterfaceStep{})
	AddStepKind(reflect.Struct, StructStep{})

	AddStepKind(reflect.Map, MapStep{})
	AddStepKind(reflect.Map, reflect.Map, MapStep{})
	AddStepKind(reflect.Map, reflect.Struct, StructToMapStep{})
	// AddStepKind(reflect.Struct, reflect.Map, MapToStructStep{})

	AddStepKind(reflect.Slice, SliceStep{})
	AddStepKind(reflect.Array, SliceStep{})

	AddStepKind(reflect.Slice, reflect.Slice, SliceStep{})
	AddStepKind(reflect.Slice, reflect.Array, SliceStep{})

	AddStepKind(reflect.Array, reflect.Slice, ToArrayStep{})
	AddStepKind(reflect.Array, reflect.Array, ToArrayStep{})

	AddStepType(reflect.TypeFor[uuid.UUID](), UUIDStep{})
}

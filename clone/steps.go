package clone

import (
	"reflect"
	"uuid"
)

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

	AddStepKind(reflect.Uint64, BaseStep{})
	AddStepKind(reflect.Uintptr, BaseStep{})

	AddStepKind(reflect.Pointer, PointerStep{})
	AddStepKind(reflect.Bool, BaseStep{})
	AddStepKind(reflect.String, BaseStep{})

	AddStepType(reflect.TypeFor[uuid.UUID](), UUIDStep{})
}

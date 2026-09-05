package danger

import "reflect"

var setDirectKinds = func() reflect.Kind {
	var n reflect.Kind
	n |= 1 << reflect.Bool
	n |= 1 << reflect.String
	n |= 1 << reflect.Int
	n |= 1 << reflect.Int8
	n |= 1 << reflect.Int16
	n |= 1 << reflect.Int32
	n |= 1 << reflect.Int64
	n |= 1 << reflect.Uint
	n |= 1 << reflect.Uint8
	n |= 1 << reflect.Uint16
	n |= 1 << reflect.Uint32
	n |= 1 << reflect.Uint64
	n |= 1 << reflect.Float32
	n |= 1 << reflect.Float64
	n |= 1 << reflect.Complex64
	n |= 1 << reflect.Complex128
	n |= 1 << reflect.Uintptr
	n |= 1 << reflect.Func
	return n
}()

// isValueType reports wether the given type is a value or reference type.
func IsValueType(k reflect.Kind) bool {
	return setDirectKinds&(1<<k) > 0
}

package versatile

import (
	re "reflect"

	"github.com/Nigel2392/versatile/internal/danger"
)

func ConvertToType(value re.Value, targetType re.Type) (re.Value, error) {
	var vt = value.Type()
	var assignableToParam = vt == targetType || vt.AssignableTo(targetType)
	var convertibleToParam = danger.IsSafeConversion(vt, targetType) && vt.ConvertibleTo(targetType)
	if !assignableToParam && !convertibleToParam {
		return re.Value{}, ErrTypeMismatch.Wrapf(
			"cannot convert type %s to %s",
			vt, targetType,
		)
	}

	if convertibleToParam && !assignableToParam {
		value = value.Convert(targetType)
	}

	if !value.Type().AssignableTo(targetType) {
		return re.Value{}, ErrTypeMismatch.Wrapf(
			"cannot assign type %s to %s",
			value.Type(), targetType,
		)
	}

	return value, nil
}

type IsZeroer interface {
	IsZero() bool
}

var _isZeroerType = re.TypeOf((*IsZeroer)(nil)).Elem()

func ReflectValue(value interface{}) re.Value {
	switch v := value.(type) {
	case re.Value:
		return v
	default:
		return re.ValueOf(value)
	}
}

func IsZero(value interface{}) bool {
	if value == nil {
		return true
	}

	if zeroer, ok := value.(IsZeroer); ok {
		return zeroer.IsZero()
	}

	var rv = ReflectValue(value)
	if !rv.IsValid() {
		return true
	}

	if rv.Kind() == re.Interface && rv.IsNil() {
		return true
	}

	if rv.Kind() == re.Ptr && rv.IsNil() {
		return true
	}

	// check if either the pointer to the value or the value itself implements isZeroer
	if rv.Type().Implements(_isZeroerType) {
		return rv.Interface().(IsZeroer).IsZero()
	} else if rv.Kind() == re.Ptr && !rv.IsNil() && rv.Elem().Type().Implements(_isZeroerType) {
		return rv.Elem().Interface().(IsZeroer).IsZero()
	}

	switch rv.Kind() {
	case re.Bool:
		return !rv.Bool()
	case re.Int, re.Int8, re.Int16, re.Int32, re.Int64:
		return rv.Int() == 0
	case re.Uint, re.Uint8, re.Uint16, re.Uint32, re.Uint64, re.Uintptr:
		return rv.Uint() == 0
	case re.Float32, re.Float64:
		return rv.Float() == 0
	case re.Complex64, re.Complex128:
		return rv.Complex() == 0
	case re.Ptr:
		if !rv.IsValid() || rv.IsNil() {
			return true
		}
		return IsZero(rv.Elem().Interface())
	case re.String:
		return rv.String() == ""
	case re.Slice, re.Array:
		if rv.Len() == 0 {
			return true
		}

		for i := 0; i < rv.Len(); i++ {
			if !IsZero(rv.Index(i).Interface()) {
				return false
			}
		}
	case re.Map:
		return rv.Len() == 0
	}

	return re.DeepEqual(rv.Interface(), re.Zero(rv.Type()).Interface())
}

package clone

import (
	"reflect"
	"sync"

	"github.com/Nigel2392/goldcrest"
)

var (
	_typCheckIdentifier = "versatile.clone.AllowedType"
	_valCheckIdentifier = "versatile.clone.AllowedValue"
	_known_types        = make(map[reflect.Type]bool)
)

func init() {
	RegisterClonableCheck(CheckFuncDisallowIface[sync.Locker]) // if implements locker, return false (block)
}

type ClonableCheckFunc interface {
	ClonableCheckValueFunc | ClonableCheckTypeFunc
}

type ClonableCheckTypeFunc = func(reflect.Type) bool
type ClonableCheckValueFunc = func(reflect.Value, reflect.Type) bool

func RegisterClonableCheck[FUNC ClonableCheckFunc](fn FUNC) {
	switch any(fn).(type) {
	case ClonableCheckTypeFunc:
		goldcrest.Register(_typCheckIdentifier, 0, fn)
	case ClonableCheckValueFunc:
		goldcrest.Register(_valCheckIdentifier, 0, fn)
	}
}

// disallowed if provided [reflect.Type] implements [T]
func CheckFuncDisallowIface[T any](t reflect.Type) bool {
	return !t.Implements(reflect.TypeFor[T]())
}

func IsAllowedType(t reflect.Type) bool {
	return _allowsCloneType(t, true)
}

func IsAllowedValue(val reflect.Value) bool {
	if val.Kind() == reflect.Invalid {
		return false
	}

	typ := val.Type()
	orig := typ
	allows, ok := _known_types[typ]
	if ok {
		return allows
	}

checkValue:
	fL := goldcrest.Get[ClonableCheckValueFunc](_valCheckIdentifier)
	for _, check := range fL {
		if !check(val, typ) {
			_known_types[typ] = false
			return false
		}
	}

	if typ.Kind() != reflect.Interface && !_allowsCloneType(typ, false) {
		return false
	}

	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		allows = _allowsCloneType(typ.Elem(), false)

	case reflect.Pointer:
		typ = typ.Elem()
		val = val.Elem()
		goto checkValue

	case reflect.Interface:
		if !val.IsNil() {
			val = val.Elem()
			typ = val.Type()
			goto checkValue
		}

		allows = _allowsCloneType(typ, false)

	default:
		allows = _allowsCloneType(typ, false)
	}

	_known_types[orig] = allows
	return true
}

func _allowsCloneType(t reflect.Type, setCache bool) bool {
	if t.Kind() == reflect.Interface && t.NumMethod() == 0 {
		return true // is literal any
	}

	orig := t

checkTypes:
	allows, ok := _known_types[t]
	if ok {
		return allows
	}

	fL := goldcrest.Get[ClonableCheckTypeFunc](_typCheckIdentifier)
	for _, check := range fL {
		if !check(t) { // always set cache, even if [t] is [interface{}]
			_known_types[t] = false
			return false
		}
	}

	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		t = t.Elem()
		goto checkTypes
	}

	if setCache && orig.Kind() != reflect.Interface {
		_known_types[orig] = true
	}

	return true
}

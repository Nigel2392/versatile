package clone

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	gc "github.com/Nigel2392/goldcrest"
	"github.com/Nigel2392/versatile/bitcheck"
)

var (
	_typCheckIdentifier = "versatile.clone.AllowedType"
	_valCheckIdentifier = "versatile.clone.AllowedValue"
)

func init() {
	// These types cannot be cloned.
	OK.Check(CheckDisallowIface[sync.Locker])
	OK.Check(CheckDisallowType[sync.Cond])
	OK.Check(CheckDisallowType[sync.Map])
	OK.Check(CheckDisallowType[sync.Once])
	OK.Check(CheckDisallowType[sync.Pool])
	OK.Check(CheckDisallowType[sync.WaitGroup])
}

type (
	allowFlag int8

	AllowList struct {
		hooks      gc.HookRegistry
		knownTypes map[reflect.Type]allowFlag
	}

	ClonableCheckTypeFunc  = func(context.Context, reflect.Type) bool
	ClonableCheckValueFunc = func(context.Context, reflect.Value, reflect.Type) bool

	ClonableCheckFunc interface {
		ClonableCheckValueFunc | ClonableCheckTypeFunc
	}
)

const (
	UNKNOWN_FLAG_FALLBACK = true

	_false   allowFlag = -1
	_unknown allowFlag = 0
	_true    allowFlag = 1
)

func flagBool(flag allowFlag) bool {
	if flag == 0 {
		return UNKNOWN_FLAG_FALLBACK
	}
	return flag > 0
}

func NewAllowList() AllowList {
	return AllowList{
		hooks:      make(gc.HookRegistry),
		knownTypes: make(map[reflect.Type]allowFlag),
	}
}

// disallowed if provided [reflect.Type] implements [T]
func CheckDisallowIface[TYPE any](_ context.Context, t reflect.Type) bool {
	disallowed := reflect.TypeFor[TYPE]()

	// if implements, return true (block)
	return t.Implements(disallowed) ||
		t.Kind() != reflect.Pointer &&
			reflect.PointerTo(t).Implements(disallowed)
}

// disallowed if provided reflect.Type is the same kind as [TYPE]
func CheckDisallowType[TYPE any](_ context.Context, t reflect.Type) bool {
	disallowed := reflect.TypeFor[TYPE]()
	return t == disallowed || t.AssignableTo(disallowed) || t.ConvertibleTo(disallowed)
}

// disallowed if provided reflect.Type is the same kind as [TYPE]
func CheckDisallowKind[TYPE any](_ context.Context, t reflect.Type) bool {
	return t.Kind() == reflect.TypeFor[TYPE]().Kind() // if kind matches, return true (block)
}

var OK = NewAllowList()

// The function must return true to block.
func (a *AllowList) Check[FUNC ClonableCheckFunc](fn FUNC) {
	switch any(fn).(type) {
	case ClonableCheckTypeFunc:
		a.hooks.Register(_typCheckIdentifier, 0, fn)
	case ClonableCheckValueFunc:
		a.hooks.Register(_valCheckIdentifier, 0, fn)
	}
}

func (a AllowList) Type(ctx context.Context, t reflect.Type) bool {
	flag := a._allowsCloneType(ctx, t, af_setCache|af_deep, nil)
	return flagBool(flag)
}

func (a AllowList) Value(ctx context.Context, val reflect.Value) (allows bool) {
	if val.Kind() == reflect.Invalid {
		return false
	}

	for (val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface) && !val.IsNil() {
		val = val.Elem()
	}

	if val.Kind() == reflect.Interface && val.IsNil() {
		return true
	}

	var (
		typ = val.Type()
		fL  = a.hooks.Get[ClonableCheckValueFunc](_valCheckIdentifier)
	)

	flag := a._allowsCloneValue(ctx, typ, val, fL)
	switch flag {
	case -1: // disallowed values
		a.knownTypes[typ] = _false
		allows = false

	case 1: // allowed values
		a.knownTypes[typ] = _true
		allows = true

	case 0: // unknown
		allows = UNKNOWN_FLAG_FALLBACK

	default:
		panic(fmt.Sprintf("unknown flag %d, must be -1, 0 or 1", flag))
	}

	return allows
}

type actionFlag = bitcheck.Flag

const (
	af_none     actionFlag = iota
	af_setCache actionFlag = 1 << iota
	af_deep
)

// 1: allowed
// 0: unsure
// -1: block
func (a AllowList) _allowsCloneType(ctx context.Context, t reflect.Type, actionFlag actionFlag, fL []ClonableCheckTypeFunc) allowFlag {
checkTypes:
	if t.Kind() == reflect.Interface && t.NumMethod() == 0 {
		return _unknown // is literal any, cannot be sure
	}

	allows, ok := a.knownTypes[t]
	if ok && allows != 0 {
		return allows
	}

	if fL == nil {
		fL = a.hooks.Get[ClonableCheckTypeFunc](_typCheckIdentifier)
	}

	if len(fL) == 0 {
		return _unknown
	}

	for _, check := range fL {
		if check(ctx, t) { // always set cache, even if [t] is [interface{}]
			a.knownTypes[t] = _false
			return _false
		}
	}

	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		t = t.Elem()
		goto checkTypes

	case reflect.Map:
		var (
			k = t.Key()
			e = t.Elem()
		)

		flag := a._allowsCloneType(ctx, k, actionFlag, fL)

		if flag == _false {
			a.knownTypes[t] = _false
			return _false
		}

		flag = a._allowsCloneType(ctx, e, actionFlag, fL)

		if flag == _false {
			a.knownTypes[t] = _false
			return _false
		}

	case reflect.Struct:
		if !actionFlag.Is(af_deep) {
			return _unknown
		}

		for index := range t.NumField() {

			typ := t.Field(index).Type
			if v, ok := a.knownTypes[typ]; ok && v != 0 {
				return v
			}

			flag := a._allowsCloneType(ctx, typ, actionFlag, fL)

			if flag == _false {
				a.knownTypes[t] = _false
				return _false
			}
		}

	}

	if actionFlag.Is(af_setCache) && t.Kind() != reflect.Interface {
		a.knownTypes[t] = _true
	}

	return _true
}

// 1: allowed
// 0: unsure
// -1: block
func (a AllowList) _allowsCloneValue(ctx context.Context, typ reflect.Type, val reflect.Value, fL []ClonableCheckValueFunc) allowFlag {
	orig := typ

checkValue:
	allows, ok := a.knownTypes[typ]
	if ok {
		return allows
	}

	for _, check := range fL {
		if check(ctx, val, typ) {
			a.knownTypes[orig] = _false
			return _false
		}
	}

	if !(typ.Kind() == reflect.Interface && typ.NumMethod() == 0) {
		flag := a._allowsCloneType(ctx, typ, af_none, nil)
		if flag < 0 {
			return _false
		}
	}

	var allowFlag allowFlag
	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		allowFlag = a._allowsCloneType(ctx, typ.Elem(), af_none, nil)

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

		allowFlag = a._allowsCloneType(ctx, typ, af_none, nil)

	case reflect.Struct:

		for i := range val.NumField() {
			v := val.Field(i)
			vt := v.Type()

			if v, ok := a.knownTypes[vt]; ok {
				return v
			}

			flag := a._allowsCloneValue(ctx, vt, v, fL)

			if flag == _false {
				a.knownTypes[vt] = _false
				a.knownTypes[typ] = _false
				a.knownTypes[orig] = _false
				return _false
			}
		}

	default:
		allowFlag = _true

	}

	if allowFlag != 0 && !(typ.Kind() == reflect.Interface && typ.NumMethod() == 0) {
		a.knownTypes[orig] = allowFlag
	}

	return allowFlag
}

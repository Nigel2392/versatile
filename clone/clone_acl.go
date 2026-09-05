package clone

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"sync"
	"unsafe"

	gc "github.com/Nigel2392/goldcrest"
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
		knownTypes map[reflect.Type]bool
	}

	ClonableCheckTypeFunc  = func(reflect.Type) bool
	ClonableCheckValueFunc = func(reflect.Value, reflect.Type) bool

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

func boolFlag(b bool) allowFlag {
	if b {
		return _true
	}
	return _false
}

func flagBool(flag allowFlag) bool {
	if flag == 0 {
		return UNKNOWN_FLAG_FALLBACK
	}
	return flag > 0
}

func NewAllowList() AllowList {
	return AllowList{
		hooks:      make(gc.HookRegistry),
		knownTypes: make(map[reflect.Type]bool),
	}
}

// disallowed if provided [reflect.Type] implements [T]
func CheckDisallowIface[TYPE any](t reflect.Type) bool {
	disallowed := reflect.TypeFor[TYPE]()

	// if implements, return true (block)
	return t.Implements(disallowed) ||
		t.Kind() != reflect.Pointer &&
			reflect.PointerTo(t).Implements(disallowed)
}

// disallowed if provided reflect.Type is the same kind as [TYPE]
func CheckDisallowType[TYPE any](t reflect.Type) bool {
	disallowed := reflect.TypeFor[TYPE]()
	return t == disallowed || t.AssignableTo(disallowed) || t.ConvertibleTo(disallowed)
}

// disallowed if provided reflect.Type is the same kind as [TYPE]
func CheckDisallowKind[TYPE any](t reflect.Type) bool {
	return t.Kind() == reflect.TypeFor[TYPE]().Kind() // if kind matches, return true (block)
}

var OK = NewAllowList()

func (a *AllowList) Check[FUNC ClonableCheckFunc](fn FUNC) {
	switch any(fn).(type) {
	case ClonableCheckTypeFunc:
		a.hooks.Register(_typCheckIdentifier, 0, fn)
	case ClonableCheckValueFunc:
		a.hooks.Register(_valCheckIdentifier, 0, fn)
	}
}

func (a AllowList) Type(t reflect.Type) (allows bool) {
	flag := a._allowsCloneType(t, true)
	return flagBool(flag)
}

func (a AllowList) Value(ctx context.Context, val reflect.Value) bool {
	if val.Kind() == reflect.Invalid {
		return false
	}

	var (
		typ   = val.Type()
		_, fL = (*(*gc.HookRegistry)(unsafe.Pointer(&a))).
			GetIter[ClonableCheckValueFunc](_valCheckIdentifier)
	)

	allows, ok := a.knownTypes[typ]
	if ok {
		return allows
	}

	flag := a._allowsCloneValue(ctx, typ, val, fL)
	switch flag {
	case -1: // disallowed values
		a.knownTypes[typ] = false
		allows = false

	case 1: // allowed values
		a.knownTypes[typ] = true
		allows = true

	case 0: // unknown
		allows = UNKNOWN_FLAG_FALLBACK

	default:
		panic(fmt.Sprintf("unknown flag %d, must be -1, 0 or 1", flag))
	}

	return allows
}

// 1: allowed
// 0: unsure
// -1: block
func (a AllowList) _allowsCloneType(t reflect.Type, setCache bool) allowFlag {

	var (
		orig = t
		fL   []ClonableCheckTypeFunc
	)

checkTypes:
	if t.Kind() == reflect.Interface && t.NumMethod() == 0 {
		return 1 // is literal any
	}

	allows, ok := a.knownTypes[t]
	if ok {
		return boolFlag(allows)
	}

	if fL == nil {
		fL = a.hooks.Get[ClonableCheckTypeFunc](_typCheckIdentifier)
	}

	if len(fL) == 0 {
		return _unknown
	}

	for _, check := range fL {
		if check(t) { // always set cache, even if [t] is [interface{}]
			a.knownTypes[orig] = false
			return _false
		}
	}

	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		t = t.Elem()
		goto checkTypes
	}

	if setCache && orig.Kind() != reflect.Interface {
		a.knownTypes[orig] = true
	}

	return _true
}

// 1: allowed
// 0: unsure
// -1: block
func (a AllowList) _allowsCloneValue(_ context.Context, typ reflect.Type, val reflect.Value, fL iter.Seq[ClonableCheckValueFunc]) allowFlag {
	orig := typ
checkValue:
	allows, ok := a.knownTypes[typ]
	if ok {
		return boolFlag(allows)
	}

	for check := range fL {
		if check(val, typ) {
			a.knownTypes[orig] = false
			return _false
		}
	}

	if typ.Kind() != reflect.Interface {
		flag := a._allowsCloneType(typ, false)
		if flag < 0 {
			return _false
		}
	}

	var allowFlag allowFlag
	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		allowFlag = a._allowsCloneType(typ.Elem(), false)

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

		allowFlag = a._allowsCloneType(typ, false)

	default:
		allowFlag = _true

	}

	if allowFlag != 0 {
		a.knownTypes[orig] = allowFlag > 0
	}

	return allowFlag
}

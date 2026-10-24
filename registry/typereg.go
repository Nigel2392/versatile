package registry

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sync"

	"github.com/Nigel2392/versatile/bitcheck"
)

type Flag bitcheck.Flag

const (
	NONE        Flag = iota
	EXACT_IFACE Flag = iota << 1
	NOTFOUND_REG
)

type _TYPE = reflect.Type

func NewTypeRegistry[T any](lock sync.Locker, flags Flag, get func(dst, src reflect.Type) []Getter[T]) *TypeRegistry[T] {
	reg := new(TypeRegistry[T]{
		flags: flags,
		get:   get,
		lock:  lock,
		types: make(map[any]T),
		iface: interfaceReg[T]{
			lookup:         make(map[reflect.Type]int),
			checkers:       make([]*interfaceRegChecker[T], 0),
			existsDstCache: make(map[reflect.Type]int),
		},
	})

	if flags&NOTFOUND_REG == NOTFOUND_REG {
		reg.notFound = make(map[RegKey[reflect.Type]]struct{})
	}

	return reg
}

type RegKey[T any] struct {
	Dst T
	Src T
}

type TypeRegistry[T any] struct {
	iface    interfaceReg[T]
	types    map[any]T
	get      func(dst, src reflect.Type) []Getter[T]
	notFound map[RegKey[reflect.Type]]struct{}
	lock     sync.Locker
	flags    Flag
}

type interfaceReg[T any] struct {
	lookup         map[reflect.Type]int
	checkers       []*interfaceRegChecker[T]
	existsDstCache map[reflect.Type]int
}

type interfaceRegChecker[T any] struct {
	chk  _TYPE
	scan T
}

func (reg *TypeRegistry[T]) getters(dst, src reflect.Type) []Getter[T] {
	if reg.get == nil {
		return nil
	}

	return reg.get(dst, src)
}

func (reg *TypeRegistry[T]) WithLock(l sync.Locker) {
	reg.lock = l
}

func (reg *TypeRegistry[T]) Clone() *TypeRegistry[T] {
	if reg.lock != nil {
		reg.lock.Lock()
		defer reg.lock.Unlock()
	}

	return &TypeRegistry[T]{
		iface: interfaceReg[T]{
			lookup:         maps.Clone(reg.iface.lookup),
			existsDstCache: maps.Clone(reg.iface.existsDstCache),
			checkers:       slices.Clone(reg.iface.checkers),
		},
		types:    maps.Clone(reg.types),
		notFound: maps.Clone(reg.notFound),
		get:      reg.get,
		lock:     nil, // do not clone lock
		flags:    reg.flags,
	}
}

func (reg *TypeRegistry[T]) Clear() int {
	if reg.lock != nil {
		reg.lock.Lock()
		defer reg.lock.Unlock()
	}

	l := len(reg.types) + len(reg.iface.checkers)
	clear(reg.types)
	clear(reg.notFound)
	reg.iface = interfaceReg[T]{
		lookup:         make(map[reflect.Type]int),
		checkers:       make([]*interfaceRegChecker[T], 0),
		existsDstCache: make(map[reflect.Type]int),
	}

	return l
}

func (reg *TypeRegistry[T]) Register(dstType reflect.Type, srcOrT ...any) (overwrite bool) {
	if reg.lock != nil {
		reg.lock.Lock()
		defer reg.lock.Unlock()
	}

	var (
		key any
		v   T
	)

	switch len(srcOrT) {
	case 1:
		key = dstType
		v = srcOrT[0].(T)
	case 2:
		key = RegKey[reflect.Type]{Dst: dstType, Src: srcOrT[0].(reflect.Type)}
		v = srcOrT[1].(T)
	default:
		panic(fmt.Sprintf("arguments invalid, expected 1 or 2, got %d", len(srcOrT)))
	}

	_, overwrite = reg.types[key]
	reg.types[key] = v
	return overwrite
}

func (reg *TypeRegistry[T]) RegisterKind(dstKind, srcKind reflect.Kind, checker T) (overwrite bool) {
	if reg.lock != nil {
		reg.lock.Lock()
		defer reg.lock.Unlock()
	}

	var key any
	switch {
	case dstKind != 0 && srcKind != 0:
		key = RegKey[reflect.Kind]{dstKind, srcKind}
	case dstKind != 0:
		key = dstKind
	case srcKind != 0:
		key = srcKind
	default:
		panic("must provide either dst or src")
	}
	_, overwrite = reg.types[key]
	reg.types[key] = checker
	return overwrite
}

func (reg *TypeRegistry[T]) InterfaceRegister(iface reflect.Type, scan T) (overwrite bool) {
	if reg.lock != nil {
		reg.lock.Lock()
		defer reg.lock.Unlock()
	}

	var (
		checker *interfaceRegChecker[T]
		pos, ok = reg.iface.lookup[iface]
	)

	if !ok {
		// init new checker subregistry
		checker = &interfaceRegChecker[T]{
			chk: iface,
		}
		reg.iface.lookup[iface] = len(reg.iface.checkers)
		reg.iface.checkers = append(reg.iface.checkers, checker)
	} else {
		// fetch directly from list index
		checker = reg.iface.checkers[pos]
		overwrite = true
	}

	checker.scan = scan

	return overwrite
}

type Getter[T any] = func(types map[any]T, key RegKey[reflect.Type], kindKey RegKey[reflect.Kind]) (T, bool)

type locker struct {
	sync.Locker
}

func (l locker) lock() {
	if l.Locker != nil {
		l.Locker.Lock()
	}
}

func (l locker) unlock() {
	if l.Locker != nil {
		l.Locker.Unlock()
	}
}

func (reg *TypeRegistry[T]) Get(dstTyp reflect.Type, srcTyp reflect.Type) (obj T, found bool) {
	if len(reg.types) == 0 && len(reg.iface.checkers) == 0 {
		return obj, false
	}

	if reg.flags&NOTFOUND_REG == NOTFOUND_REG {
		if _, notExists := reg.notFound[RegKey[reflect.Type]{Dst: dstTyp, Src: srcTyp}]; notExists {
			return obj, false
		}
	}

	var (
		idx int
		ok  bool
		mu  = locker{reg.lock}
	)

	// skip until later on if dst kind is an interface
	if reg.flags&EXACT_IFACE == EXACT_IFACE && dstTyp.Kind() != reflect.Interface {
		goto checkReg
	}

	if len(reg.iface.checkers) == 0 {
		goto checkReg
	}

	idx, ok = reg.iface.lookup[dstTyp]
	if ok {
		return reg.iface.checkers[idx].scan, true
	}

	idx, ok = reg.iface.existsDstCache[dstTyp]
	switch {
	case idx == -1:
	case ok:
		return reg.iface.checkers[idx].scan, true
	case !ok:
		for idx, iScan := range reg.iface.checkers {
			if dstTyp.Implements(iScan.chk) {
				mu.lock()
				reg.iface.existsDstCache[dstTyp] = idx
				mu.unlock()
				return iScan.scan, true
			}
		}

		mu.lock()
		reg.iface.existsDstCache[dstTyp] = -1
		mu.unlock()
	}

checkReg:
	final := reg.getters(dstTyp, srcTyp)
	key := RegKey[reflect.Type]{dstTyp, srcTyp}
	kindKey := RegKey[reflect.Kind]{dstTyp.Kind(), srcTyp.Kind()}
	if len(final) == 0 {
		return GetDefault(reg.types, key, kindKey)
	}

	for _, fn := range final {
		obj, found = fn(reg.types, key, kindKey)
		if found {
			return obj, true
		}
	}

	if reg.flags&NOTFOUND_REG == NOTFOUND_REG {
		mu.lock()
		reg.notFound[RegKey[reflect.Type]{Dst: dstTyp, Src: srcTyp}] = struct{}{}
		mu.unlock()
	}

	return obj, false
}

func GetDefault[T any](types map[any]T, key RegKey[reflect.Type], kindKey RegKey[reflect.Kind]) (obj T, found bool) {
	if obj, found = types[key]; found {
		return obj, true
	}

	obj, found = types[kindKey]
	return obj, found
}

func GetType[T any](types map[any]T, key RegKey[reflect.Type], _ RegKey[reflect.Kind]) (obj T, found bool) {
	obj, found = types[key]
	return
}

func GetKind[T any](types map[any]T, _ RegKey[reflect.Type], kindKey RegKey[reflect.Kind]) (obj T, found bool) {
	obj, found = types[kindKey]
	return
}

func GetTypeSrc[T any](types map[any]T, key RegKey[reflect.Type], _ RegKey[reflect.Kind]) (obj T, found bool) {
	obj, found = types[key.Src]
	return
}

func GetTypeDst[T any](types map[any]T, key RegKey[reflect.Type], _ RegKey[reflect.Kind]) (obj T, found bool) {
	obj, found = types[key.Dst]
	return obj, found
}

func GetKindSrc[T any](types map[any]T, _ RegKey[reflect.Type], kindKey RegKey[reflect.Kind]) (obj T, found bool) {
	obj, found = types[kindKey.Src]
	return
}

func GetKindDst[T any](types map[any]T, _ RegKey[reflect.Type], kindKey RegKey[reflect.Kind]) (obj T, found bool) {
	obj, found = types[kindKey.Dst]
	return obj, found
}

//go:build asyncmethods
// +build asyncmethods

package registry

import (
	"errors"
	"reflect"

	"github.com/go4org/hashtriemap"
)

type methodKey struct {
	rtype reflect.Type
	name  string
}

type storedMethod struct {
	derefs int
	method *reflect.Method
}

type MethodsRegistry struct {
	tCache hashtriemap.HashTrieMap[methodKey, storedMethod]
}

func NewMethods(initSize int) *MethodsRegistry {
	if initSize == 0 {
		initSize = 32
	}

	return new(MethodsRegistry{
		tCache: hashtriemap.HashTrieMap[methodKey, storedMethod]{},
	})
}

// Clears the cache and returns the total count of cleared items
func (m *MethodsRegistry) CoolDown() {
	m.tCache.Clear()
}

// Look for multiple methods for this type, cache the first one found for future uses.
//
// This cache cannot store multiple methods per type, and will always yield the cached result.
func (m *MethodsRegistry) For(rv reflect.Value, methodNames []string, validChk ...func(reflect.Value, *reflect.Method) error) (rec reflect.Value, rm *reflect.Method, err error) {
	var (
		notFound []string
		rt       = rv.Type()
		key      = methodKey{
			rtype: rt,
		}
	)

	for _, name := range methodNames {
		key.name = name
		sm, ok := m.tCache.Load(key)

		// method is cached and valid
		if ok && sm.method != nil {

			// deref until rv is acceptable as receiver value
			for range sm.derefs {
				rv = rv.Elem()
			}

			// explicitly unlock the mutex before returning
			return rv, sm.method, nil
		}

		// method is not cached, append and store name
		if !ok {
			notFound = append(notFound, name)
		}

		// `if ok && sm.method == nil` means the name will not get appended to the list
		// this is intended behaviour to skip having to do work for methods that
		// might already be cached
	}

	return m.findMethodsFor(rv, rt, notFound, validChk...)
}

func (m *MethodsRegistry) findMethodsFor(rv reflect.Value, rt reflect.Type, methodNames []string, validChk ...func(reflect.Value, *reflect.Method) error) (rec reflect.Value, rm *reflect.Method, err error) {

	var key = methodKey{rtype: rt}

findLoop:
	for _, method := range methodNames {
		var (
			// count derefs required to make rv acceptable as receiver value
			derefs int

			// shadow the variables so they get reset on next loop
			rt = rt
			rv = rv
		)
		key.name = method

		// goto reCheck occurs when the method wasn't found on the current type
		// AND the current type has a pointer reflect kind.
	reCheck:
		rMethod, ok := rt.MethodByName(method)
		switch {
		case !ok && rt.Kind() == reflect.Pointer:
			// deref type and value
			rt = rt.Elem()
			rv = rv.Elem()

			// incr deref counter
			derefs++

			// check if the method exists on the new dereferenced type
			goto reCheck

		case !ok:
			// only lock for the short while
			// yes, this is still racy as technically
			// a write could have occurred between
			// the previous call and the current one.
			//
			// no, i do not care
			//
			// worst case is a cache miss.
			m.tCache.Store(key, storedMethod{})
			continue
		}

		rm = &rMethod
	chkLoop:
		for _, valid := range validChk {
			if valid == nil {
				continue chkLoop
			}

			err = valid(rv, rm)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					continue findLoop
				} else {
					return rv, rm, err
				}
			}
		}

		// racy again.
		//
		// see above statement.
		m.tCache.Store(key, storedMethod{
			derefs: derefs,
			method: rm,
		})
		return rv, rm, nil
	}

	return rv, nil, ErrNotFound
}

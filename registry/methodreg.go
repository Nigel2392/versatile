package registry

import (
	"errors"
	"reflect"
	"sync"
)

type methodKey struct {
	rtype reflect.Type
	name  string
}

type MethodsRegistry struct {
	mu     sync.RWMutex
	tCache map[methodKey]*reflect.Method
}

func NewMethods(initSize int) *MethodsRegistry {
	if initSize == 0 {
		initSize = 32
	}

	return new(MethodsRegistry{
		mu:     sync.RWMutex{},
		tCache: make(map[methodKey]*reflect.Method, initSize),
	})
}

// Look for multiple methods for this type, cache the first one found for future uses.
//
// This cache cannot store multiple methods per type, and will always yield the cached result.
func (m *MethodsRegistry) For(rv reflect.Value, methodNames []string, validChk ...func(reflect.Value, *reflect.Method) error) (rm *reflect.Method, err error) {
	var (
		notFound []string
		ok       bool

		rt  = rv.Type()
		key = methodKey{
			rtype: rt,
		}
	)

	m.mu.RLock()
	for _, name := range methodNames {
		key.name = name
		rm, ok = m.tCache[key]

		// method is cached and valid
		if ok && rm != nil {
			return rm, nil
		}

		// method is not cached, append and store name
		if !ok {
			notFound = append(notFound, name)
		}
	}

	m.mu.RUnlock()

findLoop:
	for _, method := range notFound {
		key.name = method
		rMethod, ok := rt.MethodByName(method)
		if !ok {
			// only lock for the short while
			// yes, this is still racy
			// no, i do not care
			//
			// worst case is a cache miss.
			m.mu.Lock()
			m.tCache[key] = nil
			m.mu.Unlock()
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
					return rm, err
				}
			}
		}

		// racy again.
		//
		// see above statement.
		m.mu.Lock()
		m.tCache[key] = rm
		m.mu.Unlock()
		return rm, nil
	}

	return nil, ErrNotFound
}

package registry

import (
	"errors"
	"reflect"
	"sync"
)

type MethodsRegistry struct {
	mu     sync.RWMutex
	tCache map[reflect.Type]cachedMethodData
}

func NewMethods() *MethodsRegistry {
	return new(MethodsRegistry{
		mu:     sync.RWMutex{},
		tCache: make(map[reflect.Type]cachedMethodData),
	})
}

// Look for multiple methods for this type, cache the first one found for future uses.
//
// This cache cannot store multiple methods per type, and will always yield the cached result.
func (m *MethodsRegistry) For(rv reflect.Value, methodNames []string, validChk ...func(reflect.Value, reflect.Method) error) (rm reflect.Method, err error) {
	rt := rv.Type()

	m.mu.RLock()
	cached, ok := m.tCache[rt]
	if ok {
		m.mu.RUnlock()

		if !cached.ok {
			return cached.method, ErrNotFound
		}

		return cached.method, nil
	}
	m.mu.RUnlock()

	for _, method := range methodNames {
		rm, ok = rt.MethodByName(method)
		if !ok {
			continue
		}

		for _, valid := range validChk {
			if valid == nil {
				continue
			}

			err = valid(rv, rm)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					continue
				} else {
					return rm, err
				}
			}

		}

		// method is found, valid and can be cached
		m.mu.Lock()
		m.tCache[rt] = cachedMethodData{
			methodName: method,
			method:     rm,
			ok:         true,
		}
		m.mu.Unlock()
		return rm, nil
	}

	m.mu.Lock()
	m.tCache[rt] = cachedMethodData{}
	m.mu.Unlock()
	return reflect.Method{}, ErrNotFound
}

type cachedMethodData struct {
	method     reflect.Method
	methodName string
	ok         bool
}

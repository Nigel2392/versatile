package clone

import (
	"context"

	"github.com/Nigel2392/versatile/registry"
)

type contextKey struct{}

type stateContext struct {
	context.Context
	state *State
}

func (s *stateContext) Value(key any) any {
	if key == (contextKey{}) {
		return s
	}

	return s.Context.Value(key)
}

// A shared state can be used to provide a shared step cache, optional shared pointer cache
// and any other options that provide functionality to the state.
//
// The [Clone] function can also provide it's own options to mutate **A COPY OF** the shared state.
func SharedStateContext(ctx context.Context, opts ...func(s *State)) context.Context {
	sCtx, ok := ctx.Value(contextKey{}).(*stateContext)
	if !ok {
		// only need to override context when state is not present already
		sCtx = &stateContext{
			Context: ctx,
			state: &State{
				pointers: make(map[oldPtr]newPtr),
			},
		}
	}

	for _, opt := range opts {
		opt(sCtx.state)
	}

	sCtx.state.Flags |= cf_shared_state

	if sCtx.state.cache == nil && !sCtx.state.Flags.Is(CF_NO_CACHE) {
		sCtx.state.cache = &cacheRegistry{_steps: registry.NewTypeRegistry(nil, registry.EXACT_IFACE, _getterExactType)}
	}

	return sCtx
}

func StateFromContext(ctx context.Context) (*State, bool) {
	s, ok := ctx.Value(contextKey{}).(*stateContext)
	if ok {
		return s.state, ok
	}
	return nil, false
}

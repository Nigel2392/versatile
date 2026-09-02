package clone

import "context"

type stateContextKey struct{}

// A shared state can be used to provide a shared step cache, optional shared pointer cache
// and any other options that provide functionality to the state.
//
// The [Clone] function can also provide it's own options to mutate **A COPY OF** the shared state.
func SharedStateContext(ctx context.Context, opts ...func(s *State)) context.Context {
	var (
		state *State
		ok    bool
	)

	if state, ok = StateFromContext(ctx); !ok {
		state = new(State{
			pointers: make(map[oldPtr]newPtr),
			cache:    &cacheRegistry{steps: make(map[any]Step)},
		})

		// only need to override context when state is not present already
		ctx = context.WithValue(ctx, stateContextKey{}, state)
	}

	for _, opt := range opts {
		opt(state)
	}

	return ctx
}

func StateFromContext(ctx context.Context) (*State, bool) {
	s, ok := ctx.Value(stateContextKey{}).(*State)
	return s, ok
}

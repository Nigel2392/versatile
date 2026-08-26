package clone

import (
	"context"
	"reflect"

	"github.com/Nigel2392/versatile/bitcheck"
)

type (
	CloneFlag = bitcheck.Flag
	oldPtr    = uintptr
	newPtr    = uintptr
)

type State struct {
	Ctx   context.Context
	Flags CloneFlag

	pointers map[oldPtr]newPtr
	reg      *stepRegistry
}

func (s *State) Step(typ reflect.Type) (Step, bool) {
	if s.reg == nil {
		return stepReg.getStep(typ)
	}

	return s.reg.getStep(typ)
}

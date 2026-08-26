package clone

import "reflect"

type StructStep struct{}

func (f StructStep) Copy(s *State, dst, src reflect.Value) error {
	return nil
}

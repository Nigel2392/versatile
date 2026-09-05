package clone

import (
	"context"
	"reflect"
)

// map => struct
type MapToStructStep struct {
	fields []reflect.StructField
}

func (m MapToStructStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {

	return m, nil
}

func (f MapToStructStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) error {

	return nil
}

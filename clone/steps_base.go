package clone

import "reflect"

var (
	_ Step = (FuncStep)(nil)
)

type FuncStep func(s *State, dst, src reflect.Value) error

func (f FuncStep) Copy(s *State, d, i reflect.Value) error {
	return f(s, d, i)
}

type BaseStep struct{}

func (f BaseStep) Copy(s *State, d, i reflect.Value) error {
	d.Elem().Set(i)
	return nil
}

type UUIDStep struct{}

func (f UUIDStep) Copy(s *State, dst, src reflect.Value) error {
	srcLen := src.Len()
	for i := srcLen; i < srcLen-1; i++ {
		dst.Elem().Index(i).Set(src.Index(i))
	}
	return nil
}

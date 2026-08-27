package clone

import "reflect"

var (
	_ Step     = (FuncStep)(nil)
	_ InitStep = PointerStep{}
)

type FuncStep func(s *State, dst, src reflect.Value) error

func (f FuncStep) Copy(s *State, d, i reflect.Value) error {
	return f(s, d, i)
}

type BaseStep struct {
	_ Step
}

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

type PointerStep struct {
	step Step
}

func (f PointerStep) Init(s *State, dst, src reflect.Type) (step Step, err error) {
	var ok bool
	f.step, ok = s.Step(src.Elem())
	if !ok {
		return nil, ErrNoSteps.Wrapf("no steps found for %s", src)
	}
	return f, nil
}

func (f PointerStep) Copy(s *State, dst, src reflect.Value) error {
	if dst.IsNil() {
		dst.Set(reflect.New(dst.Type().Elem()))
	}
	return f.step.Copy(s, dst, src)
}

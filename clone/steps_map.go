package clone

import (
	"context"
	"reflect"

	"github.com/Nigel2392/versatile/internal/danger"
)

// map => map/interface
type MapStep struct {
	key *mapItem
	val *mapItem
}

func (m MapStep) Init(ctx context.Context, s *State, dst, src reflect.Type) (step Step, err error) {

	dstTyp := dst
	if dst.Kind() == reflect.Interface {
		dstTyp = src
	}

	srcKey := src.Key()
	srcElem := src.Elem()

	dstKey := dstTyp.Key()
	dstElem := dstTyp.Elem()

	if !danger.IsValueType(srcKey.Kind()) || !danger.IsValueType(dstKey.Kind()) || !srcKey.AssignableTo(dstKey) {
		step, err = s.StepInit(ctx, dstKey, srcKey)
		if err != nil {
			return nil, err
		}

		m.key = &mapItem{
			dstType: dstKey,
			step:    step,
		}
	}

	if !danger.IsValueType(srcElem.Kind()) || !danger.IsValueType(dstElem.Kind()) || !srcElem.AssignableTo(dstElem) {
		step, err = s.StepInit(ctx, dstElem, srcElem)
		if err != nil {
			return nil, err
		}

		m.val = &mapItem{
			dstType: dstElem,
			step:    step,
		}
	}

	return m, nil
}

func (f MapStep) Copy(ctx context.Context, s *State, dst, src reflect.Value) (err error) {
	if src.IsNil() {
		dst.Elem().Set(reflect.Zero(dst.Type().Elem()))
		return nil
	}

	newMap, cached := s.MakeMap(src, dst, src.Type(), src.Len())

	if cached {
		dst.Elem().Set(newMap)
		return nil
	}

	_range := src.MapRange()
	for _range.Next() {
		k := _range.Key()
		v := _range.Value()

		if f.key != nil {
			k, err = f.key.Copy(ctx, s, k)
			if err != nil {
				return err
			}
		}

		if f.val != nil {
			v, err = f.val.Copy(ctx, s, v)
			if err != nil {
				return err
			}
		}

		newMap.SetMapIndex(k, v)
	}

	dst.Elem().Set(newMap)

	return nil
}

type mapItem struct {
	dstType reflect.Type
	step    Step
}

// Copy implementation of Copy function for map item copier
func (c mapItem) Copy(ctx context.Context, s *State, src reflect.Value) (dst reflect.Value, err error) {
	dst = reflect.New(c.dstType)
	err = c.step.Copy(ctx, s, dst, src)
	return dst.Elem(), err

}

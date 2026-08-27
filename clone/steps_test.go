package clone

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"testing"
)

type baseStepTest struct {
	expected any
	dst      reflect.Value
	src      reflect.Value
}

func newBaseStepTest[T any](expect T, src T) baseStepTest {
	return baseStepTest{
		expected: expect,
		dst:      reflect.New(reflect.TypeFor[T]()),
		src:      reflect.ValueOf(src),
	}
}

func newBaseStepPtrTest[T any](expect T, src T) baseStepTest {
	var v T
	var dst = new(T)
	*dst = v

	var dstP = new(*T)
	*dstP = dst

	var exp = new(T)
	*exp = expect

	return baseStepTest{
		expected: exp,
		dst:      reflect.ValueOf(dstP),
		src:      reflect.ValueOf(&src),
	}
}

func newSliceStepTest[T any](src []T) baseStepTest {
	var zero = make([]T, 0)
	return baseStepTest{
		expected: src,
		dst:      reflect.ValueOf(&zero),
		src:      reflect.ValueOf(src),
	}
}

var stepTests = []baseStepTest{
	newBaseStepTest(55, int(55)),
	newBaseStepTest(55, int8(55)),
	newBaseStepTest(55, int16(55)),
	newBaseStepTest(55, int32(55)),
	newBaseStepTest(55, int64(55)),

	newBaseStepTest(55, uint(55)),
	newBaseStepTest(55, uint8(55)),
	newBaseStepTest(55, uint16(55)),
	newBaseStepTest(55, uint32(55)),
	newBaseStepTest(55, uint64(55)),

	newBaseStepTest(55.55, float32(55.55)),
	newBaseStepTest(55.55, float64(55.55)),

	newBaseStepTest("my string", "my string"),
	newBaseStepTest(true, true),

	newBaseStepPtrTest(55, int(55)),
	newBaseStepPtrTest(55, int8(55)),
	newBaseStepPtrTest(55, int16(55)),
	newBaseStepPtrTest(55, int32(55)),
	newBaseStepPtrTest(55, int64(55)),

	newBaseStepPtrTest(55, uint(55)),
	newBaseStepPtrTest(55, uint8(55)),
	newBaseStepPtrTest(55, uint16(55)),
	newBaseStepPtrTest(55, uint32(55)),
	newBaseStepPtrTest(55, uint64(55)),

	newBaseStepPtrTest(55.55, float32(55.55)),
	newBaseStepPtrTest(55.55, float64(55.55)),

	newBaseStepPtrTest("my string", "my string"),
	newBaseStepPtrTest(true, true),

	newSliceStepTest([]int{1, 2, 3, 4}),
	newSliceStepTest([]float64{1, 2, 3, 4}),
	newSliceStepTest([]any{1, 2, 3, 4}),
}

type myFace interface{ StructMethod() string }

type myStruct struct{ val string }

func (m myStruct) StructMethod() string { return m.val }

func TestSteps(t *testing.T) {

	t.Run("TestPointerCloneFunc", func(t *testing.T) {

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Exception while executing tests: %v", r)
			}
		}()

		var i = new(int)
		var s = 55
		if err := Copy(t.Context(), &i, &s); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if *i != s {
			t.Errorf("expected 'i' to be %d, got %d", s, i)
		} else {
			t.Logf("i == %d: %d", *i, s)
		}

		s = 5

		t.Log(*i, s)
	})

	t.Run("TestInterfaceAssignable", func(t *testing.T) {

		var testFunc = (func(fn func(t *testing.T)) func(t *testing.T) {
			return func(t *testing.T) {
				t.Helper()
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Exception while executing tests: %v: %s", r, string(debug.Stack()))
					}
				}()

				fn(t)
			}
		})

		t.Run("IfacePointer", testFunc(func(t *testing.T) {
			var i = new(any)
			var s = 55
			if err := Copy(t.Context(), i, s); err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if *i != s {
				t.Errorf("expected 'i' to be %d, got %d", s, i)
			} else {
				t.Logf("i == %d: %d", *i, s)
			}

			s = 5

			t.Log(*i, s)

		}))

		t.Run("IfaceWithMethodPointer", testFunc(func(t *testing.T) {
			var i = new(myFace)
			var s = &myStruct{"hello world"}
			if err := Copy(t.Context(), i, s); err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if (*i).StructMethod() != s.val {
				t.Errorf("expected 'i' to be %q, got %q", s.val, (*i).StructMethod())
			} else {
				t.Logf("i == %+v: %+v", *i, s)
			}

			t.Log(*i, s)

		}))

		t.Run("IfaceSlice", testFunc(func(t *testing.T) {
			var i = new(any)
			var s = 55
			if err := Copy(t.Context(), i, s); err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if *i != s {
				t.Errorf("expected 'i' to be %d, got %d", s, *i)
			} else {
				t.Logf("i == %d: %d", *i, s)
			}

			s = 5

			t.Log(*i, s)
		}))

	})

	for _, test := range stepTests {
		t.Run(fmt.Sprintf("TestBaseStep-%T", test.src.Interface()), func(t *testing.T) {
			err := rcopy(t.Context(), test.dst, test.src, []func(*State){})
			if err != nil {
				t.Error(err)
				return
			}

			dst := test.dst.Elem().Interface()
			if !reflect.DeepEqual(dst, test.expected) {
				t.Errorf("dst does not match expected: %T(%v) != %T(%v)", dst, dst, test.expected, test.expected)
				return
			}

			test.dst.Elem().Set(reflect.Zero(test.dst.Elem().Type()))

			if !reflect.DeepEqual(dst, test.expected) {
				t.Errorf("dst does not match expected: %T(%v) != %T(%v)", dst, dst, test.expected, test.expected)
				return
			}

			t.Logf("dst == src:\n\t%T(%v) == %T(%v)",
				reflect.Indirect(reflect.ValueOf(dst)).Interface(),
				reflect.Indirect(reflect.ValueOf(dst)).Interface(),
				reflect.Indirect(reflect.ValueOf(test.expected)).Interface(),
				reflect.Indirect(reflect.ValueOf(test.expected)).Interface(),
			)
		})
	}
}

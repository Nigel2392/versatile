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

func newBaseMapTest[K comparable, V any](expect, src map[K]V) baseStepTest {
	dstP := new(make(map[K]V))
	return baseStepTest{
		expected: expect,
		dst:      reflect.ValueOf(dstP),
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

	newBaseStepTest(
		biggerStruct{
			int:     16,
			string:  "yes",
			bool:    false,
			float64: 69.420,
		},
		biggerStruct{
			int:     16,
			string:  "yes",
			bool:    false,
			float64: 69.420,
		},
	),

	newBaseStepTest(
		&biggerStruct{
			int:     16,
			string:  "yes",
			bool:    false,
			float64: 69.420,
		},
		&biggerStruct{
			int:     16,
			string:  "yes",
			bool:    false,
			float64: 69.420,
		},
	),

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

	newBaseMapTest(
		map[string]int{
			"a": 1,
			"b": 2,
		},
		map[string]int{
			"a": 1,
			"b": 2,
		},
	),
}

type myFace interface{ StructMethod() string }

type myStruct struct{ val string }

func (m myStruct) StructMethod() string { return m.val }

type biggerStruct struct {
	int
	string
	bool
	float64
}

func TestSteps(t *testing.T) {
	// FLAGFN := Flag(CF_NOWRAP)
	FLAGFN := Flag(CF_INVALID)

	t.Run("TestPointerCloneFunc", func(t *testing.T) {

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Exception while executing tests: %v: %s", r, string(debug.Stack()))
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
				t.Fatalf("expected no error, got %v", err)
			}

			if *i == nil {
				t.Fatal("expected 'i' not nil, got nil")
			}

			if (*i).StructMethod() != s.val {
				t.Errorf("expected 'i' to be %q, got %q", s.val, (*i).StructMethod())
			} else {
				t.Logf("i == %+v: %+v", *i, s)
			}

			t.Log(*i, s)

		}))

		t.Run("IfaceSlice", testFunc(func(t *testing.T) {
			var i []any
			var s = []int{1, 2, 3}
			if err := Copy(t.Context(), &i, s); err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if len(i) != len(s) {
				t.Fatalf("expected len(i) to be %d, got %d", len(s), len(i))
			} else {
				t.Logf("len(i) == %d: %d", len(i), len(s))
			}

			if i[0] != s[0] || i[1] != s[1] || i[2] != s[2] {
				t.Errorf("%v != %v", i, s)
			}
		}))
	})

	t.Run("ArrayToSlice", func(t *testing.T) {
		t.Run("StrictTypes", func(t *testing.T) {
			var i []int
			var s = [3]int{1, 2, 3}

			if err := Copy(t.Context(), &i, s); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(i) != len(s) || i[0] != s[0] || i[1] != s[1] || i[2] != s[2] {
				t.Errorf("expected %v, got %v", s, i)
			}
		})

		t.Run("TypeToIface", func(t *testing.T) {
			t.Run("Slice", func(t *testing.T) {
				var i []any
				var s = [3]int{1, 2, 3}

				if err := Copy(t.Context(), &i, s); err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if len(i) != len(s) || i[0] != s[0] || i[1] != s[1] || i[2] != s[2] {
					t.Errorf("expected %v, got %v", s, i)
				}
			})

			t.Run("Interface{}", func(t *testing.T) {
				var i any
				var s = [3]int{1, 2, 3}

				if err := Copy(t.Context(), &i, s); err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if !reflect.DeepEqual(i, s) {
					t.Errorf("expected i to be %T(%v), got %T(%v)", s, s, i, i)
				}
			})
		})
	})

	for _, test := range stepTests {
		t.Run(fmt.Sprintf("TestBaseStep-%T", test.src.Interface()), func(t *testing.T) {
			err := rcopy(t.Context(), test.dst, test.src, []func(*State){FLAGFN})
			if err != nil {
				t.Errorf("%v: %+v", err, err)
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

func TestSharedPointers(t *testing.T) {
	type sharedPtrs struct {
		A *int
		B *int
	}

	val := 42
	src := sharedPtrs{
		A: &val,
		B: &val,
	}
	var dst sharedPtrs

	if err := Copy(t.Context(), &dst, src); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dst.A == src.A {
		t.Fatal("dst.A points to the same address as src.A (not deeply cloned)", src.A, dst.A)
	}

	if dst.A != dst.B {
		t.Fatalf("expected dst.A and dst.B to point to the SAME address, got %p and %p", dst.A, dst.B)
	}

	*dst.A = 99
	if *dst.B != 99 {
		t.Errorf("mutating dst.A did not mutate dst.B, shared state is broken!")
	}
}

func TestSharedSlices(t *testing.T) {
	type sharedSlices struct {
		S1 []int
		S2 []int
	}

	backingArray := []int{10, 20, 30, 40, 50}

	src := sharedSlices{
		S1: backingArray,
		S2: backingArray,
	}
	var dst sharedSlices

	if err := Copy(t.Context(), &dst, src); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Log(src)
	t.Log(dst)

	if !reflect.DeepEqual(src.S1, dst.S1) {
		t.Fatalf("Slice mismatch: %v != %v", src.S1, dst.S1)
	}

	dst.S1[0] = 999
	if src.S1[0] == 999 {
		t.Fatalf("modifying clone mutated the original array (not deeply cloned)")
	}

	dst.S1[2] = 777

	if dst.S2[2] != 777 {
		t.Fatalf(`SHARED SLICE STATE BROKEN! 
dst.S1 and dst.S2 do not share the same backing array in the clone.
Expected dst.S2[2] to be 777, got %d`, dst.S2[2])
	} else {
		t.Log("[SUCCESS]: Overlapping slices share the same backing array in the clone!")
	}

	if backingArray[2] == 777 {
		t.Fatalf("expected backing array not to be changed: %v == %v", dst.S1, backingArray)
	} else {
		t.Logf("backingArray: %v\n\t\tdst.S1: %v", backingArray, dst.S1)
	}

}

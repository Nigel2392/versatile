package clone

import (
	"fmt"
	"reflect"
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

var baseStepTests = []baseStepTest{
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
}

func TestBaseStep(t *testing.T) {
	for _, test := range baseStepTests {
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

			t.Logf("dst == src: %T(%v) == %T(%v)",
				reflect.Indirect(reflect.ValueOf(dst)).Interface(),
				reflect.Indirect(reflect.ValueOf(dst)).Interface(),
				reflect.Indirect(reflect.ValueOf(test.expected)).Interface(),
				reflect.Indirect(reflect.ValueOf(test.expected)).Interface(),
			)
		})
	}
}

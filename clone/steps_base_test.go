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
	dstV := reflect.New(reflect.TypeFor[*T]())
	dstV.Elem().Set(reflect.ValueOf(new(T)))
	return baseStepTest{
		expected: expect,
		dst:      dstV,
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

	//	newBaseStepPtrTest(55, int(55)),
	//	newBaseStepPtrTest(55, int8(55)),
	//	newBaseStepPtrTest(55, int16(55)),
	//	newBaseStepPtrTest(55, int32(55)),
	//	newBaseStepPtrTest(55, int64(55)),

	//	newBaseStepPtrTest(55, uint(55)),
	//	newBaseStepPtrTest(55, uint8(55)),
	//	newBaseStepPtrTest(55, uint16(55)),
	//	newBaseStepPtrTest(55, uint32(55)),
	//	newBaseStepPtrTest(55, uint64(55)),

	//	newBaseStepPtrTest(55.55, float32(55.55)),
	//	newBaseStepPtrTest(55.55, float64(55.55)),

	//	newBaseStepPtrTest("my string", "my string"),
	//	newBaseStepPtrTest(true, true),
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
			if dst != test.expected {
				t.Errorf("dst does not match expected: %T(%v) != %T(%v)", dst, dst, test.expected, test.expected)
				return
			}

			test.dst.Elem().Set(reflect.Zero(test.dst.Elem().Type()))

			if dst != test.expected {
				t.Errorf("dst does not match expected: %T(%v) != %T(%v)", dst, dst, test.expected, test.expected)
				return
			}
		})
	}
}

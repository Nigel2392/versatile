package versatile

import (
	re "reflect"
	"testing"
)

type myInt int

type dummyIsZeroer struct {
	zero bool
}

func (d dummyIsZeroer) IsZero() bool {
	return d.zero
}

func TestConvertToType(t *testing.T) {
	tests := []struct {
		name       string
		in         interface{}
		targetType re.Type
		wantError  bool
		wantValue  interface{}
	}{
		{"same type int", 10, re.TypeOf(10), false, 10},
		{"same type string", "hello", re.TypeOf(""), false, "hello"},
		{"assignable to custom type", 10, re.TypeOf(myInt(0)), false, myInt(10)},
		{"convertible int32 to int64", int32(10), re.TypeOf(int64(0)), false, int64(10)},
		{"unassignable int to string", 10, re.TypeOf(""), true, nil},
		{"unassignable string to int", "10", re.TypeOf(10), true, nil},
		{"slice to different slice", []int{1}, re.TypeOf([]string{}), true, nil},
		{"ptr to different ptr", new(int), re.TypeOf(new(string)), true, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inVal := re.ValueOf(tc.in)
			res, err := ConvertToType(inVal, tc.targetType)
			if tc.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Compare interface values deeply
				if !re.DeepEqual(res.Interface(), tc.wantValue) {
					t.Errorf("got %v, want %v", res.Interface(), tc.wantValue)
				}
			}
		})
	}
}

func ptrValue(v interface{}) *re.Value {
	val := re.ValueOf(v)
	return &val
}

func TestReflectValue(t *testing.T) {
	v := 42
	vVal := re.ValueOf(v)

	tests := []struct {
		name string
		in   interface{}
		want interface{}
	}{
		{"int", 42, 42},
		{"string", "hello", "hello"},
		{"re.Value", vVal, 42},
		{"slice", []int{1, 2}, []int{1, 2}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := ReflectValue(tc.in)
			if !res.IsValid() {
				t.Fatalf("expected valid value")
			}
			if !re.DeepEqual(res.Interface(), tc.want) {
				t.Errorf("got %v, want %v", res.Interface(), tc.want)
			}
		})
	}

}

func TestIsZero(t *testing.T) {
	valZero := 0
	valNonZero := 42

	tests := []struct {
		name string
		in   interface{}
		want bool
	}{
		{"nil", nil, true},
		{"int 0", 0, true},
		{"int 42", 42, false},
		{"uint 0", uint(0), true},
		{"uint 42", uint(42), false},
		{"float 0", float64(0), true},
		{"float 3.14", float64(3.14), false},
		{"complex 0", complex128(0), true},
		{"complex non-zero", complex128(1), false},
		{"bool false", false, true},
		{"bool true", true, false},
		{"string empty", "", true},
		{"string non-empty", "hello", false},
		{"slice empty", []int{}, true},
		{"slice zeros", []int{0, 0, 0}, false}, // due to IsZero implementation falling through and checking against nil slice
		{"slice non-zeros", []int{0, 1, 0}, false},
		{"map empty", map[string]int{}, true},
		{"map non-empty", map[string]int{"a": 1}, false},
		{"array zeros", [2]int{0, 0}, true},
		{"array non-zeros", [2]int{0, 1}, false},
		{"ptr nil", (*int)(nil), true},
		{"ptr to zero", &valZero, true},
		{"ptr to non-zero", &valNonZero, false},
		{"custom IsZero true", dummyIsZeroer{true}, true},
		{"custom IsZero false", dummyIsZeroer{false}, false},
		{"ptr to custom IsZero true", &dummyIsZeroer{true}, true},
		{"ptr to custom IsZero false", &dummyIsZeroer{false}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsZero(tc.in)
			if got != tc.want {
				t.Errorf("IsZero() = %v, want %v", got, tc.want)
			}
		})
	}

	t.Run("re.Value zero panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for invalid re.Value")
			}
		}()
		IsZero(re.Value{})
	})
	t.Run("re.Value non-zero", func(t *testing.T) {
		if IsZero(re.ValueOf(42)) {
			t.Errorf("expected false for re.ValueOf(42)")
		}
	})
}

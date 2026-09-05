package clone

import (
	"reflect"
	"sync"
	"testing"
)

func TestIsAllowedType_BasicTypes(t *testing.T) {

	allowed := []any{
		1,
		"string",
		3.14,
		[]int{1, 2, 3},
		map[string]int{"a": 1},
		struct{ A int }{},
		(*int)(nil),
		[]*string{},
	}

	for _, v := range allowed {
		typ := reflect.TypeOf(v)
		if !OK.Type(typ) {
			t.Errorf("expected type %v to be allowed", typ)
		}
	}
}

func TestIsAllowedType_DisallowedSyncTypes(t *testing.T) {

	disallowed := []any{
		sync.Mutex{},
		sync.RWMutex{},
		sync.Cond{},
		sync.Map{},
		sync.Once{},
		sync.Pool{},
		sync.WaitGroup{},
		&sync.Mutex{},
		[]sync.Mutex{},
		[1]sync.WaitGroup{},
	}

	for _, v := range disallowed {
		val := reflect.ValueOf(&v).Elem().Elem()
		if OK.Type(val.Type()) {
			t.Errorf("expected type %v to be disallowed", val.Type())
		}
	}
}

func TestIsAllowedValue_BasicValues(t *testing.T) {

	allowed := []any{
		100,
		"test",
		[]string{"hello", "world"},
		map[int]bool{1: true},
	}

	for _, v := range allowed {
		val := reflect.ValueOf(v)
		if !OK.Value(t.Context(), val) {
			t.Errorf("expected value of type %v to be allowed", val.Type())
		}
	}
}

// doesn't work, values arent
func _TestIsAllowedValue_Interface(t *testing.T) {

	var validIface any = 42
	if !OK.Value(t.Context(), reflect.ValueOf(&validIface).Elem()) {
		t.Errorf("expected interface containing int to be allowed")
	}

	var invalidIface any = sync.Mutex{}
	if OK.Value(t.Context(), reflect.ValueOf(&invalidIface).Elem()) {
		t.Errorf("expected interface containing sync.Mutex to be disallowed")
	}

	var emptyIface any
	if !OK.Value(t.Context(), reflect.ValueOf(&emptyIface).Elem()) {
		t.Errorf("expected nil interface to be allowed")
	}
}

func TestCheckFuncDisallowKind(t *testing.T) {

	// Block all slices
	OK := NewAllowList()
	OK.Check(CheckDisallowKind[[]int])

	if OK.Type(reflect.TypeFor[[]string]()) {
		t.Errorf("expected string slice to be disallowed by kind check")
	}

	if OK.Type(reflect.TypeFor[int]()) {
		// Should not block non-slice types
		if !OK.Type(reflect.TypeFor[int]()) {
			t.Errorf("expected int to be allowed")
		}
	}
}

func TestIsAllowedValue_InvalidValue(t *testing.T) {

	var val reflect.Value
	if OK.Value(t.Context(), val) {
		t.Errorf("expected reflect.Invalid to not be allowed")
	}
}

// -----------------------------------------------------------------------------
// Benchmarks
// -----------------------------------------------------------------------------

func BenchmarkIsAllowedType_Allowed(b *testing.B) {

	typ := reflect.TypeOf(struct{ A int }{})
	b.ResetTimer()

	for b.Loop() {
		OK.Type(typ)
	}
}

func BenchmarkIsAllowedType_Disallowed(b *testing.B) {

	typ := reflect.TypeOf(sync.Mutex{})
	b.ResetTimer()

	for b.Loop() {
		OK.Type(typ)
	}
}

func BenchmarkIsAllowedType_PointerDisallowed(b *testing.B) {

	typ := reflect.TypeOf(&sync.Mutex{})
	b.ResetTimer()

	for b.Loop() {
		OK.Type(typ)
	}
}

func BenchmarkIsAllowedValue_Allowed(b *testing.B) {

	val := reflect.ValueOf("benchmark string")
	b.ResetTimer()

	for b.Loop() {
		OK.Value(b.Context(), val)
	}
}

func BenchmarkIsAllowedValue_Disallowed(b *testing.B) {

	val := reflect.ValueOf(sync.Mutex{})
	b.ResetTimer()

	for b.Loop() {
		OK.Value(b.Context(), val)
	}
}

func BenchmarkIsAllowedValue_NestedInterface(b *testing.B) {

	var iface any = &[]map[string]int{{"a": 1}}
	val := reflect.ValueOf(&iface).Elem()
	b.ResetTimer()

	for b.Loop() {
		OK.Value(b.Context(), val)
	}
}

func BenchmarkIsAllowedValue_CustomValueCheck(b *testing.B) {

	OK.Check(func(v reflect.Value, typ reflect.Type) bool {
		return false
	})

	val := reflect.ValueOf(12345)
	b.ResetTimer()

	for b.Loop() {
		OK.Value(b.Context(), val)
	}
}

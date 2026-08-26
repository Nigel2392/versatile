package versatile_test

import (
	"errors"
	"testing"

	"github.com/Nigel2392/versatile"
)

// 1. Not a function (Source)
func TestUnhappy_NotFuncSource(t *testing.T) {
	_, err := versatile.CastFunc[func()](123)
	mustIsErr(t, err, versatile.ErrNotFunc)
}

// 2. Not a function (Target)
func TestUnhappy_NotFuncTarget(t *testing.T) {
	src := func() {}
	_, err := versatile.CastFunc[int](src)
	mustIsErr(t, err, versatile.ErrNotFunc)
}

// 3. Nil source
func TestUnhappy_NilSource(t *testing.T) {
	var src func() // Typed nil func
	_, err := versatile.CastFunc[func()](src)
	mustIsErr(t, err, versatile.ErrNotFunc)
}

// 4. Argument count mismatch (too many args in source, fixed target)
func TestUnhappy_ArgCountMismatch_TooManySrcArgs(t *testing.T) {
	src := func(a, b int) {}
	_, err := versatile.CastFunc[func(int)](src)
	mustIsErr(t, err, versatile.ErrArgCount)
}

// 5. Argument count mismatch (too few args in source, fixed target)
func TestUnhappy_ArgCountMismatch_TooFewSrcArgs(t *testing.T) {
	src := func(a int) {}
	_, err := versatile.CastFunc[func(int, int)](src)
	mustIsErr(t, err, versatile.ErrArgCount)
}

// 6. Return count mismatch (no rules allow discarding these returns)
func TestUnhappy_ReturnCountMismatch(t *testing.T) {
	src := func() (int, int) { return 1, 2 }
	_, err := versatile.CastFunc[func() int](src)
	mustIsErr(t, err, versatile.ErrReturnCount)
}

// 7. Return type mismatch (incompatible types causes fallback to ErrReturnCount in RCastFunc)
func TestUnhappy_ReturnTypeIncompatible(t *testing.T) {
	src := func() string { return "hello" }
	// string cannot be converted to func()
	_, err := versatile.CastFunc[func() func()](src)
	// Because it doesn't match the exact type or basic convertibility, the switch falls back to default.
	mustIsErr(t, err, versatile.ErrReturnCount)
}

// 8. Method on nil object
func TestUnhappy_MethodNilObject(t *testing.T) {
	_, err := versatile.Method[func()](nil, "Do")
	mustIsErr(t, err, versatile.ErrNilObject)
}

// 9. Method not found
func TestUnhappy_MethodNotFound(t *testing.T) {
	type Dummy struct{}
	_, err := versatile.Method[func()](Dummy{}, "DoesNotExist")
	mustIsErr(t, err, versatile.ErrMethodNotFound)
}

type unhDummy struct{}

func (u *unhDummy) Save(i int) {}

// 10. Method exists but signature is incompatible
func TestUnhappy_MethodIncompatibleSignature(t *testing.T) {
	// Source method requires 1 arg (int). We try to cast to 2 args (string, string).
	_, err := versatile.Method[func(string, string)](&unhDummy{}, "Save")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Method error wraps ErrArgCount
	if !errors.Is(err, versatile.ErrArgCount) {
		t.Fatalf("expected error wrapping ErrArgCount, got %v", err)
	}
}

// 11. Runtime panic on unconvertible argument
func TestUnhappy_RuntimePanic_UnconvertibleArg(t *testing.T) {
	src := func(s string) {}
	// Creation succeeds because arg types are checked at runtime
	out, err := versatile.CastFunc[func(int)](src)
	mustNoErr(t, err)

	mustPanic(t, func() {
		out(123)
	}, "could not convert") // Cannot convert int to string
}

// 12. Runtime panic on variadic extra arg unconvertible
func TestUnhappy_RuntimePanic_VariadicUnconvertible(t *testing.T) {
	src := func(s ...string) {}
	// Target is fixed with multiple args
	out, err := versatile.CastFunc[func(string, int)](src)
	mustNoErr(t, err)

	mustPanic(t, func() {
		out("hello", 123) // 123 cannot convert to string
	}, "could not convert")
}

package clone

import (
	"fmt"
	"runtime/debug"
	"testing"
)

func BenchmarkSteps(b *testing.B) {

	FLAGFN := Flag(CF_NOWRAP | CF_NOVALIDATE | CF_NO_CONVS)
	// FLAGFN := Flag(CF_INVALID)

	stateCtx := SharedStateContext(
		b.Context(), FLAGFN,
	)
	// stateCtx := b.Context()

	b.Run("TestPointerCloneFunc", func(b *testing.B) {

		defer func() {
			if r := recover(); r != nil {
				b.Errorf("Exception while executing tests: %v", r)
			}
		}()

		var i = new(int)
		var s = 55
		for b.Loop() {
			if err := Copy(stateCtx, &i, &s, FLAGFN); err != nil {
				b.Errorf("expected no error, got %v", err)
			}

			*i = 0
		}

		b.Log(*i, s)
	})

	b.Run("TestInterfaceAssignable", func(b *testing.B) {

		var testFunc = (func(fn func(b *testing.B)) func(b *testing.B) {
			return func(b *testing.B) {
				b.Helper()
				defer func() {
					if r := recover(); r != nil {
						b.Errorf("Exception while executing tests: %v: %s", r, string(debug.Stack()))
					}
				}()

				fn(b)
			}
		})

		b.Run("IfacePointer", testFunc(func(b *testing.B) {
			var i = new(any)
			var s = 55

			for b.Loop() {
				if err := Copy(stateCtx, &i, s, FLAGFN); err != nil {
					b.Errorf("expected no error, got %v", err)
				}

				*i = any(nil)
			}

			b.Log(*i, s)
		}))

		b.Run("IfaceWithMethodPointer", testFunc(func(b *testing.B) {
			var i = new(myFace)
			var s = &myStruct{"hello world"}

			for b.Loop() {
				if err := Copy(stateCtx, i, s, FLAGFN); err != nil {
					b.Errorf("expected no error, got %v", err)
				}
				*i = myStruct{}
			}

			b.Log(*i, s)
		}))

		b.Run("IfaceSlice", testFunc(func(b *testing.B) {
			var i []any
			var s = []int{1, 2, 3}

			for b.Loop() {
				if err := Copy(stateCtx, &i, s, FLAGFN); err != nil {
					b.Errorf("expected no error, got %v", err)
				}

				i = nil
			}
		}))

	})

	b.Run("ArrayToSlice", func(b *testing.B) {
		b.Run("StrictTypes", func(b *testing.B) {
			var i []int
			var s = [3]int{1, 2, 3}

			for b.Loop() {
				if err := Copy(stateCtx, &i, s, FLAGFN); err != nil {
					b.Errorf("expected no error, got %v", err)
				}

				i = nil
			}
		})

		b.Run("TypeToIface", func(b *testing.B) {
			b.Run("Slice", func(b *testing.B) {
				var i []any
				var s = [3]int{1, 2, 3}

				for b.Loop() {
					if err := Copy(stateCtx, &i, s, FLAGFN); err != nil {
						b.Errorf("expected no error, got %v", err)
					}

					i = nil
				}
			})

			b.Run("Interface{}", func(b *testing.B) {
				var i any
				var s = [3]int{1, 2, 3}

				for b.Loop() {
					if err := Copy(stateCtx, &i, s, FLAGFN); err != nil {
						b.Errorf("expected no error, got %v", err)
					}

					i = nil
				}
			})
		})
	})

	for _, test := range stepTests {
		b.Run(fmt.Sprintf("TestBaseStep-%T", test.src.Interface()), func(b *testing.B) {
			for b.Loop() {
				err := rcopy(stateCtx, test.dst, test.src, []func(*State){FLAGFN})
				if err != nil {
					b.Error(err)
					return
				}
			}
		})
	}
}

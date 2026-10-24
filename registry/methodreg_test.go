package registry

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/Nigel2392/versatile/internal/develop"
)

// Ensure ErrNotFound is defined if not already declared in another file in package registry.
// var ErrNotFound = errors.New("method not found")

// -----------------------------------------------------------------------------
// 15 Structs scaling from 0 to 255 methods via promoted method blocks
// -----------------------------------------------------------------------------

type (
	Struct01 struct{} // 0 methods
	Struct02 struct { // 1 method
		mBlock1
	}
	Struct03 struct { // 2 methods
		mBlock2
	}
	Struct04 struct { // 3 methods
		mBlock1
		mBlock2
	}
	Struct05 struct { // 4 methods
		mBlock4
	}
	Struct06 struct { // 8 methods
		mBlock8
	}
	Struct07 struct { // 12 methods
		mBlock4
		mBlock8
	}
	Struct08 struct { // 16 methods
		mBlock16
	}
	Struct09 struct { // 24 methods
		mBlock8
		mBlock16
	}
	Struct10 struct { // 32 methods
		mBlock32
	}
	Struct11 struct { // 48 methods
		mBlock16
		mBlock32
	}
	Struct12 struct { // 64 methods
		mBlock64
	}
	Struct13 struct { // 96 methods
		mBlock32
		mBlock64
	}
	Struct14 struct { // 128 methods
		mBlock128
	}
	Struct15 struct { // 255 methods
		mBlock1
		mBlock2
		mBlock4
		mBlock8
		mBlock16
		mBlock32
		mBlock64
		mBlock128
	}
)

func noopValidator(rv reflect.Value, rm *reflect.Method) error {
	if rm.Name == "" {
		return ErrNotFound
	}
	return nil
}

func TestMethodRegistry(t *testing.T) {
	cases := []struct {
		name        string
		rv          reflect.Value
		methodNames []string
	}{
		{"000_Methods", reflect.ValueOf(Struct01{}), []string{"Missing1", "Missing2", "M001"}},
		{"001_Method", reflect.ValueOf(Struct02{}), []string{"Missing1", "Missing2", "M001"}},
		{"002_Methods", reflect.ValueOf(Struct03{}), []string{"Missing1", "Missing2", "M003"}},
		{"003_Methods", reflect.ValueOf(Struct04{}), []string{"Missing1", "Missing2", "M003"}},
		{"004_Methods", reflect.ValueOf(Struct05{}), []string{"Missing1", "Missing2", "M007"}},
		{"008_Methods", reflect.ValueOf(Struct06{}), []string{"Missing1", "Missing2", "M015"}},
		{"012_Methods", reflect.ValueOf(Struct07{}), []string{"Missing1", "Missing2", "M015"}},
		{"016_Methods", reflect.ValueOf(Struct08{}), []string{"Missing1", "Missing2", "M031"}},
		{"024_Methods", reflect.ValueOf(Struct09{}), []string{"Missing1", "Missing2", "M031"}},
		{"032_Methods", reflect.ValueOf(Struct10{}), []string{"Missing1", "Missing2", "M063"}},
		{"048_Methods", reflect.ValueOf(Struct11{}), []string{"Missing1", "Missing2", "M063"}},
		{"064_Methods", reflect.ValueOf(Struct12{}), []string{"Missing1", "Missing2", "M127"}},
		{"096_Methods", reflect.ValueOf(Struct13{}), []string{"Missing1", "Missing2", "M127"}},
		{"128_Methods", reflect.ValueOf(Struct14{}), []string{"Missing1", "Missing2", "M255"}},
		{"255_Methods", reflect.ValueOf(Struct15{}), []string{"Missing1", "Missing2", "M255"}},
	}

	t.Run("Find", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(fmt.Sprintf("%s_Total_%d", tc.name, tc.rv.Type().NumMethod()), func(t *testing.T) {
				reg := NewMethods(0)
				method, err := reg.For(tc.rv, tc.methodNames, noopValidator)
				if err != nil {
					if errors.Is(err, ErrNotFound) && tc.rv.Type().NumMethod() == 0 {
						goto containsCheck
					}

					t.Fatalf("err should be %v, got %s for %v", nil, err, tc.methodNames)
				}

			containsCheck:
				if tc.rv.Type().NumMethod() > 0 && !slices.Contains(tc.methodNames, method.Name) {
					t.Fatalf("method %s not found in method names %v", method.Name, tc.methodNames)
				}
			})
		}
	})

	t.Run("AllMissing", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(fmt.Sprintf("%s_Total_%d", tc.name, tc.rv.Type().NumMethod()), func(t *testing.T) {
				reg := NewMethods(0)
				rt := tc.rv.Type()
				s := make([]string, rt.NumMethod())
				for i := range s {
					s[i] = fmt.Sprintf("Method%d", i)
				}

				_, err := reg.For(tc.rv, s, noopValidator)
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("err should be %s, got %s", ErrNotFound, err)
				}
			})
		}
	})
}

func BenchmarkMethodRegistry(b *testing.B) {
	cases := []struct {
		name        string
		rv          reflect.Value
		methodNames []string
	}{
		{"000_Methods", reflect.ValueOf(Struct01{}), []string{"Missing1", "Missing2", "M001"}},
		{"001_Method", reflect.ValueOf(Struct02{}), []string{"Missing1", "Missing2", "M001"}},
		{"002_Methods", reflect.ValueOf(Struct03{}), []string{"Missing1", "Missing2", "M003"}},
		{"003_Methods", reflect.ValueOf(Struct04{}), []string{"Missing1", "Missing2", "M003"}},
		{"004_Methods", reflect.ValueOf(Struct05{}), []string{"Missing1", "Missing2", "M007"}},
		{"008_Methods", reflect.ValueOf(Struct06{}), []string{"Missing1", "Missing2", "M015"}},
		{"012_Methods", reflect.ValueOf(Struct07{}), []string{"Missing1", "Missing2", "M015"}},
		{"016_Methods", reflect.ValueOf(Struct08{}), []string{"Missing1", "Missing2", "M031"}},
		{"024_Methods", reflect.ValueOf(Struct09{}), []string{"Missing1", "Missing2", "M031"}},
		{"032_Methods", reflect.ValueOf(Struct10{}), []string{"Missing1", "Missing2", "M063"}},
		{"048_Methods", reflect.ValueOf(Struct11{}), []string{"Missing1", "Missing2", "M063"}},
		{"064_Methods", reflect.ValueOf(Struct12{}), []string{"Missing1", "Missing2", "M127"}},
		{"096_Methods", reflect.ValueOf(Struct13{}), []string{"Missing1", "Missing2", "M127"}},
		{"128_Methods", reflect.ValueOf(Struct14{}), []string{"Missing1", "Missing2", "M255"}},
		{"255_Methods", reflect.ValueOf(Struct15{}), []string{"Missing1", "Missing2", "M255"}},
	}

	if !develop.DEV {
		goto skipVanillaReflect
	}

	b.Run("VanillaReflect", func(b *testing.B) {
		for _, tc := range cases {
			b.Run(fmt.Sprintf("%s_Total_%d", tc.name, tc.rv.Type().NumMethod()), func(b *testing.B) {

				b.ReportAllocs()

				for b.Loop() {

					t := tc.rv.Type()
					for _, m := range tc.methodNames {
						mt, ok := t.MethodByName(m)
						if !ok {
							continue
						}

						_ = mt
					}
				}
			})
		}
	})

skipVanillaReflect:
	b.Run("WarmCache", func(b *testing.B) {
		reg := NewMethods(0)
		for _, tc := range cases {
			_, _ = reg.For(tc.rv, tc.methodNames, noopValidator)
		}

		for _, tc := range cases {
			b.Run(fmt.Sprintf("%s_Total_%d", tc.name, tc.rv.Type().NumMethod()), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_, _ = reg.For(tc.rv, tc.methodNames, noopValidator)
				}
			})
		}
	})

	// 2. Cold Cache: Measures uncached reflection lookup + validator + Lock write
	b.Run("ColdCache", func(b *testing.B) {
		for _, tc := range cases {
			b.Run(fmt.Sprintf("%s_Total_%d", tc.name, tc.rv.Type().NumMethod()), func(b *testing.B) {
				b.ReportAllocs()
				reg := NewMethods(0)
				for b.Loop() {
					clear(reg.tCache)
					_, _ = reg.For(tc.rv, tc.methodNames, noopValidator)
				}
			})
		}
	})

	// 2. Cold Cache: Measures uncached reflection lookup + validator + Lock write
	b.Run("AllMissingWarmCache", func(b *testing.B) {
		for _, tc := range cases {
			reg := NewMethods(0)
			for _, tc := range cases {
				_, _ = reg.For(tc.rv, tc.methodNames, noopValidator)
			}

			s := make([]string, 3)
			for i := range s {
				s[i] = fmt.Sprintf("Method%d", i)
			}

			b.Run(fmt.Sprintf("%s_Total_%d", tc.name, tc.rv.Type().NumMethod()), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_, _ = reg.For(tc.rv, s, noopValidator)
				}
			})
		}
	})

	b.Run("AllMissingColdCache", func(b *testing.B) {
		for _, tc := range cases {
			b.Run(fmt.Sprintf("%s_Total_%d", tc.name, tc.rv.Type().NumMethod()), func(b *testing.B) {
				b.ReportAllocs()
				reg := NewMethods(0)
				s := make([]string, 3)
				for i := range s {
					s[i] = fmt.Sprintf("Method%d", i)
				}
				for b.Loop() {
					clear(reg.tCache)
					_, _ = reg.For(tc.rv, s, noopValidator)
				}
			})
		}
	})

	// 3. Concurrent Contention: Measures RWMutex scaling across goroutines
	b.Run("ParallelWarmCache", func(b *testing.B) {
		reg := NewMethods(0)
		for _, tc := range cases {
			_, _ = reg.For(tc.rv, tc.methodNames, noopValidator)
		}

		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				tc := cases[i%len(cases)]
				_, _ = reg.For(tc.rv, tc.methodNames, noopValidator)
				i++
			}
		})
	})
}

// -----------------------------------------------------------------------------
// Method Building Blocks (M001 .. M255)
// -----------------------------------------------------------------------------

type mBlock1 struct{}

func (mBlock1) M001() {}

type mBlock2 struct{}

func (mBlock2) M002() {}
func (mBlock2) M003() {}

type mBlock4 struct{}

func (mBlock4) M004() {}
func (mBlock4) M005() {}
func (mBlock4) M006() {}
func (mBlock4) M007() {}

type mBlock8 struct{}

func (mBlock8) M008() {}
func (mBlock8) M009() {}
func (mBlock8) M010() {}
func (mBlock8) M011() {}
func (mBlock8) M012() {}
func (mBlock8) M013() {}
func (mBlock8) M014() {}
func (mBlock8) M015() {}

type mBlock16 struct{}

func (mBlock16) M016() {}
func (mBlock16) M017() {}
func (mBlock16) M018() {}
func (mBlock16) M019() {}
func (mBlock16) M020() {}
func (mBlock16) M021() {}
func (mBlock16) M022() {}
func (mBlock16) M023() {}
func (mBlock16) M024() {}
func (mBlock16) M025() {}
func (mBlock16) M026() {}
func (mBlock16) M027() {}
func (mBlock16) M028() {}
func (mBlock16) M029() {}
func (mBlock16) M030() {}
func (mBlock16) M031() {}

type mBlock32 struct{}

func (mBlock32) M032() {}
func (mBlock32) M033() {}
func (mBlock32) M034() {}
func (mBlock32) M035() {}
func (mBlock32) M036() {}
func (mBlock32) M037() {}
func (mBlock32) M038() {}
func (mBlock32) M039() {}
func (mBlock32) M040() {}
func (mBlock32) M041() {}
func (mBlock32) M042() {}
func (mBlock32) M043() {}
func (mBlock32) M044() {}
func (mBlock32) M045() {}
func (mBlock32) M046() {}
func (mBlock32) M047() {}
func (mBlock32) M048() {}
func (mBlock32) M049() {}
func (mBlock32) M050() {}
func (mBlock32) M051() {}
func (mBlock32) M052() {}
func (mBlock32) M053() {}
func (mBlock32) M054() {}
func (mBlock32) M055() {}
func (mBlock32) M056() {}
func (mBlock32) M057() {}
func (mBlock32) M058() {}
func (mBlock32) M059() {}
func (mBlock32) M060() {}
func (mBlock32) M061() {}
func (mBlock32) M062() {}
func (mBlock32) M063() {}

type mBlock64 struct{}

func (mBlock64) M064() {}
func (mBlock64) M065() {}
func (mBlock64) M066() {}
func (mBlock64) M067() {}
func (mBlock64) M068() {}
func (mBlock64) M069() {}
func (mBlock64) M070() {}
func (mBlock64) M071() {}
func (mBlock64) M072() {}
func (mBlock64) M073() {}
func (mBlock64) M074() {}
func (mBlock64) M075() {}
func (mBlock64) M076() {}
func (mBlock64) M077() {}
func (mBlock64) M078() {}
func (mBlock64) M079() {}
func (mBlock64) M080() {}
func (mBlock64) M081() {}
func (mBlock64) M082() {}
func (mBlock64) M083() {}
func (mBlock64) M084() {}
func (mBlock64) M085() {}
func (mBlock64) M086() {}
func (mBlock64) M087() {}
func (mBlock64) M088() {}
func (mBlock64) M089() {}
func (mBlock64) M090() {}
func (mBlock64) M091() {}
func (mBlock64) M092() {}
func (mBlock64) M093() {}
func (mBlock64) M094() {}
func (mBlock64) M095() {}
func (mBlock64) M096() {}
func (mBlock64) M097() {}
func (mBlock64) M098() {}
func (mBlock64) M099() {}
func (mBlock64) M100() {}
func (mBlock64) M101() {}
func (mBlock64) M102() {}
func (mBlock64) M103() {}
func (mBlock64) M104() {}
func (mBlock64) M105() {}
func (mBlock64) M106() {}
func (mBlock64) M107() {}
func (mBlock64) M108() {}
func (mBlock64) M109() {}
func (mBlock64) M110() {}
func (mBlock64) M111() {}
func (mBlock64) M112() {}
func (mBlock64) M113() {}
func (mBlock64) M114() {}
func (mBlock64) M115() {}
func (mBlock64) M116() {}
func (mBlock64) M117() {}
func (mBlock64) M118() {}
func (mBlock64) M119() {}
func (mBlock64) M120() {}
func (mBlock64) M121() {}
func (mBlock64) M122() {}
func (mBlock64) M123() {}
func (mBlock64) M124() {}
func (mBlock64) M125() {}
func (mBlock64) M126() {}
func (mBlock64) M127() {}

type mBlock128 struct{}

func (mBlock128) M128() {}
func (mBlock128) M129() {}
func (mBlock128) M130() {}
func (mBlock128) M131() {}
func (mBlock128) M132() {}
func (mBlock128) M133() {}
func (mBlock128) M134() {}
func (mBlock128) M135() {}
func (mBlock128) M136() {}
func (mBlock128) M137() {}
func (mBlock128) M138() {}
func (mBlock128) M139() {}
func (mBlock128) M140() {}
func (mBlock128) M141() {}
func (mBlock128) M142() {}
func (mBlock128) M143() {}
func (mBlock128) M144() {}
func (mBlock128) M145() {}
func (mBlock128) M146() {}
func (mBlock128) M147() {}
func (mBlock128) M148() {}
func (mBlock128) M149() {}
func (mBlock128) M150() {}
func (mBlock128) M151() {}
func (mBlock128) M152() {}
func (mBlock128) M153() {}
func (mBlock128) M154() {}
func (mBlock128) M155() {}
func (mBlock128) M156() {}
func (mBlock128) M157() {}
func (mBlock128) M158() {}
func (mBlock128) M159() {}
func (mBlock128) M160() {}
func (mBlock128) M161() {}
func (mBlock128) M162() {}
func (mBlock128) M163() {}
func (mBlock128) M164() {}
func (mBlock128) M165() {}
func (mBlock128) M166() {}
func (mBlock128) M167() {}
func (mBlock128) M168() {}
func (mBlock128) M169() {}
func (mBlock128) M170() {}
func (mBlock128) M171() {}
func (mBlock128) M172() {}
func (mBlock128) M173() {}
func (mBlock128) M174() {}
func (mBlock128) M175() {}
func (mBlock128) M176() {}
func (mBlock128) M177() {}
func (mBlock128) M178() {}
func (mBlock128) M179() {}
func (mBlock128) M180() {}
func (mBlock128) M181() {}
func (mBlock128) M182() {}
func (mBlock128) M183() {}
func (mBlock128) M184() {}
func (mBlock128) M185() {}
func (mBlock128) M186() {}
func (mBlock128) M187() {}
func (mBlock128) M188() {}
func (mBlock128) M189() {}
func (mBlock128) M190() {}
func (mBlock128) M191() {}
func (mBlock128) M192() {}
func (mBlock128) M193() {}
func (mBlock128) M194() {}
func (mBlock128) M195() {}
func (mBlock128) M196() {}
func (mBlock128) M197() {}
func (mBlock128) M198() {}
func (mBlock128) M199() {}
func (mBlock128) M200() {}
func (mBlock128) M201() {}
func (mBlock128) M202() {}
func (mBlock128) M203() {}
func (mBlock128) M204() {}
func (mBlock128) M205() {}
func (mBlock128) M206() {}
func (mBlock128) M207() {}
func (mBlock128) M208() {}
func (mBlock128) M209() {}
func (mBlock128) M210() {}
func (mBlock128) M211() {}
func (mBlock128) M212() {}
func (mBlock128) M213() {}
func (mBlock128) M214() {}
func (mBlock128) M215() {}
func (mBlock128) M216() {}
func (mBlock128) M217() {}
func (mBlock128) M218() {}
func (mBlock128) M219() {}
func (mBlock128) M220() {}
func (mBlock128) M221() {}
func (mBlock128) M222() {}
func (mBlock128) M223() {}
func (mBlock128) M224() {}
func (mBlock128) M225() {}
func (mBlock128) M226() {}
func (mBlock128) M227() {}
func (mBlock128) M228() {}
func (mBlock128) M229() {}
func (mBlock128) M230() {}
func (mBlock128) M231() {}
func (mBlock128) M232() {}
func (mBlock128) M233() {}
func (mBlock128) M234() {}
func (mBlock128) M235() {}
func (mBlock128) M236() {}
func (mBlock128) M237() {}
func (mBlock128) M238() {}
func (mBlock128) M239() {}
func (mBlock128) M240() {}
func (mBlock128) M241() {}
func (mBlock128) M242() {}
func (mBlock128) M243() {}
func (mBlock128) M244() {}
func (mBlock128) M245() {}
func (mBlock128) M246() {}
func (mBlock128) M247() {}
func (mBlock128) M248() {}
func (mBlock128) M249() {}
func (mBlock128) M250() {}
func (mBlock128) M251() {}
func (mBlock128) M252() {}
func (mBlock128) M253() {}
func (mBlock128) M254() {}
func (mBlock128) M255() {}

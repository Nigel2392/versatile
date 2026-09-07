package scan

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"unsafe"
	"uuid"

	"github.com/Nigel2392/versatile/internal/danger"
)

const (
	BENCH_FLAGS = SF_CONVS
	// BENCH_FLAGS = SF_NONE
	// BENCH_FLAGS = SF_DEFAULT
)

type (
	benchInt         int
	benchInt8        int8
	benchInt16       int16
	benchInt32       int32
	benchInt64       int64
	benchUint        uint
	benchUint8       uint8
	benchUint16      uint16
	benchUint32      uint32
	benchUint64      uint64
	benchUintptr     uintptr
	benchFloat32     float32
	benchFloat64     float64
	benchString      string
	benchBool        bool
	benchBytes       []byte
	benchRunes       []rune
	benchUUID        uuid.UUID
	benchUUIDScanner uuid.UUID
)

func (b *benchUUIDScanner) Value() (any, error) {
	return uuid.UUID(*b), nil
}

func (b *benchUUIDScanner) Scan(src any) error {
	switch s := src.(type) {
	case uuid.UUID:
		*b = benchUUIDScanner(s)
		return nil
	case string:
		p, err := uuid.Parse(s)
		*b = benchUUIDScanner(p)
		return err
	case []byte:
		if len(s) == 16 {
			*b = benchUUIDScanner(s)
			return nil
		}
		p, err := uuid.Parse(string(s))
		*b = benchUUIDScanner(p)
		return err
	}

	return errors.New("type mismatch")
}

func benchDirectAssignment[T any](b *testing.B, src T) {
	b.ResetTimer()
	for b.Loop() {
		var dest any = new(T)
		*(dest.(*T)) = src
	}
}

func benchScan[T any](b *testing.B, src any) {
	var (
		err    error
		wasSet bool
	)
	b.ResetTimer()
	for b.Loop() {
		var dest T
		var destPtr = &dest
		var destV = (*T)(danger.Noescape(unsafe.Pointer(destPtr)))
		wasSet, err = Scan(destV, src, BENCH_FLAGS)
		if err != nil {
			b.Fatalf("error during benchmark %q: %v", b.Name(), err)
		}
		if !wasSet {
			b.Fatalf("value of type %T could not be scanned into %T", src, destV)
		}
	}
}

func benchRScan[T any](b *testing.B, src any) {
	var (
		err    error
		wasSet bool
	)
	b.ResetTimer()
	for b.Loop() {
		var dest T
		var destPtr = &dest
		var destV = reflect.ValueOf((*T)(danger.Noescape(unsafe.Pointer(destPtr))))
		wasSet, err = RScan(destV, src, BENCH_FLAGS)
		if err != nil {
			b.Fatalf("error during benchmark %q: %v", b.Name(), err)
		}
		if !wasSet {
			b.Fatalf("value of type %T could not be scanned into %s", src, destV.Type())
		}
	}
}

func benchInterfaceScan[T any](b *testing.B, src any) {
	var (
		err    error
		wasSet bool
	)
	var eq func(any, any) bool = func(a1, a2 any) bool {
		return a1 == a2
	}
	var rT = reflect.TypeFor[T]()
	if rT.Kind() == reflect.Slice || rT.Kind() == reflect.Map {
		eq = reflect.DeepEqual
	}

	b.ResetTimer()
	for b.Loop() {
		var _dest = new(interface{})
		var dest = (*any)(danger.Noescape(unsafe.Pointer(_dest)))
		wasSet, err = Scan(dest, src, BENCH_FLAGS)
		if err != nil {
			b.Fatalf("error during benchmark %q: %v", b.Name(), err)
		}
		if !wasSet {
			b.Fatalf("value of type %T could not be scanned into %T", src, dest)
		}
		if !eq(*_dest, src) {
			b.Fatalf("expected dst to be %v, got %v", src, *_dest)
		}
	}
}

type benchmark struct {
	benchName string
	fn        func(b *testing.B)
}

type benchmarkGroup struct {
	groupName     string
	onlyIfVerbose bool
	benchmarks    []benchmark
}

func BenchmarkScans(b *testing.B) {
	tests := []benchmarkGroup{
		// direct assignment to interface with underlying value pointer interface((*T)(value))
		{"direct", true, []benchmark{
			{"int", func(b *testing.B) { benchDirectAssignment(b, int(1)) }},
			{"int8", func(b *testing.B) { benchDirectAssignment(b, int8(1)) }},
			{"int16", func(b *testing.B) { benchDirectAssignment(b, int16(1)) }},
			{"int32", func(b *testing.B) { benchDirectAssignment(b, int32(1)) }},
			{"int64", func(b *testing.B) { benchDirectAssignment(b, int64(1)) }},
			{"uint", func(b *testing.B) { benchDirectAssignment(b, uint(1)) }},
			{"uint8", func(b *testing.B) { benchDirectAssignment(b, uint8(1)) }},
			{"uint16", func(b *testing.B) { benchDirectAssignment(b, uint16(1)) }},
			{"uint32", func(b *testing.B) { benchDirectAssignment(b, uint32(1)) }},
			{"uint64", func(b *testing.B) { benchDirectAssignment(b, uint64(1)) }},
			{"uintptr", func(b *testing.B) { benchDirectAssignment(b, uintptr(1)) }},
			{"float32", func(b *testing.B) { benchDirectAssignment(b, float32(1)) }},
			{"float64", func(b *testing.B) { benchDirectAssignment(b, float64(1)) }},
			{"string", func(b *testing.B) { benchDirectAssignment(b, "test") }},
			{"bool", func(b *testing.B) { benchDirectAssignment(b, true) }},
			{"bytes", func(b *testing.B) { benchDirectAssignment(b, []byte("test")) }},
			{"runes", func(b *testing.B) { benchDirectAssignment(b, []rune("test")) }},
			{"uuid", func(b *testing.B) { benchDirectAssignment(b, uuid.Max()) }},
		}},

		// giant switch block of *TYPE with fallback into reflect scans
		{"default", false, []benchmark{
			{"int", func(b *testing.B) { benchScan[int](b, int64(1)) }},
			{"int8", func(b *testing.B) { benchScan[int8](b, int64(1)) }},
			{"int16", func(b *testing.B) { benchScan[int16](b, int64(1)) }},
			{"int32", func(b *testing.B) { benchScan[int32](b, int64(1)) }},
			{"int64", func(b *testing.B) { benchScan[int64](b, int64(1)) }},
			{"uint", func(b *testing.B) { benchScan[uint](b, uint64(1)) }},
			{"uint8", func(b *testing.B) { benchScan[uint8](b, uint64(1)) }},
			{"uint16", func(b *testing.B) { benchScan[uint16](b, uint64(1)) }},
			{"uint32", func(b *testing.B) { benchScan[uint32](b, uint64(1)) }},
			{"uint64", func(b *testing.B) { benchScan[uint64](b, uint64(1)) }},
			{"uintptr", func(b *testing.B) { benchScan[uintptr](b, uint64(1)) }},
			{"float32", func(b *testing.B) { benchScan[float32](b, float64(1)) }},
			{"float64", func(b *testing.B) { benchScan[float64](b, float64(1)) }},
			{"string", func(b *testing.B) { benchScan[string](b, "test") }},
			{"bool", func(b *testing.B) { benchScan[bool](b, true) }},
			{"bytes", func(b *testing.B) { benchScan[[]byte](b, []byte("test")) }},
			{"runes", func(b *testing.B) { benchScan[[]rune](b, []rune("test")) }},
			{"uuid", func(b *testing.B) { benchScan[uuid.UUID](b, uuid.Max()) }},

			{"benchInt_int", func(b *testing.B) { benchScan[benchInt](b, int64(1)) }},
			{"benchInt8_int8", func(b *testing.B) { benchScan[benchInt8](b, int64(1)) }},
			{"benchInt16_int16", func(b *testing.B) { benchScan[benchInt16](b, int64(1)) }},
			{"benchInt32_int32", func(b *testing.B) { benchScan[benchInt32](b, int64(1)) }},
			{"benchInt64_int64", func(b *testing.B) { benchScan[benchInt64](b, int64(1)) }},
			{"benchUint_uint", func(b *testing.B) { benchScan[benchUint](b, uint64(1)) }},
			{"benchUint8_uint8", func(b *testing.B) { benchScan[benchUint8](b, uint64(1)) }},
			{"benchUint16_uint16", func(b *testing.B) { benchScan[benchUint16](b, uint64(1)) }},
			{"benchUint32_uint32", func(b *testing.B) { benchScan[benchUint32](b, uint64(1)) }},
			{"benchUint64_uint64", func(b *testing.B) { benchScan[benchUint64](b, uint64(1)) }},
			{"benchUintptr_uintptr", func(b *testing.B) { benchScan[benchUintptr](b, uint64(1)) }},
			{"benchFloat32_float32", func(b *testing.B) { benchScan[benchFloat32](b, float64(1)) }},
			{"benchFloat64_float64", func(b *testing.B) { benchScan[benchFloat64](b, float64(1)) }},
			{"benchString_string", func(b *testing.B) { benchScan[benchString](b, "test") }},
			{"benchBool_bool", func(b *testing.B) { benchScan[benchBool](b, true) }},
			{"benchBytes_bytes", func(b *testing.B) { benchScan[benchBytes](b, []byte("test")) }},
			{"benchRunes_runes", func(b *testing.B) { benchScan[benchRunes](b, []rune("test")) }},

			{"benchInt", func(b *testing.B) { benchScan[benchInt](b, int64(1)) }},
			{"benchInt8", func(b *testing.B) { benchScan[benchInt8](b, int64(1)) }},
			{"benchInt16", func(b *testing.B) { benchScan[benchInt16](b, int64(1)) }},
			{"benchInt32", func(b *testing.B) { benchScan[benchInt32](b, int64(1)) }},
			{"benchInt64", func(b *testing.B) { benchScan[benchInt64](b, int64(1)) }},
			{"benchUint", func(b *testing.B) { benchScan[benchUint](b, uint64(1)) }},
			{"benchUint8", func(b *testing.B) { benchScan[benchUint8](b, uint64(1)) }},
			{"benchUint16", func(b *testing.B) { benchScan[benchUint16](b, uint64(1)) }},
			{"benchUint32", func(b *testing.B) { benchScan[benchUint32](b, uint64(1)) }},
			{"benchUint64", func(b *testing.B) { benchScan[benchUint64](b, uint64(1)) }},
			{"benchUintptr", func(b *testing.B) { benchScan[benchUintptr](b, uint64(1)) }},
			{"benchFloat32", func(b *testing.B) { benchScan[benchFloat32](b, float64(1)) }},
			{"benchFloat64", func(b *testing.B) { benchScan[benchFloat64](b, float64(1)) }},
			{"benchString", func(b *testing.B) { benchScan[benchString](b, "test") }},
			{"benchBool", func(b *testing.B) { benchScan[benchBool](b, true) }},
			{"benchBytes", func(b *testing.B) { benchScan[benchBytes](b, []byte("test")) }},
			{"benchRunes", func(b *testing.B) { benchScan[benchRunes](b, []rune("test")) }},
			{"benchUUID_Direct", func(b *testing.B) { b.Helper(); benchScan[benchUUID](b, benchUUID(uuid.Max())) }},
			{"benchUUIDScanner_Direct", func(b *testing.B) { b.Helper(); benchScan[benchUUIDScanner](b, benchUUIDScanner(uuid.Max())) }},
			{"benchUUID", func(b *testing.B) { b.Helper(); benchScan[benchUUID](b, uuid.Max()) }},
			{"benchUUIDScanner", func(b *testing.B) { b.Helper(); benchScan[benchUUIDScanner](b, uuid.Max()) }},
			{"benchUUID2", func(b *testing.B) { b.Helper(); benchScan[uuid.UUID](b, benchUUID(uuid.Max())) }},
			{"benchUUIDScanner2", func(b *testing.B) { b.Helper(); benchScan[uuid.UUID](b, benchUUIDScanner(uuid.Max())) }},
		}},

		// reflect scans
		{"reflect", false, []benchmark{
			{"int", func(b *testing.B) { benchRScan[int](b, int64(1)) }},
			{"int8", func(b *testing.B) { benchRScan[int8](b, int64(1)) }},
			{"int16", func(b *testing.B) { benchRScan[int16](b, int64(1)) }},
			{"int32", func(b *testing.B) { benchRScan[int32](b, int64(1)) }},
			{"int64", func(b *testing.B) { benchRScan[int64](b, int64(1)) }},
			{"uint", func(b *testing.B) { benchRScan[uint](b, uint64(1)) }},
			{"uint8", func(b *testing.B) { benchRScan[uint8](b, uint64(1)) }},
			{"uint16", func(b *testing.B) { benchRScan[uint16](b, uint64(1)) }},
			{"uint32", func(b *testing.B) { benchRScan[uint32](b, uint64(1)) }},
			{"uint64", func(b *testing.B) { benchRScan[uint64](b, uint64(1)) }},
			{"uintptr", func(b *testing.B) { benchRScan[uintptr](b, uint64(1)) }},
			{"float32", func(b *testing.B) { benchRScan[float32](b, float64(1)) }},
			{"float64", func(b *testing.B) { benchRScan[float64](b, float64(1)) }},
			{"string", func(b *testing.B) { benchRScan[string](b, "test") }},
			{"bool", func(b *testing.B) { benchRScan[bool](b, true) }},
			{"bytes", func(b *testing.B) { benchRScan[[]byte](b, []byte("test")) }},
			{"runes", func(b *testing.B) { benchRScan[[]rune](b, []rune("test")) }},
			{"uuid", func(b *testing.B) { benchRScan[uuid.UUID](b, uuid.Max()) }},

			{"benchInt_int", func(b *testing.B) { benchRScan[benchInt](b, int64(1)) }},
			{"benchInt8_int8", func(b *testing.B) { benchRScan[benchInt8](b, int64(1)) }},
			{"benchInt16_int16", func(b *testing.B) { benchRScan[benchInt16](b, int64(1)) }},
			{"benchInt32_int32", func(b *testing.B) { benchRScan[benchInt32](b, int64(1)) }},
			{"benchInt64_int64", func(b *testing.B) { benchRScan[benchInt64](b, int64(1)) }},
			{"benchUint_uint", func(b *testing.B) { benchRScan[benchUint](b, uint64(1)) }},
			{"benchUint8_uint8", func(b *testing.B) { benchRScan[benchUint8](b, uint64(1)) }},
			{"benchUint16_uint16", func(b *testing.B) { benchRScan[benchUint16](b, uint64(1)) }},
			{"benchUint32_uint32", func(b *testing.B) { benchRScan[benchUint32](b, uint64(1)) }},
			{"benchUint64_uint64", func(b *testing.B) { benchRScan[benchUint64](b, uint64(1)) }},
			{"benchUintptr_uintptr", func(b *testing.B) { benchRScan[benchUintptr](b, uint64(1)) }},
			{"benchFloat32_float32", func(b *testing.B) { benchRScan[benchFloat32](b, float64(1)) }},
			{"benchFloat64_float64", func(b *testing.B) { benchRScan[benchFloat64](b, float64(1)) }},
			{"benchString_string", func(b *testing.B) { benchRScan[benchString](b, "test") }},
			{"benchBool_bool", func(b *testing.B) { benchRScan[benchBool](b, true) }},
			{"benchBytes_bytes", func(b *testing.B) { benchRScan[benchBytes](b, []byte("test")) }},
			{"benchRunes_runes", func(b *testing.B) { benchRScan[benchRunes](b, []rune("test")) }},

			{"benchInt", func(b *testing.B) { benchRScan[benchInt](b, int64(1)) }},
			{"benchInt8", func(b *testing.B) { benchRScan[benchInt8](b, int64(1)) }},
			{"benchInt16", func(b *testing.B) { benchRScan[benchInt16](b, int64(1)) }},
			{"benchInt32", func(b *testing.B) { benchRScan[benchInt32](b, int64(1)) }},
			{"benchInt64", func(b *testing.B) { benchRScan[benchInt64](b, int64(1)) }},
			{"benchUint", func(b *testing.B) { benchRScan[benchUint](b, uint64(1)) }},
			{"benchUint8", func(b *testing.B) { benchRScan[benchUint8](b, uint64(1)) }},
			{"benchUint16", func(b *testing.B) { benchRScan[benchUint16](b, uint64(1)) }},
			{"benchUint32", func(b *testing.B) { benchRScan[benchUint32](b, uint64(1)) }},
			{"benchUint64", func(b *testing.B) { benchRScan[benchUint64](b, uint64(1)) }},
			{"benchUintptr", func(b *testing.B) { benchRScan[benchUintptr](b, uint64(1)) }},
			{"benchFloat32", func(b *testing.B) { benchRScan[benchFloat32](b, float64(1)) }},
			{"benchFloat64", func(b *testing.B) { benchRScan[benchFloat64](b, float64(1)) }},
			{"benchString", func(b *testing.B) { benchRScan[benchString](b, "test") }},
			{"benchBool", func(b *testing.B) { benchRScan[benchBool](b, true) }},
			{"benchBytes", func(b *testing.B) { benchRScan[benchBytes](b, []byte("test")) }},
			{"benchRunes", func(b *testing.B) { benchRScan[benchRunes](b, []rune("test")) }},
			{"benchUUID_Direct", func(b *testing.B) { b.Helper(); benchRScan[benchUUID](b, benchUUID(uuid.Max())) }},
			{"benchUUIDScanner_Direct", func(b *testing.B) { b.Helper(); benchRScan[benchUUIDScanner](b, benchUUIDScanner(uuid.Max())) }},
			{"benchUUID", func(b *testing.B) { b.Helper(); benchRScan[benchUUID](b, uuid.Max()) }},
			{"benchUUIDScanner", func(b *testing.B) { b.Helper(); benchRScan[benchUUIDScanner](b, uuid.Max()) }},
			{"benchUUID2", func(b *testing.B) { b.Helper(); benchRScan[uuid.UUID](b, benchUUID(uuid.Max())) }},
			{"benchUUIDScanner2", func(b *testing.B) { b.Helper(); benchRScan[uuid.UUID](b, benchUUIDScanner(uuid.Max())) }},
		}},

		// scan into *interface{}
		{"interfaces", false, []benchmark{
			{"int", func(b *testing.B) { benchInterfaceScan[int](b, int(1)) }},
			{"int8", func(b *testing.B) { benchInterfaceScan[int8](b, int8(1)) }},
			{"int16", func(b *testing.B) { benchInterfaceScan[int16](b, int16(1)) }},
			{"int32", func(b *testing.B) { benchInterfaceScan[int32](b, int32(1)) }},
			{"int64", func(b *testing.B) { benchInterfaceScan[int64](b, int64(1)) }},
			{"uint", func(b *testing.B) { benchInterfaceScan[uint](b, uint(1)) }},
			{"uint8", func(b *testing.B) { benchInterfaceScan[uint8](b, uint8(1)) }},
			{"uint16", func(b *testing.B) { benchInterfaceScan[uint16](b, uint16(1)) }},
			{"uint32", func(b *testing.B) { benchInterfaceScan[uint32](b, uint32(1)) }},
			{"uint64", func(b *testing.B) { benchInterfaceScan[uint64](b, uint64(1)) }},
			{"uintptr", func(b *testing.B) { benchInterfaceScan[uintptr](b, uintptr(1)) }},
			{"float32", func(b *testing.B) { benchInterfaceScan[float32](b, float32(1)) }},
			{"float64", func(b *testing.B) { benchInterfaceScan[float64](b, float64(1)) }},
			{"string", func(b *testing.B) { benchInterfaceScan[string](b, "test") }},
			{"bool", func(b *testing.B) { benchInterfaceScan[bool](b, true) }},
			{"bytes", func(b *testing.B) { benchInterfaceScan[[]byte](b, []byte("test")) }},
			{"runes", func(b *testing.B) { benchInterfaceScan[[]rune](b, []rune("test")) }},
			{"uuid", func(b *testing.B) { benchInterfaceScan[uuid.UUID](b, uuid.Max()) }},
		}},
	}

	for _, t := range tests {
		if t.onlyIfVerbose && !testing.Verbose() {
			continue
		}

		b.Run(fmt.Sprintf("%s", t.groupName), func(b *testing.B) {
			for _, bench := range t.benchmarks {
				b.Run(bench.benchName, bench.fn)
			}
		})
	}
}

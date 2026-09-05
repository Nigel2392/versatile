package versatile

import (
	"reflect"
	"testing"
	"uuid"
)

type benchInt int
type benchInt8 int8
type benchInt16 int16
type benchInt32 int32
type benchInt64 int64
type benchUint uint
type benchUint8 uint8
type benchUint16 uint16
type benchUint32 uint32
type benchUint64 uint64
type benchUintptr uintptr
type benchFloat32 float32
type benchFloat64 float64
type benchString string
type benchBool bool
type benchBytes []byte
type benchRunes []rune

func benchScanTo[T any](b *testing.B, src any) {
	var dest T
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScanTo(&dest, src, SF_NONE)
	}
}

func benchRScanTo[T any](b *testing.B, src any) {
	var dest T
	destV := reflect.ValueOf(&dest)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RScanTo(destV, src, SF_NONE)
	}
}

func BenchmarkScanTo(b *testing.B) {
	b.Run("int", func(b *testing.B) { benchScanTo[int](b, int64(1)) })
	b.Run("int8", func(b *testing.B) { benchScanTo[int8](b, int64(1)) })
	b.Run("int16", func(b *testing.B) { benchScanTo[int16](b, int64(1)) })
	b.Run("int32", func(b *testing.B) { benchScanTo[int32](b, int64(1)) })
	b.Run("int64", func(b *testing.B) { benchScanTo[int64](b, int64(1)) })
	b.Run("uint", func(b *testing.B) { benchScanTo[uint](b, uint64(1)) })
	b.Run("uint8", func(b *testing.B) { benchScanTo[uint8](b, uint64(1)) })
	b.Run("uint16", func(b *testing.B) { benchScanTo[uint16](b, uint64(1)) })
	b.Run("uint32", func(b *testing.B) { benchScanTo[uint32](b, uint64(1)) })
	b.Run("uint64", func(b *testing.B) { benchScanTo[uint64](b, uint64(1)) })
	b.Run("uintptr", func(b *testing.B) { benchScanTo[uintptr](b, uint64(1)) })
	b.Run("float32", func(b *testing.B) { benchScanTo[float32](b, float64(1)) })
	b.Run("float64", func(b *testing.B) { benchScanTo[float64](b, float64(1)) })
	b.Run("string", func(b *testing.B) { benchScanTo[string](b, "test") })
	b.Run("bool", func(b *testing.B) { benchScanTo[bool](b, true) })
	b.Run("bytes", func(b *testing.B) { benchScanTo[[]byte](b, []byte("test")) })
	b.Run("runes", func(b *testing.B) { benchScanTo[[]rune](b, []rune("test")) })
	b.Run("uuid", func(b *testing.B) { benchScanTo[uuid.UUID](b, uuid.Max()) })

	b.Run("benchInt", func(b *testing.B) { benchScanTo[benchInt](b, int64(1)) })
	b.Run("benchInt8", func(b *testing.B) { benchScanTo[benchInt8](b, int64(1)) })
	b.Run("benchInt16", func(b *testing.B) { benchScanTo[benchInt16](b, int64(1)) })
	b.Run("benchInt32", func(b *testing.B) { benchScanTo[benchInt32](b, int64(1)) })
	b.Run("benchInt64", func(b *testing.B) { benchScanTo[benchInt64](b, int64(1)) })
	b.Run("benchUint", func(b *testing.B) { benchScanTo[benchUint](b, uint64(1)) })
	b.Run("benchUint8", func(b *testing.B) { benchScanTo[benchUint8](b, uint64(1)) })
	b.Run("benchUint16", func(b *testing.B) { benchScanTo[benchUint16](b, uint64(1)) })
	b.Run("benchUint32", func(b *testing.B) { benchScanTo[benchUint32](b, uint64(1)) })
	b.Run("benchUint64", func(b *testing.B) { benchScanTo[benchUint64](b, uint64(1)) })
	b.Run("benchUintptr", func(b *testing.B) { benchScanTo[benchUintptr](b, uint64(1)) })
	b.Run("benchFloat32", func(b *testing.B) { benchScanTo[benchFloat32](b, float64(1)) })
	b.Run("benchFloat64", func(b *testing.B) { benchScanTo[benchFloat64](b, float64(1)) })
	b.Run("benchString", func(b *testing.B) { benchScanTo[benchString](b, "test") })
	b.Run("benchBool", func(b *testing.B) { benchScanTo[benchBool](b, true) })
	b.Run("benchBytes", func(b *testing.B) { benchScanTo[benchBytes](b, []byte("test")) })
	b.Run("benchRunes", func(b *testing.B) { benchScanTo[benchRunes](b, []rune("test")) })
}

func BenchmarkRScanTo(b *testing.B) {
	b.Run("int", func(b *testing.B) { benchRScanTo[int](b, int64(1)) })
	b.Run("int8", func(b *testing.B) { benchRScanTo[int8](b, int64(1)) })
	b.Run("int16", func(b *testing.B) { benchRScanTo[int16](b, int64(1)) })
	b.Run("int32", func(b *testing.B) { benchRScanTo[int32](b, int64(1)) })
	b.Run("int64", func(b *testing.B) { benchRScanTo[int64](b, int64(1)) })
	b.Run("uint", func(b *testing.B) { benchRScanTo[uint](b, uint64(1)) })
	b.Run("uint8", func(b *testing.B) { benchRScanTo[uint8](b, uint64(1)) })
	b.Run("uint16", func(b *testing.B) { benchRScanTo[uint16](b, uint64(1)) })
	b.Run("uint32", func(b *testing.B) { benchRScanTo[uint32](b, uint64(1)) })
	b.Run("uint64", func(b *testing.B) { benchRScanTo[uint64](b, uint64(1)) })
	b.Run("uintptr", func(b *testing.B) { benchRScanTo[uintptr](b, uint64(1)) })
	b.Run("float32", func(b *testing.B) { benchRScanTo[float32](b, float64(1)) })
	b.Run("float64", func(b *testing.B) { benchRScanTo[float64](b, float64(1)) })
	b.Run("string", func(b *testing.B) { benchRScanTo[string](b, "test") })
	b.Run("bool", func(b *testing.B) { benchRScanTo[bool](b, true) })
	b.Run("bytes", func(b *testing.B) { benchRScanTo[[]byte](b, []byte("test")) })
	b.Run("runes", func(b *testing.B) { benchRScanTo[[]rune](b, []rune("test")) })
	b.Run("uuid", func(b *testing.B) { benchScanTo[uuid.UUID](b, uuid.Max()) })

	b.Run("benchInt", func(b *testing.B) { benchRScanTo[benchInt](b, int64(1)) })
	b.Run("benchInt8", func(b *testing.B) { benchRScanTo[benchInt8](b, int64(1)) })
	b.Run("benchInt16", func(b *testing.B) { benchRScanTo[benchInt16](b, int64(1)) })
	b.Run("benchInt32", func(b *testing.B) { benchRScanTo[benchInt32](b, int64(1)) })
	b.Run("benchInt64", func(b *testing.B) { benchRScanTo[benchInt64](b, int64(1)) })
	b.Run("benchUint", func(b *testing.B) { benchRScanTo[benchUint](b, uint64(1)) })
	b.Run("benchUint8", func(b *testing.B) { benchRScanTo[benchUint8](b, uint64(1)) })
	b.Run("benchUint16", func(b *testing.B) { benchRScanTo[benchUint16](b, uint64(1)) })
	b.Run("benchUint32", func(b *testing.B) { benchRScanTo[benchUint32](b, uint64(1)) })
	b.Run("benchUint64", func(b *testing.B) { benchRScanTo[benchUint64](b, uint64(1)) })
	b.Run("benchUintptr", func(b *testing.B) { benchRScanTo[benchUintptr](b, uint64(1)) })
	b.Run("benchFloat32", func(b *testing.B) { benchRScanTo[benchFloat32](b, float64(1)) })
	b.Run("benchFloat64", func(b *testing.B) { benchRScanTo[benchFloat64](b, float64(1)) })
	b.Run("benchString", func(b *testing.B) { benchRScanTo[benchString](b, "test") })
	b.Run("benchBool", func(b *testing.B) { benchRScanTo[benchBool](b, true) })
	b.Run("benchBytes", func(b *testing.B) { benchRScanTo[benchBytes](b, []byte("test")) })
	b.Run("benchRunes", func(b *testing.B) { benchRScanTo[benchRunes](b, []rune("test")) })
}

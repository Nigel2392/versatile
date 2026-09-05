package scan

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
type benchUUID uuid.UUID

func benchScan[T any](b *testing.B, src any) {
	var dest T
	b.ResetTimer()
	for b.Loop() {
		Scan(&dest, src, SF_NONE)
	}
}

func benchRScan[T any](b *testing.B, src any) {
	var dest T
	destV := reflect.ValueOf(&dest)
	b.ResetTimer()
	for b.Loop() {
		RScan(destV, src, SF_NONE)
	}
}

func BenchmarkScan(b *testing.B) {
	b.Run("int", func(b *testing.B) { benchScan[int](b, int64(1)) })
	b.Run("int8", func(b *testing.B) { benchScan[int8](b, int64(1)) })
	b.Run("int16", func(b *testing.B) { benchScan[int16](b, int64(1)) })
	b.Run("int32", func(b *testing.B) { benchScan[int32](b, int64(1)) })
	b.Run("int64", func(b *testing.B) { benchScan[int64](b, int64(1)) })
	b.Run("uint", func(b *testing.B) { benchScan[uint](b, uint64(1)) })
	b.Run("uint8", func(b *testing.B) { benchScan[uint8](b, uint64(1)) })
	b.Run("uint16", func(b *testing.B) { benchScan[uint16](b, uint64(1)) })
	b.Run("uint32", func(b *testing.B) { benchScan[uint32](b, uint64(1)) })
	b.Run("uint64", func(b *testing.B) { benchScan[uint64](b, uint64(1)) })
	b.Run("uintptr", func(b *testing.B) { benchScan[uintptr](b, uint64(1)) })
	b.Run("float32", func(b *testing.B) { benchScan[float32](b, float64(1)) })
	b.Run("float64", func(b *testing.B) { benchScan[float64](b, float64(1)) })
	b.Run("string", func(b *testing.B) { benchScan[string](b, "test") })
	b.Run("bool", func(b *testing.B) { benchScan[bool](b, true) })
	b.Run("bytes", func(b *testing.B) { benchScan[[]byte](b, []byte("test")) })
	b.Run("runes", func(b *testing.B) { benchScan[[]rune](b, []rune("test")) })
	b.Run("uuid", func(b *testing.B) { benchScan[uuid.UUID](b, uuid.Max()) })

	b.Run("benchInt", func(b *testing.B) { benchScan[benchInt](b, int64(1)) })
	b.Run("benchInt8", func(b *testing.B) { benchScan[benchInt8](b, int64(1)) })
	b.Run("benchInt16", func(b *testing.B) { benchScan[benchInt16](b, int64(1)) })
	b.Run("benchInt32", func(b *testing.B) { benchScan[benchInt32](b, int64(1)) })
	b.Run("benchInt64", func(b *testing.B) { benchScan[benchInt64](b, int64(1)) })
	b.Run("benchUint", func(b *testing.B) { benchScan[benchUint](b, uint64(1)) })
	b.Run("benchUint8", func(b *testing.B) { benchScan[benchUint8](b, uint64(1)) })
	b.Run("benchUint16", func(b *testing.B) { benchScan[benchUint16](b, uint64(1)) })
	b.Run("benchUint32", func(b *testing.B) { benchScan[benchUint32](b, uint64(1)) })
	b.Run("benchUint64", func(b *testing.B) { benchScan[benchUint64](b, uint64(1)) })
	b.Run("benchUintptr", func(b *testing.B) { benchScan[benchUintptr](b, uint64(1)) })
	b.Run("benchFloat32", func(b *testing.B) { benchScan[benchFloat32](b, float64(1)) })
	b.Run("benchFloat64", func(b *testing.B) { benchScan[benchFloat64](b, float64(1)) })
	b.Run("benchString", func(b *testing.B) { benchScan[benchString](b, "test") })
	b.Run("benchBool", func(b *testing.B) { benchScan[benchBool](b, true) })
	b.Run("benchBytes", func(b *testing.B) { benchScan[benchBytes](b, []byte("test")) })
	b.Run("benchRunes", func(b *testing.B) { benchScan[benchRunes](b, []rune("test")) })
	b.Run("benchUUID", func(b *testing.B) { benchScan[benchUUID](b, uuid.Max()) })
}

func BenchmarkRScan(b *testing.B) {
	b.Run("int", func(b *testing.B) { benchRScan[int](b, int64(1)) })
	b.Run("int8", func(b *testing.B) { benchRScan[int8](b, int64(1)) })
	b.Run("int16", func(b *testing.B) { benchRScan[int16](b, int64(1)) })
	b.Run("int32", func(b *testing.B) { benchRScan[int32](b, int64(1)) })
	b.Run("int64", func(b *testing.B) { benchRScan[int64](b, int64(1)) })
	b.Run("uint", func(b *testing.B) { benchRScan[uint](b, uint64(1)) })
	b.Run("uint8", func(b *testing.B) { benchRScan[uint8](b, uint64(1)) })
	b.Run("uint16", func(b *testing.B) { benchRScan[uint16](b, uint64(1)) })
	b.Run("uint32", func(b *testing.B) { benchRScan[uint32](b, uint64(1)) })
	b.Run("uint64", func(b *testing.B) { benchRScan[uint64](b, uint64(1)) })
	b.Run("uintptr", func(b *testing.B) { benchRScan[uintptr](b, uint64(1)) })
	b.Run("float32", func(b *testing.B) { benchRScan[float32](b, float64(1)) })
	b.Run("float64", func(b *testing.B) { benchRScan[float64](b, float64(1)) })
	b.Run("string", func(b *testing.B) { benchRScan[string](b, "test") })
	b.Run("bool", func(b *testing.B) { benchRScan[bool](b, true) })
	b.Run("bytes", func(b *testing.B) { benchRScan[[]byte](b, []byte("test")) })
	b.Run("runes", func(b *testing.B) { benchRScan[[]rune](b, []rune("test")) })
	b.Run("uuid", func(b *testing.B) { benchRScan[uuid.UUID](b, uuid.Max()) })

	b.Run("benchInt", func(b *testing.B) { benchRScan[benchInt](b, int64(1)) })
	b.Run("benchInt8", func(b *testing.B) { benchRScan[benchInt8](b, int64(1)) })
	b.Run("benchInt16", func(b *testing.B) { benchRScan[benchInt16](b, int64(1)) })
	b.Run("benchInt32", func(b *testing.B) { benchRScan[benchInt32](b, int64(1)) })
	b.Run("benchInt64", func(b *testing.B) { benchRScan[benchInt64](b, int64(1)) })
	b.Run("benchUint", func(b *testing.B) { benchRScan[benchUint](b, uint64(1)) })
	b.Run("benchUint8", func(b *testing.B) { benchRScan[benchUint8](b, uint64(1)) })
	b.Run("benchUint16", func(b *testing.B) { benchRScan[benchUint16](b, uint64(1)) })
	b.Run("benchUint32", func(b *testing.B) { benchRScan[benchUint32](b, uint64(1)) })
	b.Run("benchUint64", func(b *testing.B) { benchRScan[benchUint64](b, uint64(1)) })
	b.Run("benchUintptr", func(b *testing.B) { benchRScan[benchUintptr](b, uint64(1)) })
	b.Run("benchFloat32", func(b *testing.B) { benchRScan[benchFloat32](b, float64(1)) })
	b.Run("benchFloat64", func(b *testing.B) { benchRScan[benchFloat64](b, float64(1)) })
	b.Run("benchString", func(b *testing.B) { benchRScan[benchString](b, "test") })
	b.Run("benchBool", func(b *testing.B) { benchRScan[benchBool](b, true) })
	b.Run("benchBytes", func(b *testing.B) { benchRScan[benchBytes](b, []byte("test")) })
	b.Run("benchRunes", func(b *testing.B) { benchRScan[benchRunes](b, []rune("test")) })
	b.Run("benchUUID", func(b *testing.B) { benchRScan[benchUUID](b, uuid.Max()) })
}

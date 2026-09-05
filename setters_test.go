package versatile

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

type testScanner struct {
	val any
}

func (s *testScanner) Scan(src any) error {
	s.val = src
	return nil
}

type driverValuer struct {
	val any
}

func (d driverValuer) Value() (driver.Value, error) {
	return d.val, nil
}

func TestConvertToUniformType(t *testing.T) {
	cases := []struct {
		in  any
		out any
	}{
		{int(1), int64(1)},
		{int8(1), int64(1)},
		{int16(1), int64(1)},
		{int32(1), int64(1)},
		{int64(1), int64(1)},
		{uint(1), uint64(1)},
		{uint8(1), uint64(1)},
		{uint16(1), uint64(1)},
		{uint32(1), uint64(1)},
		{uint64(1), uint64(1)},
		{uintptr(1), uint64(1)},
		{float32(1.5), float64(1.5)},
		{float64(1.5), float64(1.5)},
		{complex64(1 + 2i), complex128(1 + 2i)},
		{complex128(1 + 2i), complex128(1 + 2i)},
		{[]byte("test"), []byte("test")},
		{[]rune("test"), []rune("test")},
		{"test", "test"},
		{true, true},
		{customInt(1), int64(1)},
		{customString("test"), "test"},
	}

	for _, c := range cases {
		out := ConvertToUniformType(c.in)
		if !reflect.DeepEqual(out, c.out) {
			t.Errorf("ConvertToUniformType(%T(%v)) == %T(%v), want %T(%v)", c.in, c.in, out, out, c.out, c.out)
		}
	}
}

func TestScanTo_BasicTypes(t *testing.T) {
	runesDest := []rune("old")
	bytesDest := []byte("old")
	strDest := "old"

	cases := []struct {
		src    any
		dst    any
		expect any
	}{
		{int64(1), new(int), int(1)},
		{int64(1), new(int8), int8(1)},
		{int64(1), new(int16), int16(1)},
		{int64(1), new(int32), int32(1)},
		{int64(1), new(int64), int64(1)},
		{int64(1), new(uint), uint(1)},
		{int64(1), new(uint8), uint8(1)},
		{int64(1), new(uint16), uint16(1)},
		{int64(1), new(uint32), uint32(1)},
		{int64(1), new(uint64), uint64(1)},
		{int64(1), new(uintptr), uintptr(1)},
		{uint64(1), new(int), int(1)},
		{uint64(1), new(int8), int8(1)},
		{uint64(1), new(int16), int16(1)},
		{uint64(1), new(int32), int32(1)},
		{uint64(1), new(int64), int64(1)},
		{uint64(1), new(uint), uint(1)},
		{uint64(1), new(uint8), uint8(1)},
		{uint64(1), new(uint16), uint16(1)},
		{uint64(1), new(uint32), uint32(1)},
		{uint64(1), new(uint64), uint64(1)},
		{uint64(1), new(uintptr), uintptr(1)},
		{float64(1.5), new(float32), float32(1.5)},
		{float64(1.5), new(float64), float64(1.5)},

		{"1", new(int), int(1)},
		{"1", new(int8), int8(1)},
		{"1", new(int16), int16(1)},
		{"1", new(int32), int32(1)},
		{"1", new(int64), int64(1)},
		{"1", new(uint), uint(1)},
		{"1", new(uint8), uint8(1)},
		{"1", new(uint16), uint16(1)},
		{"1", new(uint32), uint32(1)},
		{"1", new(uint64), uint64(1)},
		{"1", new(uintptr), uintptr(1)},
		{"1.5", new(float32), float32(1.5)},
		{"1.5", new(float64), float64(1.5)},

		{[]byte("test"), &bytesDest, []byte("test")},
		{"test", &bytesDest, []byte("test")},
		{[]rune("test"), &bytesDest, []byte("test")},

		{[]byte("test"), &runesDest, []rune("test")},
		{"test", &runesDest, []rune("test")},
		{[]rune("test"), &runesDest, []rune("test")},

		{[]byte("test"), &strDest, "test"},
		{"test", &strDest, "test"},
		{[]rune("test"), &strDest, "test"},

		{true, new(bool), true},
		{int64(1), new(bool), true},
		{int64(0), new(bool), false},
		{uint64(1), new(bool), true},
		{uint64(0), new(bool), false},
		{true, new(bool), true},
		{false, new(bool), false},
	}

	for _, c := range cases {
		ok, err := ScanTo(&c.dst, c.src, SF_DEFAULT)
		if !ok || err != nil {
			t.Errorf("ScanTo(%T, %T) failed: ok=%v, err=%v", c.dst, c.src, ok, err)
		}
		val := reflect.ValueOf(c.dst).Elem().Interface()
		if !reflect.DeepEqual(val, c.expect) {
			t.Errorf("ScanTo(%T, %T) == %v, want %v (%T %T)", c.dst, c.src, val, c.expect, val, c.expect)
		}
	}
}

func TestScanTo_SpecialCases(t *testing.T) {
	// any pointer
	var anyDest any
	ok, err := ScanTo(&anyDest, "test", SF_DEFAULT)
	if !ok || err != nil || anyDest != "test" {
		t.Errorf("ScanTo(*any) failed")
	}

	// SQL Scanner
	var ts testScanner
	ok, err = ScanTo(&ts, "test", SF_DEFAULT)
	if !ok || err != nil || ts.val != "test" {
		t.Errorf("ScanTo(sql.Scanner) failed")
	}

	// driver.Valuer
	var intDest int
	dv := driverValuer{val: int64(42)}
	ok, err = ScanTo(&intDest, dv, SF_DEFAULT)
	if !ok || err != nil || intDest != 42 {
		t.Errorf("ScanTo from driver.Valuer failed")
	}

	// nil source
	intDest = 42
	ok, err = ScanTo(&intDest, nil, SF_DEFAULT)
	if !ok || err != nil || intDest != 0 {
		t.Errorf("ScanTo with nil source failed")
	}
}

func TestRScanTo_BasicTypes(t *testing.T) {
	runesDest := []rune("old")
	bytesDest := []byte("old")
	strDest := "old"

	cases := []struct {
		src    any
		dst    any
		expect any
	}{
		{int64(1), new(int), int(1)},
		{int64(1), new(int8), int8(1)},
		{int64(1), new(int16), int16(1)},
		{int64(1), new(int32), int32(1)},
		{int64(1), new(int64), int64(1)},
		{int64(1), new(uint), uint(1)},
		{int64(1), new(uint8), uint8(1)},
		{int64(1), new(uint16), uint16(1)},
		{int64(1), new(uint32), uint32(1)},
		{int64(1), new(uint64), uint64(1)},
		{int64(1), new(uintptr), uintptr(1)},
		{uint64(1), new(int), int(1)},
		{uint64(1), new(int8), int8(1)},
		{uint64(1), new(int16), int16(1)},
		{uint64(1), new(int32), int32(1)},
		{uint64(1), new(int64), int64(1)},
		{uint64(1), new(uint), uint(1)},
		{uint64(1), new(uint8), uint8(1)},
		{uint64(1), new(uint16), uint16(1)},
		{uint64(1), new(uint32), uint32(1)},
		{uint64(1), new(uint64), uint64(1)},
		{uint64(1), new(uintptr), uintptr(1)},
		{float64(1.5), new(float32), float32(1.5)},
		{float64(1.5), new(float64), float64(1.5)},

		{"1", new(int), int(1)},
		{"1", new(int8), int8(1)},
		{"1", new(int16), int16(1)},
		{"1", new(int32), int32(1)},
		{"1", new(int64), int64(1)},
		{"1", new(uint), uint(1)},
		{"1", new(uint8), uint8(1)},
		{"1", new(uint16), uint16(1)},
		{"1", new(uint32), uint32(1)},
		{"1", new(uint64), uint64(1)},
		{"1", new(uintptr), uintptr(1)},
		{"1.5", new(float32), float32(1.5)},
		{"1.5", new(float64), float64(1.5)},

		{[]byte("test"), &bytesDest, []byte("test")},
		{"test", &bytesDest, []byte("test")},
		{[]rune("test"), &bytesDest, []byte("test")},

		{[]byte("test"), &runesDest, []rune("test")},
		{"test", &runesDest, []rune("test")},
		{[]rune("test"), &runesDest, []rune("test")},

		{[]byte("test"), &strDest, "test"},
		{"test", &strDest, "test"},

		{true, new(bool), true},
		{int64(1), new(bool), true},
		{int64(0), new(bool), false},
		{uint64(1), new(bool), true},
		{uint64(0), new(bool), false},
		{"true", new(bool), true},
		{"false", new(bool), false},
	}

	for _, c := range cases {
		ok, err := RScanTo(reflect.ValueOf(c.dst), c.src, SF_DEFAULT)
		if !ok || err != nil {
			t.Errorf("RScanTo(%T, %T) failed: ok=%v, err=%v", c.dst, c.src, ok, err)
		}
		val := reflect.ValueOf(c.dst).Elem().Interface()
		if !reflect.DeepEqual(val, c.expect) {
			t.Errorf("RScanTo(%T, %T) == %v, want %v", c.dst, c.src, val, c.expect)
		}
	}
}

func TestRScanTo_SpecialCases(t *testing.T) {
	// nil source
	var intDest int = 42
	ok, err := RScanTo(reflect.ValueOf(&intDest), nil, SF_DEFAULT)
	if !ok || err != nil || intDest != 0 {
		t.Errorf("RScanTo with nil source failed")
	}

	// Convertible source
	var ci customInt
	ok, err = RScanTo(reflect.ValueOf(&ci), int(42), SF_DEFAULT)
	if !ok || err != nil || ci != 42 {
		t.Errorf("RScanTo convertible type failed")
	}
}

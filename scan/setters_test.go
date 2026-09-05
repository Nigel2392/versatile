package scan

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

type customString string
type customInt int

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

func TestScan_BasicTypes(t *testing.T) {
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
		ok, err := Scan(&c.dst, c.src, SF_DEFAULT)
		if !ok || err != nil {
			t.Errorf("Scan(%T, %T) failed: ok=%v, err=%v", c.dst, c.src, ok, err)
		}
		val := reflect.ValueOf(c.dst).Elem().Interface()
		if !reflect.DeepEqual(val, c.expect) {
			t.Errorf("Scan(%T, %T) == %v, want %v (%T %T)", c.dst, c.src, val, c.expect, val, c.expect)
		}
	}
}

func TestScan_SpecialCases(t *testing.T) {
	// any pointer
	var anyDest any
	ok, err := Scan(&anyDest, "test", SF_DEFAULT)
	if !ok || err != nil || anyDest != "test" {
		t.Errorf("Scan(*any) failed")
	}

	// SQL Scanner
	var ts testScanner
	ok, err = Scan(&ts, "test", SF_DEFAULT)
	if !ok || err != nil || ts.val != "test" {
		t.Errorf("Scan(sql.Scanner) failed")
	}

	// driver.Valuer
	var intDest int
	dv := driverValuer{val: int64(42)}
	ok, err = Scan(&intDest, dv, SF_DEFAULT)
	if !ok || err != nil || intDest != 42 {
		t.Errorf("Scan from driver.Valuer failed")
	}

	// nil source
	intDest = 42
	ok, err = Scan(&intDest, nil, SF_DEFAULT)
	if !ok || err != nil || intDest != 0 {
		t.Errorf("Scan with nil source failed")
	}
}

func TestRScan_BasicTypes(t *testing.T) {
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
		ok, err := RScan(reflect.ValueOf(c.dst), c.src, SF_DEFAULT)
		if !ok || err != nil {
			t.Errorf("RScan(%T, %T) failed: ok=%v, err=%v", c.dst, c.src, ok, err)
		}
		val := reflect.ValueOf(c.dst).Elem().Interface()
		if !reflect.DeepEqual(val, c.expect) {
			t.Errorf("RScan(%T, %T) == %v, want %v", c.dst, c.src, val, c.expect)
		}
	}
}

func TestRScan_SpecialCases(t *testing.T) {
	// nil source
	var intDest int = 42
	ok, err := RScan(reflect.ValueOf(&intDest), nil, SF_DEFAULT)
	if !ok || err != nil || intDest != 0 {
		t.Errorf("RScan with nil source failed")
	}

	// Convertible source
	var ci customInt
	ok, err = RScan(reflect.ValueOf(&ci), int(42), SF_DEFAULT)
	if !ok || err != nil || ci != 42 {
		t.Errorf("RScan convertible type failed")
	}
}

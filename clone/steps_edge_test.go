package clone

import (
	"testing"
)

type unexportedStruct struct {
	a int
	b string
	c *inner
}

func TestEdge_UnexportedFields(t *testing.T) {
	src := unexportedStruct{
		a: 42,
		b: "hello",
		c: &inner{Value: 99, Label: "secret"},
	}
	var dst unexportedStruct
	mustCopy(t, &dst, src)

	if dst.a != 42 || dst.b != "hello" || dst.c == nil || dst.c.Value != 99 {
		t.Fatalf("Unexported fields not cloned properly: %+v", dst)
	}
	assertNotSamePtr(t, "unexported pointer", dst.c, src.c)
}

type emptyStruct struct{}

func TestEdge_EmptyStruct(t *testing.T) {
	src := emptyStruct{}
	var dst emptyStruct
	mustCopy(t, &dst, src)
}

func TestEdge_MultiDimensionalArray(t *testing.T) {
	src := [2][2]int{{1, 2}, {3, 4}}
	var dst [2][2]int
	mustCopy(t, &dst, src)

	if dst[0][0] != 1 || dst[1][1] != 4 {
		t.Fatalf("Multi-dimensional array mismatch: %+v", dst)
	}
}

type ifaceCycle struct {
	Val any
}

func TestEdge_InterfaceCycle(t *testing.T) {
	src := &ifaceCycle{}
	src.Val = src // cycle through interface

	var dst *ifaceCycle
	mustCopy(t, &dst, src)

	if dst == nil {
		t.Fatalf("dst is nil")
	}
	if dst.Val != dst {
		t.Fatalf("Interface cycle not preserved. dst.Val points to %p, expected %p", dst.Val, dst)
	}
	assertNotSamePtr(t, "cycle root", dst, src)
}

type zeroIfaceStruct struct {
	I any
}

func TestEdge_NilInterface(t *testing.T) {
	src := zeroIfaceStruct{I: nil}
	var dst zeroIfaceStruct
	mustCopy(t, &dst, src)

	if dst.I != nil {
		t.Fatalf("Expected nil interface, got %v", dst.I)
	}
}

func TestEdge_NilPointerInInterface(t *testing.T) {
	var p *inner = nil
	src := zeroIfaceStruct{I: p}
	var dst zeroIfaceStruct
	mustCopy(t, &dst, src)

	if dst.I != p { // should match the typed nil pointer
		t.Fatalf("Expected typed nil pointer, got %v", dst.I)
	}
}

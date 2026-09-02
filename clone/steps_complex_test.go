package clone

import (
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// Test type definitions
// ---------------------------------------------------------------------------

type inner struct {
	Value int
	Label string
}

type outer struct {
	A *inner
	B *inner
}

type node struct {
	Val      int
	Children []*node
}

type graph struct {
	Nodes []*node
	Root  *node
}

type sliceHolder struct {
	S1 []int
	S2 []int
	S3 []int
}

type mapHolder struct {
	M1 map[string]*inner
	M2 map[string]*inner
}

type ptrChain struct {
	A **int
	B **int
}

type selfRef struct {
	Name string
	Next *selfRef
}

type mixedRefs struct {
	Ptr   *inner
	Slice []*inner
	// Map   map[string]*inner
}

type nestedSlice struct {
	Matrix [][]int
}

type nestedMap struct {
	Data map[string]map[string]int
}

type ptrSlice struct {
	Items []*int
}

type ifaceHolder struct {
	Face myFace
}

type multiPtr struct {
	A *int
	B *int
	C *int
	D *int
}

type deepNest struct {
	Level1 *struct {
		Level2 *struct {
			Level3 *struct {
				Value int
			}
		}
	}
}

type sliceOfSlicePtr struct {
	Rows []*[]int
}

type mapOfSlice struct {
	Data map[string][]int
}

type sliceOfMaps struct {
	Items []map[string]int
}

type structWithArray struct {
	Arr [3]*inner
}

type linkedList struct {
	Val  int
	Next *linkedList
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func intPtr(v int) *int { return &v }

func mustCopy(t *testing.T, dst, src any) {
	FLAGFN := Flag(CF_NOWRAP)
	// FLAGFN := Flag(CF_INVALID)
	t.Helper()
	if err := Copy(t.Context(), dst, src, FLAGFN); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
}

func assertDeepEqual(t *testing.T, label string, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: got %+v, want %+v", label, got, want)
	}
}

func assertNotSamePtr(t *testing.T, label string, a, b any) {
	t.Helper()
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	if va.Pointer() == vb.Pointer() {
		t.Errorf("%s: pointers are the same (%p == %p), expected deep copy", label, a, b)
	}
}

// ---------------------------------------------------------------------------
// Part 1: Complex / Nested type tests (non-reference-sharing)
// ---------------------------------------------------------------------------

func TestNested_StructInStruct(t *testing.T) {
	src := outer{
		A: &inner{Value: 1, Label: "a"},
		B: &inner{Value: 2, Label: "b"},
	}
	var dst outer
	mustCopy(t, &dst, src)

	assertDeepEqual(t, "dst.A", *dst.A, *src.A)
	assertDeepEqual(t, "dst.B", *dst.B, *src.B)
	assertNotSamePtr(t, "dst.A", dst.A, src.A)
	assertNotSamePtr(t, "dst.B", dst.B, src.B)
}

func TestNested_NilPointerField(t *testing.T) {
	src := outer{A: &inner{Value: 1}, B: nil}
	var dst outer
	mustCopy(t, &dst, src)

	if dst.B != nil {
		t.Fatalf("expected dst.B to be nil, got %+v", dst.B)
	}
	assertDeepEqual(t, "dst.A", *dst.A, *src.A)
}

func TestNested_DoublePointer(t *testing.T) {
	val := 42
	p := &val
	src := ptrChain{A: &p, B: &p}
	var dst ptrChain
	mustCopy(t, &dst, src)

	if **dst.A != 42 || **dst.B != 42 {
		t.Fatalf("expected **dst.A == **dst.B == 42, got %d, %d", **dst.A, **dst.B)
	}
	// Must not alias the original
	**dst.A = 999
	if val == 999 {
		t.Fatal("mutating dst changed the original source value")
	}
}

func TestNested_DeepNest3Levels(t *testing.T) {
	src := deepNest{
		Level1: &struct {
			Level2 *struct {
				Level3 *struct{ Value int }
			}
		}{
			Level2: &struct {
				Level3 *struct{ Value int }
			}{
				Level3: &struct{ Value int }{Value: 777},
			},
		},
	}
	var dst deepNest
	mustCopy(t, &dst, src)

	if dst.Level1.Level2.Level3.Value != 777 {
		t.Fatalf("expected 777, got %d", dst.Level1.Level2.Level3.Value)
	}
	// Mutate clone, original must be unchanged
	dst.Level1.Level2.Level3.Value = 0
	if src.Level1.Level2.Level3.Value != 777 {
		t.Fatal("mutating dst changed the original")
	}
}

func TestNested_SliceOfSlices(t *testing.T) {
	src := nestedSlice{
		Matrix: [][]int{
			{1, 2, 3},
			{4, 5, 6},
		},
	}
	var dst nestedSlice
	mustCopy(t, &dst, src)

	assertDeepEqual(t, "matrix", dst.Matrix, src.Matrix)
	dst.Matrix[0][0] = 999
	if src.Matrix[0][0] == 999 {
		t.Fatal("inner slice not deeply copied")
	}
}

//	func TestNested_MapOfMaps(t *testing.T) {
//		src := nestedMap{
//			Data: map[string]map[string]int{
//				"a": {"x": 1, "y": 2},
//				"b": {"z": 3},
//			},
//		}
//		var dst nestedMap
//		mustCopy(t, &dst, src)
//
//		assertDeepEqual(t, "data", dst.Data, src.Data)
//		dst.Data["a"]["x"] = 999
//		if src.Data["a"]["x"] == 999 {
//			t.Fatal("inner map not deeply copied")
//		}
//	}

func TestNested_SliceOfPointers(t *testing.T) {
	a, b, c := 10, 20, 30
	src := ptrSlice{Items: []*int{&a, &b, &c}}
	var dst ptrSlice
	mustCopy(t, &dst, src)

	for i, p := range dst.Items {
		if *p != *src.Items[i] {
			t.Errorf("item %d: got %d, want %d", i, *p, *src.Items[i])
		}
		assertNotSamePtr(t, "item ptr", p, src.Items[i])
	}
}

//	func TestNested_MapOfSlices(t *testing.T) {
//		src := mapOfSlice{
//			Data: map[string][]int{
//				"primes": {2, 3, 5, 7},
//				"evens":  {2, 4, 6, 8},
//			},
//		}
//		var dst mapOfSlice
//		mustCopy(t, &dst, src)
//
//		assertDeepEqual(t, "data", dst.Data, src.Data)
//		dst.Data["primes"][0] = 999
//		if src.Data["primes"][0] == 999 {
//			t.Fatal("slice inside map not deeply copied")
//		}
//	}
//
//	func TestNested_SliceOfMaps(t *testing.T) {
//		src := sliceOfMaps{
//			Items: []map[string]int{
//				{"a": 1, "b": 2},
//				{"c": 3},
//			},
//		}
//		var dst sliceOfMaps
//		mustCopy(t, &dst, src)
//
//		assertDeepEqual(t, "items", dst.Items, src.Items)
//		dst.Items[0]["a"] = 999
//		if src.Items[0]["a"] == 999 {
//			t.Fatal("map inside slice not deeply copied")
//		}
//	}

//	func TestNested_MixedRefsAllFields(t *testing.T) {
//		shared := &inner{Value: 42, Label: "shared"}
//		src := mixedRefs{
//			Ptr:   shared,
//			Slice: []*inner{shared, {Value: 2}},
//			Map:   map[string]*inner{"key": shared},
//		}
//		var dst mixedRefs
//		mustCopy(t, &dst, src)
//
//		assertDeepEqual(t, "ptr", *dst.Ptr, *src.Ptr)
//		assertNotSamePtr(t, "ptr", dst.Ptr, src.Ptr)
//		if len(dst.Slice) != 2 || len(dst.Map) != 1 {
//			t.Fatalf("wrong lengths: slice=%d, map=%d", len(dst.Slice), len(dst.Map))
//		}
//	}

func TestNested_InterfaceWithStruct(t *testing.T) {
	src := ifaceHolder{Face: myStruct{val: "deep"}}
	var dst ifaceHolder
	mustCopy(t, &dst, src)

	if dst.Face == nil {
		t.Fatal("expected non-nil interface")
	}
	if dst.Face.StructMethod() != "deep" {
		t.Errorf("expected 'deep', got %q", dst.Face.StructMethod())
	}
}

func TestNested_ArrayOfPointers(t *testing.T) {
	src := structWithArray{
		Arr: [3]*inner{
			{Value: 1, Label: "a"},
			{Value: 2, Label: "b"},
			{Value: 3, Label: "c"},
		},
	}
	var dst structWithArray
	mustCopy(t, &dst, src)

	for i := range 3 {
		assertDeepEqual(t, "arr elem", *dst.Arr[i], *src.Arr[i])
		assertNotSamePtr(t, "arr elem ptr", dst.Arr[i], src.Arr[i])
	}
}

func TestNested_LinkedList(t *testing.T) {
	src := &linkedList{
		Val: 1,
		Next: &linkedList{
			Val: 2,
			Next: &linkedList{
				Val:  3,
				Next: nil,
			},
		},
	}
	var dst linkedList
	mustCopy(t, &dst, *src)

	cur := &dst
	expected := []int{1, 2, 3}
	for i, v := range expected {
		if cur == nil {
			t.Fatalf("list ended early at index %d", i)
		}
		if cur.Val != v {
			t.Errorf("node %d: got %d, want %d", i, cur.Val, v)
		}
		cur = cur.Next
	}
	// Mutate clone, original must not change
	dst.Next.Val = 999
	if src.Next.Val == 999 {
		t.Fatal("mutating clone changed original linked list")
	}
}

func TestNested_TreeStructure(t *testing.T) {
	leaf1 := &node{Val: 10}
	leaf2 := &node{Val: 20}
	root := &node{Val: 1, Children: []*node{leaf1, leaf2}}
	src := graph{
		Nodes: []*node{root, leaf1, leaf2},
		Root:  root,
	}
	var dst graph
	mustCopy(t, &dst, src)

	if dst.Root.Val != 1 || len(dst.Root.Children) != 2 {
		t.Fatalf("root not cloned correctly: val=%d, children=%d", dst.Root.Val, len(dst.Root.Children))
	}
	assertNotSamePtr(t, "root", dst.Root, src.Root)
	assertDeepEqual(t, "tree values", dst.Root.Children[0].Val, 10)
	assertDeepEqual(t, "tree values", dst.Root.Children[1].Val, 20)
}

func TestNested_EmptyCollections(t *testing.T) {
	src := mixedRefs{
		Ptr:   nil,
		Slice: []*inner{},
		// Map:   map[string]*inner{},
	}
	var dst mixedRefs
	mustCopy(t, &dst, src)

	if dst.Ptr != nil {
		t.Error("expected nil ptr")
	}
	if len(dst.Slice) != 0 {
		t.Error("expected empty slice")
	}
	// if dst.Map == nil || len(dst.Map) != 0 {
	// 	t.Error("expected empty (non-nil) map")
	// }
}

func TestNested_NilSlicePreserved(t *testing.T) {
	src := sliceHolder{S1: nil, S2: []int{1}, S3: nil}
	var dst sliceHolder
	mustCopy(t, &dst, src)

	if dst.S1 != nil {
		t.Error("expected nil S1")
	}
	assertDeepEqual(t, "S2", dst.S2, src.S2)
	if dst.S3 != nil {
		t.Error("expected nil S3")
	}
}

func TestNested_SliceOfSlicePointers(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{4, 5, 6}
	src := sliceOfSlicePtr{Rows: []*[]int{&a, &b}}
	var dst sliceOfSlicePtr
	mustCopy(t, &dst, src)

	if len(dst.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(dst.Rows))
	}
	assertDeepEqual(t, "row 0", *dst.Rows[0], a)
	assertDeepEqual(t, "row 1", *dst.Rows[1], b)
	(*dst.Rows[0])[0] = 999
	if a[0] == 999 {
		t.Fatal("mutating dst row changed original")
	}
}

func TestNested_BiggerStructWithPointers(t *testing.T) {
	src := struct {
		Name    string
		Details *struct {
			Age  int
			Tags []string
		}
	}{
		Name: "test",
		Details: &struct {
			Age  int
			Tags []string
		}{
			Age:  30,
			Tags: []string{"go", "clone"},
		},
	}
	var dst struct {
		Name    string
		Details *struct {
			Age  int
			Tags []string
		}
	}
	mustCopy(t, &dst, src)

	assertDeepEqual(t, "name", dst.Name, "test")
	assertDeepEqual(t, "age", dst.Details.Age, 30)
	assertDeepEqual(t, "tags", dst.Details.Tags, []string{"go", "clone"})
	dst.Details.Tags[0] = "rust"
	if src.Details.Tags[0] == "rust" {
		t.Fatal("slice in nested struct not deeply copied")
	}
}

// ---------------------------------------------------------------------------
// Part 2: Shared reference correctness tests (25 tests)
// ---------------------------------------------------------------------------

func TestSharedRef_TwoFieldsSamePointer(t *testing.T) {
	v := 42
	src := struct{ A, B *int }{A: &v, B: &v}
	var dst struct{ A, B *int }
	mustCopy(t, &dst, src)

	if dst.A != dst.B {
		t.Fatalf("expected A and B to share the same pointer, got %p and %p", dst.A, dst.B)
	}
	assertNotSamePtr(t, "vs original", dst.A, src.A)
	*dst.A = 100
	if *dst.B != 100 {
		t.Fatal("shared pointer broken after mutation")
	}
}

func TestSharedRef_ThreeFieldsSamePointer(t *testing.T) {
	v := 7
	src := multiPtr{A: &v, B: &v, C: &v, D: &v}
	var dst multiPtr
	mustCopy(t, &dst, src)

	if dst.A != dst.B || dst.B != dst.C || dst.C != dst.D {
		t.Fatal("expected all four pointers to be shared")
	}
	*dst.A = 123
	if *dst.B != 123 || *dst.C != 123 || *dst.D != 123 {
		t.Fatal("shared pointer state is broken")
	}
	if v == 123 {
		t.Fatal("original was mutated")
	}
}

func TestSharedRef_PointerInStructAndSlice(t *testing.T) {
	shared := &inner{Value: 42, Label: "shared"}
	src := mixedRefs{
		Ptr:   shared,
		Slice: []*inner{shared},
		// Map:   map[string]*inner{"s": shared},
	}
	var dst mixedRefs
	mustCopy(t, &dst, src)

	// All three should point to the same cloned inner
	if dst.Ptr != dst.Slice[0] {
		t.Fatal("Ptr and Slice[0] should share the same pointer in the clone")
	}
	// if dst.Ptr != dst.Map["s"] {
	// 	t.Fatal("Ptr and Map['s'] should share the same pointer in the clone")
	// }
	assertNotSamePtr(t, "vs original", dst.Ptr, src.Ptr)
}

func TestSharedRef_SliceSameBackingArray(t *testing.T) {
	backing := []int{1, 2, 3, 4, 5}
	src := sliceHolder{
		S1: backing,
		S2: backing,
		S3: backing,
	}
	var dst sliceHolder
	mustCopy(t, &dst, src)

	assertDeepEqual(t, "S1", dst.S1, src.S1)
	dst.S1[0] = 999
	if dst.S2[0] != 999 || dst.S3[0] != 999 {
		t.Fatal("shared backing array not preserved in clone")
	}
	if src.S1[0] == 999 {
		t.Fatal("original mutated")
	}
}

func _TestSharedRef_MapSamePointerValues(t *testing.T) {
	shared := &inner{Value: 99}
	src := mapHolder{
		M1: map[string]*inner{"x": shared, "y": shared},
		M2: map[string]*inner{"z": shared},
	}
	var dst mapHolder
	mustCopy(t, &dst, src)

	// All entries should point to the same clone
	if dst.M1["x"] != dst.M1["y"] {
		t.Fatal("M1['x'] and M1['y'] should share the same pointer")
	}
	if dst.M1["x"] != dst.M2["z"] {
		t.Fatal("M1['x'] and M2['z'] should share the same pointer")
	}
	assertNotSamePtr(t, "vs original", dst.M1["x"], src.M1["x"])

	dst.M1["x"].Value = 1000
	if dst.M2["z"].Value != 1000 {
		t.Fatal("shared pointer in map is broken")
	}
}

func TestSharedRef_DoublePointerShared(t *testing.T) {
	val := 42
	p := &val
	src := ptrChain{A: &p, B: &p}
	var dst ptrChain
	mustCopy(t, &dst, src)

	// The inner pointer (*int) should be shared
	if *dst.A != *dst.B {
		t.Fatal("expected inner pointers to be shared")
	}
	**dst.A = 555
	if **dst.B != 555 {
		t.Fatal("shared inner pointer broken")
	}
	if val == 555 {
		t.Fatal("original mutated")
	}
}

func TestSharedRef_TreeSharedLeaf(t *testing.T) {
	sharedLeaf := &node{Val: 99}
	root := &node{
		Val: 1,
		Children: []*node{
			{Val: 2, Children: []*node{sharedLeaf}},
			{Val: 3, Children: []*node{sharedLeaf}},
		},
	}
	var dst node
	mustCopy(t, &dst, *root)

	leafA := dst.Children[0].Children[0]
	leafB := dst.Children[1].Children[0]
	if leafA != leafB {
		t.Fatal("shared leaf node not preserved in clone")
	}
	leafA.Val = 1234
	if leafB.Val != 1234 {
		t.Fatal("shared leaf mutation not reflected")
	}
	if sharedLeaf.Val == 1234 {
		t.Fatal("original leaf mutated")
	}
}

func TestSharedRef_GraphNodesMatchRoot(t *testing.T) {
	root := &node{Val: 1}
	src := graph{
		Nodes: []*node{root},
		Root:  root,
	}
	var dst graph
	mustCopy(t, &dst, src)

	if dst.Root != dst.Nodes[0] {
		t.Fatal("Root and Nodes[0] should be the same pointer in clone")
	}
	assertNotSamePtr(t, "vs original root", dst.Root, src.Root)
}

func TestSharedRef_GraphMultipleNodesShared(t *testing.T) {
	n1 := &node{Val: 1}
	n2 := &node{Val: 2}
	src := graph{
		Nodes: []*node{n1, n2, n1, n2},
		Root:  n1,
	}
	var dst graph
	mustCopy(t, &dst, src)

	if dst.Nodes[0] != dst.Nodes[2] {
		t.Fatal("Nodes[0] and Nodes[2] should be shared")
	}
	if dst.Nodes[1] != dst.Nodes[3] {
		t.Fatal("Nodes[1] and Nodes[3] should be shared")
	}
	if dst.Root != dst.Nodes[0] {
		t.Fatal("Root should be Nodes[0]")
	}
}

func TestSharedRef_StructFieldAndSliceElement(t *testing.T) {
	shared := &inner{Value: 55, Label: "both"}
	type combo struct {
		Direct *inner
		List   []*inner
	}
	src := combo{
		Direct: shared,
		List:   []*inner{shared, shared},
	}
	var dst combo
	mustCopy(t, &dst, src)

	if dst.Direct != dst.List[0] || dst.Direct != dst.List[1] {
		t.Fatal("Direct, List[0], and List[1] should all share the same pointer")
	}
	dst.Direct.Value = 9999
	if dst.List[0].Value != 9999 || dst.List[1].Value != 9999 {
		t.Fatal("shared pointer mutation not reflected in all locations")
	}
}

func TestSharedRef_OuterStructSharedInner(t *testing.T) {
	shared := &inner{Value: 100, Label: "ref"}
	src := outer{A: shared, B: shared}
	var dst outer
	mustCopy(t, &dst, src)

	if dst.A != dst.B {
		t.Fatal("A and B should share the same pointer")
	}
	dst.A.Value = 200
	if dst.B.Value != 200 {
		t.Fatal("shared pointer broken")
	}
	if shared.Value == 200 {
		t.Fatal("original mutated")
	}
}

func TestSharedRef_LinkedListCycle2(t *testing.T) {
	// Non-cyclic but with shared tail
	tail := &linkedList{Val: 3, Next: nil}
	src := struct {
		A *linkedList
		B *linkedList
	}{
		A: &linkedList{Val: 1, Next: tail},
		B: &linkedList{Val: 2, Next: tail},
	}
	var dst struct {
		A *linkedList
		B *linkedList
	}
	mustCopy(t, &dst, src)

	if dst.A.Next != dst.B.Next {
		t.Fatal("shared tail not preserved in clone")
	}
	dst.A.Next.Val = 999
	if dst.B.Next.Val != 999 {
		t.Fatal("shared tail mutation not reflected")
	}
	if tail.Val == 999 {
		t.Fatal("original tail mutated")
	}
}

func _TestSharedRef_MapTwoKeysOneValue(t *testing.T) {
	shared := &inner{Value: 11}
	src := struct {
		M map[string]*inner
	}{
		M: map[string]*inner{
			"alpha": shared,
			"beta":  shared,
		},
	}
	var dst struct {
		M map[string]*inner
	}
	mustCopy(t, &dst, src)

	if dst.M["alpha"] != dst.M["beta"] {
		t.Fatal("map values should share the same pointer")
	}
	dst.M["alpha"].Value = 500
	if dst.M["beta"].Value != 500 {
		t.Fatal("shared pointer in map broken")
	}
}

func TestSharedRef_SliceOfPointersAllSame(t *testing.T) {
	v := 42
	src := ptrSlice{Items: []*int{&v, &v, &v}}
	var dst ptrSlice
	mustCopy(t, &dst, src)

	if dst.Items[0] != dst.Items[1] || dst.Items[1] != dst.Items[2] {
		t.Fatal("all slice items should share the same pointer")
	}
	*dst.Items[0] = 888
	if *dst.Items[1] != 888 || *dst.Items[2] != 888 {
		t.Fatal("shared pointer broken")
	}
	if v == 888 {
		t.Fatal("original mutated")
	}
}

func TestSharedRef_NestedStructsShareDeepPointer(t *testing.T) {
	type deep struct {
		Core *inner
	}
	type top struct {
		X deep
		Y deep
	}
	shared := &inner{Value: 7, Label: "core"}
	src := top{
		X: deep{Core: shared},
		Y: deep{Core: shared},
	}
	var dst top
	mustCopy(t, &dst, src)

	if dst.X.Core != dst.Y.Core {
		t.Fatal("nested Core pointers should be shared")
	}
	dst.X.Core.Value = 77
	if dst.Y.Core.Value != 77 {
		t.Fatal("shared nested pointer broken")
	}
}

func TestSharedRef_ArrayElementsShared(t *testing.T) {
	shared := &inner{Value: 5, Label: "arr"}
	src := structWithArray{
		Arr: [3]*inner{shared, shared, shared},
	}
	var dst structWithArray
	mustCopy(t, &dst, src)

	if dst.Arr[0] != dst.Arr[1] || dst.Arr[1] != dst.Arr[2] {
		t.Fatal("array elements should share the same pointer")
	}
	dst.Arr[0].Label = "modified"
	if dst.Arr[1].Label != "modified" || dst.Arr[2].Label != "modified" {
		t.Fatal("shared pointer in array broken")
	}
}

func _TestSharedRef_SliceAndMapCrossReference(t *testing.T) {
	shared := &inner{Value: 33}
	type cross struct {
		List []*inner
		Dict map[int]*inner
	}
	src := cross{
		List: []*inner{shared, shared},
		Dict: map[int]*inner{0: shared, 1: shared},
	}
	var dst cross
	mustCopy(t, &dst, src)

	// All six references should be the same clone
	ptrs := []*inner{dst.List[0], dst.List[1], dst.Dict[0], dst.Dict[1]}
	for i := 1; i < len(ptrs); i++ {
		if ptrs[i] != ptrs[0] {
			t.Fatalf("reference %d is not shared with reference 0", i)
		}
	}
	dst.List[0].Value = 4444
	for i, p := range ptrs {
		if p.Value != 4444 {
			t.Fatalf("reference %d not updated after mutation", i)
		}
	}
}

func TestSharedRef_MutationIsolation(t *testing.T) {
	// Verify that after cloning, modifying the clone in any way does not touch the source
	shared := &inner{Value: 1, Label: "original"}
	src := outer{A: shared, B: shared}
	var dst outer
	mustCopy(t, &dst, src)

	dst.A.Value = 999
	dst.A.Label = "cloned"
	if src.A.Value != 1 || src.A.Label != "original" {
		t.Fatalf("source was mutated: %+v", src.A)
	}
	if src.B.Value != 1 || src.B.Label != "original" {
		t.Fatalf("source B was mutated: %+v", src.B)
	}
}

func _TestSharedRef_MapOfSlicesSharedBacking(t *testing.T) {
	backing := []int{1, 2, 3}
	src := struct {
		M map[string][]int
	}{
		M: map[string][]int{
			"a": backing,
			"b": backing,
		},
	}
	var dst struct {
		M map[string][]int
	}
	mustCopy(t, &dst, src)

	assertDeepEqual(t, "a", dst.M["a"], backing)
	assertDeepEqual(t, "b", dst.M["b"], backing)

	dst.M["a"][0] = 999
	if dst.M["b"][0] != 999 {
		t.Fatal("shared backing array in map values not preserved")
	}
	if backing[0] == 999 {
		t.Fatal("original mutated")
	}
}

func TestSharedRef_SliceOfSlicePtrsSameTarget(t *testing.T) {
	shared := []int{10, 20, 30}
	src := sliceOfSlicePtr{Rows: []*[]int{&shared, &shared}}
	var dst sliceOfSlicePtr
	mustCopy(t, &dst, src)

	if dst.Rows[0] != dst.Rows[1] {
		t.Fatal("expected Rows[0] and Rows[1] to share the same pointer")
	}
	(*dst.Rows[0])[0] = 999
	if (*dst.Rows[1])[0] != 999 {
		t.Fatal("shared pointer broken")
	}
	if shared[0] == 999 {
		t.Fatal("original mutated")
	}
}

func TestSharedRef_SiblingNodesShareParent(t *testing.T) {
	parent := &node{Val: 0}
	child1 := &node{Val: 1, Children: []*node{parent}}
	child2 := &node{Val: 2, Children: []*node{parent}}
	src := struct{ Kids []*node }{Kids: []*node{child1, child2}}
	var dst struct{ Kids []*node }
	mustCopy(t, &dst, src)

	p1 := dst.Kids[0].Children[0]
	p2 := dst.Kids[1].Children[0]
	if p1 != p2 {
		t.Fatal("shared parent node not preserved across siblings")
	}
	p1.Val = 555
	if p2.Val != 555 {
		t.Fatal("shared parent mutation not reflected")
	}
	if parent.Val == 555 {
		t.Fatal("original parent mutated")
	}
}

func TestSharedRef_InterfaceSliceSharedConcrete(t *testing.T) {
	shared := myStruct{val: "hello"}
	src := struct{ Items []myFace }{Items: []myFace{shared, shared}}
	var dst struct{ Items []myFace }
	mustCopy(t, &dst, src)

	if len(dst.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(dst.Items))
	}
	if dst.Items[0].StructMethod() != "hello" || dst.Items[1].StructMethod() != "hello" {
		t.Fatal("interface values not cloned correctly")
	}
}

package clone

import (
	"testing"
)

// Generate a deeply nested structure for benchmarking.
func createLargeTree(depth int) *node {
	if depth <= 0 {
		return &node{Val: 42}
	}
	return &node{
		Val: depth,
		Children: []*node{
			createLargeTree(depth - 1),
			createLargeTree(depth - 1),
			createLargeTree(depth - 1),
		},
	}
}

// Generate a highly connected graph.
func createComplexGraph(nodesCount int) *graph {
	nodes := make([]*node, nodesCount)
	for i := range nodes {
		nodes[i] = &node{Val: i}
	}
	// Connect every node to multiple other nodes to create many shared pointers and cycles
	for i := range nodes {
		for j := 1; j <= 5; j++ {
			target := (i + j) % nodesCount
			nodes[i].Children = append(nodes[i].Children, nodes[target])
		}
	}
	return &graph{Root: nodes[0], Nodes: nodes}
}

func BenchmarkClone_ComplexStruct(b *testing.B) {
	src := mixedRefs{
		Ptr:   &inner{Value: 1},
		Slice: []*inner{{Value: 2}, {Value: 3}},
	}
	stateCtx := SharedStateContext(
		b.Context(), Flag(CF_NOWRAP),
	)

	b.ResetTimer()
	for b.Loop() {
		var dst mixedRefs
		err := Copy(stateCtx, &dst, src)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClone_LargeTree(b *testing.B) {
	// 3^6 = 729 nodes
	src := createLargeTree(6)
	FLAGFN := Flag(CF_NOWRAP)
	stateCtx := SharedStateContext(
		b.Context(), FLAGFN,
	)

	b.ResetTimer()
	for b.Loop() {
		var dst *node
		err := Copy(stateCtx, &dst, src, FLAGFN)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClone_ComplexGraph(b *testing.B) {
	// Graph with 200 interconnected nodes
	src := createComplexGraph(200)
	stateCtx := SharedStateContext(
		b.Context(), Flag(CF_NOWRAP),
	)

	b.ResetTimer()
	for b.Loop() {
		var dst *graph
		err := Copy(stateCtx, &dst, src)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClone_DeeplyNestedSlice(b *testing.B) {
	src := [][][][]int{
		{
			{{1, 2}, {3, 4}},
			{{5, 6}, {7, 8}},
		},
		{
			{{9, 10}, {11, 12}},
			{{13, 14}, {15, 16}},
		},
	}
	stateCtx := SharedStateContext(
		b.Context(), Flag(CF_NOWRAP),
	)

	b.ResetTimer()
	for b.Loop() {
		var dst [][][][]int
		err := Copy(stateCtx, &dst, src)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClone_SharedPointersSlice(b *testing.B) {
	shared := &inner{Value: 42}
	src := make([]*inner, 1000)
	for i := range src {
		src[i] = shared
	}
	stateCtx := SharedStateContext(
		b.Context(), Flag(CF_NOWRAP),
	)

	b.ResetTimer()
	for b.Loop() {
		var dst []*inner
		err := Copy(stateCtx, &dst, src)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClone_InterfacePointers(b *testing.B) {
	src := []any{
		&inner{Value: 1},
		&inner{Value: 2},
		&inner{Value: 3},
		&inner{Value: 4},
		&inner{Value: 5},
	}
	stateCtx := SharedStateContext(
		b.Context(), Flag(CF_NOWRAP),
	)

	b.ResetTimer()
	for b.Loop() {
		var dst []any
		err := Copy(stateCtx, &dst, src)
		if err != nil {
			b.Fatal(err)
		}
	}
}

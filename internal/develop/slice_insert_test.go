package develop

import (
	"slices"
	"testing"
)

const PREBUILT int = 500

func BenchmarkSliceAppend0(b *testing.B) {
	var slice = make([]int, 0, PREBUILT)
	for i := range PREBUILT {
		slice = append(slice, i)
	}

	b.Log("Starting append test...")

	for b.Loop() {
		slice = append([]int{69}, slice...)
	}

	b.Logf("append(int{...}, ...) slice size: %d", len(slice))
}

func BenchmarkSliceInsert0(b *testing.B) {
	var slice = make([]int, 0, PREBUILT)
	for i := range PREBUILT {
		slice = append(slice, i)
	}

	b.Log("Starting slices.Insert test...")

	for b.Loop() {
		slice = slices.Insert(slice, 0, 69)
	}

	b.Logf("slices.Insert slice size: %d", len(slice))
}

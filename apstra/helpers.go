package tfapstra

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math/rand"
	"unsafe"
)

// sliceWithoutElement returns a copy of in with all occurrences of e removed.
// the returned int indicates the number of occurrences removed.
func sliceWithoutElement[A comparable](in []A, e A) ([]A, int) {
	result := make([]A, len(in))
	var resultIdx int
	for inIdx := range in {
		if in[inIdx] == e {
			continue
		}
		result[resultIdx] = in[inIdx]
		resultIdx++
	}
	return result[:resultIdx], len(in) - resultIdx
}

func FillWithRandomIntegers[A constraints.Integer](a []A) {
	if len(a) == 0 {
		return
	}

	var nominal A
	maxSize := unsafe.Sizeof(uint64(0))
	nominalSize := unsafe.Sizeof(nominal)
	if nominalSize > maxSize {
		panic(fmt.Sprintf("FillWithRandomIntegers got unexpectedly large integer type %v", nominal))
	}

	for i := 0; i < len(a); i++ {
		a[i] = A(rand.Uint64())
	}
}

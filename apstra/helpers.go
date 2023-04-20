package tfapstra

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

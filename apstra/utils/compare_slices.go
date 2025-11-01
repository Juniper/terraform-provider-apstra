package utils

import "fmt"

func SliceStringersMatch[A fmt.Stringer](a, b []A) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}

	return true
}

// DiffSliceSets compares two slices of un-ordered data. The returned values are
// un-ordered slices of values found in only one of the supplied slices.
// It might be reasonable to think of the returned slices as "missing from a"
// and "missing from b"
//
//	Example:
//	 a: []int{1,2,3,4}
//	 b: []int{2,1,5}
//	 return: []int{5}, []int{4, 3}
func DiffSliceSets[A comparable](a, b []A) ([]A, []A) {
	ma := make(map[A]bool, len(a))
	for i := range a {
		ma[a[i]] = true
	}

	mb := make(map[A]bool, len(b))
	for i := range b {
		mb[b[i]] = true
	}

	var resultA []A
	for i := range mb {
		if ma[i] {
			continue
		}
		resultA = append(resultA, i)
	}

	var resultB []A
	for i := range ma {
		if mb[i] {
			continue
		}
		resultB = append(resultB, i)
	}

	return resultA, resultB
}

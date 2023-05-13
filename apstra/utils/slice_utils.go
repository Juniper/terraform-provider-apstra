package utils

import "fmt"

func ItemInSlice[A comparable](item A, slice []A) bool {
	for i := range slice {
		if item == slice[i] {
			return true
		}
	}
	return false
}

// Uniq returns the supplied slice with duplicates removed. Order is not
// preserved. Name is taken from '/usr/bin/uniq' because of similar function.
func Uniq[A comparable](in []A) []A {
	m := make(map[A]struct{})           // unique values from 'A' key this map
	for i := len(in) - 1; i >= 0; i-- { // work backwards through the slice
		if _, ok := m[in[i]]; ok {
			in[i] = in[len(in)-1] // overwrite duplicate with item from end of slice
			in = in[:len(in)-1]   // shorten slice to eliminate newly dup'ed last item
			continue
		}
		m[in[i]] = struct{}{} // unique element used as map key
	}

	return in
}

// UniqueElementsFromA returns elements of slice 'a' which do not appear in
// slice 'b'. Order is not preserved.
func UniqueElementsFromA[E comparable](a, b []E) []E {
	bMap := make(map[E]struct{})
	for i := range b {
		bMap[b[i]] = struct{}{}
	}

	var result []E
	for i := range a {
		if _, ok := bMap[a[i]]; !ok {
			result = append(result, a[i])
		}
	}
	return result
}

func SliceContains[A comparable](a A, s []A) bool {
	for i := range s {
		if s[i] == a {
			return true
		}
	}
	return false
}

func StringersToStrings[A fmt.Stringer](in []A) []string {
	result := make([]string, len(in))
	for i := range in {
		result[i] = in[i].String()
	}
	return result
}

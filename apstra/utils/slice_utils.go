package utils

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"sort"
)

func ItemInSlice[A comparable](item A, slice []A) bool {
	for i := range slice {
		if item == slice[i] {
			return true
		}
	}
	return false
}

// Uniq returns the supplied slice with consecutive duplicates removed. Order is
// not preserved. Name is taken from '/usr/bin/uniq' because of similar function.
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

// UniqStringers returns the supplied slice with consecutive duplicates removed. Order
// is not preserved. Name is taken from '/usr/bin/uniq' because of similar function.
func UniqStringers[A fmt.Stringer](in []A) []A {
	m := make(map[string]struct{})      // unique string results from 'A' key this map
	for i := len(in) - 1; i >= 0; i-- { // work backwards through the slice
		if _, ok := m[in[i].String()]; ok {
			in[i] = in[len(in)-1] // overwrite duplicate with item from end of slice
			in = in[:len(in)-1]   // shorten slice to eliminate newly dup'ed last item
			continue
		}
		m[in[i].String()] = struct{}{} // unique element used as map key
	}

	return in
}

func Swap[A any](a, b int, in []A) {
	x := in[a]
	in[a] = in[b]
	in[b] = x
}

func Reverse[A any](in []A) {
	for i := 0; i < len(in)/2; i++ {
		Swap(i, len(in)-1-i, in)
	}
}

func SliceDeleteUnOrdered[A any](i int, a *[]A) {
	s := *a
	s[i] = s[len(s)-1] // copy item from end of slice to position i
	s = s[:len(s)-1]
	*a = s
}

// SliceComplementOfA returns items from b which do not appear in a
func SliceComplementOfA[T comparable](a, b []T) []T {
	mapA := make(map[T]bool, len(a))
	for _, t := range a {
		mapA[t] = true
	}

	mapB := make(map[T]bool, len(b))
	for _, t := range b {
		mapB[t] = true
	}

	var result []T
	for t := range mapB {
		if !mapA[t] {
			result = append(result, t)
		}
	}

	return result
}

// SliceIntersectionOfAB returns items which appear in both a and b
func SliceIntersectionOfAB[T comparable](a, b []T) []T {
	mapA := make(map[T]bool, len(a))
	for _, t := range a {
		mapA[t] = true
	}

	mapB := make(map[T]bool, len(b))
	for _, t := range b {
		mapB[t] = true
	}

	var result []T
	for t := range mapB {
		if mapA[t] {
			result = append(result, t)
		}
	}

	return result
}

func Sort[A constraints.Ordered](in []A) {
	if in == nil {
		return
	}

	sort.Slice(in, func(i, j int) bool {
		return in[i] < in[j]
	})
}

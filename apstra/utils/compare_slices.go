package utils

import "fmt"

func SlicesMatch[A comparable](a, b []A) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

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

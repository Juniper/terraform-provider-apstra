package utils

import (
	"cmp"
	"slices"
)

func MapKeys[K comparable, V interface{}](m map[K]V) []K {
	keys := make([]K, len(m))
	var i int
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

func MapKeysSorted[K cmp.Ordered, V interface{}](m map[K]V) []K {
	keys := MapKeys(m)
	slices.Sort(keys)
	return keys
}

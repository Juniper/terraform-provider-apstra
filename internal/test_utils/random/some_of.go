//go:build integration

package random

import (
	"fmt"
	"math/rand"
)

func SomeOf[T any](t []T, min_max ...int) []T {
	min := 0
	max := len(t)
	if len(min_max) > 0 {
		min = len(t)
	}
	if len(min_max) > 1 {
		max = min_max[1]
	}

	if max < min {
		panic(fmt.Sprintf("max (%d) cannot be less than min (%d)", max, min))
	}

	result := make([]T, len(t))
	copy(result, t)

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result[:rand.Intn(max-min)+min+1]
}

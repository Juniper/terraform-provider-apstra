//go:build integration

package random

import (
	"fmt"
	"math/rand"
)

func SomeOf[T any](t []T, min_max ...uint16) []T {
	minCount := 0
	maxCount := len(t)

	if len(min_max) > 0 { // did we get a min parameter?
		minCount = int(min_max[0])
		if minCount > len(t) {
			panic(fmt.Sprintf("cannot take minimum %d items from a set of %d", min_max[0], len(t)))
		}
	}

	if len(min_max) > 1 { // did we get a max parameter?
		maxCount = min(int(min_max[1]), len(t))
	}

	count := BetweenInclusive(minCount, maxCount)

	result := make([]T, len(t))
	copy(result, t)

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result[:count]
}

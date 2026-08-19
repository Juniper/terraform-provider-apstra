//go:build integration

package random

import "math/rand"

func BetweenInclusive(a, b int) int {
	if a > b {
		a, b = b, a
	}
	return rand.Intn(b-a+1) + a
}

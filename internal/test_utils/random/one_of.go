//go:build integration

package random

import "math/rand"

func OneOf[T any](t ...T) T {
	return t[rand.Intn(len(t))]
}

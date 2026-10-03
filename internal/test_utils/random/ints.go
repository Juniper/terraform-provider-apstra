package random

import (
	"fmt"
	"math/rand"
)

func Ints(min, max, count int) ([]int, error) {
	if max-min+1 < count {
		return nil, fmt.Errorf("cannot generate %d random numbers between %d and %d inclusive", count, min, max)
	}
	intMap := make(map[int]struct{}, count)
	for len(intMap) < count {
		intMap[rand.Intn(max+1-min)+min] = struct{}{}
	}
	result := make([]int, count)
	i := 0
	for k := range intMap {
		result[i] = k
		i++
	}
	return result, nil
}

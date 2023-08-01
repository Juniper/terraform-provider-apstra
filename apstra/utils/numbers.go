package utils

import (
	"golang.org/x/exp/constraints"
	"math/big"
)

func BigIntToBigFloat(in *big.Int) *big.Float {
	bigval := new(big.Float)
	bigval.SetInt(in)
	return bigval
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

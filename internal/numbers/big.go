package numbers

import "math/big"

func BigIntToBigFloat(in *big.Int) *big.Float {
	bigval := new(big.Float)
	bigval.SetInt(in)
	return bigval
}

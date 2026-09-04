package random

import (
	"fmt"
	"math/rand"
	"testing"
)

func RouteTarget(t testing.TB) string {
	// three syntactic styles for RTs
	r := rand.Intn(3)
	switch r {
	case 0: // 16-bits:32-bits
		return fmt.Sprintf("%d:%d", uint16(rand.Uint32()), rand.Uint32())
	case 1: // 32-bits:16-bits
		return fmt.Sprintf("%d:%d", rand.Uint32(), uint16(rand.Uint32()))
	case 2: // IPv4:16-bits
		return fmt.Sprintf("%s:%d", NetIP(t, "192.0.2.0/24"), uint16(rand.Uint32()))
	}

	panic(nil)
}

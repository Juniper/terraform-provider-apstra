package random

import (
	"math/rand"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

// NetIP returns a random net.IP from within the specified block.
// Works for both IPv4 and IPv6 blocks.
func NetIP(t testing.TB, block string) net.IP {
	t.Helper()

	_, ipNet, err := net.ParseCIDR(block)
	require.NoError(t, err)
	require.NotNil(t, ipNet)

	return netIP(*ipNet)
}

// netIP returns a random net.IP from within the specified block.
// Works for both IPv4 and IPv6 blocks.
func netIP(block net.IPNet) net.IP {
	// ipNet.IP and ipNet.Mask are always the same length: 4 for IPv4, 16 for IPv6.
	// For each byte, fixed (masked) bits come from the network address;
	// free (host) bits are randomized.
	result := make(net.IP, len(block.IP))
	for i := range block.IP {
		result[i] = block.IP[i] | (byte(rand.Intn(256)) &^ block.Mask[i])
	}

	return result
}

// NetIPNet returns a net.IPNet from within the specified block.
func NetIPNet(t testing.TB, block string) net.IPNet {
	t.Helper()

	_, ipNet, err := net.ParseCIDR(block)
	require.NoError(t, err)
	require.NotNil(t, ipNet)

	ip := netIP(*ipNet)
	ones, bits := ipNet.Mask.Size()

	if ones == bits {
		panic("cannot create a new block within " + block)
	}

	// Increase the prefix length by a random amount (1 to remaining host bits).
	newOnes := ones + rand.Intn(bits-ones) + 1
	newMask := net.CIDRMask(newOnes, bits)

	// Mask the random IP down to the new network address.
	network := ip.Mask(newMask)

	return net.IPNet{
		IP:   network,
		Mask: newMask,
	}
}

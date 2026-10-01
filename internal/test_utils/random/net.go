package random

import (
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"net"
	"net/netip"
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

// RandomPrefixes returns n random prefixes of the specified size (bits) from within the given CIDR block.
func RandomPrefixes(t testing.TB, cidrBlock string, bits, n int) []netip.Prefix {
	t.Helper()

	if n < 0 {
		t.Fatalf("number of prefixes must not be negative: %d", n)
	}
	if n == 0 {
		return nil
	}

	container, err := netip.ParsePrefix(cidrBlock)
	if err != nil {
		t.Fatalf("invalid CIDR block %q: %v", cidrBlock, err)
	}
	if container != container.Masked() {
		t.Fatalf("CIDR block %q is not a network address", cidrBlock)
	}

	addrBits := container.Addr().BitLen()
	containerBits := container.Bits()

	if bits < containerBits || bits > addrBits {
		t.Fatalf(
			"output prefix /%d is invalid for containing prefix %s",
			bits, container,
		)
	}

	// There are 2^(bits-containerBits) possible output prefixes.
	indexBits := bits - containerBits
	numPrefixes := new(big.Int).Lsh(
		big.NewInt(1),
		uint(indexBits),
	)

	if new(big.Int).SetInt64(int64(n)).Cmp(numPrefixes) > 0 {
		t.Fatalf(
			"cannot generate %d distinct /%d prefixes inside %s: only %s exist",
			n, bits, container, numPrefixes,
		)
	}

	// Pick distinct indexes.
	indexes := make(map[string]*big.Int, n)
	for len(indexes) < n {
		index, err := crand.Int(crand.Reader, numPrefixes)
		if err != nil {
			t.Fatalf("generating random prefix index: %v", err)
		}

		indexes[index.String()] = index
	}

	// Convert the unique indexes to prefixes.
	result := make([]netip.Prefix, 0, n)

	for _, index := range indexes {
		result = append(result, prefixAtIndex(container, bits, index))
	}

	return result
}

// prefixAtIndex returns the bits-sized prefix identified by index within
// container. The index is interpreted as the bits between container's prefix
// length and bits, in network-address order. The caller must ensure that
// container is a network address and that container.Bits() <= bits <= the
// address length.
func prefixAtIndex(container netip.Prefix, bits int, index *big.Int) netip.Prefix {
	addrBits := container.Addr().BitLen()
	hostBits := addrBits - bits

	addr := container.Addr().As16()
	base := new(big.Int).SetBytes(addr[:])

	offset := new(big.Int).Lsh(new(big.Int).Set(index), uint(hostBits))
	base.Add(base, offset)

	var bytes [16]byte
	base.FillBytes(bytes[:])

	var result netip.Addr

	if container.Addr().Is4() {
		var v4 [4]byte
		copy(v4[:], bytes[12:])
		result = netip.AddrFrom4(v4)
	} else {
		result = netip.AddrFrom16(bytes)
	}

	return netip.PrefixFrom(result, bits)
}

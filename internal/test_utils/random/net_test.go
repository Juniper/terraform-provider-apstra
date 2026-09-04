package random_test

import (
	"net"
	"testing"

	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/stretchr/testify/require"
)

func TestNetIP(t *testing.T) {
	t.Parallel()

	type testCase struct {
		block string
	}

	testCases := []testCase{
		{block: "10.0.0.0/8"},
		{block: "192.168.1.0/24"},
		{block: "172.16.0.0/12"},
		{block: "10.0.0.0/31"}, // only two addresses
		{block: "2001:db8::/32"},
		{block: "fd00::/8"},
		{block: "2001:db8:abcd::/48"},
	}

	for _, tCase := range testCases {
		t.Run(tCase.block, func(t *testing.T) {
			t.Parallel()

			_, ipNet, err := net.ParseCIDR(tCase.block)
			require.NoError(t, err)

			// Run several times to increase confidence the result is always within the block.
			for range 100 {
				got := random.NetIP(t, tCase.block)
				require.NotNil(t, got)

				if !ipNet.Contains(got) {
					t.Errorf("NetIP returned %q which is not within block %q", got, tCase.block)
				}
			}
		})
	}
}

func TestNetIPNet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		block string
	}

	testCases := []testCase{
		{block: "10.0.0.0/8"},
		{block: "192.168.1.0/24"},
		{block: "172.16.0.0/12"},
		{block: "10.0.0.0/30"}, // only two host bits, so newOnes must be /31 or /32
		{block: "2001:db8::/32"},
		{block: "fd00::/8"},
		{block: "2001:db8:abcd::/48"},
	}

	for _, tCase := range testCases {
		t.Run(tCase.block, func(t *testing.T) {
			t.Parallel()

			_, outerNet, err := net.ParseCIDR(tCase.block)
			if err != nil {
				t.Fatalf("failed to parse block %q: %v", tCase.block, err)
			}
			outerOnes, _ := outerNet.Mask.Size()

			// Run several times to increase confidence the result is always valid.
			for range 100 {
				result := random.NetIPNet(t, tCase.block)

				// The result's base address must be within the outer block.
				require.True(t, outerNet.Contains(result.IP), "NetIPNet returned network %s whose address is not within block %s", result, tCase.block)

				// The result's prefix length must be strictly longer than the outer block's.
				resultOnes, _ := result.Mask.Size()
				require.Greater(t, resultOnes, outerOnes, "NetIPNet returned network %s whose prefix length is not longer than block %s's prefix length", result, tCase.block)

				// The result must be a proper network address (IP & mask == IP).
				require.True(t, result.IP.Equal(result.IP.Mask(result.Mask)), "NetIPNet returned a malformed IPNet: Base address mismatch")
			}
		})
	}
}

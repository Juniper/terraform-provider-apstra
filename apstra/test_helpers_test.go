package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"golang.org/x/exp/constraints"
)

func systemIds(ctx context.Context, t testing.TB, bp *apstra.TwoStageL3ClosClient, role string) []string {
	t.Helper()

	query := new(apstra.PathQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "role", Value: apstra.QEStringVal(role)},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		})

	var result struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	err := query.Do(ctx, &result)
	if err != nil {
		t.Fatal(err)
	}

	ids := make([]string, len(result.Items))
	for i, item := range result.Items {
		ids[i] = item.System.Id
	}

	return ids
}

//func stringPtrOrNull(in *string) string {
//	if in == nil {
//		return "null"
//	}
//	return `"` + *in + `"`
//}

func stringOrNull(in string) string {
	if in == "" {
		return "null"
	}
	return `"` + in + `"`
}

func boolPtrOrNull(b *bool) string {
	if b == nil {
		return "null"
	}
	return strconv.FormatBool(*b)
}

func cidrOrNull(in *net.IPNet) string {
	if in == nil {
		return "null"
	}
	return `"` + in.String() + `"`
}

//func ipOrNull(in *net.IP) string {
//	if in == nil {
//		return "null"
//	}
//	return `"` + in.String() + `"`
//}

func intPtrOrNull[A constraints.Integer](in *A) string {
	if in == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *in)
}

func intZeroAsNull[A constraints.Integer](in A) string {
	if in == 0 {
		return "null"
	}
	return fmt.Sprintf("%d", in)
}

func stringSetOrNull(in []string) string {
	if in == nil {
		return "null"
	}

	if len(in) == 0 {
		return "[]"
	}

	return `["` + strings.Join(in, `","`) + `"]`
}

func randIpv4NetWithPrefixLen(t testing.TB, cidrBlock string, cidrBits int) *net.IPNet {
	t.Helper()
	ip := randIpv4Address(t, cidrBlock)

	return &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(cidrBits, 32),
	}
}

func randIpv6NetWithPrefixLen(t testing.TB, cidrBlock string, cidrBits int) *net.IPNet {
	t.Helper()
	ip := randIpv6Address(t, cidrBlock)

	return &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(cidrBits, 128),
	}
}

func randIpv4Net(t testing.TB, cidrBlock string) *net.IPNet {
	t.Helper()
	ip := randIpv4Address(t, cidrBlock)

	_, ipNet, _ := net.ParseCIDR(cidrBlock)
	cidrBlockPrefixLen, _ := ipNet.Mask.Size()
	targetPrefixLen := rand.Intn(31-cidrBlockPrefixLen) + cidrBlockPrefixLen

	_, result, _ := net.ParseCIDR(fmt.Sprintf("%s/%d", ip.String(), targetPrefixLen))

	return result
}

func randIpv4Address(t testing.TB, cidrBlock string) net.IP {
	t.Helper()

	s, err := acctest.RandIpAddress(cidrBlock)
	require.NoError(t, err)

	ip := net.ParseIP(s)
	if ip == nil {
		t.Fatalf("randIpv4Address failed to parse IP address %q", s)
	}

	return ip
}

func randIpv6Address(t testing.TB, cidrBlock string) net.IP {
	t.Helper()

	s, err := acctest.RandIpAddress(cidrBlock)
	require.NoError(t, err)

	ip := net.ParseIP(s)
	if ip == nil {
		t.Fatalf("randIpv6Address failed to parse IP address %q", s)
	}

	return ip
}

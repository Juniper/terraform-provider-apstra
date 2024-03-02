package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"golang.org/x/exp/constraints"
)

func systemIds(ctx context.Context, t *testing.T, bp *apstra.TwoStageL3ClosClient, role string) []string {
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

func intPtrOrNull[A constraints.Integer](in *A) string {
	if in == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *in)
}

func stringSetOrNull(in []string) string {
	if in == nil {
		return "null"
	}

	sb := new(strings.Builder)
	for i, s := range in {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("%q", s))
		} else {
			sb.WriteString(fmt.Sprintf(", %q", s))
		}
	}
	return "[ " + sb.String() + " ]"
}

func randIpv4NetMust(t *testing.T, cidrBlock string) *net.IPNet {
	ip := randIpv4AddressMust(t, cidrBlock)

	_, ipNet, _ := net.ParseCIDR(cidrBlock)
	cidrBlockPrefixLen, _ := ipNet.Mask.Size()
	targetPrefixLen := rand.Intn(31-cidrBlockPrefixLen) + cidrBlockPrefixLen

	_, result, _ := net.ParseCIDR(fmt.Sprintf("%s/%d", ip.String(), targetPrefixLen))

	return result
}

func randIpv4AddressMust(t *testing.T, cidrBlock string) net.IP {
	s, err := acctest.RandIpAddress(cidrBlock)
	if err != nil {
		t.Fatal(err)
	}

	ip := net.ParseIP(s)
	if ip == nil {
		t.Fatalf("randIpv4AddressMust failed to parse IP address %q", s)
	}

	return ip
}

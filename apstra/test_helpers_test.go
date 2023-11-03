package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"golang.org/x/exp/constraints"
	"testing"
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

//func stringOrNull(in string) string {
//	if in == "" {
//		return "null"
//	}
//	return `"` + in + `"`
//}

func intPtrOrNull[A constraints.Integer](in *A) string {
	if in == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *in)
}

//func ipOrNull(in *net.IPNet) string {
//	if in == nil {
//		return "null"
//	}
//	return `"` + in.String() + `"`
//}

//func randIpAddressMust(t *testing.T, cidrBlock string) net.IP {
//	s, err := acctest.RandIpAddress(cidrBlock)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ip := net.ParseIP(s)
//	if ip == nil {
//		t.Fatalf("randIpAddressMust failed to parse IP address %q", s)
//	}
//
//	return ip
//}

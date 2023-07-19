package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

func randomIpv4Net(prefixLength int) net.IPNet {
	if prefixLength < 0 || prefixLength > 24 {
		panic("ipv4 prefix length " + strconv.Itoa(prefixLength) + " is invalid.")
	}

	octets := []string{
		strconv.Itoa(rand.Intn(223)),
		strconv.Itoa(rand.Intn(255)),
		strconv.Itoa(rand.Intn(255)),
		strconv.Itoa(rand.Intn(255)),
	}
	_, ipNet, err := net.ParseCIDR(strings.Join(octets, ".") + "/" + strconv.Itoa(prefixLength))
	if err != nil {
		panic("ipNet returned an error: " + err.Error())
	}

	return *ipNet
}

func RoutingPolicyA(ctx context.Context, client *apstra.TwoStageL3ClosClient) (apstra.ObjectId, func(context.Context) error, error) {
	deleteFunc := func(_ context.Context) error { return nil }
	id, err := client.CreateRoutingPolicy(ctx, &apstra.DcRoutingPolicyData{
		Label:        acctest.RandString(10),
		Description:  acctest.RandString(10),
		PolicyType:   apstra.DcRoutingPolicyTypeUser,
		ImportPolicy: apstra.DcRoutingPolicyImportPolicyAll,
		ExportPolicy: apstra.DcRoutingExportPolicy{
			StaticRoutes:         acctest.RandInt()%2 == 0,
			Loopbacks:            acctest.RandInt()%2 == 0,
			SpineSuperspineLinks: acctest.RandInt()%2 == 0,
			L3EdgeServerLinks:    acctest.RandInt()%2 == 0,
			SpineLeafLinks:       acctest.RandInt()%2 == 0,
			L2EdgeSubnets:        acctest.RandInt()%2 == 0,
		},
		ExpectDefaultIpv4Route: acctest.RandInt()%2 == 0,
		ExpectDefaultIpv6Route: acctest.RandInt()%2 == 0,
		AggregatePrefixes:      []net.IPNet{randomIpv4Net(8)},
	})
	if err != nil {
		return "", deleteFunc, err
	}
	deleteFunc = func(ctx context.Context) error {
		return client.DeleteRoutingPolicy(ctx, id)
	}

	return id, deleteFunc, nil
}

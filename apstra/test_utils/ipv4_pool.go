package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"math"
	"math/big"
	"net"
	"testing"
)

func Ipv4PoolA(t testing.TB, ctx context.Context) *apstra.IpPool {
	client := GetTestClient(t, ctx)

	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	request := apstra.NewIpPoolRequest{
		DisplayName: name,
		Subnets: []apstra.NewIpSubnet{
			{Network: "192.168.0.0/16"},
		},
	}

	id, err := client.CreateIp4Pool(ctx, &request)
	require.NoError(t, err)

	var z *big.Int
	var ok bool

	bigZero := new(big.Int)
	z, ok = bigZero.SetString("0", 10)
	if z == nil || !ok {
		t.Fatal("error setting 'bigZero' value")
	}

	subnets := make([]apstra.IpSubnet, len(request.Subnets))
	total := new(big.Int)
	for i := range request.Subnets {
		_, n, err := net.ParseCIDR(request.Subnets[i].Network)
		require.NoError(t, err)

		subnets[i] = apstra.IpSubnet{
			Network:        n,
			Status:         "pool_element_available",
			Used:           *bigZero,
			UsedPercentage: 0,
		}
		maskOnes, maskBits := n.Mask.Size()
		subnets[i].Total.SetInt64(int64(math.Pow(2, float64(maskBits-maskOnes))))
		total.Add(total, &subnets[i].Total)
	}

	pool, err := client.GetIp4Pool(ctx, id)
	require.NoError(t, err)

	return pool
}

func Ipv4PoolB(t testing.TB, ctx context.Context) *apstra.IpPool {
	client := GetTestClient(t, ctx)

	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	request := apstra.NewIpPoolRequest{
		DisplayName: name,
		Subnets: []apstra.NewIpSubnet{
			{Network: "192.168.0.0/24"},
			{Network: "192.168.1.0/24"},
			{Network: "192.168.2.0/23"},
		},
	}
	id, err := client.CreateIp4Pool(ctx, &request)
	require.NoError(t, err)

	var z *big.Int
	var ok bool

	bigZero := new(big.Int)
	z, ok = bigZero.SetString("0", 10)
	if z == nil || !ok {
		t.Fatal("error setting 'bigZero' value")
	}

	subnets := make([]apstra.IpSubnet, len(request.Subnets))
	total := new(big.Int)
	for i := range request.Subnets {
		_, n, err := net.ParseCIDR(request.Subnets[i].Network)
		require.NoError(t, err)

		subnets[i] = apstra.IpSubnet{
			Network:        n,
			Status:         "pool_element_available",
			Used:           *bigZero,
			UsedPercentage: 0,
		}
		maskOnes, maskBits := n.Mask.Size()
		subnets[i].Total.SetInt64(int64(math.Pow(2, float64(maskBits-maskOnes))))
		total.Add(total, &subnets[i].Total)
	}

	pool, err := client.GetIp4Pool(ctx, id)
	require.NoError(t, err)

	return pool
}

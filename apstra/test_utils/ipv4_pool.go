package testutils

import (
	"context"
	"errors"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"math"
	"math/big"
	"net"
)

func Ipv4PoolA(ctx context.Context) (*apstra.IpPool, func(context.Context) error, error) {
	client, err := GetTestClient()
	if err != nil {
		return nil, nil, err
	}

	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	request := apstra.NewIpPoolRequest{
		DisplayName: name,
		Subnets: []apstra.NewIpSubnet{
			{Network: "192.168.0.0/16"},
		},
	}
	id, err := client.CreateIp4Pool(ctx, &request)
	if err != nil {
		return nil, nil, err
	}

	var z *big.Int
	var ok bool

	bigZero := new(big.Int)
	z, ok = bigZero.SetString("0", 10)
	if z == nil || !ok {
		return nil, nil, errors.New("error setting 'bigZero' value")
	}

	subnets := make([]apstra.IpSubnet, len(request.Subnets))
	total := new(big.Int)
	for i := range request.Subnets {
		_, net, err := net.ParseCIDR(request.Subnets[i].Network)
		if err != nil {
			return nil, nil, err
		}
		subnets[i] = apstra.IpSubnet{
			Network:        net,
			Status:         "pool_element_available",
			Used:           *bigZero,
			UsedPercentage: 0,
		}
		maskOnes, maskBits := net.Mask.Size()
		subnets[i].Total.SetInt64(int64(math.Pow(2, float64(maskBits-maskOnes))))
		total.Add(total, &subnets[i].Total)
	}

	pool := apstra.IpPool{
		Id:             id,
		DisplayName:    name,
		Status:         apstra.PoolStatusUnused,
		Used:           *bigZero,
		Total:          *total,
		UsedPercentage: 0,
		Subnets:        subnets,
	}

	deleteFunc := func(ctx context.Context) error {
		err := client.DeleteIp4Pool(ctx, id)
		if err != nil {
			return err
		}
		return nil
	}

	return &pool, deleteFunc, nil
}

func Ipv4PoolB(ctx context.Context) (*apstra.IpPool, func(context.Context) error, error) {
	client, err := GetTestClient()
	if err != nil {
		return nil, nil, err
	}

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
	if err != nil {
		return nil, nil, err
	}

	var z *big.Int
	var ok bool

	bigZero := new(big.Int)
	z, ok = bigZero.SetString("0", 10)
	if z == nil || !ok {
		return nil, nil, errors.New("error setting 'bigZero' value")
	}

	subnets := make([]apstra.IpSubnet, len(request.Subnets))
	total := new(big.Int)
	for i := range request.Subnets {
		_, net, err := net.ParseCIDR(request.Subnets[i].Network)
		if err != nil {
			return nil, nil, err
		}
		subnets[i] = apstra.IpSubnet{
			Network:        net,
			Status:         "pool_element_available",
			Used:           *bigZero,
			UsedPercentage: 0,
		}
		maskOnes, maskBits := net.Mask.Size()
		subnets[i].Total.SetInt64(int64(math.Pow(2, float64(maskBits-maskOnes))))
		total.Add(total, &subnets[i].Total)
	}

	pool := apstra.IpPool{
		Id:             id,
		DisplayName:    name,
		Status:         apstra.PoolStatusUnused,
		Used:           *bigZero,
		Total:          *total,
		UsedPercentage: 0,
		Subnets:        subnets,
	}

	deleteFunc := func(ctx context.Context) error {
		err := client.DeleteIp4Pool(ctx, id)
		if err != nil {
			return err
		}
		return nil
	}

	return &pool, deleteFunc, nil
}

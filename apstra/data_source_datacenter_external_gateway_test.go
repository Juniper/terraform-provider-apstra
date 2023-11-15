package tfapstra_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"math/rand"
	"testing"
)

const (
	dataSourceDataCenterExternalGatewayByIdHCL = `
data "apstra_datacenter_external_gateway" "test" {
  blueprint_id = "%s"
  id = "%s"
}
`

	dataSourceDataCenterExternalGatewayByNameHCL = `
data "apstra_datacenter_external_gateway" "test" {
  blueprint_id = "%s"
  name = "%s"
}
`
)

func TestDatacenterExternalGateway(t *testing.T) {
	ctx := context.Background()

	bp, bpDelete, err := testutils.BlueprintC(ctx)
	if err != nil {
		t.Fatal(errors.Join(err, bpDelete(ctx)))
	}
	defer func() {
		err = bpDelete(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	leafIdStrings := systemIds(ctx, t, bp, "leaf")
	leafIds := make([]apstra.ObjectId, len(leafIdStrings))
	for i, id := range leafIdStrings {
		leafIds[i] = apstra.ObjectId(id)
	}

	ttl := uint8(5)
	keepalive := uint16(6)
	hold := uint16(18)
	password := "big secret"

	rgConfigs := []apstra.RemoteGatewayData{
		{
			RouteTypes:     apstra.RemoteGatewayRouteTypesAll,
			LocalGwNodes:   leafIds,
			GwAsn:          rand.Uint32(),
			GwIp:           randIpv4AddressMust(t, "10.0.0.0/8"),
			GwName:         acctest.RandString(5),
			Ttl:            &ttl,
			KeepaliveTimer: &keepalive,
			HoldtimeTimer:  &hold,
			Password:       &password,
		},
		{
			RouteTypes:     apstra.RemoteGatewayRouteTypesFiveOnly,
			LocalGwNodes:   leafIds,
			GwAsn:          rand.Uint32(),
			GwIp:           randIpv4AddressMust(t, "10.0.0.0/8"),
			GwName:         acctest.RandString(5),
			Ttl:            &ttl,
			KeepaliveTimer: &keepalive,
			HoldtimeTimer:  &hold,
			Password:       &password,
		},
	}

	rgIds := make([]apstra.ObjectId, len(rgConfigs))
	for i, rgConfig := range rgConfigs {
		rgIds[i], err = bp.CreateRemoteGateway(ctx, &rgConfig)
		if err != nil {
			t.Fatal(err)
		}
	}

	rgs := make([]apstra.RemoteGateway, len(rgIds))
	for i, rgData := range rgConfigs {
		rgData := rgData
		rgs[i] = apstra.RemoteGateway{
			Id:   rgIds[i],
			Data: &rgData,
		}
	}

	genTestCheckFuncs := func(rg apstra.RemoteGateway) []resource.TestCheckFunc {
		result := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "id", rg.Id.String()),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "blueprint_id", bp.Id().String()),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "name", rg.Data.GwName),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "ip_address", rg.Data.GwIp.String()),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "asn", fmt.Sprintf("%d", rg.Data.GwAsn)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "ttl", fmt.Sprintf("%d", *rg.Data.Ttl)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "keepalive_time", fmt.Sprintf("%d", *rg.Data.KeepaliveTimer)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "hold_time", fmt.Sprintf("%d", *rg.Data.HoldtimeTimer)),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "evpn_route_types", rg.Data.RouteTypes.Value),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "local_gateway_nodes.#", fmt.Sprintf("%d", len(rg.Data.LocalGwNodes))),
			resource.TestCheckResourceAttr("data.apstra_datacenter_external_gateway.test", "password", *rg.Data.Password),
		}

		for _, id := range leafIdStrings {
			tcf := resource.TestCheckTypeSetElemAttr(
				"data.apstra_datacenter_external_gateway.test",
				"local_gateway_nodes.*", id,
			)
			result = append(result, tcf)
		}

		return result
	}

	testCheckFuncsByRgId := make(map[apstra.ObjectId][]resource.TestCheckFunc, len(rgs))
	for _, rg := range rgs {
		testCheckFuncsByRgId[rg.Id] = genTestCheckFuncs(rg)
	}

	stepsById := make([]resource.TestStep, len(rgs))
	for i, rg := range rgs {
		stepsById[i] = resource.TestStep{
			Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDataCenterExternalGatewayByIdHCL, bp.Id(), rg.Id),
			Check:  resource.ComposeAggregateTestCheckFunc(testCheckFuncsByRgId[rg.Id]...),
		}
	}

	stepsByName := make([]resource.TestStep, len(rgs))
	for i, rg := range rgs {
		stepsByName[i] = resource.TestStep{
			Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDataCenterExternalGatewayByNameHCL, bp.Id(), rg.Data.GwName),
			Check:  resource.ComposeAggregateTestCheckFunc(genTestCheckFuncs(rg)...),
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		//Steps:                    stepsById,
		Steps: append(stepsById, stepsByName...),
	})
}

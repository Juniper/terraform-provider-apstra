//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceDataCenterExternalGatewayHCL = `
resource %q %q {
  blueprint_id        = %q
  name                = %q
  ip_address          = %q
  asn                 = %d
  evpn_route_types    = %s
  local_gateway_nodes = %s
  ttl                 = %s
  keepalive_time      = %s
  hold_time           = %s
  password            = %s
}
`
)

type resourceDataCenterExternalGateway struct {
	blueprintID   apstra.ObjectId
	name          string
	ipAddress     net.IP
	asn           uint32
	routeTypes    *enum.RemoteGatewayRouteType
	nodes         []string
	ttl           *uint8
	keepaliveTime *uint16
	holdTime      *uint16
	password      string
}

func (o resourceDataCenterExternalGateway) render(rType, rName string) string {
	return fmt.Sprintf(resourceDataCenterExternalGatewayHCL,
		rType, rName,
		o.blueprintID,
		o.name,
		o.ipAddress.String(),
		o.asn,
		stringerOrNull(o.routeTypes),
		stringSliceOrNull(o.nodes),
		intPtrOrNull(o.ttl),
		intPtrOrNull(o.keepaliveTime),
		intPtrOrNull(o.holdTime),
		stringOrNull(o.password),
	)
}

func (o resourceDataCenterExternalGateway) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// ensure ID has been set
	result.append(t, "TestCheckResourceAttrSet", "id")

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "ip_address", o.ipAddress.String())
	result.append(t, "TestCheckResourceAttr", "asn", strconv.Itoa(int(o.asn)))

	if o.routeTypes == nil {
		result.append(t, "TestCheckResourceAttr", "evpn_route_types", enum.RemoteGatewayRouteTypeAll.Value)
	} else {
		result.append(t, "TestCheckResourceAttr", "evpn_route_types", o.routeTypes.String())
	}

	result.append(t, "TestCheckResourceAttr", "local_gateway_nodes.#", strconv.Itoa(len(o.nodes)))
	for _, node := range o.nodes {
		result.append(t, "TestCheckTypeSetElemAttr", "local_gateway_nodes.*", node)
	}

	if o.ttl == nil {
		result.append(t, "TestCheckResourceAttr", "ttl", "30")
	} else {
		result.append(t, "TestCheckResourceAttr", "ttl", strconv.Itoa(int(*o.ttl)))
	}

	if o.keepaliveTime == nil {
		result.append(t, "TestCheckResourceAttr", "keepalive_time", "10")
	} else {
		result.append(t, "TestCheckResourceAttr", "keepalive_time", strconv.Itoa(int(*o.keepaliveTime)))
	}

	if o.holdTime == nil {
		result.append(t, "TestCheckResourceAttr", "hold_time", "30")
	} else {
		result.append(t, "TestCheckResourceAttr", "hold_time", strconv.Itoa(int(*o.holdTime)))
	}

	if o.password == "" {
		result.append(t, "TestCheckNoResourceAttr", "password")
	} else {
		result.append(t, "TestCheckResourceAttr", "password", o.password)
	}

	return result
}

func TestResourceDatacenterExternalGateway(t *testing.T) {
	ctx := context.Background()

	var bp, bp6 *apstra.TwoStageL3ClosClient
	var leafIDs, leafIDs6 []string

	bp = testutils.BlueprintC(t, ctx)
	leafIDs = systemIds(ctx, t, bp, "leaf")

	if compatibility.BPDefaultRoutingZoneAddressingOK.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
		bp6 = testutils.BlueprintJ(t, ctx)
		leafIDs6 = systemIds(ctx, t, bp6, "leaf")
	}

	type testStep struct {
		config resourceDataCenterExternalGateway
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
		bp                 *apstra.TwoStageL3ClosClient
		leafIDs            []string
	}

	testCases := map[string]testCase{
		"ipv6": {
			versionConstraints: compatibility.BPDefaultRoutingZoneAddressingOK.Constraints,
			bp:                 bp6,
			steps: []testStep{
				{
					config: resourceDataCenterExternalGateway{
						name:      acctest.RandString(6),
						ipAddress: randIpvAddressMust(t, "3fff::/64"),
						asn:       uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:     leafIDs6,
					},
				},
			},
		},
		"minimal_with_changes": {
			bp: bp,
			steps: []testStep{
				{
					config: resourceDataCenterExternalGateway{
						name:      acctest.RandString(6),
						ipAddress: randIpvAddressMust(t, "10.0.0.0/8"),
						asn:       uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:     []string{leafIDs[rand.Intn(len(leafIDs))]},
					},
				},
				{
					config: resourceDataCenterExternalGateway{
						name:      acctest.RandString(6),
						ipAddress: randIpvAddressMust(t, "10.0.0.0/8"),
						asn:       uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:     []string{leafIDs[rand.Intn(len(leafIDs))]},
					},
				},
			},
		},
		"start_minimal": {
			bp: bp,
			steps: []testStep{
				{
					config: resourceDataCenterExternalGateway{
						name:      acctest.RandString(6),
						ipAddress: randIpvAddressMust(t, "10.0.0.0/8"),
						asn:       uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:     []string{leafIDs[rand.Intn(len(leafIDs))]},
					},
				},
				{
					config: resourceDataCenterExternalGateway{
						name:          acctest.RandString(6),
						ipAddress:     randIpvAddressMust(t, "10.0.0.0/8"),
						asn:           uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						routeTypes:    pointer.To(enum.RemoteGatewayRouteTypeFiveOnly),
						nodes:         leafIDs,
						ttl:           pointer.To(uint8(rand.Intn(254) + 2)),               // 2-255
						keepaliveTime: pointer.To(uint16(rand.Intn(math.MaxUint16) + 1)),   // 1-65535
						holdTime:      pointer.To(uint16(rand.Intn(math.MaxUint16-2) + 3)), // 3-65535
						password:      acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterExternalGateway{
						name:          acctest.RandString(6),
						ipAddress:     randIpvAddressMust(t, "10.0.0.0/8"),
						asn:           uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						routeTypes:    pointer.To(enum.RemoteGatewayRouteTypeFiveOnly),
						nodes:         leafIDs,
						ttl:           pointer.To(uint8(rand.Intn(254) + 2)),               // 2-255
						keepaliveTime: pointer.To(uint16(rand.Intn(math.MaxUint16) + 1)),   // 1-65535
						holdTime:      pointer.To(uint16(rand.Intn(math.MaxUint16-2) + 3)), // 3-65535
						password:      acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterExternalGateway{
						name:      acctest.RandString(6),
						ipAddress: randIpvAddressMust(t, "10.0.0.0/8"),
						asn:       uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:     []string{leafIDs[rand.Intn(len(leafIDs))]},
					},
				},
			},
		},
		"start_maximal": {
			bp: bp,
			steps: []testStep{
				{
					config: resourceDataCenterExternalGateway{
						name:          acctest.RandString(6),
						ipAddress:     randIpvAddressMust(t, "10.0.0.0/8"),
						asn:           uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						routeTypes:    pointer.To(enum.RemoteGatewayRouteTypeFiveOnly),
						nodes:         leafIDs,
						ttl:           pointer.To(uint8(rand.Intn(254) + 2)),               // 2-255
						keepaliveTime: pointer.To(uint16(rand.Intn(math.MaxUint16) + 1)),   // 1-65535
						holdTime:      pointer.To(uint16(rand.Intn(math.MaxUint16-2) + 3)), // 3-65535
						password:      acctest.RandString(6),
					},
				},
				{
					config: resourceDataCenterExternalGateway{
						name:      acctest.RandString(6),
						ipAddress: randIpvAddressMust(t, "10.0.0.0/8"),
						asn:       uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:     []string{leafIDs[rand.Intn(len(leafIDs))]},
					},
				},
				{
					config: resourceDataCenterExternalGateway{
						name:      acctest.RandString(6),
						ipAddress: randIpvAddressMust(t, "10.0.0.0/8"),
						asn:       uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:     []string{leafIDs[rand.Intn(len(leafIDs))]},
					},
				},
				{
					config: resourceDataCenterExternalGateway{
						name:          acctest.RandString(6),
						ipAddress:     randIpvAddressMust(t, "10.0.0.0/8"),
						asn:           uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						routeTypes:    pointer.To(enum.RemoteGatewayRouteTypeFiveOnly),
						nodes:         leafIDs,
						ttl:           pointer.To(uint8(rand.Intn(254) + 2)),               // 2-255
						keepaliveTime: pointer.To(uint16(rand.Intn(math.MaxUint16) + 1)),   // 1-65535
						holdTime:      pointer.To(uint16(rand.Intn(math.MaxUint16-2) + 3)), // 3-65535
						password:      acctest.RandString(6),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterExternalGateway)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			if tCase.bp == nil {
				t.Skip("test case has no blueprint - probably a version incompatibility")
			}
			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(tCase.bp.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				step.config.blueprintID = tCase.bp.Id()
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, tCase.bp.Id(), resourceType, tName)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checks.checks...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}

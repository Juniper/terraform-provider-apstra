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
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceDataCenterInterconnectDomainGatewayHCL = `
resource %q %q {
  blueprint_id           = %q
  name                   = %q
  ip_address             = %q
  asn                    = %d
  interconnect_domain_id = %q
  local_gateway_nodes    = %s
  ttl                    = %s
  keepalive_time         = %s
  hold_time              = %s
  password               = %s
}
`
)

type resourceDataCenterInterconnectDomainGateway struct {
	blueprintID          apstra.ObjectId
	name                 string
	ipAddress            net.IP
	asn                  uint32
	interconnectDomainId apstra.ObjectId
	nodes                []string
	ttl                  *uint8
	keepaliveTime        *uint16
	holdTime             *uint16
	password             string
}

func (o resourceDataCenterInterconnectDomainGateway) render(rType, rName string) string {
	return fmt.Sprintf(resourceDataCenterInterconnectDomainGatewayHCL,
		rType, rName,
		o.blueprintID,
		o.name,
		o.ipAddress.String(),
		o.asn,
		o.interconnectDomainId,
		stringSliceOrNull(o.nodes),
		intPtrOrNull(o.ttl),
		intPtrOrNull(o.keepaliveTime),
		intPtrOrNull(o.holdTime),
		stringOrNull(o.password),
	)
}

func (o resourceDataCenterInterconnectDomainGateway) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// ensure ID has been set
	result.append(t, "TestCheckResourceAttrSet", "id")

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "ip_address", o.ipAddress.String())
	result.append(t, "TestCheckResourceAttr", "asn", strconv.Itoa(int(o.asn)))
	result.append(t, "TestCheckResourceAttr", "interconnect_domain_id", o.interconnectDomainId.String())

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

func TestResourceDatacenterInterconnectDomainGateway(t *testing.T) {
	ctx := context.Background()

	bp := testutils.BlueprintC(t, ctx)

	interconnectDomainIds := make([]apstra.ObjectId, 2)
	for i := range interconnectDomainIds {
		id, err := bp.CreateEvpnInterconnectGroup(ctx, &apstra.EvpnInterconnectGroupData{
			Label:       acctest.RandStringFromCharSet(6, acctest.CharSetAlpha),
			RouteTarget: randomRT(t),
		})
		require.NoError(t, err)
		interconnectDomainIds[i] = id
	}

	leafIds := systemIds(ctx, t, bp, "leaf")

	type testStep struct {
		config resourceDataCenterInterconnectDomainGateway
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"minimal_with_changes": {
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                []string{leafIds[rand.Intn(len(leafIds))]},
						interconnectDomainId: interconnectDomainIds[0],
					},
				},
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                []string{leafIds[rand.Intn(len(leafIds))]},
						interconnectDomainId: interconnectDomainIds[1],
					},
				},
			},
		},
		"start_minimal": {
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                []string{leafIds[rand.Intn(len(leafIds))]},
						interconnectDomainId: interconnectDomainIds[0],
					},
				},
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                leafIds,
						ttl:                  utils.ToPtr(uint8(rand.Intn(254) + 2)),               // 2-255
						keepaliveTime:        utils.ToPtr(uint16(rand.Intn(math.MaxUint16) + 1)),   // 1-65535
						holdTime:             utils.ToPtr(uint16(rand.Intn(math.MaxUint16-2) + 3)), // 3-65535
						password:             acctest.RandString(6),
						interconnectDomainId: interconnectDomainIds[1],
					},
				},
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                leafIds,
						ttl:                  utils.ToPtr(uint8(rand.Intn(254) + 2)),               // 2-255
						keepaliveTime:        utils.ToPtr(uint16(rand.Intn(math.MaxUint16) + 1)),   // 1-65535
						holdTime:             utils.ToPtr(uint16(rand.Intn(math.MaxUint16-2) + 3)), // 3-65535
						password:             acctest.RandString(6),
						interconnectDomainId: interconnectDomainIds[0],
					},
				},
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                []string{leafIds[rand.Intn(len(leafIds))]},
						interconnectDomainId: interconnectDomainIds[1],
					},
				},
			},
		},
		"start_maximal": {
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                leafIds,
						ttl:                  utils.ToPtr(uint8(rand.Intn(254) + 2)),               // 2-255
						keepaliveTime:        utils.ToPtr(uint16(rand.Intn(math.MaxUint16) + 1)),   // 1-65535
						holdTime:             utils.ToPtr(uint16(rand.Intn(math.MaxUint16-2) + 3)), // 3-65535
						password:             acctest.RandString(6),
						interconnectDomainId: interconnectDomainIds[0],
					},
				},
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                []string{leafIds[rand.Intn(len(leafIds))]},
						interconnectDomainId: interconnectDomainIds[1],
					},
				},
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                []string{leafIds[rand.Intn(len(leafIds))]},
						interconnectDomainId: interconnectDomainIds[0],
					},
				},
				{
					config: resourceDataCenterInterconnectDomainGateway{
						blueprintID:          bp.Id(),
						name:                 acctest.RandString(6),
						ipAddress:            randIpvAddressMust(t, "10.0.0.0/8"),
						asn:                  uint32(rand.Intn(math.MaxUint32) + 1), // not zero
						nodes:                leafIds,
						ttl:                  utils.ToPtr(uint8(rand.Intn(254) + 2)),               // 2-255
						keepaliveTime:        utils.ToPtr(uint16(rand.Intn(math.MaxUint16) + 1)),   // 1-65535
						holdTime:             utils.ToPtr(uint16(rand.Intn(math.MaxUint16-2) + 3)), // 3-65535
						password:             acctest.RandString(6),
						interconnectDomainId: interconnectDomainIds[1],
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterInterconnectDomainGateway)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, bp.Id(), resourceType, tName)

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

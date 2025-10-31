//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/compatibility"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const resourceDatacenterRoutingZoneLoopbackAddressHCL = `{
      ipv4_addr = %s
      ipv6_addr = %s
    }
`

type resourceDatacenterRoutingZoneLoopbackAddress struct {
	iPv4Addr string
	iPv6Addr string
}

func (o resourceDatacenterRoutingZoneLoopbackAddress) render() string {
	return fmt.Sprintf(resourceDatacenterRoutingZoneLoopbackAddressHCL,
		stringOrNull(o.iPv4Addr),
		stringOrNull(o.iPv6Addr),
	)
}

const resourceDatacenterRoutingZoneLoopbackAddressesHCL = `resource %q %q {
  blueprint_id    = %q
  routing_zone_id = %q
  loopbacks       = %s
}
`

type resourceDatacenterRoutingZoneLoopbackAddresses struct {
	blueprintID   string
	routingZoneID string
	loopbacks     map[string]resourceDatacenterRoutingZoneLoopbackAddress
}

func (o resourceDatacenterRoutingZoneLoopbackAddresses) render(rType, rName string) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	for k, v := range o.loopbacks {
		sb.WriteString(fmt.Sprintf("    %q = %s", k, v.render()))
	}
	sb.WriteString("  }")

	return fmt.Sprintf(resourceDatacenterRoutingZoneLoopbackAddressesHCL,
		rType, rName,
		o.blueprintID,
		o.routingZoneID,
		sb.String(),
	)
}

func (o resourceDatacenterRoutingZoneLoopbackAddresses) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	result.append(t, "TestCheckResourceAttr", "routing_zone_id", o.routingZoneID)

	result.append(t, "TestCheckResourceAttr", "loopbacks.%", strconv.Itoa(len(o.loopbacks)))
	for k, v := range o.loopbacks {
		if v.iPv4Addr != "" {
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("loopbacks.%s.ipv4_addr", k), v.iPv4Addr)
		} else {
			result.append(t, "TestCheckNoResourceAttr", fmt.Sprintf("loopbacks.%s.ipv4_addr", k))
		}

		if v.iPv6Addr != "" {
			result.append(t, "TestCheckResourceAttr", fmt.Sprintf("loopbacks.%s.ipv6_addr", k), v.iPv6Addr)
		} else {
			result.append(t, "TestCheckNoResourceAttr", fmt.Sprintf("loopbacks.%s.ipv6_addr", k))
		}

		result.append(t, "TestCheckResourceAttrSet", fmt.Sprintf("loopback_ids.%s", k))
	}

	return result
}

func TestResourceDatacenterRoutingZoneLoopbackAddresses(t *testing.T) {
	ctx := context.Background()
	cleanup := true

	client := testutils.GetTestClient(t, ctx)
	if !compatibility.SecurityZoneLoopbackApiSupported.Check(version.Must(version.NewVersion(client.ApiVersion()))) {
		t.Skipf("skipping test due to version %s", client.ApiVersion())
	}

	// create a blueprint
	bp := testutils.BlueprintG(t, ctx, cleanup)

	// enable ipv6
	settings, err := bp.GetFabricSettings(ctx)
	require.NoError(t, err)
	settings.Ipv6Enabled = pointer.To(true)
	require.NoError(t, bp.SetFabricSettings(ctx, settings))

	// create a routing zone
	rzLabel := acctest.RandString(6)
	rzId, err := bp.CreateSecurityZone(ctx, &apstra.SecurityZoneData{
		Label:   rzLabel,
		SzType:  apstra.SecurityZoneTypeEVPN,
		VrfName: rzLabel,
	})
	require.NoError(t, err)

	// discover leaf switch IDs
	leafs := testutils.GetSystemIds(t, ctx, bp, "leaf")

	// create a VN to ensure the leaf switches participate in the RZ
	vnBindings := make([]apstra.VnBinding, len(leafs))
	var i int
	for _, leafId := range leafs {
		vnBindings[i] = apstra.VnBinding{SystemId: leafId}
		i++
	}
	_, err = bp.CreateVirtualNetwork(ctx, &apstra.VirtualNetworkData{
		Label:          acctest.RandString(6),
		SecurityZoneId: rzId,
		VnType:         enum.VnTypeVxlan,
		VnBindings:     vnBindings,
	})
	require.NoError(t, err)

	// make a slice of leafIds
	i = 0
	leafIds := make([]apstra.ObjectId, len(leafs))
	for _, v := range leafs {
		leafIds[i] = v
		i++
	}

	type testStep struct {
		config resourceDatacenterRoutingZoneLoopbackAddresses
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"empty_populated_empty": {
			steps: []testStep{
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[0].String(): {
							iPv4Addr: "",
							iPv6Addr: "",
						},
					},
				}},
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[0].String(): {
							iPv4Addr: pointer.To(randomPrefix(t, "10.1.0.0/16", 32)).String(),
							iPv6Addr: pointer.To(randomPrefix(t, "3fff:1::/32", 128)).String(),
						},
					},
				}},
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[0].String(): {
							iPv4Addr: "",
							iPv6Addr: "",
						},
					},
				}},
			},
		},
		"populated_empty_populated": {
			steps: []testStep{
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[1].String(): {
							iPv4Addr: pointer.To(randomPrefix(t, "10.2.0.0/16", 32)).String(),
							iPv6Addr: pointer.To(randomPrefix(t, "3fff:2::/32", 128)).String(),
						},
					},
				}},
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[1].String(): {
							iPv4Addr: "",
							iPv6Addr: "",
						},
					},
				}},
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[1].String(): {
							iPv4Addr: pointer.To(randomPrefix(t, "10.2.0.0/16", 32)).String(),
							iPv6Addr: pointer.To(randomPrefix(t, "3fff:2::/32", 128)).String(),
						},
					},
				}},
			},
		},
		"two_leafs_at_once_random_values": {
			steps: []testStep{
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[2].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.30.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:30::/32", 128)).String(), ""),
						},
						leafIds[3].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.31.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:31::/32", 128)).String(), ""),
						},
					},
				}},
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[2].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.30.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:30::/32", 128)).String(), ""),
						},
						leafIds[3].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.31.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:31::/32", 128)).String(), ""),
						},
					},
				}},
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[2].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.30.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:30::/32", 128)).String(), ""),
						},
						leafIds[3].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.31.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:31::/32", 128)).String(), ""),
						},
					},
				}},
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[2].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.30.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:30::/32", 128)).String(), ""),
						},
						leafIds[3].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.31.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:31::/32", 128)).String(), ""),
						},
					},
				}},
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[2].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.30.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:30::/32", 128)).String(), ""),
						},
						leafIds[3].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.31.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:31::/32", 128)).String(), ""),
						},
					},
				}},
				{config: resourceDatacenterRoutingZoneLoopbackAddresses{
					blueprintID:   bp.Id().String(),
					routingZoneID: rzId.String(),
					loopbacks: map[string]resourceDatacenterRoutingZoneLoopbackAddress{
						leafIds[2].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.30.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:30::/32", 128)).String(), ""),
						},
						leafIds[3].String(): {
							iPv4Addr: oneOf(pointer.To(randomPrefix(t, "10.31.0.0/16", 32)).String(), ""),
							iPv6Addr: oneOf(pointer.To(randomPrefix(t, "3fff:31::/32", 128)).String(), ""),
						},
					},
				}},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterRoutingZoneLoopbackAddresses)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

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

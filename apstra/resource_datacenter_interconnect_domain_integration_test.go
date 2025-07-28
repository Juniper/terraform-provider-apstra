//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
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
	resourceDatacenterInterconnectDomainHCL = `
resource %q %q {
  blueprint_id = %q
  name         = %q
  route_target = %q
  esi_mac      = %s
}`
)

type resourceDatacenterInterconnectDomain struct {
	blueprintId apstra.ObjectId
	name        string
	routeTarget string
	esiMac      net.HardwareAddr
}

func (o resourceDatacenterInterconnectDomain) render(rType, rName string) string {
	return fmt.Sprintf(resourceDatacenterInterconnectDomainHCL,
		rType, rName,
		o.blueprintId,
		o.name,
		o.routeTarget,
		stringerOrNull(o.esiMac),
	)
}

func (o resourceDatacenterInterconnectDomain) testChecks(t testing.TB, bpId apstra.ObjectId, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// ensure ID has been set
	result.append(t, "TestCheckResourceAttrSet", "id")

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpId.String())
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "route_target", o.routeTarget)

	if o.esiMac == nil {
		result.append(t, "TestCheckResourceAttrSet", "esi_mac")
	} else {
		result.append(t, "TestCheckResourceAttr", "esi_mac", o.esiMac.String())
	}

	return result
}

func TestResourceDatacenterInterconnectDomain(t *testing.T) {
	ctx := context.Background()

	// create test blueprint
	bp := testutils.BlueprintA(t, ctx)

	// set the ESI MAC MSB
	fs, err := bp.GetFabricSettings(ctx)
	require.NoError(t, err)
	fs.EsiMacMsb = utils.ToPtr(uint8((rand.Int() & 254) | 2))
	err = bp.SetFabricSettings(ctx, fs)
	require.NoError(t, err)

	type step struct {
		config resourceDatacenterInterconnectDomain
	}

	type testCase struct {
		steps              []step
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"start_without_mac": {
			steps: []step{
				{
					config: resourceDatacenterInterconnectDomain{
						blueprintId: bp.Id(),
						name:        "a" + acctest.RandString(6),
						routeTarget: randomRT(t),
					},
				},
				{
					config: resourceDatacenterInterconnectDomain{
						blueprintId: bp.Id(),
						name:        "b" + acctest.RandString(6),
						routeTarget: randomRT(t),
						esiMac:      randomHardwareAddr([]byte{*fs.EsiMacMsb}, []byte{^*fs.EsiMacMsb}), // match policy MAC MSB
					},
				},
				{
					config: resourceDatacenterInterconnectDomain{
						blueprintId: bp.Id(),
						name:        "c" + acctest.RandString(6),
						routeTarget: randomRT(t),
						esiMac:      randomHardwareAddr([]byte{*fs.EsiMacMsb}, []byte{^*fs.EsiMacMsb}), // match policy MAC MSB
					},
				},
				{
					config: resourceDatacenterInterconnectDomain{
						blueprintId: bp.Id(),
						name:        "d" + acctest.RandString(6),
						routeTarget: randomRT(t),
					},
				},
			},
		},
		"start_with_mac": {
			steps: []step{
				{
					config: resourceDatacenterInterconnectDomain{
						blueprintId: bp.Id(),
						name:        "a" + acctest.RandString(6),
						routeTarget: randomRT(t),
						esiMac:      randomHardwareAddr([]byte{*fs.EsiMacMsb}, []byte{^*fs.EsiMacMsb}), // match policy MAC MSB
					},
				},
				{
					config: resourceDatacenterInterconnectDomain{
						blueprintId: bp.Id(),
						name:        "b" + acctest.RandString(6),
						routeTarget: randomRT(t),
					},
				},
				{
					config: resourceDatacenterInterconnectDomain{
						blueprintId: bp.Id(),
						name:        "c" + acctest.RandString(6),
						routeTarget: randomRT(t),
					},
				},
				{
					config: resourceDatacenterInterconnectDomain{
						blueprintId: bp.Id(),
						name:        "d" + acctest.RandString(6),
						routeTarget: randomRT(t),
						esiMac:      randomHardwareAddr([]byte{*fs.EsiMacMsb}, []byte{^*fs.EsiMacMsb}), // match policy MAC MSB
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterInterconnectDomain)

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

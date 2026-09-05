//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/compatibility"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	dctestobj "github.com/Juniper/terraform-provider-apstra/internal/test_utils/datacenter_test_objects"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceDataCenterInterconnectDomainConnectionTypeHCL = `resource %q %q {
  blueprint_id           = %q // required attribute
  interconnect_domain_id = %q // required attribute
  virtual_network_id     = %q // required attribute
  layer_2_enabled        = %t // required attribute
  layer_3_enabled        = %t // required attribute
  translation_vni        = %s // optional attribute
}
`

	datasourceDataCenterInterconnectDomainConnectionTypeHCL = `data %q %q {
  blueprint_id           = %q        // required attribute
  interconnect_domain_id = %q        // required attribute
  virtual_network_id     = %q        // required attribute
  depends_on             = [ %s.%s ] // ensure the data source runs after resource Create() or Update()
}
`
)

type resourceDataCenterInterconnectDomainConnectionType struct {
	InterconnectDomainID string
	VirtualNetworkID     string
	Layer2Enabled        bool
	Layer3Enabled        bool
	TranslationVNI       *uint32
}

func (ct *resourceDataCenterInterconnectDomainConnectionType) normalize() {
	if ct.TranslationVNI != nil && *ct.TranslationVNI > 0 {
		ct.Layer2Enabled = true // if translation_vni is set, layer_2_enabled must be true
	}
}

func (ct *resourceDataCenterInterconnectDomainConnectionType) render(rType, rName, bpID string) string {
	r := fmt.Sprintf(resourceDataCenterInterconnectDomainConnectionTypeHCL, rType, rName,
		bpID,
		ct.InterconnectDomainID,
		ct.VirtualNetworkID,
		ct.Layer2Enabled,
		ct.Layer3Enabled,
		intPtrOrNull(ct.TranslationVNI),
	)

	d := fmt.Sprintf(datasourceDataCenterInterconnectDomainConnectionTypeHCL, rType, rName,
		bpID,
		ct.InterconnectDomainID,
		ct.VirtualNetworkID,
		rType, rName,
	)

	return r + d
}

func (ct resourceDataCenterInterconnectDomainConnectionType) testChecks(t testing.TB, rType, rName, bpID string) []testChecks {
	rChecks := newTestChecks(rType + "." + rName)
	dChecks := newTestChecks("data." + rType + "." + rName)

	// required and computed attributes can always be checked
	rChecks.append(t, "TestCheckResourceAttr", "blueprint_id", bpID)
	dChecks.append(t, "TestCheckResourceAttr", "blueprint_id", bpID)
	rChecks.append(t, "TestCheckResourceAttr", "interconnect_domain_id", ct.InterconnectDomainID)
	dChecks.append(t, "TestCheckResourceAttr", "interconnect_domain_id", ct.InterconnectDomainID)
	rChecks.append(t, "TestCheckResourceAttr", "virtual_network_id", ct.VirtualNetworkID)
	dChecks.append(t, "TestCheckResourceAttr", "virtual_network_id", ct.VirtualNetworkID)
	rChecks.append(t, "TestCheckResourceAttr", "layer_2_enabled", strconv.FormatBool(ct.Layer2Enabled))
	dChecks.append(t, "TestCheckResourceAttr", "layer_2_enabled", strconv.FormatBool(ct.Layer2Enabled))
	rChecks.append(t, "TestCheckResourceAttr", "layer_3_enabled", strconv.FormatBool(ct.Layer3Enabled))
	dChecks.append(t, "TestCheckResourceAttr", "layer_3_enabled", strconv.FormatBool(ct.Layer3Enabled))
	if ct.TranslationVNI == nil {
		rChecks.append(t, "TestCheckNoResourceAttr", "translation_vni")
		dChecks.append(t, "TestCheckNoResourceAttr", "translation_vni")
	} else {
		rChecks.append(t, "TestCheckResourceAttr", "translation_vni", strconv.Itoa(int(*ct.TranslationVNI)))
		dChecks.append(t, "TestCheckResourceAttr", "translation_vni", strconv.Itoa(int(*ct.TranslationVNI)))

	}

	return []testChecks{rChecks, dChecks}
}

func TestACCResourceDatacenterInterconnectDomainConnectionType(t *testing.T) {
	ctx := context.Background()

	client := testutils.GetTestClient(t, ctx)
	ver := version.Must(version.NewVersion(client.ApiVersion()))
	if !compatibility.EmptyVnBindingsOk.Check(ver) {
		t.Skipf("skipping test because Apstra version %s does not support empty virtual network bindings", ver.String())
	}

	// Create a Blueprint, Routing Zone and DCI Domain
	bp := testutils.BlueprintA(t, ctx)
	rzID := dctestobj.RoutingZoneB(t, ctx, bp, false)
	dciID, err := bp.CreateEVPNInterconnectGroup(ctx, apstra.EVPNInterconnectGroup{
		Label:       pointer.To(acctest.RandString(6)),
		RouteTarget: pointer.To(random.RouteTarget(t)),
	})
	require.NoError(t, err)

	type testStep struct {
		config resourceDataCenterInterconnectDomainConnectionType
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	testCases := map[string]testCase{
		"random_x3": {
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        random.OneOf(true, false),
						Layer3Enabled:        random.OneOf(true, false),
						TranslationVNI:       random.OneOf((*uint32)(nil), pointer.To(uint32(rand.Intn(100000)+1))),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        random.OneOf(true, false),
						Layer3Enabled:        random.OneOf(true, false),
						TranslationVNI:       random.OneOf((*uint32)(nil), pointer.To(uint32(rand.Intn(100000)+1))),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        random.OneOf(true, false),
						Layer3Enabled:        random.OneOf(true, false),
						TranslationVNI:       random.OneOf((*uint32)(nil), pointer.To(uint32(rand.Intn(100000)+1))),
					},
				},
			},
		},
		"empty_empty": {
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        false,
						Layer3Enabled:        false,
					},
				},
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        false,
						Layer3Enabled:        false,
					},
				},
			},
		},
		"set_clear_set_translation_vni": {
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        true,
						Layer3Enabled:        random.OneOf(true, false),
						TranslationVNI:       pointer.To(uint32(rand.Intn(100000) + 1)),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        true,
						Layer3Enabled:        random.OneOf(true, false),
						TranslationVNI:       pointer.To(uint32(rand.Intn(100000) + 1)),
					},
				},
			},
		},
		"clear_set_clear_translation_vni": {
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        true,
						Layer3Enabled:        random.OneOf(true, false),
						TranslationVNI:       pointer.To(uint32(rand.Intn(100000) + 1)),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
					},
				},
			},
		},
		"min_max_translation_vni": {
			steps: []testStep{
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        true,
						Layer3Enabled:        true,
						TranslationVNI:       pointer.To(uint32(1)),
					},
				},
				{
					config: resourceDataCenterInterconnectDomainConnectionType{
						InterconnectDomainID: dciID,
						VirtualNetworkID:     dctestobj.VirtualNetworkA(t, ctx, bp, rzID),
						Layer2Enabled:        true,
						Layer3Enabled:        true,
						TranslationVNI:       pointer.To(uint32(1<<24 - 1)),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterInterconnectDomainConnectionType)
	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				step.config.normalize()
				config := step.config.render(resourceType, tName, string(bp.Id()))
				checks := step.config.testChecks(t, resourceType, tName, string(bp.Id()))

				var checkLog string
				var checkFuncs []resource.TestCheckFunc

				for _, checkList := range checks {
					checkLog = checkLog + checkList.string(len(checkFuncs))
					checkFuncs = append(checkFuncs, checkList.checks...)
				}

				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, checkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checkFuncs...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}

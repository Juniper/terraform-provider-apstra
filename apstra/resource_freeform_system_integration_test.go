//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceFreeformSystemHcl = `
resource %q %q {
  blueprint_id      = %q
  name              = %q
  device_profile_id = %s
  hostname          = %q
  type              = %q
  deploy_mode       = %s
  tags              = %s
}
`
)

type resourceFreeformSystem struct {
	blueprintId     string
	name            string
	deviceProfileId string
	hostname        string
	systemType      string
	deployMode      string
	tags            []string
}

func (o resourceFreeformSystem) render(rType, rName string) string {
	return fmt.Sprintf(resourceFreeformSystemHcl,
		rType, rName,
		o.blueprintId,
		o.name,
		stringOrNull(o.deviceProfileId),
		o.hostname,
		o.systemType,
		stringOrNull(o.deployMode),
		stringSliceOrNull(o.tags),
	)
}

func (o resourceFreeformSystem) testChecks(t testing.TB, rType, rName string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)
	result.append(t, "TestCheckResourceAttr", "type", o.systemType)
	if o.deviceProfileId != "" {
		result.append(t, "TestCheckResourceAttr", "device_profile_id", o.deviceProfileId)
	} else {
		result.append(t, "TestCheckNoResourceAttr", "device_profile_id")
	}
	result.append(t, "TestCheckResourceAttr", "hostname", o.hostname)
	if o.deployMode != "" {
		result.append(t, "TestCheckResourceAttr", "deploy_mode", o.deployMode)
	}
	if len(o.tags) > 0 {
		result.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
		for _, tag := range o.tags {
			result.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		}
	}

	return result
}

func TestResourceFreeformSystem(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// create a blueprint
	bp := testutils.FfBlueprintA(t, ctx)

	// import a device profile
	dpId, _ := bp.ImportDeviceProfile(ctx, "Juniper_vEX")

	type testStep struct {
		config resourceFreeformSystem
	}
	type testCase struct {
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"external_start_minimal": {
			steps: []testStep{
				{
					config: resourceFreeformSystem{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						hostname:    acctest.RandString(6),
						systemType:  apstra.SystemTypeExternal.String(),
					},
				},
				{
					config: resourceFreeformSystem{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						hostname:    acctest.RandString(6),
						deployMode:  enum.DeployModeDeploy.String(),
						systemType:  apstra.SystemTypeExternal.String(),
						tags:        randomStrings(rand.Intn(10)+2, 6),
					},
				},
				{
					config: resourceFreeformSystem{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						hostname:    acctest.RandString(6),
						systemType:  apstra.SystemTypeExternal.String(),
					},
				},
			},
		},
		"external_start_maximal": {
			steps: []testStep{
				{
					config: resourceFreeformSystem{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						hostname:    acctest.RandString(6),
						deployMode:  enum.DeployModeDeploy.String(),
						systemType:  apstra.SystemTypeExternal.String(),
						tags:        randomStrings(rand.Intn(10)+2, 6),
					},
				},
				{
					config: resourceFreeformSystem{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						hostname:    acctest.RandString(6),
						systemType:  apstra.SystemTypeExternal.String(),
					},
				},
				{
					config: resourceFreeformSystem{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						hostname:    acctest.RandString(6),
						deployMode:  enum.DeployModeDeploy.String(),
						systemType:  apstra.SystemTypeExternal.String(),
						tags:        randomStrings(rand.Intn(10)+2, 6),
					},
				},
			},
		},
		"internal_start_minimal": {
			steps: []testStep{
				{
					config: resourceFreeformSystem{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						hostname:        acctest.RandString(6),
						deviceProfileId: string(dpId),
						systemType:      apstra.SystemTypeInternal.String(),
					},
				},
				{
					config: resourceFreeformSystem{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						hostname:        acctest.RandString(6),
						deviceProfileId: string(dpId),
						deployMode:      enum.DeployModeDeploy.String(),
						systemType:      apstra.SystemTypeInternal.String(),
						tags:            randomStrings(rand.Intn(10)+2, 6),
					},
				},
				{
					config: resourceFreeformSystem{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						hostname:        acctest.RandString(6),
						systemType:      apstra.SystemTypeInternal.String(),
						deviceProfileId: string(dpId),
					},
				},
			},
		},
		"internal_start_maxmial": {
			steps: []testStep{
				{
					config: resourceFreeformSystem{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						hostname:        acctest.RandString(6),
						deviceProfileId: string(dpId),
						deployMode:      enum.DeployModeDeploy.String(),
						systemType:      apstra.SystemTypeInternal.String(),
						tags:            randomStrings(rand.Intn(10)+2, 6),
					},
				},
				{
					config: resourceFreeformSystem{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						hostname:        acctest.RandString(6),
						systemType:      apstra.SystemTypeInternal.String(),
						deviceProfileId: string(dpId),
					},
				},
				{
					config: resourceFreeformSystem{
						blueprintId:     bp.Id().String(),
						name:            acctest.RandString(6),
						hostname:        acctest.RandString(6),
						deviceProfileId: string(dpId),
						deployMode:      enum.DeployModeDeploy.String(),
						systemType:      apstra.SystemTypeInternal.String(),
						tags:            randomStrings(rand.Intn(10)+2, 6),
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformSystem)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()
			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.apiVersionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

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

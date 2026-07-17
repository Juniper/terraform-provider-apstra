//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/compatibility"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const datasourceDatacenterSwitchingZoneHCL = `
data %q %q {
  blueprint_id = %q
  id           = %s
  name         = %s
  mac_vrf_name = %s
}
`

const resourceDatacenterSwitchingZoneHCL = `resource %q %q {
  blueprint_id = %q
  name         = %q
  mac_vrf_name = %q
  description  = %s
  service_type = %q
  route_target = %s
  tags         = %s
}
`

type resourceDatacenterSwitchingZone struct {
	blueprintID string
	name        string
	macVRFName  string
	description string
	serviceType string
	routeTarget string
	tags        []string
}

func (o resourceDatacenterSwitchingZone) render(rType, rName string) string {
	resourceBlock := fmt.Sprintf(resourceDatacenterSwitchingZoneHCL, rType, rName,
		o.blueprintID,
		o.name,
		o.macVRFName,
		stringOrNull(o.description),
		o.serviceType,
		stringOrNull(o.routeTarget),
		stringSliceOrNull(o.tags),
	)

	datasourceBlockByID := fmt.Sprintf(datasourceDatacenterSwitchingZoneHCL, rType, rName+"_by_id", o.blueprintID, fmt.Sprintf("%s.%s.id", rType, rName), "null", "null")
	datasourceBlockByName := fmt.Sprintf(datasourceDatacenterSwitchingZoneHCL, rType, rName+"_by_name", o.blueprintID, "null", fmt.Sprintf("%s.%s.name", rType, rName), "null")
	datasourceBlockByMACVRFName := fmt.Sprintf(datasourceDatacenterSwitchingZoneHCL, rType, rName+"_by_mac_vrf_name", o.blueprintID, "null ", "null", fmt.Sprintf("%s.%s.mac_vrf_name", rType, rName))

	return resourceBlock + datasourceBlockByID + datasourceBlockByName + datasourceBlockByMACVRFName
}

func (o resourceDatacenterSwitchingZone) testChecks(t testing.TB, rType, rName string) []testChecks {
	resourceChecks := newTestChecks(rType + "." + rName)
	dataByIDChecks := newTestChecks("data." + rType + "." + rName + "_by_id")
	dataByNameChecks := newTestChecks("data." + rType + "." + rName + "_by_name")
	dataByMACVRFNameChecks := newTestChecks("data." + rType + "." + rName + "_by_mac_vrf_name")

	// required and computed attributes can always be checked
	resourceChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByMACVRFNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	resourceChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	dataByMACVRFNameChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	resourceChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByMACVRFNameChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	resourceChecks.append(t, "TestCheckResourceAttr", "mac_vrf_name", o.macVRFName)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "mac_vrf_name", o.macVRFName)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "mac_vrf_name", o.macVRFName)
	dataByMACVRFNameChecks.append(t, "TestCheckResourceAttr", "mac_vrf_name", o.macVRFName)

	if o.description != "" {
		resourceChecks.append(t, "TestCheckResourceAttr", "description", o.description)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "description", o.description)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "description", o.description)
		dataByMACVRFNameChecks.append(t, "TestCheckResourceAttr", "description", o.description)
	} else {
		resourceChecks.append(t, "TestCheckNoResourceAttr", "description")
		dataByIDChecks.append(t, "TestCheckNoResourceAttr", "description")
		dataByNameChecks.append(t, "TestCheckNoResourceAttr", "description")
		dataByMACVRFNameChecks.append(t, "TestCheckNoResourceAttr", "description")
	}

	resourceChecks.append(t, "TestCheckResourceAttr", "service_type", o.serviceType)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "service_type", o.serviceType)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "service_type", o.serviceType)
	dataByMACVRFNameChecks.append(t, "TestCheckResourceAttr", "service_type", o.serviceType)

	if o.routeTarget != "" {
		resourceChecks.append(t, "TestCheckResourceAttr", "route_target", o.routeTarget)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "route_target", o.routeTarget)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "route_target", o.routeTarget)
		dataByMACVRFNameChecks.append(t, "TestCheckResourceAttr", "route_target", o.routeTarget)
	} else {
		resourceChecks.append(t, "TestCheckResourceAttrSet", "route_target")
		dataByIDChecks.append(t, "TestCheckResourceAttrSet", "route_target")
		dataByNameChecks.append(t, "TestCheckResourceAttrSet", "route_target")
		dataByMACVRFNameChecks.append(t, "TestCheckResourceAttrSet", "route_target")
	}

	resourceChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByIDChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByNameChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByMACVRFNameChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	for _, tag := range o.tags {
		resourceChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByIDChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByNameChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByMACVRFNameChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
	}

	return []testChecks{resourceChecks, dataByIDChecks, dataByNameChecks, dataByMACVRFNameChecks}
}

func TestResourceDatacenterSwitchingZone(t *testing.T) {
	ctx := context.Background()

	// create the blueprint
	bp := testutils.BlueprintA(t, ctx)

	type testCase struct {
		versionconstraints []compatibility.Constraint
		steps              []resourceDatacenterSwitchingZone
	}

	testCases := map[string]testCase{
		"start_minimal": {
			versionconstraints: []compatibility.Constraint{compatibility.SwitchingZoneSupported},
			steps: []resourceDatacenterSwitchingZone{
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
					macVRFName:  acctest.RandString(6),
					serviceType: oneOf(enum.SwitchingZoneMACVRFServiceTypeVLANAware, enum.SwitchingZoneMACVRFServiceTypeVLANBundle).String(),
				},
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
					macVRFName:  acctest.RandString(6),
					serviceType: oneOf(enum.SwitchingZoneMACVRFServiceTypeVLANAware, enum.SwitchingZoneMACVRFServiceTypeVLANBundle).String(),
					description: acctest.RandString(6),
					routeTarget: randomRT(t),
					tags:        randomStrings(3, 6),
				},
			},
		},
		"start_maximal": {
			versionconstraints: []compatibility.Constraint{compatibility.SwitchingZoneSupported},
			steps: []resourceDatacenterSwitchingZone{
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
					macVRFName:  acctest.RandString(6),
					serviceType: oneOf(enum.SwitchingZoneMACVRFServiceTypeVLANAware, enum.SwitchingZoneMACVRFServiceTypeVLANBundle).String(),
					description: acctest.RandString(6),
					routeTarget: randomRT(t),
					tags:        randomStrings(3, 6),
				},
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
					macVRFName:  acctest.RandString(6),
					serviceType: oneOf(enum.SwitchingZoneMACVRFServiceTypeVLANAware, enum.SwitchingZoneMACVRFServiceTypeVLANBundle).String(),
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterSwitchingZone)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			for _, versionConstraint := range tCase.versionconstraints {
				if !versionConstraint.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
					t.Skipf("Skipping %s: version not supported", tName)
				}
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.render(resourceType, tName)
				checks := step.testChecks(t, resourceType, tName)

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

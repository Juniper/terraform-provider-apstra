//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/speed"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	versionconstraints "github.com/chrismarget-j/version-constraints"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

const (
	datasourceDesignLogicalDeviceHCL = `data %q %q {
  id   = %s
  name = %s
}
`

	resourceDesignLogicalDeviceHCL = `resource %q %q {
  name   = %q
  panels = [%s
  ]
}
`

	resourceDesignLogicalDevicePanelHCL = `
    {
      rows        = %d
      columns     = %d
      port_groups = [%s
      ]
    },`

	resourceDesignLogicalDevicePanelPortGroupHCL = `
        {
          port_count = %d
          port_speed = %q
          port_roles = %s
        },`
)

type resourceDesignLogicalDevice struct {
	name   string
	panels []resourceDesignLogicalDevicePanel
}

func (ld resourceDesignLogicalDevice) render(rType, rName string) string {
	sb := new(strings.Builder)
	for _, p := range ld.panels {
		sb.WriteString(p.render())
	}

	resourceBlock := fmt.Sprintf(resourceDesignLogicalDeviceHCL, rType, rName,
		ld.name,
		sb.String(),
	)
	datasourceBlockByID := fmt.Sprintf(datasourceDesignLogicalDeviceHCL, rType, rName+"_by_id", fmt.Sprintf("%s.%s.id", rType, rName), "null")
	datasourceBlockByName := fmt.Sprintf(datasourceDesignLogicalDeviceHCL, rType, rName+"_by_name", "null", fmt.Sprintf("%s.%s.name", rType, rName))

	return resourceBlock + "\n" + datasourceBlockByID + "\n" + datasourceBlockByName
}

func (ld resourceDesignLogicalDevice) testChecks(t testing.TB, rType, rName string) []testChecks {
	resourceChecks := newTestChecks(rType + "." + rName)
	dataByIDChecks := newTestChecks("data." + rType + "." + rName + "_by_id")
	dataByNameChecks := newTestChecks("data." + rType + "." + rName + "_by_name")

	// required and computed attributes can always be checked
	resourceChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	resourceChecks.append(t, "TestCheckResourceAttr", "name", ld.name)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "name", ld.name)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "name", ld.name)
	resourceChecks.append(t, "TestCheckResourceAttr", "definition.name", ld.name)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "definition.name", ld.name)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "definition.name", ld.name)

	for i, panel := range ld.panels {
		panel.testChecks(t, &resourceChecks, i)
	}

	return []testChecks{resourceChecks, dataByIDChecks, dataByNameChecks}
}

type resourceDesignLogicalDevicePanel struct {
	rows       int
	columns    int
	portGroups []resourceDesignLogicalDevicePanelPortGroup
}

func (ldp resourceDesignLogicalDevicePanel) render() string {
	sb := new(strings.Builder)
	for _, ldppg := range ldp.portGroups {
		sb.WriteString(ldppg.render())
	}

	return fmt.Sprintf(resourceDesignLogicalDevicePanelHCL,
		ldp.rows,
		ldp.columns,
		sb.String(),
	)
}

func (ldp resourceDesignLogicalDevicePanel) testChecks(t testing.TB, checks *testChecks, idx int) {
	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("panels.%d.rows", idx), strconv.Itoa(ldp.rows))
	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("definition.panels.%d.rows", idx), strconv.Itoa(ldp.rows))
	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("panels.%d.columns", idx), strconv.Itoa(ldp.columns))
	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("definition.panels.%d.columns", idx), strconv.Itoa(ldp.columns))
	for i, pg := range ldp.portGroups {
		pg.testChecks(t, checks, idx, i)
	}
}

type resourceDesignLogicalDevicePanelPortGroup struct {
	count int
	speed speed.Speed
	roles []string
}

func (ldppg resourceDesignLogicalDevicePanelPortGroup) render() string {
	return fmt.Sprintf(resourceDesignLogicalDevicePanelPortGroupHCL,
		ldppg.count,
		ldppg.speed,
		stringSliceOrNull(ldppg.roles),
	)
}

func (ldppg resourceDesignLogicalDevicePanelPortGroup) testChecks(t testing.TB, checks *testChecks, panelIdx, idx int) {
	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("panels.%d.port_groups.%d.port_count", panelIdx, idx), strconv.Itoa(ldppg.count))
	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("definition.panels.%d.port_groups.%d.port_count", panelIdx, idx), strconv.Itoa(ldppg.count))
	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("panels.%d.port_groups.%d.port_speed", panelIdx, idx), string(ldppg.speed))
	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("definition.panels.%d.port_groups.%d.port_speed", panelIdx, idx), string(ldppg.speed))

	// empty roles default to having all roles set
	if len(ldppg.roles) == 0 {
		var allPortRoles apstra.LogicalDevicePortRoles
		allPortRoles.IncludeAllUses()
		ldppg.roles = allPortRoles.Strings()
	}

	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("panels.%d.port_groups.%d.port_roles.#", panelIdx, idx), strconv.Itoa(len(ldppg.roles)))
	checks.append(t, "TestCheckResourceAttr", fmt.Sprintf("definition.panels.%d.port_groups.%d.port_roles.#", panelIdx, idx), strconv.Itoa(len(ldppg.roles)))

	for _, role := range ldppg.roles {
		checks.append(t, "TestCheckTypeSetElemAttr", fmt.Sprintf("panels.%d.port_groups.%d.port_roles.*", panelIdx, idx), role)
		checks.append(t, "TestCheckTypeSetElemAttr", fmt.Sprintf("definition.panels.%d.port_groups.%d.port_roles.*", panelIdx, idx), role)
	}
}

func TestAccResourceDesignLogicalDevice(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	clientVersion, err := version.NewVersion(client.ApiVersion())
	require.NoError(t, err)

	type testStep struct {
		config                  resourceDesignLogicalDevice
		preApplyResourceActions []plancheck.ResourceActionType
	}

	type testCase struct {
		versionConstraints []versionconstraints.Constraints
		steps              []testStep
	}

	testCases := map[string]testCase{
		"random": {
			versionConstraints: nil,
			steps: []testStep{
				{config: newResourceDesignLogicalDevice()},
				{config: newResourceDesignLogicalDevice()},
				{config: newResourceDesignLogicalDevice()},
			},
		},
		"begin_without_roles": {
			steps: []testStep{
				{
					config: resourceDesignLogicalDevice{
						name: acctest.RandString(8),
						panels: []resourceDesignLogicalDevicePanel{
							{
								rows:    1,
								columns: 1,
								portGroups: []resourceDesignLogicalDevicePanelPortGroup{
									{
										count: 1,
										speed: "100G",
									},
								},
							},
						},
					},
				},
				{
					config: resourceDesignLogicalDevice{
						name: acctest.RandString(8),
						panels: []resourceDesignLogicalDevicePanel{
							{
								rows:    1,
								columns: 1,
								portGroups: []resourceDesignLogicalDevicePanelPortGroup{
									{
										count: 1,
										speed: "100G",
										roles: []string{"spine", "leaf"},
									},
								},
							},
						},
					},
				},
				{
					config: resourceDesignLogicalDevice{
						name: acctest.RandString(8),
						panels: []resourceDesignLogicalDevicePanel{
							{
								rows:    1,
								columns: 1,
								portGroups: []resourceDesignLogicalDevicePanelPortGroup{
									{
										count: 1,
										speed: "100G",
									},
								},
							},
						},
					},
				},
			},
		},
		"begin_with_roles": {
			steps: []testStep{
				{
					config: resourceDesignLogicalDevice{
						name: acctest.RandString(8),
						panels: []resourceDesignLogicalDevicePanel{
							{
								rows:    1,
								columns: 1,
								portGroups: []resourceDesignLogicalDevicePanelPortGroup{
									{
										count: 1,
										speed: "100G",
										roles: []string{"spine", "leaf"},
									},
								},
							},
						},
					},
				},
				{
					config: resourceDesignLogicalDevice{
						name: acctest.RandString(8),
						panels: []resourceDesignLogicalDevicePanel{
							{
								rows:    1,
								columns: 1,
								portGroups: []resourceDesignLogicalDevicePanelPortGroup{
									{
										count: 1,
										speed: "100G",
									},
								},
							},
						},
					},
				},
				{
					config: resourceDesignLogicalDevice{
						name: acctest.RandString(8),
						panels: []resourceDesignLogicalDevicePanel{
							{
								rows:    1,
								columns: 1,
								portGroups: []resourceDesignLogicalDevicePanelPortGroup{
									{
										count: 1,
										speed: "100G",
										roles: []string{"spine", "leaf"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDesignLogicalDevice)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			for _, versionConstraint := range tCase.versionConstraints {
				if !versionConstraint.Check(clientVersion) {
					t.Skipf("Skipping %s: version %s not supported by test case", tName, clientVersion)
				}
			}

			steps := make([]resource.TestStep, 0, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

				var checkLog string
				var checkFuncs []resource.TestCheckFunc
				for _, checkList := range checks {
					checkLog = checkLog + checkList.string(len(checkFuncs))
					checkFuncs = append(checkFuncs, checkList.checks...)
				}

				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, checkLog, stepName)

				preApplyPlanChecks := make([]plancheck.PlanCheck, len(step.preApplyResourceActions))
				for i, preApplyPlanCheck := range step.preApplyResourceActions {
					preApplyPlanChecks[i] = plancheck.ExpectResourceAction(resourceType+"."+tName, preApplyPlanCheck)
				}

				steps = append(steps, resource.TestStep{
					Config:           insecureProviderConfigHCL + config,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: preApplyPlanChecks},
					Check:            resource.ComposeAggregateTestCheckFunc(checkFuncs...),
				})
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}

func newResourceDesignLogicalDevice() resourceDesignLogicalDevice {
	panels := make([]resourceDesignLogicalDevicePanel, rand.Intn(10)+1)
	for i := range panels {
		panels[i] = newResourceDesignLogicalDevicePanel()
	}

	return resourceDesignLogicalDevice{
		name:   acctest.RandString(8),
		panels: panels,
	}
}

func newResourceDesignLogicalDevicePanel() resourceDesignLogicalDevicePanel {
	portGroups := make([]resourceDesignLogicalDevicePanelPortGroup, rand.Intn(10)+1)
	portCount := 0
	for i := range portGroups {
		portGroups[i] = newResourceDesignLogicalDevicePanelPortGroup()
		portCount += portGroups[i].count
	}

	factorPairs := func(n int) [][2]int {
		var pairs [][2]int
		for i := 1; i*i <= n; i++ {
			if n%i == 0 {
				pairs = append(pairs, [2]int{i, n / i})
			}
		}
		return pairs
	}

	pair := random.OneOf(factorPairs(portCount)...)

	return resourceDesignLogicalDevicePanel{
		rows:       pair[0],
		columns:    pair[1],
		portGroups: portGroups,
	}
}

func newResourceDesignLogicalDevicePanelPortGroup() resourceDesignLogicalDevicePanelPortGroup {
	var allPortRoles apstra.LogicalDevicePortRoles
	allPortRoles.IncludeAllUses()

	return resourceDesignLogicalDevicePanelPortGroup{
		count: rand.Intn(20) + 1,
		speed: newRandomSpeed(),
		roles: random.SomeOf(allPortRoles.Strings()),
	}
}

func newRandomSpeed() speed.Speed {
	return speed.Speed(random.OneOf("100M", "1G", "10G", "25G", "40G", "50G", "100G", "200G", "400G", "800G"))
}

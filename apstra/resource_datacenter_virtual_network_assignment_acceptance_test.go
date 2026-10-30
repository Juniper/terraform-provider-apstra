package tfapstra_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	dctestobj "github.com/Juniper/terraform-provider-apstra/internal/test_utils/datacenter_test_objects"
	versionconstraints "github.com/chrismarget-j/version-constraints"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

const (
	resourceDatacenterVirtualNetworkAssignmentHCL = `resource %q %q {
  blueprint_id       = %q // required attribute
  virtual_network_id = %q // required attribute
  leaf_switch_id     = %q // required attribute
  vlan               = %s // optional integer attribute
  ipv4_mode          = %s // optional string attribute
  ipv4_address       = %s // optional string attribute
  ipv6_mode          = %s // optional string attribute
  ipv6_address       = %s // optional string attribute
  access_switch_ids  = %s // optional set attribute
}`
)

type resourceDatacenterVirtualNetworkAssignment struct {
	virtualNetworkID      string
	leafSwitchNickname    string
	vlan                  *int
	ipv4Mode              *enum.IPv4SVIMode
	ipv4Address           string
	ipv6Mode              *enum.IPv6SVIMode
	ipv6Address           string
	accessSwitchNicknames []string
}

func (r *resourceDatacenterVirtualNetworkAssignment) render(rType, rName, bpID string, switches map[string]string) string {
	accessSwitchIDs := make([]string, len(r.accessSwitchNicknames))
	for i, nickname := range r.accessSwitchNicknames {
		accessSwitchIDs[i] = switches[nickname]
	}

	return fmt.Sprintf(resourceDatacenterVirtualNetworkAssignmentHCL, rType, rName,
		bpID,
		r.virtualNetworkID,
		switches[r.leafSwitchNickname],
		intPtrOrNull(r.vlan),
		stringerOrNull(r.ipv4Mode),
		stringOrNull(r.ipv4Address),
		stringerOrNull(r.ipv6Mode),
		stringOrNull(r.ipv6Address),
		stringSliceOrNull(accessSwitchIDs),
	)
}

func (r resourceDatacenterVirtualNetworkAssignment) testChecks(t testing.TB, rType, rName, bpID string, switches map[string]string) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttr", "blueprint_id", bpID)
	result.append(t, "TestCheckResourceAttr", "virtual_network_id", r.virtualNetworkID)
	result.append(t, "TestCheckResourceAttr", "leaf_switch_id", switches[r.leafSwitchNickname])
	if r.vlan == nil {
		result.append(t, "TestCheckNoResourceAttr", "vlan")
	} else {
		result.append(t, "TestCheckResourceAttr", "vlan", strconv.Itoa(*r.vlan))
	}
	if r.ipv4Mode == nil {
		result.append(t, "TestCheckNoResourceAttr", "ipv4_mode")
	} else {
		result.append(t, "TestCheckResourceAttr", "ipv4_mode", r.ipv4Mode.String())
	}
	if r.ipv4Address == "" {
		result.append(t, "TestCheckNoResourceAttr", "ipv4_address")
	} else {
		result.append(t, "TestCheckResourceAttr", "ipv4_address", r.ipv4Address)
	}
	if r.ipv6Mode == nil {
		result.append(t, "TestCheckNoResourceAttr", "ipv6_mode")
	} else {
		result.append(t, "TestCheckResourceAttr", "ipv6_mode", r.ipv6Mode.String())
	}
	if r.ipv6Address == "" {
		result.append(t, "TestCheckNoResourceAttr", "ipv6_address")
	} else {
		result.append(t, "TestCheckResourceAttr", "ipv6_address", r.ipv6Address)
	}
	result.append(t, "TestCheckResourceAttr", "access_switch_ids.#", strconv.Itoa(len(r.accessSwitchNicknames)))
	for _, accessSwitchNickname := range r.accessSwitchNicknames {
		result.append(t, "TestCheckTypeSetElemAttr", "access_switch_ids.*", switches[accessSwitchNickname])
	}

	return result
}

func TestAccResourceDatacenterVirtualNetworkAssignment(t *testing.T) {
	ctx := t.Context()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))
	if !compatibility.BPDefaultRoutingZoneAddressingOK.Check(apiVersion) {
		t.Skipf("skipping test because Apstra version %s does not support default routing zone addressing", apiVersion.String())
	}

	bp := dctestobj.BlueprintB(t, ctx, client)
	rzID := dctestobj.RoutingZoneA(t, ctx, bp, false)
	ipv4Nets := []string{
		"10.0.1.0/24",
		"10.0.2.0/24",
		"10.0.3.0/24",
	}
	ipv6Nets := []string{
		"3fff:0:0:1::/64",
		"3fff:0:0:2::/64",
		"3fff:0:0:3::/64",
	}
	require.Equal(t, len(ipv4Nets), len(ipv6Nets))
	vnIDs := make([]string, max(len(ipv4Nets), len(ipv6Nets)))
	for i := range vnIDs {
		vnIDs[i] = dctestobj.VirtualNetworkC(t, ctx, bp, rzID, ipv4Nets[i], ipv6Nets[i])
	}

	query := new(apstra.PathQuery).SetClient(bp.Client()).SetBlueprintId(bp.Id()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "system_type", Value: apstra.QEStringVal("switch")},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		})
	var result struct {
		Items []struct {
			System struct {
				ID    string `json:"id"`
				Label string `json:"label"`
			} `json:"n_system"`
		} `json:"Items"`
	}
	require.NoError(t, query.Do(ctx, &result))

	// Build a map of switch nickname (access5, leaf2) to system ID for use in the test cases below.
	switches := make(map[string]string, len(result.Items))
	for _, item := range result.Items {
		parts := strings.Split(item.System.Label, "_")
		switches[parts[len(parts)-1]] = item.System.ID
	}

	type testStep struct {
		config                     map[string]resourceDatacenterVirtualNetworkAssignment
		preApplyResourceActionType plancheck.ResourceActionType
	}

	type testCase struct {
		apiVersionConstraints []versionconstraints.Constraints
		steps                 []testStep
	}

	// Leaf/Access topology:
	//  - leaf1				(standalone leaf switch)
	//    - access1			(single-homed standalone access switch)
	//    - access2			(single-homed standalone access switch)
	//    - access3			(single-homed standalone access switch)
	//  - leaf2				(standalone leaf switch)
	//    - access4			(single-homed standalone access switch)
	//    - access5			(single-homed standalone access switch)
	//    - access6			(single-homed standalone access switch)
	//  - leaf3				(member of leaf pair with leaf4)
	//    - access7			(single-homed standalone access switch)
	//    - access8			(single-homed standalone access switch)
	//    - access17		(single-homed member of access pair with access18)
	//    - access19		(single-homed member of access pair with access20)
	//  - leaf4				(member of leaf pair with leaf3)
	//    - access9			(single-homed standalone access switch)
	//    - access10		(single-homed standalone access switch)
	//    - access18		(single-homed member of access pair with access17)
	//    - access20		(single-homed member of access pair with access19)
	//  - leaf3/4			(leaf pair)
	//    - access11		(dual-homed standalone access switch)
	//    - access12		(dual-homed standalone access switch)
	//    - access13		(dual-homed member of access pair with access14)
	//    - access14		(dual-homed member of access pair with access13)
	//    - access15		(dual-homed member of access pair with access16)
	//    - access16		(dual-homed member of access pair with access15)

	testCases := map[string]testCase{
		"single_leaf_start_minimal": {
			steps: []testStep{
				{
					config: map[string]resourceDatacenterVirtualNetworkAssignment{
						"leaf1": {
							virtualNetworkID:   vnIDs[0],
							leafSwitchNickname: "leaf1",
						},
					},
				},
				{
					preApplyResourceActionType: plancheck.ResourceActionUpdate,
					config: map[string]resourceDatacenterVirtualNetworkAssignment{
						"leaf1": {
							virtualNetworkID:      vnIDs[0],
							leafSwitchNickname:    "leaf1",
							vlan:                  pointer.To(100),
							ipv4Mode:              pointer.To(enum.IPv4SVIModeForced),
							ipv4Address:           "10.0.1.11/24",
							ipv6Mode:              pointer.To(enum.IPv6SVIModeForced),
							ipv6Address:           "3fff:0:0:1::11/64",
							accessSwitchNicknames: []string{"access1", "access2", "access3"},
						},
					},
				},
			},
		},
		"single_leaf_start_maximal": {
			steps: []testStep{
				{
					config: map[string]resourceDatacenterVirtualNetworkAssignment{
						"leaf2": {
							virtualNetworkID:      vnIDs[1],
							leafSwitchNickname:    "leaf2",
							vlan:                  pointer.To(200),
							ipv4Mode:              pointer.To(enum.IPv4SVIModeForced),
							ipv4Address:           "10.0.2.12/24",
							ipv6Mode:              pointer.To(enum.IPv6SVIModeForced),
							ipv6Address:           "3fff:0:0:2::12/64",
							accessSwitchNicknames: []string{"access4", "access5", "access6"},
						},
					},
				},
				{
					config: map[string]resourceDatacenterVirtualNetworkAssignment{
						"leaf2": {
							virtualNetworkID:   vnIDs[1],
							leafSwitchNickname: "leaf2",
						},
					},
				},
			},
		},
		"leaf_pair_with_single_homed_access_switches": {
			steps: []testStep{
				{
					config: map[string]resourceDatacenterVirtualNetworkAssignment{
						"leaf3": {
							virtualNetworkID:      vnIDs[2],
							leafSwitchNickname:    "leaf3",
							vlan:                  pointer.To(300),
							ipv4Mode:              pointer.To(enum.IPv4SVIModeForced),
							ipv4Address:           "10.0.3.13/24",
							ipv6Mode:              pointer.To(enum.IPv6SVIModeForced),
							ipv6Address:           "3fff:0:0:3::13/64",
							accessSwitchNicknames: []string{"access7", "access8", "access9", "access10"},
						},
						"leaf4": {
							virtualNetworkID:      vnIDs[2],
							leafSwitchNickname:    "leaf4",
							vlan:                  pointer.To(300),
							ipv4Mode:              pointer.To(enum.IPv4SVIModeForced),
							ipv4Address:           "10.0.3.14/24",
							ipv6Mode:              pointer.To(enum.IPv6SVIModeForced),
							ipv6Address:           "3fff:0:0:3::14/64",
							accessSwitchNicknames: []string{"access7", "access8", "access9", "access10"},
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterVirtualNetworkAssignment)

	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() // Not using parallelism because the test cases may use the same resource groups.

			for _, constraint := range tCase.apiVersionConstraints {
				if !constraint.Check(apiVersion) {
					t.Skipf("test case %s requires Apstra %s", tName, constraint)
				}
			}

			// These variables represent *one of* the resources in the test case, for use
			// in the import step below. Map order will dictate the winner. We don't care
			// which one wins, we're only importing one.
			var importName, importVN, importLeaf string

			steps := make([]resource.TestStep, len(tCase.steps), len(tCase.steps)+2) // 2 extra step capacity for import by ID and by Identity.
			for i, step := range tCase.steps {
				config := new(strings.Builder)
				chkLog := new(strings.Builder)
				var testCheckFuncs []resource.TestCheckFunc
				for k, v := range step.config {
					config.WriteString(v.render(resourceType, tName+"_"+k, bp.Id().String(), switches))
					config.WriteString("\n\n")
					checks := v.testChecks(t, resourceType, tName+"_"+k, bp.Id().String(), switches)
					chkLog.WriteString(checks.string())
					chkLog.WriteString("\n")
					testCheckFuncs = append(testCheckFuncs, checks.checks...)
					importName = fmt.Sprintf("%s.%s", resourceType, tName+"_"+k) // last one wins. we don't care which, we're only importing one.
					importVN = v.virtualNetworkID
					importLeaf = switches[v.leafSwitchNickname]
				}

				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config.String(),
					Check:  resource.ComposeAggregateTestCheckFunc(testCheckFuncs...),
				}
			}

			// Import by composite ID string (legacy import ID)
			steps = append(steps, resource.TestStep{
				ExpectNonEmptyPlan: true, // because ImportState() will set attributes which were not present in the state.
				ImportState:        true,
				ResourceName:       importName,
				ImportStateKind:    resource.ImportCommandWithID,
				ImportStateId:      fmt.Sprintf(`{"blueprint_id":"%s","virtual_network_id":"%s","leaf_switch_id":"%s"}`, bp.Id().String(), importVN, importLeaf),
				ImportStateVerify:  false, // There's no "id" field, and we're a composite ID resource anyway, so we can't verify the import state with the current library. We can only verify that the import didn't error.
			})

			// Import by Identity (preferred import method)
			steps = append(steps, resource.TestStep{
				ExpectNonEmptyPlan:      true, // because ImportState() will set attributes which were not present in the state.
				ImportState:             true,
				ResourceName:            importName,
				ImportStateKind:         resource.ImportBlockWithResourceIdentity,
				ImportStateVerify:       false, // There's no "id" field, and we're a composite ID resource anyway, so we can't verify the import state with the current library. We can only verify that the import didn't error.
				ImportStateVerifyIgnore: []string{"vlan"},
			})

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}

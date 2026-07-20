//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/compatibility"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceDataCenterSecurityPolicyRefName = "apstra_datacenter_security_policy.test"

	datasourceDataCenterSecurityPolicyHCL = `
data %q %q {
  blueprint_id = %q
  id           = %s
  name         = %s
}
`

	resourceDataCenterSecurityPolicyHCL = `
resource %q %q {
  blueprint_id                     = %q
  name                             = %q
  description                      = %s
  enabled                          = %s
  ip_version                       = %s
  source_application_point_id      = %s
  destination_application_point_id = %s
  tags                             = %s
  rules                            = %s
}
`
	resourceDataCenterSecurityPolicyRuleHCL = `
    {
      name              = %q
      description       = %s
      action            = %q
      protocol          = %q
      source_ports      = %s
      destination_ports = %s
      established       = %s
    },
`
	resourceDataCenterSecurityPolicyRulePortHCL = `
        {
          from_port = %d
          to_port   = %d
        },
`
)

type resourceDataCenterSecurityPolicy struct {
	blueprintID                   string
	name                          string
	description                   string
	enabled                       *bool
	ipVersion                     *enum.PolicyAddressFamily
	sourceApplicationPointID      string
	destinationApplicationPointID string
	tags                          []string
	rules                         []resourceDataCenterSecurityPolicyRule
}

func (r resourceDataCenterSecurityPolicy) render(rType, rName string) string {
	rules := "null"
	if len(r.rules) > 0 {
		sb := new(strings.Builder)
		sb.WriteString("[\n")
		for _, rule := range r.rules {
			sb.WriteString(rule.render())
		}
		sb.WriteString("  ]")
	}

	ipVersion := "null"
	if r.ipVersion != nil {
		ipVersion = `"` + r.ipVersion.String() + `"`
	}

	resource := fmt.Sprintf(resourceDataCenterSecurityPolicyHCL, rType, rName,
		r.blueprintID,
		r.name,
		stringOrNull(r.description),
		boolPtrOrNull(r.enabled),
		ipVersion,
		stringOrNull(r.sourceApplicationPointID),
		stringOrNull(r.destinationApplicationPointID),
		stringSliceOrNull(r.tags),
		rules,
	)
	datasourceByID := fmt.Sprintf(datasourceDataCenterSecurityPolicyHCL, rType, rName+"_by_id",
		r.blueprintID,
		rType+"."+rName+".id",
		"null",
	)
	datasourceByName := fmt.Sprintf(datasourceDataCenterSecurityPolicyHCL, rType, rName+"_by_name",
		r.blueprintID,
		"null",
		rType+"."+rName+".name",
	)

	return resource + datasourceByID + datasourceByName
}

func (o resourceDataCenterSecurityPolicy) testChecks(t testing.TB, rType, rName string) []testChecks {
	resourceChecks := newTestChecks(rType + "." + rName)
	dataByIDChecks := newTestChecks("data." + rType + "." + rName + "_by_id")
	dataByNameChecks := newTestChecks("data." + rType + "." + rName + "_by_name")

	// required and computed attributes can always be checked
	resourceChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByIDChecks.append(t, "TestCheckResourceAttrSet", "id")
	dataByNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	resourceChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintID)
	resourceChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByIDChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	dataByNameChecks.append(t, "TestCheckResourceAttr", "name", o.name)
	if o.description == "" {
		resourceChecks.append(t, "TestCheckNoResourceAttr", "description")
		dataByIDChecks.append(t, "TestCheckNoResourceAttr", "description")
		dataByNameChecks.append(t, "TestCheckNoResourceAttr", "description")
	} else {
		resourceChecks.append(t, "TestCheckResourceAttr", "description", o.description)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "description", o.description)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "description", o.description)
	}
	if o.enabled == nil {
		resourceChecks.append(t, "TestCheckResourceAttr", "enabled", "true")
		dataByIDChecks.append(t, "TestCheckResourceAttr", "enabled", "true")
		dataByNameChecks.append(t, "TestCheckResourceAttr", "enabled", "true")
	} else {
		resourceChecks.append(t, "TestCheckResourceAttr", "enabled", strconv.FormatBool(*o.enabled))
		dataByIDChecks.append(t, "TestCheckResourceAttr", "enabled", strconv.FormatBool(*o.enabled))
		dataByNameChecks.append(t, "TestCheckResourceAttr", "enabled", strconv.FormatBool(*o.enabled))
	}
	if o.ipVersion == nil {
		resourceChecks.append(t, "TestCheckNoResourceAttr", "ip_version")
		dataByIDChecks.append(t, "TestCheckNoResourceAttr", "ip_version")
		dataByNameChecks.append(t, "TestCheckNoResourceAttr", "ip_version")
	} else {
		resourceChecks.append(t, "TestCheckResourceAttr", "ip_version", o.ipVersion.String())
		dataByIDChecks.append(t, "TestCheckResourceAttr", "ip_version", o.ipVersion.String())
		dataByNameChecks.append(t, "TestCheckResourceAttr", "ip_version", o.ipVersion.String())
	}
	if o.sourceApplicationPointID == "" {
		resourceChecks.append(t, "TestCheckNoResourceAttr", "source_application_point_id")
		dataByIDChecks.append(t, "TestCheckNoResourceAttr", "source_application_point_id")
		dataByNameChecks.append(t, "TestCheckNoResourceAttr", "source_application_point_id")
	} else {
		resourceChecks.append(t, "TestCheckResourceAttr", "source_application_point_id", o.sourceApplicationPointID)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "source_application_point_id", o.sourceApplicationPointID)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "source_application_point_id", o.sourceApplicationPointID)
	}
	if o.destinationApplicationPointID == "" {
		resourceChecks.append(t, "TestCheckNoResourceAttr", "destination_application_point_id")
		dataByIDChecks.append(t, "TestCheckNoResourceAttr", "destination_application_point_id")
		dataByNameChecks.append(t, "TestCheckNoResourceAttr", "destination_application_point_id")
	} else {
		resourceChecks.append(t, "TestCheckResourceAttr", "destination_application_point_id", o.destinationApplicationPointID)
		dataByIDChecks.append(t, "TestCheckResourceAttr", "destination_application_point_id", o.destinationApplicationPointID)
		dataByNameChecks.append(t, "TestCheckResourceAttr", "destination_application_point_id", o.destinationApplicationPointID)
	}
	resourceChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByIDChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	dataByNameChecks.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
	for _, tag := range o.tags {
		resourceChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByIDChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		dataByNameChecks.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
	}
	resourceChecks.append(t, "TestCheckResourceAttr", "rules.#", strconv.Itoa(len(o.rules)))
	dataByIDChecks.append(t, "TestCheckResourceAttr", "rules.#", strconv.Itoa(len(o.rules)))
	dataByNameChecks.append(t, "TestCheckResourceAttr", "rules.#", strconv.Itoa(len(o.rules)))
	for i, rule := range o.rules {
		rule.addTestChecks(t, fmt.Sprintf("rules.%d.", i), &resourceChecks)
		rule.addTestChecks(t, fmt.Sprintf("rules.%d.", i), &dataByIDChecks)
		rule.addTestChecks(t, fmt.Sprintf("rules.%d.", i), &dataByNameChecks)
	}

	return []testChecks{resourceChecks, dataByIDChecks, dataByNameChecks}
}

type resourceDataCenterSecurityPolicyRule struct {
	name             string
	description      string
	action           *enum.PolicyRuleAction
	protocol         enum.PolicyRuleProtocol
	sourcePorts      []resourceDataCenterSecurityPolicyRulePort
	destinationPorts []resourceDataCenterSecurityPolicyRulePort
	established      *bool
}

func (r resourceDataCenterSecurityPolicyRule) render() string {
	var action string
	if r.action != nil {
		action = rosetta.StringersToFriendlyString(*r.action)
	}

	portsToString := func(ports []resourceDataCenterSecurityPolicyRulePort) string {
		if len(ports) == 0 {
			return "null"
		}

		result := new(strings.Builder)
		result.WriteString("[\n")
		for _, port := range ports {
			result.WriteString(port.render())
		}
		result.WriteString("      ]")
		return result.String()
	}

	return fmt.Sprintf(resourceDataCenterSecurityPolicyRuleHCL,
		r.name,
		stringOrNull(r.description),
		stringOrNull(action),
		r.protocol,
		portsToString(r.sourcePorts),
		portsToString(r.destinationPorts),
		boolPtrOrNull(r.established),
	)
}

func (r resourceDataCenterSecurityPolicyRule) addTestChecks(t testing.TB, p string, c *testChecks) {
	c.append(t, "TestCheckResourceAttrSet", p+"id")
	c.append(t, "TestCheckResourceAttr", p+"name", r.name)
	if r.description == "" {
		c.append(t, "TestCheckNoResourceAttr", p+"description")
	} else {
		c.append(t, "TestCheckResourceAttr", p+"description", r.description)
	}
	c.append(t, "TestCheckResourceAttr", p+"protocol", rosetta.StringersToFriendlyString(r.protocol))
	c.append(t, "TestCheckResourceAttr", p+"action", r.action.String())
	c.append(t, "TestCheckResourceAttr", p+"source_ports.#", strconv.Itoa(len(r.sourcePorts)))
	for _, port := range r.sourcePorts {
		c.appendSetNestedCheck(t, p+"source_ports", map[string]string{
			"from_port": strconv.Itoa(port.from),
			"to_port":   strconv.Itoa(port.to),
		})
	}
	c.append(t, "TestCheckResourceAttr", p+"destination_ports.#", strconv.Itoa(len(r.destinationPorts)))
	for _, port := range r.destinationPorts {
		c.appendSetNestedCheck(t, p+"destination_ports", map[string]string{
			"from_port": strconv.Itoa(port.from),
			"to_port":   strconv.Itoa(port.to),
		})
	}
}

type resourceDataCenterSecurityPolicyRulePort struct {
	from int
	to   int
}

func (r resourceDataCenterSecurityPolicyRulePort) render() string {
	return fmt.Sprintf(resourceDataCenterSecurityPolicyRulePortHCL, r.from, r.to)
}

func TestResourceDatacenterSecurityPolicy(t *testing.T) {
	ctx := context.Background()

	// create the blueprint
	bp := testutils.BlueprintA(t, ctx)

	type testCase struct {
		versionconstraints []compatibility.Constraint
		steps              []resourceDataCenterSecurityPolicy
	}

	testCases := map[string]testCase{
		"start_minimal_620_and_later": {
			versionconstraints: []compatibility.Constraint{compatibility.DatacenterPolicyAddressFamilyRequired},
			steps: []resourceDataCenterSecurityPolicy{
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
					ipVersion:   pointer.To(random.OneOf(enum.PolicyAddressFamilyIPv4, enum.PolicyAddressFamilyIPv6, enum.PolicyAddressFamilyIPv6)),
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceDatacenterSecurityPolicy)

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

//type resourceDataCenterSecurityPolicy struct {
//	config     datacenter.Policy
//	checks     []resource.TestCheckFunc
//	minVersion *version.Version
//}
//
//func (o resourceDataCenterSecurityPolicy) renderConfig(bpId apstra.ObjectId) string {
//	renderPort := func(port datacenter.PortRange) string {
//		return fmt.Sprintf(resourceDataCenterSecurityPolicyRulePortHCL, port.First, port.Last)
//	}
//
//	renderPorts := func(ports datacenter.PortRanges) string {
//		if len(ports) == 0 {
//			return "null"
//		}
//
//		var sb strings.Builder
//		sb.WriteString("[\n")
//		for _, port := range ports {
//			sb.WriteString(renderPort(port))
//		}
//		sb.WriteString("      \n]")
//		return sb.String()
//	}
//
//	renderEstablished := func(tsq *enum.TcpStateQualifier) string {
//		if tsq == nil {
//			return "null"
//		}
//
//		return strconv.FormatBool(*tsq == enum.TcpStateQualifierEstablished)
//	}
//
//	renderRule := func(rule datacenter.PolicyRule) string {
//		return fmt.Sprintf(resourceDataCenterSecurityPolicyRuleHCL,
//			rule.Label,
//			stringOrNull(rule.Description),
//			rule.Action.Value,
//			rosetta.StringersToFriendlyString(rule.Protocol),
//			renderPorts(rule.SrcPort),
//			renderPorts(rule.DstPort),
//			renderEstablished(rule.TcpStateQualifier),
//		)
//	}
//
//	renderRules := func(rules []datacenter.PolicyRule) string {
//		if len(rules) == 0 {
//			return "null"
//		}
//
//		var sb strings.Builder
//		sb.WriteString("[\n")
//		for _, rule := range rules {
//			sb.WriteString(renderRule(rule))
//		}
//		sb.WriteString("    ]\n")
//		return sb.String()
//	}
//
//	renderTags := func(s []string) string {
//		if len(s) == 0 {
//			return "null"
//		}
//		return `["` + strings.Join(s, `","`) + `"]`
//	}
//
//	return insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterSecurityPolicyHCL,
//		bpId,
//		o.config.Label,
//		stringOrNull(o.config.Description),
//		strconv.FormatBool(o.config.Enabled),
//		stringPtrOrNull(o.config.SrcApplicationPoint),
//		stringPtrOrNull(o.config.DstApplicationPoint),
//		renderTags(o.config.Tags),
//		renderRules(o.config.Rules),
//	)
//}
//
//func TestResourceDatacenterSecurityPolicy(t *testing.T) {
//	ctx := context.Background()
//
//	bpClient := testutils.BlueprintA(t, ctx)
//
//	// collect leaf switch IDs
//	leafIds := systemIds(ctx, t, bpClient, "leaf")
//
//	// create virtual networks
//	vnIds := make([]string, 2)
//	for i := range vnIds {
//		id, err := bpClient.CreateVirtualNetwork(ctx, datacenter.VirtualNetwork{
//			IPv4Enabled: true,
//			Label:       acctest.RandString(5),
//			Bindings:    []datacenter.VNBinding{{SystemID: leafIds[i]}},
//			Type:        enum.VnTypeVlan,
//		})
//		if err != nil {
//			t.Fatal(err)
//		}
//		vnIds[i] = id
//	}
//
//	tests := []resourceDataCenterSecurityPolicy{
//		{
//			config: datacenter.Policy{
//				Label: "1",
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "1"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
//			},
//		},
//		{
//			config: datacenter.Policy{
//				Label:       "2",
//				Enabled:     true,
//				Description: "two",
//				Tags:        []string{"a", "b", "c"},
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "2"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "description", "two"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "tags.#", "3"),
//				resource.TestCheckTypeSetElemAttr(resourceDataCenterSecurityPolicyRefName, "tags.*", "c"),
//				resource.TestCheckTypeSetElemAttr(resourceDataCenterSecurityPolicyRefName, "tags.*", "a"),
//				resource.TestCheckTypeSetElemAttr(resourceDataCenterSecurityPolicyRefName, "tags.*", "b"),
//			},
//		},
//		{
//			config: datacenter.Policy{
//				Label:   "3",
//				Enabled: false,
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "3"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
//			},
//		},
//		{
//			config: datacenter.Policy{
//				Label:               "4",
//				Enabled:             true,
//				SrcApplicationPoint: &vnIds[0],
//				DstApplicationPoint: &vnIds[1],
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "4"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "source_application_point_id", vnIds[0]),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "destination_application_point_id", vnIds[1]),
//			},
//		},
//		{
//			config: datacenter.Policy{
//				Label:               "5",
//				Enabled:             false,
//				SrcApplicationPoint: &vnIds[1],
//				DstApplicationPoint: &vnIds[0],
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "5"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "source_application_point_id", vnIds[1]),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "destination_application_point_id", vnIds[0]),
//			},
//		},
//		{
//			config: datacenter.Policy{
//				Label:   "6",
//				Enabled: true,
//				Rules: []datacenter.PolicyRule{
//					{
//						Label:       "60",
//						Description: "",
//						Protocol:    enum.PolicyRuleProtocolIcmp,
//						Action:      enum.PolicyRuleActionDeny,
//					},
//				},
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "6"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.#", "1"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.name", "60"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "icmp"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.action", "deny"),
//			},
//		},
//		{
//			config: datacenter.Policy{
//				Label:   "7",
//				Enabled: false,
//				Rules: []datacenter.PolicyRule{
//					{
//						Label:       "70",
//						Description: "seventy",
//						Protocol:    enum.PolicyRuleProtocolIp,
//						Action:      enum.PolicyRuleActionDenyLog,
//					},
//				},
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "7"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.#", "1"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.name", "70"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.description", "seventy"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "ip"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.action", "deny_log"),
//			},
//		},
//		{
//			config: datacenter.Policy{
//				Label:               "8",
//				Enabled:             true,
//				SrcApplicationPoint: &vnIds[0],
//				DstApplicationPoint: &vnIds[1],
//				Rules: []datacenter.PolicyRule{
//					{
//						Label:       "80",
//						Description: "eighty",
//						Protocol:    enum.PolicyRuleProtocolUdp,
//						Action:      enum.PolicyRuleActionPermit,
//					},
//					{
//						Label:       "81",
//						Description: "eightyone",
//						Protocol:    enum.PolicyRuleProtocolTcp,
//						Action:      enum.PolicyRuleActionPermitLog,
//					},
//					{
//						Label:             "82",
//						Description:       "eightytwo",
//						Protocol:          enum.PolicyRuleProtocolTcp,
//						Action:            enum.PolicyRuleActionPermit,
//						TcpStateQualifier: &enum.TcpStateQualifierEstablished,
//						SrcPort: datacenter.PortRanges{
//							{First: 1, Last: 1},
//							{First: 3, Last: 5},
//							{First: 7, Last: 9},
//							{First: 11, Last: 11},
//						},
//						DstPort: datacenter.PortRanges{
//							{First: 50, Last: 50},
//						},
//					},
//				},
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "8"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "source_application_point_id", vnIds[0]),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "destination_application_point_id", vnIds[1]),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.#", "3"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.name", "80"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.description", "eighty"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "udp"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.action", "permit"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.name", "81"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.description", "eightyone"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.protocol", "tcp"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.action", "permit_log"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.established", "false"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.name", "82"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.description", "eightytwo"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.protocol", "tcp"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.action", "permit"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.established", "true"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.source_ports.#", "4"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.destination_ports.#", "1"),
//			},
//		},
//		{
//			config: datacenter.Policy{
//				Label:               "9",
//				Enabled:             true,
//				SrcApplicationPoint: &vnIds[0],
//				DstApplicationPoint: &vnIds[1],
//				Rules: []datacenter.PolicyRule{
//					{
//						Label:       "90",
//						Description: "ninety",
//						Protocol:    enum.PolicyRuleProtocolUdp,
//						Action:      enum.PolicyRuleActionPermit,
//					},
//					{
//						Label:       "91",
//						Description: "ninetyone",
//						Protocol:    enum.PolicyRuleProtocolTcp,
//						Action:      enum.PolicyRuleActionPermitLog,
//					},
//					{
//						Label:             "92",
//						Description:       "ninetytwo",
//						Protocol:          enum.PolicyRuleProtocolTcp,
//						Action:            enum.PolicyRuleActionPermit,
//						TcpStateQualifier: &enum.TcpStateQualifierEstablished,
//						SrcPort: datacenter.PortRanges{
//							{First: 1, Last: 1},
//							{First: 7, Last: 9},
//							{First: 11, Last: 11},
//						},
//						DstPort: datacenter.PortRanges{
//							{First: 50, Last: 50},
//							{First: 3, Last: 5},
//						},
//					},
//				},
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "9"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "source_application_point_id", vnIds[0]),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "destination_application_point_id", vnIds[1]),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.#", "3"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.name", "90"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.description", "ninety"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "udp"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.action", "permit"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.name", "91"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.description", "ninetyone"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.protocol", "tcp"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.action", "permit_log"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.established", "false"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.name", "92"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.description", "ninetytwo"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.protocol", "tcp"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.action", "permit"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.established", "true"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.source_ports.#", "3"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.destination_ports.#", "2"),
//			},
//		},
//		{
//			config: datacenter.Policy{
//				Label:   "10",
//				Enabled: false,
//			},
//			checks: []resource.TestCheckFunc{
//				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "10"),
//				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
//			},
//		},
//	}
//
//	var steps []resource.TestStep
//	for _, test := range tests {
//		if test.minVersion != nil && version.Must(version.NewVersion(bpClient.Client().ApiVersion())).LessThan(test.minVersion) {
//			continue
//		}
//		steps = append(steps, resource.TestStep{
//			Config: test.renderConfig(bpClient.Id()),
//			Check:  resource.ComposeAggregateTestCheckFunc(test.checks...),
//		})
//	}
//
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		Steps:                    steps,
//	})
//}
//

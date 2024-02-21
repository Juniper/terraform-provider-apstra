package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"strconv"
	"strings"
	"testing"
)

const (
	resourceDataCenterSecurityPolicyRefName = "apstra_datacenter_security_policy.test"
	resourceDataCenterSecurityPolicyHCL     = `
resource "apstra_datacenter_security_policy" "test" {
  blueprint_id                     = "%s"
  name                             = "%s"
  description                      = %s
  enabled                          = %s
  source_application_point_id      = %s
  destination_application_point_id = %s
  tags                             = %s
  rules                            = %s
}
`
	resourceDataCenterSecurityPolicyRuleHCL = `
    {
      name              = "%s"
      description       = %s
      action            = "%s"
      protocol          = "%s"
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

type testCaseResourceSecurityPolicy struct {
	config     apstra.PolicyData
	checks     []resource.TestCheckFunc
	minVersion *version.Version
}

func (o testCaseResourceSecurityPolicy) renderConfig(bpId apstra.ObjectId) string {
	renderPort := func(port apstra.PortRange) string {
		return fmt.Sprintf(resourceDataCenterSecurityPolicyRulePortHCL, port.First, port.Last)
	}

	renderPorts := func(ports apstra.PortRanges) string {
		if len(ports) == 0 {
			return "null"
		}

		var sb strings.Builder
		sb.WriteString("[\n")
		for _, port := range ports {
			sb.WriteString(renderPort(port))
		}
		sb.WriteString("      \n]")
		return sb.String()
	}

	renderEstablished := func(tsq *apstra.TcpStateQualifier) string {
		if tsq == nil {
			return "null"
		}

		return strconv.FormatBool(*tsq == apstra.TcpStateQualifierEstablished)
	}

	renderRule := func(rule apstra.PolicyRule) string {
		return fmt.Sprintf(resourceDataCenterSecurityPolicyRuleHCL,
			rule.Data.Label,
			stringOrNull(rule.Data.Description),
			rule.Data.Action.Value,
			utils.StringersToFriendlyString(rule.Data.Protocol),
			renderPorts(rule.Data.SrcPort),
			renderPorts(rule.Data.DstPort),
			renderEstablished(rule.Data.TcpStateQualifier),
		)
	}

	renderRules := func(rules []apstra.PolicyRule) string {
		if len(rules) == 0 {
			return "null"
		}

		var sb strings.Builder
		sb.WriteString("[\n")
		for _, rule := range rules {
			sb.WriteString(renderRule(rule))
		}
		sb.WriteString("    ]\n")
		return sb.String()
	}

	renderApplicationPoint := func(p *apstra.PolicyApplicationPointData) string {
		if p == nil {
			return "null"
		}

		return `"` + p.Id.String() + `"`
	}

	renderTags := func(s []string) string {
		if len(s) == 0 {
			return "null"
		}
		return `["` + strings.Join(s, `","`) + `"]`
	}

	return insecureProviderConfigHCL + fmt.Sprintf(resourceDataCenterSecurityPolicyHCL,
		bpId,
		o.config.Label,
		stringOrNull(o.config.Description),
		strconv.FormatBool(o.config.Enabled),
		renderApplicationPoint(o.config.SrcApplicationPoint),
		renderApplicationPoint(o.config.DstApplicationPoint),
		renderTags(o.config.Tags),
		renderRules(o.config.Rules),
	)
}

func TestResourceDatacenterSecurityPolicy(t *testing.T) {
	ctx := context.Background()

	bpClient := testutils.BlueprintA(t, ctx)

	// collect leaf switch IDs
	leafIds := systemIds(ctx, t, bpClient, "leaf")

	// create virtual networks
	vnIds := make([]apstra.ObjectId, 2)
	for i := range vnIds {
		id, err := bpClient.CreateVirtualNetwork(ctx, &apstra.VirtualNetworkData{
			Ipv4Enabled: true,
			Label:       acctest.RandString(5),
			VnBindings:  []apstra.VnBinding{{SystemId: apstra.ObjectId(leafIds[i])}},
			VnType:      apstra.VnTypeVlan,
		})
		if err != nil {
			t.Fatal(err)
		}
		vnIds[i] = id
	}

	tests := []testCaseResourceSecurityPolicy{
		{
			config: apstra.PolicyData{
				Label: "1",
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "1"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
			},
		},
		{
			config: apstra.PolicyData{
				Label:       "2",
				Enabled:     true,
				Description: "two",
				Tags:        []string{"a", "b", "c"},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "2"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "description", "two"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "tags.#", "3"),
				resource.TestCheckTypeSetElemAttr(resourceDataCenterSecurityPolicyRefName, "tags.*", "c"),
				resource.TestCheckTypeSetElemAttr(resourceDataCenterSecurityPolicyRefName, "tags.*", "a"),
				resource.TestCheckTypeSetElemAttr(resourceDataCenterSecurityPolicyRefName, "tags.*", "b"),
			},
		},
		{
			config: apstra.PolicyData{
				Label:   "3",
				Enabled: false,
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "3"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
			},
		},
		{
			config: apstra.PolicyData{
				Label:               "4",
				Enabled:             true,
				SrcApplicationPoint: &apstra.PolicyApplicationPointData{Id: vnIds[0]},
				DstApplicationPoint: &apstra.PolicyApplicationPointData{Id: vnIds[1]},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "4"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "source_application_point_id", vnIds[0].String()),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "destination_application_point_id", vnIds[1].String()),
			},
		},
		{
			config: apstra.PolicyData{
				Label:               "5",
				Enabled:             false,
				SrcApplicationPoint: &apstra.PolicyApplicationPointData{Id: vnIds[1]},
				DstApplicationPoint: &apstra.PolicyApplicationPointData{Id: vnIds[0]},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "5"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "source_application_point_id", vnIds[1].String()),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "destination_application_point_id", vnIds[0].String()),
			},
		},
		{
			config: apstra.PolicyData{
				Label:   "6",
				Enabled: true,
				Rules: []apstra.PolicyRule{
					{
						Data: &apstra.PolicyRuleData{
							Label:       "60",
							Description: "",
							Protocol:    apstra.PolicyRuleProtocolIcmp,
							Action:      apstra.PolicyRuleActionDeny,
						},
					},
				},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "6"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.#", "1"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.name", "60"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "icmp"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.action", "deny"),
			},
		},
		{
			config: apstra.PolicyData{
				Label:   "7",
				Enabled: false,
				Rules: []apstra.PolicyRule{
					{
						Data: &apstra.PolicyRuleData{
							Label:       "70",
							Description: "seventy",
							Protocol:    apstra.PolicyRuleProtocolIp,
							Action:      apstra.PolicyRuleActionDenyLog,
						},
					},
				},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "7"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.#", "1"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.name", "70"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.description", "seventy"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "ip"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.action", "deny_log"),
			},
		},
		{
			config: apstra.PolicyData{
				Label:               "8",
				Enabled:             true,
				SrcApplicationPoint: &apstra.PolicyApplicationPointData{Id: vnIds[0]},
				DstApplicationPoint: &apstra.PolicyApplicationPointData{Id: vnIds[1]},
				Rules: []apstra.PolicyRule{
					{
						Data: &apstra.PolicyRuleData{
							Label:       "80",
							Description: "eighty",
							Protocol:    apstra.PolicyRuleProtocolUdp,
							Action:      apstra.PolicyRuleActionPermit,
						},
					},
					{
						Data: &apstra.PolicyRuleData{
							Label:       "81",
							Description: "eightyone",
							Protocol:    apstra.PolicyRuleProtocolTcp,
							Action:      apstra.PolicyRuleActionPermitLog,
						},
					},
					{
						Data: &apstra.PolicyRuleData{
							Label:             "82",
							Description:       "eightytwo",
							Protocol:          apstra.PolicyRuleProtocolTcp,
							Action:            apstra.PolicyRuleActionPermit,
							TcpStateQualifier: &apstra.TcpStateQualifierEstablished,
							SrcPort: apstra.PortRanges{
								{First: 1, Last: 1},
								{First: 3, Last: 5},
								{First: 7, Last: 9},
								{First: 11, Last: 11},
							},
							DstPort: apstra.PortRanges{
								{First: 50, Last: 50},
							},
						},
					},
				},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "8"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "source_application_point_id", vnIds[0].String()),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "destination_application_point_id", vnIds[1].String()),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.#", "3"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.name", "80"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.description", "eighty"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "udp"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.action", "permit"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.name", "81"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.description", "eightyone"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.protocol", "tcp"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.action", "permit_log"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.established", "false"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.name", "82"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.description", "eightytwo"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.protocol", "tcp"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.action", "permit"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.established", "true"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.source_ports.#", "4"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.destination_ports.#", "1"),
			},
			minVersion: version.Must(version.NewVersion(apiversions.Apstra412)),
		},
		{
			config: apstra.PolicyData{
				Label:               "9",
				Enabled:             true,
				SrcApplicationPoint: &apstra.PolicyApplicationPointData{Id: vnIds[0]},
				DstApplicationPoint: &apstra.PolicyApplicationPointData{Id: vnIds[1]},
				Rules: []apstra.PolicyRule{
					{
						Data: &apstra.PolicyRuleData{
							Label:       "90",
							Description: "ninety",
							Protocol:    apstra.PolicyRuleProtocolUdp,
							Action:      apstra.PolicyRuleActionPermit,
						},
					},
					{
						Data: &apstra.PolicyRuleData{
							Label:       "91",
							Description: "ninetyone",
							Protocol:    apstra.PolicyRuleProtocolTcp,
							Action:      apstra.PolicyRuleActionPermitLog,
						},
					},
					{
						Data: &apstra.PolicyRuleData{
							Label:             "92",
							Description:       "ninetytwo",
							Protocol:          apstra.PolicyRuleProtocolTcp,
							Action:            apstra.PolicyRuleActionPermit,
							TcpStateQualifier: &apstra.TcpStateQualifierEstablished,
							SrcPort: apstra.PortRanges{
								{First: 1, Last: 1},
								{First: 7, Last: 9},
								{First: 11, Last: 11},
							},
							DstPort: apstra.PortRanges{
								{First: 50, Last: 50},
								{First: 3, Last: 5},
							},
						},
					},
				},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "9"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "true"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "source_application_point_id", vnIds[0].String()),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "destination_application_point_id", vnIds[1].String()),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.#", "3"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.name", "90"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.description", "ninety"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "udp"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.0.action", "permit"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.name", "91"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.description", "ninetyone"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.protocol", "tcp"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.action", "permit_log"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.1.established", "false"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.name", "92"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.description", "ninetytwo"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.protocol", "tcp"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.action", "permit"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.established", "true"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.source_ports.#", "3"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "rules.2.destination_ports.#", "2"),
			},
			minVersion: version.Must(version.NewVersion(apiversions.Apstra412)),
		},
		{
			config: apstra.PolicyData{
				Label:   "10",
				Enabled: false,
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceDataCenterSecurityPolicyRefName, "id"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "name", "10"),
				resource.TestCheckResourceAttr(resourceDataCenterSecurityPolicyRefName, "enabled", "false"),
			},
		},
	}

	var steps []resource.TestStep
	for _, test := range tests {
		if test.minVersion != nil && version.Must(version.NewVersion(bpClient.Client().ApiVersion())).LessThan(test.minVersion) {
			continue
		}
		steps = append(steps, resource.TestStep{
			Config: test.renderConfig(bpClient.Id()),
			Check:  resource.ComposeAggregateTestCheckFunc(test.checks...),
		})
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}

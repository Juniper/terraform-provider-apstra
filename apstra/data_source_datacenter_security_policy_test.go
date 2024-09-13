package tfapstra_test

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	dataSourceDataCenterSecurityPolicyRefName = "data.apstra_datacenter_security_policy.test"
	dataSourceDataCenterSecurityPolicyByIdHCL = `
data "apstra_datacenter_security_policy" "test" {
  blueprint_id = "%s"
  id = "%s"
}
`

	dataSourceDataCenterSecurityPolicyByNameHCL = `
data "apstra_datacenter_security_policy" "test" {
  blueprint_id = "%s"
  name = "%s"
}
`
)

func TestDatacenterSecurityPolicy(t *testing.T) {
	ctx := context.Background()

	bp := testutils.BlueprintC(t, ctx)

	szs, err := bp.GetAllSecurityZones(ctx)
	require.NoError(t, err)
	if len(szs) != 1 {
		t.Fatalf("expected one security zone, got %d", len(szs))
	}

	szId := szs[0].Id

	testCases := map[string]struct {
		id         string
		policyData apstra.PolicyData
		checks     []resource.TestCheckFunc
	}{
		"minimal_enabled": {
			policyData: apstra.PolicyData{
				Label:   "minimal_enabled",
				Enabled: true,
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "id", "bogus placeholder fixed in loop below. this check must be first"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "enabled", "true"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "name", "minimal_enabled"),
			},
		},
		"minimal_disabled": {
			policyData: apstra.PolicyData{
				Label:   "minimal_disabled",
				Enabled: false,
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "id", "bogus placeholder fixed in loop below. this check must be first"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "enabled", "false"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "name", "minimal_disabled"),
			},
		},
		"with_description": {
			policyData: apstra.PolicyData{
				Label:       "with_description",
				Description: "with_description ... description",
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "id", "bogus placeholder fixed in loop below. this check must be first"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "name", "with_description"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "description", "with_description ... description"),
			},
		},
		"with_src_app_point": {
			policyData: apstra.PolicyData{
				Label:               "with_src_app_point",
				SrcApplicationPoint: &apstra.PolicyApplicationPointData{Id: szId},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "id", "bogus placeholder fixed in loop below. this check must be first"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "name", "with_src_app_point"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "source_application_point_id", szId.String()),
			},
		},
		"with_dst_app_point": {
			policyData: apstra.PolicyData{
				Label:               "with_dst_app_point",
				DstApplicationPoint: &apstra.PolicyApplicationPointData{Id: szId},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "id", "bogus placeholder fixed in loop below. this check must be first"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "name", "with_dst_app_point"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "destination_application_point_id", szId.String()),
			},
		},
		"with_tags": {
			policyData: apstra.PolicyData{
				Label:               "with_tags",
				DstApplicationPoint: &apstra.PolicyApplicationPointData{Id: szId},
				Tags:                []string{"foo", "bar"},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "id", "bogus placeholder fixed in loop below. this check must be first"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "name", "with_tags"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "tags.#", "2"),
				resource.TestCheckTypeSetElemAttr(dataSourceDataCenterSecurityPolicyRefName, "tags.*", "foo"),
				resource.TestCheckTypeSetElemAttr(dataSourceDataCenterSecurityPolicyRefName, "tags.*", "bar"),
			},
		},
		"with_rules": {
			policyData: apstra.PolicyData{
				Label: "with_rules",
				Rules: []apstra.PolicyRule{
					{
						Data: &apstra.PolicyRuleData{
							Label:       "name_0",
							Description: "description_0",
							Protocol:    enum.PolicyRuleProtocolIcmp,
							Action:      enum.PolicyRuleActionDeny,
						},
					},
					{
						Data: &apstra.PolicyRuleData{
							Label:       "name_1",
							Description: "description_1",
							Protocol:    enum.PolicyRuleProtocolIp,
							Action:      enum.PolicyRuleActionDenyLog,
						},
					},
					{
						Data: &apstra.PolicyRuleData{
							Label:       "name_2",
							Description: "description_2",
							Protocol:    enum.PolicyRuleProtocolTcp,
							Action:      enum.PolicyRuleActionPermitLog,
						},
					},
					{
						Data: &apstra.PolicyRuleData{
							Label:       "name_3",
							Description: "description_3",
							Protocol:    enum.PolicyRuleProtocolTcp,
							Action:      enum.PolicyRuleActionPermitLog,
							SrcPort: []apstra.PortRange{
								{First: 11, Last: 11},
								{First: 13, Last: 13},
								{First: 15, Last: 19},
							},
							DstPort: []apstra.PortRange{
								{First: 21, Last: 21},
								{First: 23, Last: 23},
								{First: 25, Last: 25},
								{First: 27, Last: 29},
							},
						},
					},
				},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "id", "bogus placeholder fixed in loop below. this check must be first"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "name", "with_rules"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.#", "4"),

				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.name", "name_0"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.description", "description_0"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "icmp"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.action", "deny"),

				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.1.name", "name_1"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.1.description", "description_1"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.1.protocol", "ip"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.1.action", "deny_log"),

				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.2.name", "name_2"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.2.description", "description_2"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.2.protocol", "tcp"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.2.action", "permit_log"),

				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.3.name", "name_3"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.3.description", "description_3"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.3.protocol", "tcp"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.3.action", "permit_log"),

				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.3.source_ports.#", "3"),
				resource.TestCheckTypeSetElemNestedAttrs(dataSourceDataCenterSecurityPolicyRefName, "rules.3.source_ports.*", map[string]string{"from_port": "11", "to_port": "11"}),
				resource.TestCheckTypeSetElemNestedAttrs(dataSourceDataCenterSecurityPolicyRefName, "rules.3.source_ports.*", map[string]string{"from_port": "13", "to_port": "13"}),
				resource.TestCheckTypeSetElemNestedAttrs(dataSourceDataCenterSecurityPolicyRefName, "rules.3.source_ports.*", map[string]string{"from_port": "15", "to_port": "19"}),

				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.3.destination_ports.#", "4"),
				resource.TestCheckTypeSetElemNestedAttrs(dataSourceDataCenterSecurityPolicyRefName, "rules.3.destination_ports.*", map[string]string{"from_port": "21", "to_port": "21"}),
				resource.TestCheckTypeSetElemNestedAttrs(dataSourceDataCenterSecurityPolicyRefName, "rules.3.destination_ports.*", map[string]string{"from_port": "23", "to_port": "23"}),
				resource.TestCheckTypeSetElemNestedAttrs(dataSourceDataCenterSecurityPolicyRefName, "rules.3.destination_ports.*", map[string]string{"from_port": "25", "to_port": "25"}),
				resource.TestCheckTypeSetElemNestedAttrs(dataSourceDataCenterSecurityPolicyRefName, "rules.3.destination_ports.*", map[string]string{"from_port": "27", "to_port": "29"}),
			},
		},
		"with_established": {
			policyData: apstra.PolicyData{
				Label: "with_established",
				Rules: []apstra.PolicyRule{
					{
						Data: &apstra.PolicyRuleData{
							Label:             "ssh_established",
							Description:       "ssh established",
							Protocol:          enum.PolicyRuleProtocolTcp,
							Action:            enum.PolicyRuleActionPermitLog,
							SrcPort:           []apstra.PortRange{{First: 22, Last: 22}},
							TcpStateQualifier: &enum.TcpStateQualifierEstablished,
						},
					},
				},
			},
			checks: []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "id", "bogus placeholder fixed in loop below. this check must be first"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "name", "with_established"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.#", "1"),

				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.name", "ssh_established"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.description", "ssh established"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.protocol", "tcp"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.action", "permit_log"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.established", "true"),
				resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "rules.0.source_ports.#", "1"),
				resource.TestCheckTypeSetElemNestedAttrs(dataSourceDataCenterSecurityPolicyRefName, "rules.0.source_ports.*", map[string]string{"from_port": "22", "to_port": "22"}),
			},
		},
	}

	for i, tc := range testCases {
		id, err := bp.CreatePolicy(ctx, &tc.policyData)
		if err != nil {
			t.Fatal(err)
		}

		tc.id = id.String()
		tc.checks[0] = resource.TestCheckResourceAttr(dataSourceDataCenterSecurityPolicyRefName, "id", id.String())
		testCases[i] = tc
	}

	for tName, tCase := range testCases {

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDataCenterSecurityPolicyByIdHCL, bp.Id(), tCase.id),
					Check:  resource.ComposeAggregateTestCheckFunc(tCase.checks...),
				},
				{
					Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceDataCenterSecurityPolicyByNameHCL, bp.Id(), tName),
					Check:  resource.ComposeAggregateTestCheckFunc(tCase.checks...),
				},
			},
		})
	}
}

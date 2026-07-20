//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/compatibility"
	"github.com/Juniper/apstra-go-sdk/datacenter"
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

	resourceDataCenterSecurityPolicyHCL = `resource %q %q {
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
		sb.WriteString("[")
		for _, rule := range r.rules {
			sb.WriteString(rule.render())
		}
		sb.WriteString("  ]")
		rules = sb.String()
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
	action           enum.PolicyRuleAction
	protocol         enum.PolicyRuleProtocol
	sourcePorts      []resourceDataCenterSecurityPolicyRulePort
	destinationPorts []resourceDataCenterSecurityPolicyRulePort
	established      *bool
}

func (r resourceDataCenterSecurityPolicyRule) render() string {
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
		r.action,
		rosetta.StringersToFriendlyString(r.protocol),
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
		c.appendSetNestedCheck(t, p+"source_ports.*", map[string]string{
			"from_port": strconv.Itoa(port.from),
			"to_port":   strconv.Itoa(port.to),
		})
	}
	c.append(t, "TestCheckResourceAttr", p+"destination_ports.#", strconv.Itoa(len(r.destinationPorts)))
	for _, port := range r.destinationPorts {
		c.appendSetNestedCheck(t, p+"destination_ports.*", map[string]string{
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

	// collect leaf switch IDs
	leafIds := systemIds(ctx, t, bp, "leaf")

	// create virtual networks
	vnIds := make([]string, 2)
	for i := range vnIds {
		id, err := bp.CreateVirtualNetwork(ctx, datacenter.VirtualNetwork{
			IPv4Enabled: true,
			Label:       acctest.RandString(5),
			Bindings:    []datacenter.VNBinding{{SystemID: leafIds[i]}},
			Type:        enum.VnTypeVlan,
		})
		if err != nil {
			t.Fatal(err)
		}
		vnIds[i] = id
	}

	type testCase struct {
		versionconstraints []compatibility.Constraint
		steps              []resourceDataCenterSecurityPolicy
	}

	testCases := map[string]testCase{
		"start_minimal_pre_620": {
			versionconstraints: []compatibility.Constraint{compatibility.DatacenterPolicyAddressFamilyNotSupported},
			steps: []resourceDataCenterSecurityPolicy{
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
				},
				{
					blueprintID:                   string(bp.Id()),
					name:                          acctest.RandString(6),
					description:                   acctest.RandString(10),
					enabled:                       random.OneOf(pointer.To(true), pointer.To(false), nil),
					sourceApplicationPointID:      vnIds[0],
					destinationApplicationPointID: vnIds[1],
					tags:                          randomStrings(3, 6),
					rules: []resourceDataCenterSecurityPolicyRule{
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIcmp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10000, to: 10000},
								{from: 10010, to: 10020},
								{from: 10030, to: 10040},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20000, to: 20000},
								{from: 20010, to: 20020},
								{from: 20030, to: 20040},
							},
							established: pointer.To(true),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10050, to: 10050},
								{from: 10060, to: 10070},
								{from: 10080, to: 10090},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20050, to: 20050},
								{from: 20060, to: 20070},
								{from: 20080, to: 20090},
							},
							established: pointer.To(false),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10100, to: 10100},
								{from: 10110, to: 10120},
								{from: 10130, to: 10140},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20100, to: 20100},
								{from: 20110, to: 20120},
								{from: 20130, to: 20140},
							},
							established: nil,
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolUdp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10150, to: 10150},
								{from: 10160, to: 10170},
								{from: 10180, to: 10190},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20150, to: 20150},
								{from: 20160, to: 20170},
								{from: 20180, to: 20190},
							},
						},
					},
				},
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
				},
			},
		},
		"start_maximal_pre_620": {
			versionconstraints: []compatibility.Constraint{compatibility.DatacenterPolicyAddressFamilyNotSupported},
			steps: []resourceDataCenterSecurityPolicy{
				{
					blueprintID:                   string(bp.Id()),
					name:                          acctest.RandString(6),
					description:                   acctest.RandString(10),
					enabled:                       random.OneOf(pointer.To(true), pointer.To(false), nil),
					sourceApplicationPointID:      vnIds[0],
					destinationApplicationPointID: vnIds[1],
					tags:                          randomStrings(3, 6),
					rules: []resourceDataCenterSecurityPolicyRule{
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIcmp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10000, to: 10000},
								{from: 10010, to: 10020},
								{from: 10030, to: 10040},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20000, to: 20000},
								{from: 20010, to: 20020},
								{from: 20030, to: 20040},
							},
							established: pointer.To(true),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10050, to: 10050},
								{from: 10060, to: 10070},
								{from: 10080, to: 10090},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20050, to: 20050},
								{from: 20060, to: 20070},
								{from: 20080, to: 20090},
							},
							established: pointer.To(false),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10100, to: 10100},
								{from: 10110, to: 10120},
								{from: 10130, to: 10140},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20100, to: 20100},
								{from: 20110, to: 20120},
								{from: 20130, to: 20140},
							},
							established: nil,
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolUdp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10150, to: 10150},
								{from: 10160, to: 10170},
								{from: 10180, to: 10190},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20150, to: 20150},
								{from: 20160, to: 20170},
								{from: 20180, to: 20190},
							},
						},
					},
				},
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
				},
				{
					blueprintID:                   string(bp.Id()),
					name:                          acctest.RandString(6),
					description:                   acctest.RandString(10),
					enabled:                       random.OneOf(pointer.To(true), pointer.To(false), nil),
					sourceApplicationPointID:      vnIds[0],
					destinationApplicationPointID: vnIds[1],
					tags:                          randomStrings(3, 6),
					rules: []resourceDataCenterSecurityPolicyRule{
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIcmp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10000, to: 10000},
								{from: 10010, to: 10020},
								{from: 10030, to: 10040},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20000, to: 20000},
								{from: 20010, to: 20020},
								{from: 20030, to: 20040},
							},
							established: pointer.To(true),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10050, to: 10050},
								{from: 10060, to: 10070},
								{from: 10080, to: 10090},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20050, to: 20050},
								{from: 20060, to: 20070},
								{from: 20080, to: 20090},
							},
							established: pointer.To(false),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10100, to: 10100},
								{from: 10110, to: 10120},
								{from: 10130, to: 10140},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20100, to: 20100},
								{from: 20110, to: 20120},
								{from: 20130, to: 20140},
							},
							established: nil,
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolUdp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10150, to: 10150},
								{from: 10160, to: 10170},
								{from: 10180, to: 10190},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20150, to: 20150},
								{from: 20160, to: 20170},
								{from: 20180, to: 20190},
							},
						},
					},
				},
			},
		},
		"start_minimal_620_and_later": {
			versionconstraints: []compatibility.Constraint{compatibility.DatacenterPolicyAddressFamilyRequired},
			steps: []resourceDataCenterSecurityPolicy{
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
					ipVersion:   pointer.To(random.OneOf(enum.PolicyAddressFamilyIPv4, enum.PolicyAddressFamilyIPv6, enum.PolicyAddressFamilyIPv6)),
				},
				{
					blueprintID:                   string(bp.Id()),
					name:                          acctest.RandString(6),
					description:                   acctest.RandString(10),
					enabled:                       random.OneOf(pointer.To(true), pointer.To(false), nil),
					ipVersion:                     pointer.To(random.OneOf(enum.PolicyAddressFamilyIPv4, enum.PolicyAddressFamilyIPv6, enum.PolicyAddressFamilyIPv6)),
					sourceApplicationPointID:      vnIds[0],
					destinationApplicationPointID: vnIds[1],
					tags:                          randomStrings(3, 6),
					rules: []resourceDataCenterSecurityPolicyRule{
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIcmp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10000, to: 10000},
								{from: 10010, to: 10020},
								{from: 10030, to: 10040},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20000, to: 20000},
								{from: 20010, to: 20020},
								{from: 20030, to: 20040},
							},
							established: pointer.To(true),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10050, to: 10050},
								{from: 10060, to: 10070},
								{from: 10080, to: 10090},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20050, to: 20050},
								{from: 20060, to: 20070},
								{from: 20080, to: 20090},
							},
							established: pointer.To(false),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10100, to: 10100},
								{from: 10110, to: 10120},
								{from: 10130, to: 10140},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20100, to: 20100},
								{from: 20110, to: 20120},
								{from: 20130, to: 20140},
							},
							established: nil,
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolUdp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10150, to: 10150},
								{from: 10160, to: 10170},
								{from: 10180, to: 10190},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20150, to: 20150},
								{from: 20160, to: 20170},
								{from: 20180, to: 20190},
							},
						},
					},
				},
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
					ipVersion:   pointer.To(random.OneOf(enum.PolicyAddressFamilyIPv4, enum.PolicyAddressFamilyIPv6, enum.PolicyAddressFamilyIPv6)),
				},
			},
		},
		"start_maximal_620_and_later": {
			versionconstraints: []compatibility.Constraint{compatibility.DatacenterPolicyAddressFamilyRequired},
			steps: []resourceDataCenterSecurityPolicy{
				{
					blueprintID:                   string(bp.Id()),
					name:                          acctest.RandString(6),
					description:                   acctest.RandString(10),
					enabled:                       random.OneOf(pointer.To(true), pointer.To(false), nil),
					ipVersion:                     pointer.To(random.OneOf(enum.PolicyAddressFamilyIPv4, enum.PolicyAddressFamilyIPv6, enum.PolicyAddressFamilyIPv6)),
					sourceApplicationPointID:      vnIds[0],
					destinationApplicationPointID: vnIds[1],
					tags:                          randomStrings(3, 6),
					rules: []resourceDataCenterSecurityPolicyRule{
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIcmp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10000, to: 10000},
								{from: 10010, to: 10020},
								{from: 10030, to: 10040},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20000, to: 20000},
								{from: 20010, to: 20020},
								{from: 20030, to: 20040},
							},
							established: pointer.To(true),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10050, to: 10050},
								{from: 10060, to: 10070},
								{from: 10080, to: 10090},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20050, to: 20050},
								{from: 20060, to: 20070},
								{from: 20080, to: 20090},
							},
							established: pointer.To(false),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10100, to: 10100},
								{from: 10110, to: 10120},
								{from: 10130, to: 10140},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20100, to: 20100},
								{from: 20110, to: 20120},
								{from: 20130, to: 20140},
							},
							established: nil,
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolUdp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10150, to: 10150},
								{from: 10160, to: 10170},
								{from: 10180, to: 10190},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20150, to: 20150},
								{from: 20160, to: 20170},
								{from: 20180, to: 20190},
							},
						},
					},
				},
				{
					blueprintID: string(bp.Id()),
					name:        acctest.RandString(6),
					ipVersion:   pointer.To(random.OneOf(enum.PolicyAddressFamilyIPv4, enum.PolicyAddressFamilyIPv6, enum.PolicyAddressFamilyIPv6)),
				},
				{
					blueprintID:                   string(bp.Id()),
					name:                          acctest.RandString(6),
					description:                   acctest.RandString(10),
					enabled:                       random.OneOf(pointer.To(true), pointer.To(false), nil),
					ipVersion:                     pointer.To(random.OneOf(enum.PolicyAddressFamilyIPv4, enum.PolicyAddressFamilyIPv6, enum.PolicyAddressFamilyIPv6)),
					sourceApplicationPointID:      vnIds[0],
					destinationApplicationPointID: vnIds[1],
					tags:                          randomStrings(3, 6),
					rules: []resourceDataCenterSecurityPolicyRule{
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIcmp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:     acctest.RandString(5),
							protocol: enum.PolicyRuleProtocolIp,
							action:   random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10000, to: 10000},
								{from: 10010, to: 10020},
								{from: 10030, to: 10040},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20000, to: 20000},
								{from: 20010, to: 20020},
								{from: 20030, to: 20040},
							},
							established: pointer.To(true),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10050, to: 10050},
								{from: 10060, to: 10070},
								{from: 10080, to: 10090},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20050, to: 20050},
								{from: 20060, to: 20070},
								{from: 20080, to: 20090},
							},
							established: pointer.To(false),
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolTcp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10100, to: 10100},
								{from: 10110, to: 10120},
								{from: 10130, to: 10140},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20100, to: 20100},
								{from: 20110, to: 20120},
								{from: 20130, to: 20140},
							},
							established: nil,
						},
						{
							name:        acctest.RandString(5),
							description: acctest.RandString(10),
							action:      random.OneOf(enum.PolicyRuleActionDeny, enum.PolicyRuleActionDenyLog, enum.PolicyRuleActionPermit, enum.PolicyRuleActionPermitLog),
							protocol:    enum.PolicyRuleProtocolUdp,
							sourcePorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 10150, to: 10150},
								{from: 10160, to: 10170},
								{from: 10180, to: 10190},
							},
							destinationPorts: []resourceDataCenterSecurityPolicyRulePort{
								{from: 20150, to: 20150},
								{from: 20160, to: 20170},
								{from: 20180, to: 20190},
							},
						},
					},
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

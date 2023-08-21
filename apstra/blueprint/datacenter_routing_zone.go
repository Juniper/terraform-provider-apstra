package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	"regexp"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/resources"
	"terraform-provider-apstra/apstra/utils"
)

const (
	errInvalidConfig = "invalid configuration"
)

type DatacenterRoutingZone struct {
	Id                   types.String `tfsdk:"id"`
	BlueprintId          types.String `tfsdk:"blueprint_id"`
	Name                 types.String `tfsdk:"name"`
	VlanId               types.Int64  `tfsdk:"vlan_id"`
	HadPriorVlanIdConfig types.Bool   `tfsdk:"had_prior_vlan_id_config"`
	Vni                  types.Int64  `tfsdk:"vni"`
	HadPriorVniConfig    types.Bool   `tfsdk:"had_prior_vni_config"`
	DhcpServers          types.Set    `tfsdk:"dhcp_servers"`
	RoutingPolicyId      types.String `tfsdk:"routing_policy_id"`
}

func (o DatacenterRoutingZone) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	nameRE := regexp.MustCompile("^[A-Za-z0-9_-]+$")
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID. Required when `name` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Required when `id` is omitted.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "VRF name displayed in thw Apstra web UI.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(0, 18),
				stringvalidator.RegexMatches(nameRE, "only underscore, dash and alphanumeric characters allowed."),
			},
		},
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Used for VLAN tagged Layer 3 links on external connections. " +
				"Leave this field blank to have it automatically assigned from a static pool in the " +
				"range of 2-4094), or enter a specific value.",
			Computed: true,
		},
		"had_prior_vlan_id_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vlan_id` has been removed from the " +
				"configuration in managed resource context, this attribute will always be `null` and should be " +
				"ignored in data source context.",
			Computed: true,
		},
		"vni": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "VxLAN VNI associated with the routing zone. Leave this field blank to have it " +
				"automatically assigned from an allocated resource pool, or enter a specific value.",
			Computed: true,
		},
		"had_prior_vni_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vni` has been removed from the " +
				"configuration in managed resource context, this attribute will always be `null` and should be " +
				"ignored in data source context.",
			Computed: true,
		},
		"dhcp_servers": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of DHCP server IPv4 or IPv6 addresses of DHCP servers.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"routing_policy_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Non-EVPN blueprints must use the default policy, so this field must be null. " +
				"Set this attribute in an EVPN blueprint to use a non-default policy.",
			Computed: true,
		},
	}
}

func (o DatacenterRoutingZone) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "VRF name displayed in thw Apstra web UI.",
			Optional:            true,
		},
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Used for VLAN tagged Layer 3 links on external connections.",
			Optional:            true,
		},
		"had_prior_vlan_id_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"vni": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "VxLAN VNI associated with the routing zone.",
			Optional:            true,
		},
		"had_prior_vni_config": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"dhcp_servers": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of addresses of DHCP servers (IPv4 or IPv6) which must be configured " +
				"in the Routing Zone. This is a list of *required* servers, not an exact-match list.",
			ElementType: types.StringType,
			Optional:    true,
		},
		"routing_policy_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Non-EVPN blueprints must use the default policy, so this field must be null. " +
				"Set this attribute in an EVPN blueprint to use a non-default policy.",
			Optional: true,
		},
	}
}

func (o DatacenterRoutingZone) ResourceAttributes() map[string]resourceSchema.Attribute {
	nameRE := regexp.MustCompile("^[A-Za-z0-9_-]+$")
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "VRF name displayed in thw Apstra web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(nameRE, "only underscore, dash and alphanumeric characters allowed."),
				stringvalidator.LengthBetween(0, 15),
			},
		},
		"vlan_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "Used for VLAN tagged Layer 3 links on external connections. " +
				"Leave this field blank to have it automatically assigned from a static pool in the " +
				"range of 2-4094, or enter a specific value.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.Int64{int64validator.Between(design.VlanMin, design.VlanMax)},
		},
		"had_prior_vlan_id_config": resourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vlan_id` has been removed from the " +
				"configuration, this attribute can be ignored.",
			Computed: true,
		},
		"vni": resourceSchema.Int64Attribute{
			MarkdownDescription: "VxLAN VNI associated with the routing zone. Leave this field blank to have it " +
				"automatically assigned from an allocated resource pool, or enter a specific value.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.Int64{int64validator.Between(resources.VniMin, resources.VniMax)},
		},
		"had_prior_vni_config": resourceSchema.BoolAttribute{
			MarkdownDescription: "Used to trigger plan modification when `vni` has been removed from the " +
				"configuration, this attribute can be ignored.",
			Computed: true,
		},
		"dhcp_servers": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of DHCP server IPv4 or IPv6 addresses of DHCP servers.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseIp(false, false)),
			},
		},
		"routing_policy_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Non-EVPN blueprints must use the default policy, so this field must be null. " +
				"Set this attribute in an EVPN blueprint to use a non-default policy.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o *DatacenterRoutingZone) Request(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) *apstra.SecurityZoneData {
	var vlan *apstra.Vlan
	if !o.VlanId.IsNull() && !o.VlanId.IsUnknown() {
		v := apstra.Vlan(o.VlanId.ValueInt64())
		vlan = &v
	}

	var vni *int
	if !o.Vni.IsNull() && !o.Vni.IsUnknown() {
		v := int(o.Vni.ValueInt64())
		vni = &v
	}

	ocp, err := client.BlueprintOverlayControlProtocol(ctx, apstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		diags.AddError(fmt.Sprintf("API error querying for blueprint %q Overlay Control Protocol",
			o.BlueprintId.ValueString()), err.Error())
		return nil
	}

	if ocp != apstra.OverlayControlProtocolEvpn {
		diags.AddAttributeError(
			path.Root("blueprint_id"),
			errInvalidConfig,
			fmt.Sprintf("cannot create routing zone in blueprints with overlay control protocol %q", ocp.String())) // todo: need rosetta treatment
	}

	return &apstra.SecurityZoneData{
		SzType:          apstra.SecurityZoneTypeEVPN,
		VrfName:         o.Name.ValueString(),
		Label:           o.Name.ValueString(),
		RoutingPolicyId: apstra.ObjectId(o.RoutingPolicyId.ValueString()),
		VlanId:          vlan,
		VniId:           vni,
	}
}

func (o *DatacenterRoutingZone) DhcpServerRequest(_ context.Context, _ *diag.Diagnostics) []net.IP {
	dhcpServers := o.DhcpServers.Elements()
	request := make([]net.IP, len(dhcpServers))
	for i, dhcpServer := range dhcpServers {
		request[i] = net.ParseIP(dhcpServer.(types.String).ValueString())
	}
	return request
}

func (o *DatacenterRoutingZone) LoadApiData(_ context.Context, sz *apstra.SecurityZoneData, _ *diag.Diagnostics) {
	o.Name = types.StringValue(sz.VrfName)

	if sz.VlanId != nil {
		o.VlanId = types.Int64Value(int64(*sz.VlanId))
	} else {
		o.VlanId = types.Int64Null()
	}

	if sz.RoutingPolicyId != "" {
		o.RoutingPolicyId = types.StringValue(sz.RoutingPolicyId.String())
	} else {
		o.RoutingPolicyId = types.StringNull()
	}

	if sz.VniId != nil {
		o.Vni = types.Int64Value(int64(*sz.VniId))
	} else {
		o.Vni = types.Int64Null()
	}
}

func (o *DatacenterRoutingZone) LoadApiDhcpServers(ctx context.Context, IPs []net.IP, diags *diag.Diagnostics) {
	dhcpServers := make([]string, len(IPs))
	for i, ip := range IPs {
		dhcpServers[i] = ip.String()
	}
	o.DhcpServers = utils.SetValueOrNull(ctx, types.StringType, dhcpServers, diags)
}

func (o *DatacenterRoutingZone) Query(szResultName, policyResultName string) *apstra.PathQuery {
	query := new(apstra.PathQuery)

	if utils.Known(o.RoutingPolicyId) {
		query.Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeRoutingPolicy.String())},
			{Key: "id", Value: apstra.QEStringVal(o.RoutingPolicyId.ValueString())},
		})
		query.In([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("policy")},
		})
	}

	query.Node(o.szNodeQueryAttributes(szResultName))

	if utils.Known(o.DhcpServers) {
		query.Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("policy")},
		})
		query.Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypePolicy.String())},
			{Key: "policy_type", Value: apstra.QEStringVal("dhcp_relay")},
			{Key: "name", Value: apstra.QEStringVal(policyResultName)},
		})
	}

	for _, dhcpServerVal := range o.DhcpServers.Elements() {
		query.Where("lambda " +
			policyResultName +
			": '" +
			dhcpServerVal.(types.String).ValueString() +
			"' in " +
			policyResultName +
			".dhcp_servers")
	}
	return query
}

func (o *DatacenterRoutingZone) szNodeQueryAttributes(name string) []apstra.QEEAttribute {
	result := []apstra.QEEAttribute{
		{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeSecurityZone.String())},
	}

	if name != "" {
		result = append(result, apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal(name)})
	}

	if utils.Known(o.Name) {
		result = append(result, apstra.QEEAttribute{Key: "label", Value: apstra.QEStringVal(o.Name.ValueString())})
	}

	if utils.Known(o.Vni) {
		result = append(result, apstra.QEEAttribute{Key: "vni_id", Value: apstra.QEIntVal(int(o.Vni.ValueInt64()))})
	}

	if utils.Known(o.VlanId) {
		result = append(result, apstra.QEEAttribute{Key: "vlan_id", Value: apstra.QEIntVal(int(o.VlanId.ValueInt64()))})
	}

	return result
}

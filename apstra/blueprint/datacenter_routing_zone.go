package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/resources"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
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
)

var junosIrbModeDefault = apstra.JunosEvpnIrbModeAsymmetric.Value

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
	ImportRouteTargets   types.Set    `tfsdk:"import_route_targets"`
	ExportRouteTargets   types.Set    `tfsdk:"export_route_targets"`
	JunosIrbMode         types.String `tfsdk:"junos_irb_mode"`
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
			MarkdownDescription: "VRF name displayed in the Apstra web UI.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 17),
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
		"import_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Used to import routes into the EVPN VRF.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"export_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Used to export routes from the EVPN VRF.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"junos_irb_mode": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Symmetric IRB Routing for EVPN on Junos devices makes use of an L3 VNI for " +
				"inter-subnet routing which is embedded into EVPN Type2-routes to support better scaling for " +
				"networks with large amounts of VLANs.",
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
			MarkdownDescription: "VRF name displayed in the Apstra web UI.",
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
		"import_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "This is a set of *required* RTs, not an exact-match list.",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"export_route_targets": dataSourceSchema.SetAttribute{
			MarkdownDescription: "This is a set of *required* RTs, not an exact-match list.",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"junos_irb_mode": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Symmetric IRB Routing for EVPN on Junos devices makes use of an L3 VNI for " +
				"inter-subnet routing which is embedded into EVPN Type2-routes to support better scaling for " +
				"networks with large amounts of VLANs.",
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
			MarkdownDescription: "VRF name displayed in the Apstra web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(nameRE, "only underscore, dash and alphanumeric characters allowed."),
				stringvalidator.LengthBetween(1, 15),
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
		"import_route_targets": resourceSchema.SetAttribute{
			MarkdownDescription: "Used to import routes into the EVPN VRF.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseRT()),
			},
		},
		"export_route_targets": resourceSchema.SetAttribute{
			MarkdownDescription: "Used to export routes from the EVPN VRF.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseRT()),
			},
		},
		"junos_irb_mode": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Symmetric IRB Routing for EVPN on Junos devices makes use of "+
				"an L3 VNI for inter-subnet routing which is embedded into EVPN Type2-routes to support better "+
				"scaling for networks with large amounts of VLANs. Applicable only to Apstra 4.2.0+. When omitted, "+
				"Routing Zones in Apstra 4.2.0 and later will be configured with mode `%s`.", junosIrbModeDefault),
			Optional: true,
			Computed: true,
			Validators: []validator.String{stringvalidator.OneOf(
				apstra.JunosEvpnIrbModeSymmetric.Value,
				apstra.JunosEvpnIrbModeAsymmetric.Value,
			)},
			// Default: DO NOT USE stringdefault.StaticString(apstra.JunosEvpnIrbModeAsymmetric.Value) here
			// because that will set the attribute for Apstra < 4.2.0 (which do not support it) leading to
			// confusion.
		},
	}
}

func (o *DatacenterRoutingZone) Request(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) *apstra.SecurityZoneData {
	o.setDefaults(ctx, client, diags)
	if diags.HasError() {
		return nil
	}

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

	var importRTs, exportRTs []string
	diags.Append(o.ImportRouteTargets.ElementsAs(ctx, &importRTs, false)...)
	diags.Append(o.ExportRouteTargets.ElementsAs(ctx, &exportRTs, false)...)
	if diags.HasError() {
		return nil
	}

	var junosEvpnIrbMode *apstra.JunosEvpnIrbMode
	if !o.JunosIrbMode.IsNull() {
		junosEvpnIrbMode = &apstra.JunosEvpnIrbModeAsymmetric
	}

	return &apstra.SecurityZoneData{
		SzType:           apstra.SecurityZoneTypeEVPN,
		VrfName:          o.Name.ValueString(),
		Label:            o.Name.ValueString(),
		RoutingPolicyId:  apstra.ObjectId(o.RoutingPolicyId.ValueString()),
		VlanId:           vlan,
		VniId:            vni,
		JunosEvpnIrbMode: junosEvpnIrbMode,
		RtPolicy: &apstra.RtPolicy{
			ImportRTs: importRTs,
			ExportRTs: exportRTs,
		},
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

func (o *DatacenterRoutingZone) LoadApiData(ctx context.Context, sz *apstra.SecurityZoneData, diags *diag.Diagnostics) {
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

	if sz.RtPolicy != nil && sz.RtPolicy.ImportRTs != nil {
		o.ImportRouteTargets = utils.SetValueOrNull(ctx, types.StringType, sz.RtPolicy.ImportRTs, diags)
	}

	if sz.RtPolicy != nil && sz.RtPolicy.ExportRTs != nil {
		o.ExportRouteTargets = utils.SetValueOrNull(ctx, types.StringType, sz.RtPolicy.ExportRTs, diags)
	}

	o.JunosIrbMode = types.StringNull()
	if sz.JunosEvpnIrbMode != nil {
		o.JunosIrbMode = types.StringValue(sz.JunosEvpnIrbMode.Value)
	}
}

func (o *DatacenterRoutingZone) LoadApiDhcpServers(ctx context.Context, IPs []net.IP, diags *diag.Diagnostics) {
	dhcpServers := make([]string, len(IPs))
	for i, ip := range IPs {
		dhcpServers[i] = ip.String()
	}
	o.DhcpServers = utils.SetValueOrNull(ctx, types.StringType, dhcpServers, diags)
}

func (o *DatacenterRoutingZone) Query(szResultName string) *apstra.MatchQuery {
	matchQuery := new(apstra.MatchQuery)
	nodeQuery := new(apstra.PathQuery).Node(o.szNodeQueryAttributes(szResultName))
	matchQuery.Match(nodeQuery)

	if utils.Known(o.RoutingPolicyId) {
		q := new(apstra.PathQuery)
		q.Node([]apstra.QEEAttribute{
			{Key: "name", Value: apstra.QEStringVal(szResultName)},
		})
		q.Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.RelationshipTypePolicy.String())},
		})
		q.Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeRoutingPolicy.String())},
			{Key: "id", Value: apstra.QEStringVal(o.RoutingPolicyId.ValueString())},
		})
		matchQuery.Match(q)
	}

	if utils.Known(o.DhcpServers) {
		q := new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				{Key: "name", Value: apstra.QEStringVal(szResultName)},
			}).
			Out([]apstra.QEEAttribute{
				{Key: "type", Value: apstra.QEStringVal(apstra.RelationshipTypePolicy.String())},
			}).
			Node([]apstra.QEEAttribute{
				{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypePolicy.String())},
				{Key: "policy_type", Value: apstra.QEStringVal("dhcp_relay")},
				{Key: "name", Value: apstra.QEStringVal("dhcp_policy")},
				{Key: "dhcp_servers", Value: apstra.QENone(false)},
			})

		for _, dhcpServerVal := range o.DhcpServers.Elements() {
			q.Where(fmt.Sprintf(
				"lambda %s: '%s' in %s.dhcp_servers",
				"dhcp_policy",
				dhcpServerVal.(types.String).ValueString(),
				"dhcp_policy",
			))
		}

		matchQuery.Match(q)
	}

	if utils.Known(o.ImportRouteTargets) || utils.Known(o.ExportRouteTargets) {
		rtPolicyNodeAttrs := []apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeRouteTargetPolicy.String())},
			{Key: "name", Value: apstra.QEStringVal("rt_policy")},
		}

		if utils.Known(o.ImportRouteTargets) {
			rtPolicyNodeAttrs = append(rtPolicyNodeAttrs, apstra.QEEAttribute{
				Key:   "import_RTs",
				Value: apstra.QENone(false),
			})
		}

		if utils.Known(o.ExportRouteTargets) {
			rtPolicyNodeAttrs = append(rtPolicyNodeAttrs, apstra.QEEAttribute{
				Key:   "export_RTs",
				Value: apstra.QENone(false),
			})
		}

		q := new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				{Key: "name", Value: apstra.QEStringVal(szResultName)},
			}).
			Out([]apstra.QEEAttribute{
				{Key: "type", Value: apstra.QEStringVal(apstra.RelationshipTypeRouteTargetPolicy.String())},
			}).
			Node(rtPolicyNodeAttrs)

		for _, importRT := range o.ImportRouteTargets.Elements() {
			q.Where(fmt.Sprintf(
				"lambda %s: '%s' in %s.import_RTs",
				"rt_policy",
				importRT.(types.String).ValueString(),
				"rt_policy",
			))
		}

		for _, exportRT := range o.ExportRouteTargets.Elements() {
			q.Where(fmt.Sprintf(
				"lambda %s: '%s' in %s.export_RTs",
				"rt_policy",
				exportRT.(types.String).ValueString(),
				"rt_policy",
			))
		}

		matchQuery.Match(q)
	}

	return matchQuery
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

	if utils.Known(o.JunosIrbMode) {
		result = append(result, apstra.QEEAttribute{Key: "junos_evpn_irb_mode", Value: apstra.QEStringVal(o.JunosIrbMode.ValueString())})
	}

	return result
}

func (o *DatacenterRoutingZone) setDefaults(_ context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	apiVersion, err := version.NewVersion(client.ApiVersion())
	if err != nil {
		diags.AddError("failed parsing API version", err.Error())
		return
	}
	junosIrbModeMinVersion, _ := version.NewVersion("4.2.0")

	if utils.Known(o.JunosIrbMode) && apiVersion.LessThan(junosIrbModeMinVersion) {
		// junos_irb_mode is set, but Apstra version < 4.2.0
		diags.Append(validatordiag.InvalidAttributeValueDiagnostic(
			path.Root("junos_irb_mode"),
			errApiCompatibility,
			fmt.Sprintf("junos_irb_mode must be set with Apstra API version %s", apiVersion),
		))
	}

	if !utils.Known(o.JunosIrbMode) && apiVersion.GreaterThanOrEqual(junosIrbModeMinVersion) {
		// junos_irb_mode not set, Apstra version >= 4.2.0, so set the default value
		o.JunosIrbMode = types.StringValue(junosIrbModeDefault)
	}

}

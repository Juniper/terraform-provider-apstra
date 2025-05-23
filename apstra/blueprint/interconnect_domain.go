package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"net"

	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InterconnectDomain struct {
	Id                 types.String `tfschema:"id"`
	BlueprintId        types.String `tfsdk:"blueprint_id"`
	Name               types.String `tfsdk:"name"`
	RouteTarget        types.String `tfsdk:"route_target"`
	InterconnectEsiMac types.String `tfsdk:"interconnect_esi_mac"`
}

func (o InterconnectDomain) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI. Required when `id` is omitted.",
			Computed:            true,
			Optional:            true,
		},
		"route_target": dataSourceSchema.StringAttribute{
			MarkdownDescription: "All interconnect gateways MUST use the same Interconnect Route Target (iRT).  The " +
				"iRT is an additional unique RT for the interconnect domain.",
			Computed: true,
		},
		"interconnect_esi_mac": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Each site requires a unique site id iESI at the MAC-VRF level. This can be " +
				"auto-derived or manually set.",
			Computed: true,
		},
	}
}

func (o InterconnectDomain) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Optional:            true,
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI. Required when `id` is omitted.",
			Optional:            true,
		},
		"route_target": dataSourceSchema.StringAttribute{
			MarkdownDescription: "All interconnect gateways MUST use the same Interconnect Route Target (iRT).  The " +
				"iRT is an additional unique RT for the interconnect domain.",
			Optional: true,
		},
		"interconnect_esi_mac": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Each site requires a unique site id iESI at the MAC-VRF level. This can be " +
				"auto-derived or manually set.",
			Optional: true,
		},
	}
}

func (o InterconnectDomain) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(apstraregexp.AlphaCharsRequiredConstraint, apstraregexp.AlphaCharsRequiredConstraintMsg),
			},
		},
		"route_target": resourceSchema.StringAttribute{
			MarkdownDescription: "All interconnect gateways MUST use the same Interconnect Route Target (iRT).  The " +
				"iRT is an additional unique RT for the interconnect domain.",
			Required:   true,
			Validators: []validator.String{apstravalidator.ParseRT()},
		},
		"interconnect_esi_mac": resourceSchema.StringAttribute{
			MarkdownDescription: "Each site requires a unique site id iESI at the MAC-VRF level. This can be " +
				"auto-derived or manually set.",
			Required:   true,
			Validators: []validator.String{apstravalidator.ParseMac()},
		},
	}
}

func (o InterconnectDomain) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.EvpnInterconnectGroupData {
	var esiMac net.HardwareAddr

	if !o.InterconnectEsiMac.IsNull() {
		var err error
		esiMac, err = net.ParseMAC(o.InterconnectEsiMac.String())
		if err != nil {
			diags.AddError("failed to parse interconnect esi mac: "+o.InterconnectEsiMac.String(), err.Error())
		}
	}

	return &apstra.EvpnInterconnectGroupData{
		Label:       o.Name.ValueString(),
		RouteTarget: o.RouteTarget.ValueString(),
		EsiMac:      esiMac,
	}
}

func (o *InterconnectDomain) LoadApiData(_ context.Context, data *apstra.EvpnInterconnectGroupData, _ *diag.Diagnostics) {
	o.InterconnectEsiMac = types.StringValue(data.EsiMac.String())
	o.Name = types.StringValue(data.Label)
	o.RouteTarget = types.StringValue(data.RouteTarget)
}

func (o *InterconnectDomain) Query(resultName string) apstra.QEQuery {
	attributes := []apstra.QEEAttribute{
		apstra.NodeTypeEvpnInterconnectGroup.QEEAttribute(),
		{Key: "name", Value: apstra.QEStringVal("n_evpn_interconnect_group")},
	}

	if !o.Name.IsNull() {
		attributes = append(attributes, apstra.QEEAttribute{Key: "label", Value: apstra.QEStringVal(o.Name.ValueString())})
	}

	if !o.InterconnectEsiMac.IsNull() {
		attributes = append(attributes, apstra.QEEAttribute{Key: "interconnect_esi_mac", Value: apstra.QEStringVal(o.InterconnectEsiMac.ValueString())})
	}

	if !o.RouteTarget.IsNull() {
		attributes = append(attributes, apstra.QEEAttribute{Key: "interconnect_route_target", Value: apstra.QEStringVal(o.RouteTarget.ValueString())})
	}

	return new(apstra.PathQuery).
		Node(attributes)
}

package blueprint

import (
	"context"
	"net"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/hwtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InterconnectDomain struct {
	Id          types.String       `tfsdk:"id"`
	BlueprintId types.String       `tfsdk:"blueprint_id"`
	Name        types.String       `tfsdk:"name"`
	RouteTarget types.String       `tfsdk:"route_target"`
	EsiMac      hwtypes.MACAddress `tfsdk:"esi_mac"`
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
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"route_target": dataSourceSchema.StringAttribute{
			MarkdownDescription: "All interconnect gateways MUST use the same Interconnect Route Target (iRT).  The " +
				"iRT is an additional unique RT for the interconnect domain.",
			Computed: true,
		},
		"esi_mac": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Each site requires a unique site id iESI at the MAC-VRF level. This can be " +
				"auto-derived or manually set.",
			CustomType: hwtypes.MACAddressType{},
			Computed:   true,
		},
	}
}

func (o InterconnectDomain) DataSourceAttributesAsFilter() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI. Required when `id` is omitted.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"route_target": dataSourceSchema.StringAttribute{
			MarkdownDescription: "All interconnect gateways MUST use the same Interconnect Route Target (iRT).  The " +
				"iRT is an additional unique RT for the interconnect domain.",
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"esi_mac": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Each site requires a unique site id iESI at the MAC-VRF level. This can be " +
				"auto-derived or manually set.",
			CustomType: hwtypes.MACAddressType{},
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o InterconnectDomain) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
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
		"esi_mac": resourceSchema.StringAttribute{
			MarkdownDescription: "Each site requires a unique site id iESI at the MAC-VRF level. This can be " +
				"auto-derived or manually set.",
			Optional:   true,
			Computed:   true,
			CustomType: hwtypes.MACAddressType{},
		},
	}
}

func (o InterconnectDomain) Request(_ context.Context, diags *diag.Diagnostics) *apstra.EvpnInterconnectGroupData {
	var esiMac net.HardwareAddr

	if utils.HasValue(o.EsiMac) {
		var err error
		esiMac, err = net.ParseMAC(o.EsiMac.ValueString())
		if err != nil {
			diags.AddError("failed to parse interconnect esi mac: "+o.EsiMac.ValueString(), err.Error())
		}
	}

	return &apstra.EvpnInterconnectGroupData{
		Label:       o.Name.ValueString(),
		RouteTarget: o.RouteTarget.ValueString(),
		EsiMac:      esiMac,
	}
}

func (o *InterconnectDomain) LoadApiData(_ context.Context, data *apstra.EvpnInterconnectGroupData, _ *diag.Diagnostics) {
	o.EsiMac = hwtypes.NewMACAddressValue(data.EsiMac.String())
	o.Name = types.StringValue(data.Label)
	o.RouteTarget = types.StringValue(data.RouteTarget)
}

func (o *InterconnectDomain) Query(resultName string) apstra.QEQuery {
	attributes := []apstra.QEEAttribute{
		apstra.NodeTypeEvpnInterconnectGroup.QEEAttribute(),
		{Key: "name", Value: apstra.QEStringVal(resultName)},
	}

	if !o.Id.IsNull() {
		attributes = append(attributes, apstra.QEEAttribute{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())})
	}

	if !o.Name.IsNull() {
		attributes = append(attributes, apstra.QEEAttribute{Key: "label", Value: apstra.QEStringVal(o.Name.ValueString())})
	}

	if !o.EsiMac.IsNull() {
		mac, _ := net.ParseMAC(o.EsiMac.ValueString()) // ignore error because value is already validated
		attributes = append(attributes, apstra.QEEAttribute{Key: "interconnect_esi_mac", Value: apstra.QEStringVal(mac.String())})
	}

	if !o.RouteTarget.IsNull() {
		attributes = append(attributes, apstra.QEEAttribute{Key: "interconnect_route_target", Value: apstra.QEStringVal(o.RouteTarget.ValueString())})
	}

	return new(apstra.PathQuery).
		Node(attributes)
}

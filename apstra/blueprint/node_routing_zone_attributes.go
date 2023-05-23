package blueprint

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/resources"
	"terraform-provider-apstra/apstra/utils"
)

type NodeTypeRoutingZoneAttributes struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	VlanID           types.Int64  `tfsdk:"vlan_id"`
	Vni              types.Int64  `tfsdk:"vni"`
	VrfID            types.Int64  `tfsdk:"vrf_id"`
	VrfName          types.String `tfsdk:"vrf_name"`
	DhcpServers      types.Set    `tfsdk:"dhcp_servers"`
	RoutingPolicydId types.String `tfsdk:"routing_policy_id"`
}

func (o NodeTypeRoutingZoneAttributes) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.StringType,
		"name":              types.StringType,
		"type":              types.StringType,
		"vlan_id":           types.Int64Type,
		"vni":               types.Int64Type,
		"vrf_id":            types.Int64Type,
		"vrf_name":          types.StringType,
		"dhcp_servers":      types.SetType{ElemType: types.StringType},
		"routing_policy_id": types.StringType,
	}
}

func (o NodeTypeRoutingZoneAttributes) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node `id` field",
			Required:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node `label` field",
			Computed:            true,
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node `sz_type` field",
			Computed:            true,
		},
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Apstra graph datastore node `vlan_id` field",
			Computed:            true,
		},
		"vni": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Apstra graph datastore node `vni_id` field",
			Computed:            true,
		},
		"vrf_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Apstra graph datastore node `vrf_id` field",
			Computed:            true,
		},
		"vrf_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node `vrf_name` field",
			Computed:            true,
		},
		"dhcp_servers": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of DHCP servers used by the Routing Zone",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"routing_policy_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node ID of the Routing Policy applied to this Routing Zone",
			Computed:            true,
		},
	}
}

func (o NodeTypeRoutingZoneAttributes) DataSourceAttributesAsFilter() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node `id` field",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node `label` field",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node `sz_type` field",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Apstra graph datastore node `vlan_id` field",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(design.VlanMin-1, design.VlanMax+1)},
		},
		"vni": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Apstra graph datastore node `vni_id` field",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(resources.VniMin-1, resources.VniMax+1)},
		},
		"vrf_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Apstra graph datastore node `vrf_id` field",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(vrfIdMin-1, vrfIdMax+1)},
		},
		"vrf_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node `vrf_name` field",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"dhcp_servers": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of DHCP servers used by the Routing Zone",
			Optional:            true,
			Validators:          []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
		"routing_policy_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node ID of the Routing Policy applied to this Routing Zone",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o NodeTypeRoutingZoneAttributes) QEEAttributes() []apstra.QEEAttribute {
	var result []apstra.QEEAttribute

	if utils.Known(o.Id) {
		result = append(result, apstra.QEEAttribute{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())})
	}

	if utils.Known(o.Name) {
		result = append(result, apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal(o.Name.ValueString())})
	}

	if utils.Known(o.Type) {
		result = append(result, apstra.QEEAttribute{Key: "sz_type", Value: apstra.QEStringVal(o.Type.ValueString())})
	}

	if utils.Known(o.VlanID) {
		result = append(result, apstra.QEEAttribute{Key: "vlan_id", Value: o.VlanID})
	}

	if utils.Known(o.Vni) {
		result = append(result, apstra.QEEAttribute{Key: "vni_id", Value: o.Vni})
	}

	if utils.Known(o.VrfID) {
		result = append(result, apstra.QEEAttribute{Key: "vrf_id", Value: o.VrfID})
	}

	if utils.Known(o.Name) {
		result = append(result, apstra.QEEAttribute{Key: "vrf_name", Value: apstra.QEStringVal(o.VrfName.ValueString())})
	}

	return result
}

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
	VlanId           types.Int64  `tfsdk:"vlan_id"`
	Vni              types.Int64  `tfsdk:"vni"`
	DhcpServers      types.Set    `tfsdk:"dhcp_servers"`
	RoutingPolicydId types.String `tfsdk:"routing_policy_id"`
}

func (o NodeTypeRoutingZoneAttributes) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.StringType,
		"name":              types.StringType,
		"vlan_id":           types.Int64Type,
		"vni":               types.Int64Type,
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
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Apstra graph datastore node `vlan_id` field",
			Computed:            true,
		},
		"vni": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Apstra graph datastore node `vni_id` field",
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

	if utils.Known(o.VlanId) {
		result = append(result, apstra.QEEAttribute{Key: "vlan_id", Value: o.VlanId})
	}

	if utils.Known(o.Vni) {
		result = append(result, apstra.QEEAttribute{Key: "vni_id", Value: o.Vni})
	}

	return result
}

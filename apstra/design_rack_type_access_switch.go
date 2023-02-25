package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateAccessSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	as := rt.Data.AccessSwitches[i]
	if as.RedundancyProtocol == goapstra.AccessRedundancyProtocolEsi && as.EsiLagInfo == nil {
		diags.AddError("access switch ESI LAG Info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, as.Label, as.RedundancyProtocol.String()))
		return
	}
	if as.LogicalDevice == nil {
		diags.AddError("access switch logical device info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' logical device is nil",
				rt.Id, as.Label))
		return
	}
}

type accessSwitch struct {
	LogicalDevice      types.Object `tfsdk:"logical_device"`
	EsiLagInfo         types.Object `tfsdk:"esi_lag_info"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	Count              types.Int64  `tfsdk:"count"`
	Links              types.Set    `tfsdk:"links"`
	Tags               types.Set    `tfsdk:"tags"`
}

func (o accessSwitch) dataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"count": schema.Int64Attribute{
			MarkdownDescription: "Count of Access Switches of this type.",
			Computed:            true,
		},
		"redundancy_protocol": schema.StringAttribute{
			MarkdownDescription: "Indicates whether 'the switch' is actually a LAG-capable redundant pair and if so, what type.",
			Computed:            true,
		},
		"esi_lag_info": schema.SingleNestedAttribute{
			MarkdownDescription: "Interconnect information for Access Switches in ESI-LAG redundancy mode.",
			Computed:            true,
			Attributes:          esiLagInfo{}.schemaAsDataSource(),
		},
		"logical_device": schema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          logicalDevice{}.dataSourceAttributesNested(),
		},
		"tags": schema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Access Switch.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tag{}.dataSourceAttributesNested(),
			},
		},
		"links": schema.SetNestedAttribute{
			MarkdownDescription: "Details links from this Access Switch to upstream switches within this Rack Type.",
			Computed:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: schema.NestedAttributeObject{
				Attributes: rackLink{}.dataSourceAttributes(),
			},
		},
	}
}

func (o accessSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device":      types.ObjectType{AttrTypes: logicalDevice{}.attrTypes()},
		"esi_lag_info":        types.ObjectType{AttrTypes: esiLagInfo{}.attrTypes()},
		"redundancy_protocol": types.StringType,
		"count":               types.Int64Type,
		"links":               types.SetType{ElemType: types.ObjectType{AttrTypes: rackLink{}.attrTypes()}},
		"tags":                types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *accessSwitch) loadApiResponse(ctx context.Context, in *goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) {
	o.LogicalDevice = newLogicalDeviceObject(ctx, in.LogicalDevice, diags)
	o.EsiLagInfo = newEsiLagInfo(ctx, in.EsiLagInfo, diags)

	if in.RedundancyProtocol == goapstra.AccessRedundancyProtocolNone {
		o.RedundancyProtocol = types.StringNull()
	} else {
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	}

	o.Count = types.Int64Value(int64(in.InstanceCount))
	o.Links = newLinkSet(ctx, in.Links, diags)
	o.Tags = newTagSet(ctx, in.Tags, diags)
}

package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	LogicalDeviceID    types.String `tfsdk:"logical_device_id"`
	LogicalDevice      types.Object `tfsdk:"logical_device"`
	EsiLagInfo         types.Object `tfsdk:"esi_lag_info"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	Count              types.Int64  `tfsdk:"count"`
	Links              types.Set    `tfsdk:"links"`
	TagIds             types.Set    `tfsdk:"tag_ids"`
	Tags               types.Set    `tfsdk:"tags"`
}

func (o accessSwitch) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"logical_device_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in data source contexts.",
			Computed:            true,
		},
		"logical_device": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          logicalDevice{}.dataSourceAttributesNested(),
		},
		"esi_lag_info": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Interconnect information for Access Switches in ESI-LAG redundancy mode.",
			Computed:            true,
			Attributes:          esiLagInfo{}.dataSourceAttributes(),
		},
		"redundancy_protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Indicates whether 'the switch' is actually a LAG-capable redundant pair and if so, what type.",
			Computed:            true,
		},
		"count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Count of Access Switches of this type.",
			Computed:            true,
		},
		"links": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Details links from this Access Switch to upstream switches within this Rack Type.",
			Computed:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: rackLink{}.dataSourceAttributes(),
			},
		},
		"tag_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs will always be `<null>` in data source contexts.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tags": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Access Switch.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: tag{}.dataSourceAttributesNested(),
			},
		},
	}
}

func (o accessSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device_id":   types.StringType,
		"logical_device":      types.ObjectType{AttrTypes: logicalDevice{}.attrTypes()},
		"esi_lag_info":        types.ObjectType{AttrTypes: esiLagInfo{}.attrTypes()},
		"redundancy_protocol": types.StringType,
		"count":               types.Int64Type,
		"links":               types.SetType{ElemType: types.ObjectType{AttrTypes: rackLink{}.attrTypes()}},
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tags":                types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *accessSwitch) loadApiData(ctx context.Context, in *goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) {
	o.LogicalDeviceID = types.StringNull()
	o.LogicalDevice = newLogicalDeviceObject(ctx, in.LogicalDevice, diags)
	o.EsiLagInfo = newEsiLagInfo(ctx, in.EsiLagInfo, diags)

	if in.RedundancyProtocol == goapstra.AccessRedundancyProtocolNone {
		o.RedundancyProtocol = types.StringNull()
	} else {
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	}

	o.Count = types.Int64Value(int64(in.InstanceCount))
	o.Links = newLinkSet(ctx, in.Links, diags)
	o.TagIds = types.SetNull(types.StringType)
	o.Tags = newTagSet(ctx, in.Tags, diags)
}

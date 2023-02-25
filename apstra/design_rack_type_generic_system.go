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

func validateGenericSystem(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	gs := rt.Data.GenericSystems[i]
	if gs.LogicalDevice == nil {
		diags.AddError("generic system logical device info missing",
			fmt.Sprintf("rack type '%s', generic system '%s' logical device is nil",
				rt.Id, gs.Label))
		return
	}
}

type genericSystem struct {
	LogicalDevice    types.Object `tfsdk:"logical_device"`
	PortChannelIdMin types.Int64  `tfsdk:"port_channel_id_min"`
	PortChannelIdMax types.Int64  `tfsdk:"port_channel_id_max"`
	Count            types.Int64  `tfsdk:"count"`
	Links            types.Set    `tfsdk:"links"`
	Tags             types.Set    `tfsdk:"tags"`
}

func (o genericSystem) dataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"count": schema.Int64Attribute{
			MarkdownDescription: "Number of Generic Systems of this type.",
			Computed:            true,
		},
		"port_channel_id_min": schema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Computed:            true,
		},
		"port_channel_id_max": schema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Computed:            true,
		},
		"logical_device": schema.SingleNestedAttribute{
			MarkdownDescription: "Logical Device attributes as represented in the Global Catalog.",
			Computed:            true,
			Attributes:          logicalDevice{}.dataSourceAttributesNested(),
		},
		"tags": schema.SetNestedAttribute{
			MarkdownDescription: "Details any tags applied to this Generic System.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: tag{}.dataSourceAttributesNested(),
			},
		},
		"links": schema.SetNestedAttribute{
			MarkdownDescription: "Details links from this Generic System to upstream switches within this Rack Type.",
			Computed:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: schema.NestedAttributeObject{
				Attributes: dRackLink{}.attributes(),
			},
		},
	}
}

func (o genericSystem) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_device":      types.ObjectType{AttrTypes: logicalDevice{}.attrTypes()},
		"port_channel_id_min": types.Int64Type,
		"port_channel_id_max": types.Int64Type,
		"count":               types.Int64Type,
		"links":               types.SetType{ElemType: types.ObjectType{AttrTypes: dRackLink{}.attrTypes()}},
		"tags":                types.SetType{ElemType: types.ObjectType{AttrTypes: tag{}.attrTypes()}},
	}
}

func (o *genericSystem) loadApiResponse(ctx context.Context, in *goapstra.RackElementGenericSystem, diags *diag.Diagnostics) {
	o.LogicalDevice = newLogicalDeviceObject(ctx, in.LogicalDevice, diags)
	o.PortChannelIdMin = types.Int64Value(int64(in.PortChannelIdMin))
	o.PortChannelIdMax = types.Int64Value(int64(in.PortChannelIdMax))
	o.Count = types.Int64Value(int64(in.Count))
	o.Links = newDataSourceLinkSet(ctx, in.Links, diags)
	o.Tags = newTagSet(ctx, in.Tags, diags)
}

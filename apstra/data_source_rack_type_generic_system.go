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
	}
}

type dRackTypeGenericSystem struct {
	Name             types.String `tfsdk:"name"`
	Count            types.Int64  `tfsdk:"count"`
	PortChannelIdMin types.Int64  `tfsdk:"port_channel_id_min"`
	PortChannelIdMax types.Int64  `tfsdk:"port_channel_id_max"`
	LogicalDevice    types.Object `tfsdk:"logical_device"`
	TagData          types.Set    `tfsdk:"tag_data"`
	Links            types.Set    `tfsdk:"links"`
}

func (o dRackTypeGenericSystem) schema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Generic name, must be unique within the rack-type.",
			Computed:            true,
		},
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
		"logical_device": logicalDeviceData{}.schemaAsDataSource(),
		"tag_data": schema.SetNestedAttribute{
			NestedObject:        tagData{}.schema(),
			MarkdownDescription: "Details any tags applied to this Generic System.",
			Computed:            true,
		},
		"links": schema.SetNestedAttribute{
			MarkdownDescription: "Details links from this Generic System to upstream switches within this Rack Type.",
			Computed:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject:        dRackLink{}.schema(),
		},
	}
}

func (o dRackTypeGenericSystem) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"count":               types.Int64Type,
		"port_channel_id_min": types.Int64Type,
		"port_channel_id_max": types.Int64Type,
		"logical_device":      logicalDeviceData{}.attrType(),
		"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
		"links":               types.SetType{ElemType: dRackLink{}.attrType()},
	}
}

func (o dRackTypeGenericSystem) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *dRackTypeGenericSystem) loadApiResponse(ctx context.Context, in *goapstra.RackElementGenericSystem, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Count = types.Int64Value(int64(in.Count))
	o.PortChannelIdMin = types.Int64Value(int64(in.PortChannelIdMin))
	o.PortChannelIdMax = types.Int64Value(int64(in.PortChannelIdMax))

	o.LogicalDevice = newLogicalDeviceObject(ctx, in.LogicalDevice, diags)
	if diags.HasError() {
		return
	}

	o.Links = newDataSourceLinkSet(ctx, in.Links, diags)
	if diags.HasError() {
		return
	}

	o.TagData = newTagSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}
}

package design

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TemplateCollapsed struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	RackType      types.Object `tfsdk:"rack_type"`
	RackTypeId    types.String `tfsdk:"rack_type_id"`
	MeshLinkCount types.Int64  `tfsdk:"mesh_link_count"`
	MeshLinkSpeed types.String `tfsdk:"mesh_link_speed"`
}

func (o TemplateCollapsed) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":              types.StringType,
		"name":            types.StringType,
		"rack_type":       types.ObjectType{AttrTypes: RackType{}.AttrTypes()},
		"rack_type_id":    types.StringType,
		"mesh_link_count": types.Int64Type,
		"mesh_link_speed": types.StringType,
	}
}

func (o TemplateCollapsed) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Template ID. Required when `id` is omitted.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Web UI name of the Template. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"rack_type": dataSourceSchema.ObjectAttribute{
			MarkdownDescription: "rack_type details",
			Computed:            true,
		},
		"rack_type_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "rack_type_id details ",
			Computed:            true,
		},
		"mesh_link_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "mesh_link_count integer ",
			Computed:            true,
		},
		"mesh_link_speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "mesh_link_speed details ",
			Computed:            true,
		},
	}
}

func (o TemplateCollapsed) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Collapsed Template.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra name of the Collapsed Template.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"rack_type": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "rack type layer details",
			Computed:            true,
			Attributes:          RackType{}.ResourceAttributesNested(),
		},
		"rack_type_id": resourceSchema.StringAttribute{
			MarkdownDescription: "rack type id ",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"mesh_link_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "mesh_link_count integer ",
			Required:            true,
		},
		"mesh_link_speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "mesh_link_speed details ",
			Required:            true,
		},
	}
}

func (o *TemplateCollapsed) Request(_ context.Context, _ *diag.Diagnostics) *apstra.CreateL3CollapsedTemplateRequest {

	return &apstra.CreateL3CollapsedTemplateRequest{
		DisplayName:          o.Name.ValueString(),
		RackTypeIds:          []apstra.ObjectId{apstra.ObjectId(o.RackTypeId.ValueString())},
		RackTypeCounts:       []apstra.RackTypeCount{{RackTypeId: apstra.ObjectId(o.RackTypeId.ValueString()), Count: 1}},
		MeshLinkCount:        int(o.MeshLinkCount.ValueInt64()),
		MeshLinkSpeed:        apstra.LogicalDevicePortSpeed(o.MeshLinkSpeed.ValueString()),
		DhcpServiceIntent:    apstra.DhcpServiceIntent{Active: true},
		VirtualNetworkPolicy: apstra.VirtualNetworkPolicy{OverlayControlProtocol: apstra.OverlayControlProtocolEvpn},
	}
}

func (o *TemplateCollapsed) LoadApiData(ctx context.Context, in *apstra.TemplateL3CollapsedData, diags *diag.Diagnostics) {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to load TemplateCollapsed from nil source")
		return
	}
	if len(in.RackTypes) != 1 {
		diags.AddError("cannot load the RackType", "API response load RackTypes was not 1 element")
		return
	}
	if in.RackTypes[0].Data == nil {
		diags.AddError("cannot load the RackType", "API response contains NIL RackType Data")
		return
	}
	o.Name = types.StringValue(in.DisplayName)
	o.RackType = NewRackTypeObject(ctx, in.RackTypes[0].Data, diags)
	o.MeshLinkCount = types.Int64Value(int64(in.MeshLinkCount))
	o.MeshLinkSpeed = types.StringValue(string(in.MeshLinkSpeed))
}

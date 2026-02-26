package freeform

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
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
)

type AggregateLink struct {
	ID             types.String `tfsdk:"id"`
	BlueprintID    types.String `tfsdk:"blueprint_id"`
	Name           types.String `tfsdk:"name"`
	EndpointGroup1 types.Object `tfsdk:"endpoint_group_1"`
	EndpointGroup2 types.Object `tfsdk:"endpoint_group_2"`
	MemberLinkIDs  types.Set    `tfsdk:"member_link_ids"`
	Tags           types.Set    `tfsdk:"tags"`
}

func (o AggregateLink) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Aggregate Link by ID. Required when `name` is omitted.",
			Optional:            true,
			Computed:            true,
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
			MarkdownDescription: "Populate this field to look up the Freeform Aggregate Link by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"endpoint_group_1": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "An Endpoint Group represents the system(s) on one end of the Aggregate Link.",
			Attributes:          AggregateLinkEndpointGroup{}.dataSourceAttributes(),
			Computed:            true,
		},
		"endpoint_group_2": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "An Endpoint Group represents the system(s) on one end of the Aggregate Link.",
			Attributes:          AggregateLinkEndpointGroup{}.dataSourceAttributes(),
			Computed:            true,
		},
		"member_link_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Link IDs used in the aggregation.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of tags associated with this Aggregate Link.",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

func (o AggregateLink) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Aggregate Link by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Name of the Aggregate Link.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"endpoint_group_1": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "An Endpoint Group represents the system(s) on one end of the Aggregate Link.",
			Attributes:          AggregateLinkEndpointGroup{}.resourceAttributes(),
			Required:            true,
		},
		"endpoint_group_2": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "An Endpoint Group represents the system(s) on one end of the Aggregate Link.",
			Attributes:          AggregateLinkEndpointGroup{}.resourceAttributes(),
			Required:            true,
		},
		"member_link_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Link IDs to be used in the aggregation",
			ElementType:         types.StringType,
			Required:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of unique case-insensitive tag labels",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *AggregateLink) Request(ctx context.Context, diags *diag.Diagnostics) apstra.FreeformAggregateLink {
	var result apstra.FreeformAggregateLink

	if !o.Name.IsUnknown() || !o.Name.IsNull() { // leave nil when no value set
		result.Label = o.Name.ValueStringPointer()
	}

	diags.Append(o.MemberLinkIDs.ElementsAs(ctx, &result.MemberLinkIds, false)...)

	var endpointGroup1, endpointGroup2 AggregateLinkEndpointGroup
	diags.Append(o.EndpointGroup1.As(ctx, &endpointGroup1, basetypes.ObjectAsOptions{})...)
	diags.Append(o.EndpointGroup2.As(ctx, &endpointGroup2, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return result
	}

	result.EndpointGroups[0] = endpointGroup1.Request(ctx, path.Root("endpoint_group_1"), diags)
	result.EndpointGroups[1] = endpointGroup2.Request(ctx, path.Root("endpoint_group_2"), diags)
	if diags.HasError() {
		return result
	}

	diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)

	if !o.ID.IsNull() && !o.ID.IsUnknown() {
		result.SetID(o.ID.ValueString())
	}

	return result
}

func (o *AggregateLink) LoadApiData(ctx context.Context, in apstra.FreeformAggregateLink, diags *diag.Diagnostics) {
	o.ID = types.StringPointerValue(in.ID())
	o.Name = types.StringPointerValue(in.Label)

	var endpointGroup1, endpointGroup2 AggregateLinkEndpointGroup
	endpointGroup1.LoadAPIData(ctx, in.EndpointGroups[0], diags)
	endpointGroup2.LoadAPIData(ctx, in.EndpointGroups[1], diags)
	if diags.HasError() {
		return
	}

	var d diag.Diagnostics
	o.EndpointGroup1, d = types.ObjectValueFrom(ctx, endpointGroup1.attrTypes(), &endpointGroup1)
	diags.Append(d...)
	o.EndpointGroup2, d = types.ObjectValueFrom(ctx, endpointGroup2.attrTypes(), &endpointGroup2)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	o.MemberLinkIDs = value.SetOrNull(ctx, types.StringType, in.MemberLinkIds, diags)
	o.Tags = value.SetOrNull(ctx, types.StringType, in.Tags, diags)
}

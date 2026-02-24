package freeform

import (
	"context"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/Juniper/apstra-go-sdk/apstra"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AggregateLink struct {
	BlueprintID    types.String `tfsdk:"blueprint_id"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	EndpointGroups types.List   `tfsdk:"endpoint_groups"`
	MemberLinkIDs  types.Set    `tfsdk:"member_link_ids"`
	Tags           types.Set    `tfsdk:"tags"`
}

func (o AggregateLink) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
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
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Aggregate Link by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"endpoint_groups": dataSourceSchema.ListAttribute{
			MarkdownDescription: "A list of endpoint objects, must be exactly two items. Each represents the " +
				"system(s) on one end of the Aggregate Link.",
			ElementType: types.ObjectType{},
			Computed:    true,
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
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Aggregate Link by ID. Required when `name` is omitted.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name of the Aggregate Link.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"endpoint_groups": resourceSchema.ListAttribute{
			MarkdownDescription: "A list of endpoint objects, must be exactly two items. Each group represents the " +
				"system(s) on one end of the Aggregate Link.",
			Required:    true,
			ElementType: types.ObjectType{},
			Validators:  []validator.List{listvalidator.SizeBetween(2, 2)},
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
	result := apstra.FreeformAggregateLink{
		Label: o.Name.ValueString(),
		// MemberLinkIds:  nil,                                            // see below
		// EndpointGroups: [2]apstra.FreeformAggregateLinkEndpointGroup{}, // see below
		// Tags:           nil,                                            // see below
	}

	diags.Append(o.MemberLinkIDs.ElementsAs(ctx, &result.MemberLinkIds, false)...)
	diags.Append(o.EndpointGroups.ElementsAs(ctx, &result.EndpointGroups, false)...)
	diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)

	return result
}

func (o *AggregateLink) LoadApiData(ctx context.Context, in apstra.FreeformAggregateLink, diags *diag.Diagnostics) {
	o.ID = types.StringPointerValue(in.ID())
	o.Name = types.StringValue(in.Label)
	endpointGroups := make([]AggregateLinkEndpointGroup, 2)
	endpointGroups[0].LoadAPIData(ctx, in.EndpointGroups[0], diags)
	endpointGroups[1].LoadAPIData(ctx, in.EndpointGroups[1], diags)
	if diags.HasError() {
		return
	}
	o.EndpointGroups = value.ListOrNull(ctx, types.ObjectType{}, endpointGroups, diags)
	o.MemberLinkIDs = value.SetOrNull(ctx, types.StringType, in.MemberLinkIds, diags)
	o.Tags = value.SetOrNull(ctx, types.StringType, in.Tags, diags)
}

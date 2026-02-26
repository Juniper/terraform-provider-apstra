package freeform

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AggregateLinkEndpointGroup struct {
	Endpoints types.List   `tfsdk:"endpoints"`
	Label     types.String `tfsdk:"label"`
	Tags      types.Set    `tfsdk:"tags"`

	ID types.String `tfsdk:"id"`
}

func (o AggregateLinkEndpointGroup) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"endpoints": types.ListType{ElemType: types.ObjectType{AttrTypes: AggregateLinkEndpoint{}.attrTypes()}},
		"label":     types.StringType,
		"tags":      types.SetType{ElemType: types.StringType},
		"id":        types.StringType,
	}
}

func (o AggregateLinkEndpointGroup) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"endpoints": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "List of Aggregate Link Endpoints associated with this Aggregate Link Endpoint Group.",
			NestedObject:        dataSourceSchema.NestedAttributeObject{Attributes: AggregateLinkEndpoint{}.dataSourceAttributes()},
			Computed:            true,
		},
		"label": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Endpoint Group Label as displayed in the web UI.",
			Computed:            true,
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Tags associated with this Aggregate Link Endpoint Group.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Aggregate Link Endpoint Group.",
			Computed:            true,
		},
	}
}

func (o AggregateLinkEndpointGroup) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"endpoints": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "List of Aggregate Link Endpoints associated with this Aggregate Link Endpoint Group.",
			NestedObject:        resourceSchema.NestedAttributeObject{Attributes: AggregateLinkEndpoint{}.resourceAttributes()},
			Required:            true,
			Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
		},
		"label": resourceSchema.StringAttribute{
			MarkdownDescription: "Endpoint Group Label as displayed in the web UI.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of tags associated with this Aggregate Link Endpoint Group.",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Aggregate Link Endpoint Group.",
			Computed:            true,
		},
	}
}

func (o AggregateLinkEndpointGroup) Request(ctx context.Context, path path.Path, diags *diag.Diagnostics) apstra.FreeformAggregateLinkEndpointGroup {
	var result apstra.FreeformAggregateLinkEndpointGroup

	if !o.Label.IsNull() || !o.Label.IsUnknown() { // leave nil when no value set
		result.Label = o.Label.ValueStringPointer()
	}

	diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)

	var endpoints []AggregateLinkEndpoint
	diags.Append(o.Endpoints.ElementsAs(ctx, &endpoints, false)...)
	if diags.HasError() {
		return result
	}

	result.Endpoints = make([]apstra.FreeformAggregateLinkEndpoint, len(endpoints))
	for i, endpoint := range endpoints {
		result.Endpoints[i] = endpoint.Request(ctx, path.AtName("endpoints").AtSetValue(o.Endpoints.Elements()[i]), diags)
	}
	if diags.HasError() {
		return result
	}

	return result
}

func (o *AggregateLinkEndpointGroup) LoadAPIData(ctx context.Context, in apstra.FreeformAggregateLinkEndpointGroup, diags *diag.Diagnostics) {
	endpoints := make([]AggregateLinkEndpoint, len(in.Endpoints))
	for i, endpoint := range in.Endpoints {
		endpoints[i].LoadAPIData(ctx, endpoint, diags)
	}
	o.Endpoints = value.ListOrNull(ctx, types.ObjectType{AttrTypes: AggregateLinkEndpoint{}.attrTypes()}, endpoints, diags)
	o.Label = types.StringPointerValue(in.Label)
	o.Tags = value.SetOrNull(ctx, types.StringType, in.Tags, diags)
	o.ID = types.StringPointerValue(in.ID())
}

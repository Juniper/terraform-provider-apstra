package blueprint

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"regexp"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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

type FreeformLink struct {
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	Id              types.String `tfsdk:"id"`
	Speed           types.String `tfsdk:"speed"`
	Type            types.String `tfsdk:"type"`
	Name            types.String `tfsdk:"name"`
	AggregateLinkId types.String `tfsdk:"aggregate_link_id"`
	Endpoints       types.Set    `tfsdk:"endpoints"`
	InterfaceIds    types.Set    `tfsdk:"interface_ids"`
	Tags            types.Set    `tfsdk:"tags"`
}

func (o FreeformLink) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint where the Link lives.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up the Freeform Link by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Populate this field to look up the Link by Name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"speed": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Speed of the Link",
			Computed:            true,
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Link type",
			Computed:            true,
		},
		"aggregate_link_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "aggregate link  ID",
			Computed:            true,
		},
		"endpoints": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Endpoints assigned to the Link",
			Computed:            true,
		},
		"interface_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Interface IDs associated with the link",
			Computed:            true,
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

func (o FreeformLink) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Freeform Link.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Freeform Link name as shown in the Web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9.-_]+$"), "name may consist only of the following characters : a-zA-Z0-9.-_")},
		},
		"speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of the Freeform Link.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Deploy mode of the Link",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"aggregate_link_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Aggregate ID of the Link",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"endpoints": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: freeformEndpoint{}.ResourceAttributes(),
			},
			PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
			MarkdownDescription: "Endpoints of the  Link",
			Required:            true,
			Validators:          []validator.Set{setvalidator.SizeBetween(2, 2)},
		},
		"interface_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Interface IDs associated with the link",
			Computed:            true,
			ElementType:         types.StringType,
			Validators:          []validator.Set{setvalidator.SizeBetween(2, 2)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Optional:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *FreeformLink) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.FreeformLinkRequest {
	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		return nil
	}

	var endpoints []freeformEndpoint
	diags.Append(o.Endpoints.ElementsAs(ctx, &endpoints, false)...)
	if diags.HasError() {
		return nil
	}

	var epArray [2]apstra.FreeformEndpoint
	for i, endpoint := range endpoints {
		epArray[i] = *endpoint.request()
	}

	return &apstra.FreeformLinkRequest{
		Label:     o.Name.ValueString(),
		Tags:      tags,
		Endpoints: epArray,
	}
}

func (o *FreeformLink) GetInterfaceIds(ctx context.Context, bp *apstra.FreeformClient, diags *diag.Diagnostics) {
	var endpoints []freeformEndpoint
	diags.Append(o.Endpoints.ElementsAs(ctx, &endpoints, false)...)
	if diags.HasError() {
		return
	}

	query := new(apstra.PathQuery).
		SetClient(bp.Client()).
		SetBlueprintId(bp.Id()).
		//node('system', id='v7A_2gUtP9vHq2DYhug')
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(endpoints[0].SystemId.ValueString())},
		}).
		//.out('hosted_interfaces')
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		//.node('interface', name='n_interface_0')
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_interface_0")},
		}).
		//.out('link')
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		//.node('link', id='P6oXpH9ho-_m41LLcSY')
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeLink.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
		}).
		//.in_('link')
		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		//.node('interface', name='n_interface_1')
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_interface_1")},
			//.in_('hosted_interfaces')
		}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		//.node('system', id='uPeP0h4d-q8OAIomCpY')
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(endpoints[1].SystemId.ValueString())},
		})

	var response struct {
		Items []struct {
			Interface0 struct {
				Id string `json:"id"`
			} `json:"n_interface_0"`
			Interface1 struct {
				Id string `json:"id"`
			} `json:"n_interface_1"`
		} `json:"items"`
	}

	err := query.Do(ctx, &response)
	if err != nil {
		diags.AddError("unable to perform query.do()", "query.do()")
		return
	}
	if len(response.Items) != 1 {
		diags.AddError("the Query response is incorrect", "the query is not 1 response ")
		return
	}

	interfaceIds := make([]attr.Value, 2)
	interfaceIds[0] = types.StringValue(response.Items[0].Interface0.Id)
	interfaceIds[1] = types.StringValue(response.Items[0].Interface1.Id)

	var d diag.Diagnostics
	o.InterfaceIds, d = types.SetValue(types.StringType, interfaceIds)
	diags.Append(d...)
}

func (o *FreeformLink) LoadApiData(ctx context.Context, in *apstra.FreeformLinkData, diags *diag.Diagnostics) {
	o.Speed = types.StringValue(string(in.Speed))
	o.Type = types.StringValue(in.Type.String())
	o.Name = types.StringValue(in.Label)
	o.Endpoints = newFreeformEndpointSet(ctx, in.Endpoints, diags) // safe to ignore diagnostic here
	o.AggregateLinkId = types.StringValue(in.AggregateLinkId.String())
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags) // safe to ignore diagnostic here
}

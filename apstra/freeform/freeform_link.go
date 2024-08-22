package freeform

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Link struct {
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	Id              types.String `tfsdk:"id"`
	Speed           types.String `tfsdk:"speed"`
	Type            types.String `tfsdk:"type"`
	Name            types.String `tfsdk:"name"`
	AggregateLinkId types.String `tfsdk:"aggregate_link_id"`
	Endpoints       types.Map    `tfsdk:"endpoints"`
	Tags            types.Set    `tfsdk:"tags"`
}

func (o Link) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
			MarkdownDescription: "Speed of the Link " +
				"200G | 5G | 1G | 100G | 150g | 40g | 2500M | 25G | 25g | 10G | 50G | 800G " +
				"| 10M | 100m | 2500m | 50g | 400g | 400G | 200g | 5g | 800g | 100M | 10g " +
				"| 150G | 10m | 100g | 1g | 40G",
			Computed: true,
		},
		"type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "aggregate_link | ethernet\n" +
				"Link Type. An 'ethernet' link is a normal front-panel interface. " +
				"An 'aggregate_link' is a bonded interface which is typically used for LACP or Static LAGs. " +
				"Note that the lag_mode parameter is a property of the interface and not the link, " +
				"since interfaces may have different lag modes on opposite sides of the link - " +
				"eg lacp_passive <-> lacp_active",
			Computed: true,
		},
		"aggregate_link_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of aggregate link node that the current link belongs to",
			Computed:            true,
		},
		"endpoints": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Endpoints of the  Link, a Map keyed by System ID.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: FreeformEndpoint{}.DatasourceAttributes(),
			},
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of unique case-insensitive tag labels",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

func (o Link) ResourceAttributes() map[string]resourceSchema.Attribute {
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
				stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9.-_]+$"), "name may consist only of the following characters : a-zA-Z0-9.-_"),
			},
		},
		"speed": resourceSchema.StringAttribute{
			MarkdownDescription: "Speed of the Freeform Link.",
			Computed:            true,
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Deploy mode of the Link",
			Computed:            true,
		},
		"aggregate_link_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of aggregate link node that the current link belongs to",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"endpoints": resourceSchema.MapNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: FreeformEndpoint{}.ResourceAttributes(),
			},
			PlanModifiers:       []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			MarkdownDescription: "Endpoints of the  Link, a Map keyed by System ID.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeBetween(2, 2)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *Link) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.FreeformLinkRequest {
	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		return nil
	}

	var endpoints map[string]FreeformEndpoint
	diags.Append(o.Endpoints.ElementsAs(ctx, &endpoints, false)...)
	if diags.HasError() {
		return nil
	}

	var epArray [2]apstra.FreeformEndpoint
	var i int
	for systemId, endpoint := range endpoints {
		epArray[i] = *endpoint.request(ctx, systemId, diags)
		i++
	}

	return &apstra.FreeformLinkRequest{
		Label:     o.Name.ValueString(),
		Tags:      tags,
		Endpoints: epArray,
	}
}

func (o *Link) LoadApiData(ctx context.Context, in *apstra.FreeformLinkData, diags *diag.Diagnostics) {
	interfaceIds := make([]string, len(in.Endpoints))
	for i, endpoint := range in.Endpoints {
		if endpoint.Interface.Id == nil {
			diags.AddError(
				fmt.Sprintf("api returned null interface id for system %s", endpoint.SystemId),
				"link endpoints should always have an interface id.",
			)
			return
		}
		interfaceIds[i] = endpoint.Interface.Id.String()
	}

	o.Speed = types.StringValue(string(in.Speed))
	o.Type = types.StringValue(in.Type.String())
	o.Name = types.StringValue(in.Label)
	o.Endpoints = newFreeformEndpointMap(ctx, in.Endpoints, diags) // safe to ignore diagnostic here
	o.AggregateLinkId = types.StringPointerValue((*string)(in.AggregateLinkId))
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags) // safe to ignore diagnostic here
}

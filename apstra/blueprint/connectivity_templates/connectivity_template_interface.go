package connectivitytemplates

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint/connectivity_templates/primitives"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectivityTemplateInterface struct {
	Id                      types.String `tfsdk:"id"`
	BlueprintId             types.String `tfsdk:"blueprint_id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	Tags                    types.Set    `tfsdk:"tags"`
	IpLinks                 types.Map    `tfsdk:"ip_links"`
	RoutingZoneConstraints  types.Map    `tfsdk:"routing_zone_constraints"`
	VirtualNetworkMultiples types.Map    `tfsdk:"virtual_network_multiples"`
	VirtualNetworkSingles   types.Map    `tfsdk:"virtual_network_singles"`
}

func (o ConnectivityTemplateInterface) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Connectivity Template Name displayed in the web UI",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Connectivity Template Description displayed in the web UI",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags associated with this Connectivity Template",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"ip_links": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of *IP Link* Primitives in this Connectivity Template",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: primitives.IpLink{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
		"routing_zone_constraints": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of *Routing Zone Constraint* Primitives in this Connectivity Template",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: primitives.RoutingZoneConstraint{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
		"virtual_network_multiples": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of *Virtual Network (Multiple)* Primitives in this Connectivity Template",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: primitives.VirtualNetworkMultiple{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
		"virtual_network_singles": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of *Virtual Network (Single)* Primitives in this Connectivity Template",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: primitives.VirtualNetworkSingle{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
	}
}

func (o ConnectivityTemplateInterface) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplate {
	result := apstra.ConnectivityTemplate{
		Label:       o.Name.ValueString(),
		Description: o.Description.ValueString(),
		// Tags:        // set below
		// Subpolicies: // set below
	}

	// Set tags
	diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)

	// Set subpolicies
	result.Subpolicies = append(result.Subpolicies, primitives.IpLinkSubpolicies(ctx, o.IpLinks, diags)...)
	result.Subpolicies = append(result.Subpolicies, primitives.RoutingZoneConstraintSubpolicies(ctx, o.RoutingZoneConstraints, diags)...)
	result.Subpolicies = append(result.Subpolicies, primitives.VirtualNetworkMultipleSubpolicies(ctx, o.VirtualNetworkMultiples, diags)...)
	result.Subpolicies = append(result.Subpolicies, primitives.VirtualNetworkSingleSubpolicies(ctx, o.VirtualNetworkSingles, diags)...)

	// try to set the root batch policy ID from o.Id
	if !o.Id.IsUnknown() {
		result.Id = utils.ToPtr(apstra.ObjectId(o.Id.ValueString()))
	}

	// set remaining policy IDs
	err := result.SetIds()
	if err != nil {
		diags.AddError("Failed while generating Connectivity Template IDs", err.Error())
		return nil
	}

	// set user data
	err = result.SetUserData()
	if err != nil {
		diags.AddError("Failed while generating Connectivity Template User Data", err.Error())
		return nil
	}

	return &result
}

func (o *ConnectivityTemplateInterface) LoadApiData(ctx context.Context, in *apstra.ConnectivityTemplate, diags *diag.Diagnostics) {
	o.Id = types.StringPointerValue((*string)(in.Id))
	o.Name = types.StringValue(in.Label)
	o.Description = utils.StringValueOrNull(ctx, in.Description, diags)
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags)
	o.IpLinks = primitives.IpLinkPrimitivesFromSubpolicies(ctx, in.Subpolicies, diags)
	o.RoutingZoneConstraints = primitives.RoutingZoneConstraintPrimitivesFromSubpolicies(ctx, in.Subpolicies, diags)
	o.VirtualNetworkMultiples = primitives.VirtualNetworkMultiplePrimitivesFromSubpolicies(ctx, in.Subpolicies, diags)
	o.VirtualNetworkSingles = primitives.VirtualNetworkSinglePrimitivesFromSubpolicies(ctx, in.Subpolicies, diags)
}

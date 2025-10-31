package connectivitytemplates

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint/connectivity_templates/primitives"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
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

type ConnectivityTemplateProtocolEndpoint struct {
	Id              types.String `tfsdk:"id"`
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Tags            types.Set    `tfsdk:"tags"`
	RoutingPolicies types.Map    `tfsdk:"routing_policies"`
}

func (o ConnectivityTemplateProtocolEndpoint) ResourceAttributes() map[string]resourceSchema.Attribute {
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
		"routing_policies": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Routing Policy Primitives to be used with this *Protocol Endpoint*.",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: primitives.RoutingPolicy{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
	}
}

func (o ConnectivityTemplateProtocolEndpoint) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplate {
	result := apstra.ConnectivityTemplate{
		Label:       o.Name.ValueString(),
		Description: o.Description.ValueString(),
		// Tags:        // set below
		// Subpolicies: // set below
	}

	// Set tags
	diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)

	// Set subpolicies
	result.Subpolicies = append(result.Subpolicies, primitives.RoutingPolicySubpolicies(ctx, o.RoutingPolicies, diags)...)

	// try to set the root batch policy ID from o.Id
	if !o.Id.IsUnknown() {
		result.Id = (*apstra.ObjectId)(o.Id.ValueStringPointer()) // nil when null
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

func (o *ConnectivityTemplateProtocolEndpoint) LoadApiData(ctx context.Context, in *apstra.ConnectivityTemplate, diags *diag.Diagnostics) {
	o.Id = types.StringPointerValue((*string)(in.Id))
	o.Name = types.StringValue(in.Label)
	o.Description = value.StringOrNull(ctx, in.Description, diags)
	o.Tags = value.SetOrNull(ctx, types.StringType, in.Tags, diags)
	o.RoutingPolicies = primitives.RoutingPolicyPrimitivesFromSubpolicies(ctx, in.Subpolicies, diags)
}

func (o *ConnectivityTemplateProtocolEndpoint) LoadPrimitiveIds(ctx context.Context, in *apstra.ConnectivityTemplate, diags *diag.Diagnostics) {
	o.Id = types.StringPointerValue((*string)(in.Id))

	o.RoutingPolicies = primitives.LoadIDsIntoRoutingPolicyMap(ctx, in.Subpolicies, o.RoutingPolicies, diags)
}

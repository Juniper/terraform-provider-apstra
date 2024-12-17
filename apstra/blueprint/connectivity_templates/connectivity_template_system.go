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
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectivityTemplateSystem struct {
	Id                 types.String `tfsdk:"id"`
	BlueprintId        types.String `tfsdk:"blueprint_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Tags               types.Set    `tfsdk:"tags"`
	CustomStaticRoutes types.Map    `tfsdk:"custom_static_routes"`
}

func (o ConnectivityTemplateSystem) ResourceAttributes() map[string]resourceSchema.Attribute {
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
			MarkdownDescription: "Connectivity Template Name displayed in the web UI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Connectivity Template Description displayed in the web UI.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags associated with this Connectivity Template.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"custom_static_routes": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of *Custom Static Route* Primitives in this Connectivity Template.",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: primitives.CustomStaticRoute{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *ConnectivityTemplateSystem) ValidateConfig(ctx context.Context, diags *diag.Diagnostics) {
	if o.CustomStaticRoutes.IsUnknown() {
		return
	}

	customStaticRoutes := make(map[string]primitives.CustomStaticRoute)
	diags.Append(o.CustomStaticRoutes.ElementsAs(ctx, &customStaticRoutes, false)...)
	if diags.HasError() {
		return
	}

	for k, v := range o.CustomStaticRoutes.Elements() {
		if v.IsUnknown() {
			continue
		}

		customStaticRoute := customStaticRoutes[k]
		customStaticRoute.ValidateConfig(ctx, path.Root("custom_static_routes").AtMapKey(k), diags)
	}
}

func (o ConnectivityTemplateSystem) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplate {
	result := apstra.ConnectivityTemplate{
		Label:       o.Name.ValueString(),
		Description: o.Description.ValueString(),
		// Tags:        // set below
		// Subpolicies: // set below
	}

	// Set tags
	diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)

	// Set subpolicies
	result.Subpolicies = append(result.Subpolicies, primitives.CustomStaticRouteSubpolicies(ctx, o.CustomStaticRoutes, diags)...)

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

func (o *ConnectivityTemplateSystem) LoadApiData(ctx context.Context, in *apstra.ConnectivityTemplate, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.Description = utils.StringValueOrNull(ctx, in.Description, diags)
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags)
	o.CustomStaticRoutes = primitives.CustomStaticRoutePrimitivesFromSubpolicies(ctx, in.Subpolicies, diags)
}

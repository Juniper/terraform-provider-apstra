package connectivitytemplates

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint/connectivity_templates/primitives"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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
	CustomStaticRoutes types.Set    `tfsdk:"custom_static_routes"`
}

func (o ConnectivityTemplateSystem) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID",
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
		"custom_static_routes": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Set of Custom Static Route Primitives in this Connectivity Template",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: primitives.CustomStaticRoute{}.ResourceAttributes(),
			},
			Optional:   true,
			Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *ConnectivityTemplateSystem) ValidateConfig(ctx context.Context, diags *diag.Diagnostics) {
	if o.CustomStaticRoutes.IsUnknown() {
		return
	}

	var customStaticRoutes []primitives.CustomStaticRoute
	diags.Append(o.CustomStaticRoutes.ElementsAs(ctx, &customStaticRoutes, false)...)
	if diags.HasError() {
		return
	}

	for i, attrVal := range o.CustomStaticRoutes.Elements() {
		if attrVal.IsUnknown() {
			continue
		}

		customStaticRoutes[i].ValidateConfig(ctx, path.Root("custom_static_routes").AtSetValue(attrVal), diags)
	}
}

func (o ConnectivityTemplateSystem) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplate {
	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)

	var customStaticRoutes []primitives.CustomStaticRoute
	diags.Append(o.CustomStaticRoutes.ElementsAs(ctx, &customStaticRoutes, false)...)

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(customStaticRoutes))
	for i, customStaticRoute := range customStaticRoutes {
		subpolicies[i] = customStaticRoute.Request()
	}

	result := apstra.ConnectivityTemplate{
		Label:       o.Name.ValueString(),
		Description: o.Description.ValueString(),
		Tags:        tags,
		Subpolicies: subpolicies,
	}

	// try to set the root batch policy ID from o.Id
	if !o.Id.IsUnknown() {
		id := apstra.ObjectId(o.Id.ValueString())
		result.Id = &id
	}

	// set remaining policy IDs
	err := result.SetIds()
	if err != nil {
		diags.AddError("Failed while generating Connectivity Template IDs", err.Error())
		return nil
	}

	err = result.SetUserData()
	if err != nil {
		diags.AddError("Failed while generating Connectivity Template User Data", err.Error())
		return nil
	}

	return &result
}

func (o *ConnectivityTemplateSystem) LoadApiData(ctx context.Context, in *apstra.ConnectivityTemplate, diags *diag.Diagnostics) {
	var customStaticRoutes []apstra.ConnectivityTemplatePrimitive

	for i, primitive := range in.Subpolicies {
		if primitive == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		switch primitive.Attributes.(type) {
		case *apstra.ConnectivityTemplatePrimitiveAttributesAttachCustomStaticRoute:
			customStaticRoutes = append(customStaticRoutes, *primitive)
		default:
			diags.AddError(
				"Connectivity Template contains Primitives incompatible with the parent policy node (System type)",
				fmt.Sprintf("subpolicy %d has unhandled attribute type %s", i, primitive.Attributes.PolicyTypeName()),
			)
		}

	}
	if diags.HasError() {
		return
	}

	o.Name = types.StringValue(in.Label)
	o.Description = types.StringValue(in.Description)
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags)
	o.CustomStaticRoutes = primitives.NewSetCustomStaticRoutes(ctx, customStaticRoutes, diags)
}

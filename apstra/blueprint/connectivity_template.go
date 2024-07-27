package blueprint

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	connectivitytemplate "github.com/Juniper/terraform-provider-apstra/apstra/connectivity_template"
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

type ConnectivityTemplate struct {
	Id          types.String `tfsdk:"id"`
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.Set    `tfsdk:"tags"`
	Primitives  types.Set    `tfsdk:"primitives"`
}

func (o ConnectivityTemplate) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in web UI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Description displayed in web UI.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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
		"primitives": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Connectivity Template Primitives expressed as JSON strings.",
			ElementType:         types.StringType,
			Required:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

func (o ConnectivityTemplate) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplate {
	var childPrimitivesAsJson []string
	diags.Append(o.Primitives.ElementsAs(ctx, &childPrimitivesAsJson, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := connectivitytemplate.ChildPrimitivesFromListOfJsonStrings(ctx, childPrimitivesAsJson, path.Root("primitives"), diags)
	if diags.HasError() {
		return nil
	}

	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		return nil
	}

	ct := apstra.ConnectivityTemplate{
		Label:       o.Name.ValueString(),
		Description: o.Description.ValueString(),
		Subpolicies: subpolicies,
		Tags:        tags,
	}

	// try to set the root batch policy ID from o.Id
	if !o.Id.IsUnknown() {
		id := apstra.ObjectId(o.Id.ValueString())
		ct.Id = &id
	}

	// set remaining policy IDs
	err := ct.SetIds()
	if err != nil {
		diags.AddError("failed to set CT IDs", err.Error())
		return nil
	}

	ct.SetUserData()

	return &ct
}

func (o *ConnectivityTemplate) LoadApiData(ctx context.Context, in *apstra.ConnectivityTemplate, diags *diag.Diagnostics) {
	// JSON string primitives already part of this object (state)
	oPrimitives := o.Primitives.Elements()

	// JSON string primitives derived from in (api data)
	inPrimitives := connectivitytemplate.SdkPrimitivesToJsonStrings(ctx, in.Subpolicies, diags)
	if diags.HasError() {
		return
	}

	// loop over element indexes common to both oPrimitives an inPrimitives
	for i := 0; i < utils.Min(len(inPrimitives), len(oPrimitives)); i++ {
		// overwrite the state primitive when they're not semantically equal
		if !utils.JSONEqual(inPrimitives[i].(types.String), oPrimitives[i].(types.String), diags) {
			oPrimitives[i] = inPrimitives[i]
		}
		if diags.HasError() {
			return
		}
	}

	// shorten state primitives to match API response length if necessary
	if len(oPrimitives) > len(inPrimitives) {
		oPrimitives = oPrimitives[:len(inPrimitives)]
	}

	// extend state primitives to match API response length if necessary
	if len(inPrimitives) > len(oPrimitives) {
		oPrimitives = append(oPrimitives, inPrimitives[len(oPrimitives):]...)
	}

	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Label)
	o.Description = utils.StringValueOrNull(ctx, in.Description, diags)            // safe to ignore diagnostic here
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags)           // safe to ignore diagnostic here
	o.Primitives = utils.SetValueOrNull(ctx, types.StringType, oPrimitives, diags) // safe to ignore diagnostic here
}

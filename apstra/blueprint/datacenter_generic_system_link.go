package blueprint

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterGenericSystemLink struct {
	TargetSwitchId            types.String `tfsdk:"target_switch_id"`
	TargetSwitchIfName        types.String `tfsdk:"target_switch_if_name"`
	TargetSwitchIfTransformId types.Int64  `tfsdk:"target_switch_if_transform_id"`
	GroupLabel                types.String `tfsdk:"group_label"`
	LagMode                   types.String `tfsdk:"lag_mode"`
	Tags                      types.Set    `tfsdk:"tags"`
}

func (o DatacenterGenericSystemLink) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"target_switch_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Graph Node ID of the Leaf Switch or Access Switch where the link connects.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"target_switch_if_name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name of the physical interface where the link connects on the Leaf Switch " +
				"or Access Switch (\"ge-0/0/1\" or similar).",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"target_switch_if_transform_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "Transformation ID sets the operational mode of an interface.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
		},
		"group_label": resourceSchema.StringAttribute{
			MarkdownDescription: "This field is used to collect multiple links into aggregation " +
				"groups. For example, to create two LAG pairs from four physical links, you might " +
				"use `group_label` value \"bond0\" on two links and \"bond1\" on the other two " +
				"links. Apstra assigns a unique group ID to each link by default.",
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"lag_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "LAG negotiation mode of the Link. All links with the same " +
				"`group_label` must use the value.",
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					apstra.RackLinkLagModeActive.String(),
					apstra.RackLinkLagModePassive.String(),
					apstra.RackLinkLagModeStatic.String(),
				),
			},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Names of Tag to be applied to this Link. If a Tag doesn't exist " +
				"in the Blueprint it will be created automatically.",
			ElementType: types.StringType,
			Optional:    true,
			Validators:  []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
	}
}

func (o DatacenterGenericSystemLink) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"target_switch_id":              types.StringType,
		"target_switch_if_name":         types.StringType,
		"target_switch_if_transform_id": types.Int64Type,
		"group_label":                   types.StringType,
		"lag_mode":                      types.StringType,
		"tags":                          types.SetType{ElemType: types.StringType},
	}
}

func (o DatacenterGenericSystemLink) request(ctx context.Context, diags *diag.Diagnostics) *apstra.CreateLinkRequest {
	result := apstra.CreateLinkRequest{
		SwitchEndpoint: apstra.SwitchLinkEndpoint{
			TransformationId: int(o.TargetSwitchIfTransformId.ValueInt64()),
			SystemId:         apstra.ObjectId(o.TargetSwitchId.ValueString()),
			IfName:           o.TargetSwitchIfName.ValueString(),
		},
		GroupLabel: o.GroupLabel.ValueString(),
	}

	diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)

	err := result.LagMode.FromString(o.LagMode.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("failed parsing lag mode %s", o.LagMode), err.Error())
	}

	return &result
}

func (o *DatacenterGenericSystemLink) digest() string {
	return o.TargetSwitchId.ValueString() + ":" + o.TargetSwitchIfName.ValueString()
}

func (o *DatacenterGenericSystemLink) loadApiData(ctx context.Context, in *apstra.CablingMapLink, genericSystemId apstra.ObjectId, diags *diag.Diagnostics) {
	switchEndpoint := in.OppositeEndpointBySystemId(genericSystemId)

	o.TargetSwitchId = types.StringValue(switchEndpoint.System.Id.String())
	o.TargetSwitchIfName = types.StringValue(*switchEndpoint.Interface.IfName)
	o.GroupLabel = types.StringValue(in.GroupLabel)
	o.LagMode = utils.StringValueOrNull(ctx, switchEndpoint.Interface.LagMode.String(), diags)
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.TagLabels, diags)
}

func (o *DatacenterGenericSystemLink) getTransformId(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	if !utils.Known(o.TargetSwitchId) {
		diags.AddError(
			"provider bug",
			"attempt to get interface transform ID without TargetSwitchId - please report this issue to the maintainers")
		return
	}

	if !utils.Known(o.TargetSwitchIfName) {
		diags.AddError(
			"provider bug",
			"attempt to get interface transform ID without TargetSwitchIfName - please report this issue to the maintainers")
		return
	}

	transformId, err := client.GetTransformationIdByIfName(ctx, apstra.ObjectId(o.TargetSwitchId.ValueString()), o.TargetSwitchIfName.ValueString())
	if err != nil {
		var ace apstra.ClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			o.TargetSwitchIfTransformId = types.Int64Null()
			return
		}
		diags.AddError(fmt.Sprintf("failed to get transform ID for %q", o.digest()), err.Error())
		return
	}

	o.TargetSwitchIfTransformId = types.Int64Value(int64(transformId))
}

// updateParams checks/updates the following link parameters.
// - transform ID
// - group label
// - LAG mode
// - tags
// Because the DatacenterGenericSystemLink object doesn't know the link ID,
// the ID of the link's graph node is passed as a function argument.
func (o *DatacenterGenericSystemLink) updateParams(ctx context.Context, id apstra.ObjectId, state *DatacenterGenericSystemLink, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// set the transform ID if it has changed
	if !o.TargetSwitchIfTransformId.Equal(state.TargetSwitchIfTransformId) {
		err := client.SetTransformIdByIfName(ctx, apstra.ObjectId(o.TargetSwitchId.ValueString()),
			o.TargetSwitchIfName.ValueString(), int(o.TargetSwitchIfTransformId.ValueInt64()))
		if err != nil {
			var ace apstra.ClientErr
			if errors.As(err, &ace) && ace.Type() == apstra.ErrCannotChangeTransform {
				diags.AddWarning("could not change interface transform", err.Error())
			} else {
				diags.AddError("failed to set interface transform", err.Error())
				return
			}
		}
	}

	// set the tags, lag mode and group label if any have changed
	if !o.Tags.Equal(state.Tags) || !o.LagMode.Equal(state.LagMode) || !o.GroupLabel.Equal(state.GroupLabel) {
		var tags []string
		diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
		if tags == nil {
			tags = []string{} // convert nil -> empty slice to clear tags
		}

		var lagMode apstra.RackLinkLagMode
		err := lagMode.FromString(o.LagMode.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to parse lag mode %s", o.LagMode), err.Error())
			return
		}

		// set lag params + tag set
		err = client.SetLinkLagParams(ctx, &apstra.SetLinkLagParamsRequest{id: apstra.LinkLagParams{
			GroupLabel: o.GroupLabel.ValueString(),
			LagMode:    lagMode,
			Tags:       tags,
		}})
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to set link %s LAG parameters", id), err.Error())
		}
	}
}

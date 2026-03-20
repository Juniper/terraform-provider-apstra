package blueprint

import (
	"context"
	"errors"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterGenericSystemLink struct {
	TargetSwitchId            types.String `tfsdk:"target_switch_id"`
	TargetSwitchIfName        types.String `tfsdk:"target_switch_if_name"`
	TargetSwitchIfTransformId types.Int64  `tfsdk:"target_switch_if_transform_id"`
	GenericSystemIfName       types.String `tfsdk:"generic_system_if_name"`
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
		},
		"generic_system_if_name": resourceSchema.StringAttribute{
			DeprecationMessage: "This is an experimental feature, not fully supported in the current release - use at your own risk",
			MarkdownDescription: "Name of the physical interface where the link connects. This attribute is " +
				"reflected in the cabling map and is informational only. Apstra doesn't require it, but including a " +
				"value here  may be useful for scoping Configlets or in other scenarios that rely on recording the " +
				"server's interface name (for example, `enp5s0d1`). Note that populating this field will slow Generic " +
				"Server creation.",
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
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
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o DatacenterGenericSystemLink) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"target_switch_id":              types.StringType,
		"target_switch_if_name":         types.StringType,
		"target_switch_if_transform_id": types.Int64Type,
		"generic_system_if_name":        types.StringType,
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

// Digest returns a string composed of the switch ID and Interface name joined by ':'.
// For example: "scausZjtxhDFyRatlQ:xe-0/0/0"
func (o *DatacenterGenericSystemLink) Digest() string {
	return o.TargetSwitchId.ValueString() + ":" + o.TargetSwitchIfName.ValueString()
}

func (o *DatacenterGenericSystemLink) loadApiData(ctx context.Context, in *apstra.CablingMapLink, genericSystemId string, diags *diag.Diagnostics) {
	switchEndpoint := in.OppositeEndpointBySystemID(genericSystemId)
	if switchEndpoint != nil {
		if switchEndpoint.System != nil {
			o.TargetSwitchId = types.StringValue(switchEndpoint.System.ID)
		}
		if switchEndpoint.Interface.Name != nil {
			o.TargetSwitchIfName = types.StringPointerValue(switchEndpoint.Interface.Name)
		}
		if switchEndpoint.Interface.LAGMode != nil {
			o.LagMode = value.StringOrNull(ctx, switchEndpoint.Interface.LAGMode.String(), diags)
		}
	}

	serverEndpoint := in.EndpointBySystemID(genericSystemId)
	if serverEndpoint != nil {
		o.GenericSystemIfName = types.StringPointerValue(serverEndpoint.Interface.Name)
	}
	o.GroupLabel = types.StringPointerValue(in.GroupLabel)
	o.Tags = value.SetOrNull(ctx, types.StringType, in.TagLabels, diags)
}

func (o *DatacenterGenericSystemLink) getTransformId(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	if !utils.HasValue(o.TargetSwitchId) {
		diags.AddError(
			"provider bug",
			"attempt to get interface transform ID without TargetSwitchId - please report this issue to the maintainers")
		return
	}

	if !utils.HasValue(o.TargetSwitchIfName) {
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
		diags.AddError(fmt.Sprintf("failed to get transform ID for %q", o.Digest()), err.Error())
		return
	}

	o.TargetSwitchIfTransformId = types.Int64Value(int64(transformId))
}

func (o *DatacenterGenericSystemLink) lagParams(ctx context.Context, id apstra.ObjectId, state *DatacenterGenericSystemLink, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) *apstra.LinkLagParams {
	if o.Tags.Equal(state.Tags) && o.LagMode.Equal(state.LagMode) && o.GroupLabel.Equal(state.GroupLabel) {
		return nil // nothing to do
	}

	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if tags == nil {
		tags = []string{} // convert nil -> empty slice to clear tags
	}

	var lagMode apstra.RackLinkLagMode
	err := lagMode.FromString(o.LagMode.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse lag mode %s", o.LagMode), err.Error())
		return nil
	}

	return &apstra.LinkLagParams{
		GroupLabel: o.GroupLabel.ValueString(),
		LagMode:    lagMode,
		Tags:       tags,
	}
}

func (o *DatacenterGenericSystemLink) updateTransformId(ctx context.Context, state *DatacenterGenericSystemLink, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
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
}

package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
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
	Id                        types.String `tfsdk:"id"`
	TargetSwitchId            types.String `tfsdk:"target_switch_id"`
	TargetSwitchIfName        types.String `tfsdk:"target_switch_if_name"`
	TargetSwitchIfTransformId types.Int64  `tfsdk:"target_switch_if_transform_id"`
	GroupLabel                types.String `tfsdk:"group_label"`
	LagMode                   types.String `tfsdk:"lag_mode"`
	Tags                      types.Set    `tfsdk:"tags"`
}

func (o DatacenterGenericSystemLink) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Graph datastore ID of the link node.",
			Computed:            true,
		},
		"lag_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "LAG negotiation mode of the Link.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					apstra.RackLinkLagModeActive.String(),
					apstra.RackLinkLagModePassive.String(),
					apstra.RackLinkLagModeStatic.String(),
				),
			},
		},
		"group_label": resourceSchema.StringAttribute{
			MarkdownDescription: "This field is used to collect multiple links into aggregation " +
				"groups. For example, to create two LAG pairs from four physical links, you might " +
				"use `group_label` value \"bond0\" on two links and \"bond1\" on the other two links",
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Names of Tag to be applied to this Link. If a Tag doesn't exist " +
				"in the Blueprint it will be created automatically.",
			ElementType: types.StringType,
			Optional:    true,
			Validators:  []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
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
	}
}

func (o DatacenterGenericSystemLink) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"lag_mode":                      types.StringType,
		"group_label":                   types.StringType,
		"id":                            types.StringType,
		"tags":                          types.SetType{ElemType: types.StringType},
		"target_switch_id":              types.StringType,
		"target_switch_if_name":         types.StringType,
		"target_switch_if_transform_id": types.Int64Type,
	}
}

func (o DatacenterGenericSystemLink) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.CreateLinksWithNewServerRequestLink {
	result := apstra.CreateLinksWithNewServerRequestLink{
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

//func (o *DatacenterGenericSystemLink) GenericSystemQuery() *apstra.PathQuery {
//	query := new(apstra.PathQuery).
//		Node([]apstra.QEEAttribute{apstra.NodeTypeLink.QEEAttribute(),
//			{"id", apstra.QEStringVal(o.Id.ValueString())},
//		}).
//		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
//		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
//		In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
//		Node([]apstra.QEEAttribute{apstra.NodeTypeSystem.QEEAttribute(),
//			{"role", apstra.QEStringVal("generic")},
//			{"name", apstra.QEStringVal("n_generic")},
//		})
//
//	return query
//}

//func (o DatacenterGenericSystemLink) GenericSystemQueryResponse() struct {
//	Items []struct {
//		Generic struct {
//			Hostname string `json:"hostname"`
//			Id       string `json:"id"`
//			Label    string `json:"label"`
//		} `json:"n_generic"`
//	} `json:"items"`
//} {
//	return struct {
//		Items []struct {
//			Generic struct {
//				Hostname string `json:"hostname"`
//				Id       string `json:"id"`
//				Label    string `json:"label"`
//			} `json:"n_generic"`
//		} `json:"items"`
//	}{}
//}

//func (o *DatacenterGenericSystemLink) endpoint(ctx context.Context, diags *diag.Diagnostics) *DatacenterGenericSystemLinkEndpoint {
//	var result DatacenterGenericSystemLinkEndpoint
//	diags.Append(o.SwitchEndpoint.As(ctx, &result, basetypes.ObjectAsOptions{})...)
//	return &result
//}

//func (o *DatacenterGenericSystemLink) loadApiData(ctx context.Context, in apstra.CablingMapLink, diags *diag.Diagnostics) {
//	//SwitchEndpoint types.Object `tfsdk:"switch_endpoint"`
//	//LagMode        types.String `tfsdk:"lag_mode"`
//
//	o.Id = types.StringValue(in.Id.String())
//	o.GroupLabel = utils.StringValueOrNull(ctx, in.GroupLabel, diags)
//	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.TagLabels, diags)
//	//o.LagMode = types.StringValue(in.)
//}

func (o *DatacenterGenericSystemLink) digest() string {
	return o.TargetSwitchId.ValueString() + ":" + o.TargetSwitchIfName.ValueString()
}

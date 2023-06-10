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
	"terraform-provider-apstra/apstra/utils"
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

	query := new(apstra.PathQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(client.Id()).
		SetClient(client.Client()).
		Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(o.TargetSwitchId.ValueString())}}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeInterfaceMap.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterfaceMap.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_interface_map")},
		})

	var queryResponse struct {
		Items []struct {
			InterfaceMap struct {
				Id         string `json:"id"`
				Interfaces []struct {
					Mapping []int  `json:"mapping"`
					Name    string `json:"name"`
				} `json:"interfaces"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResponse)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed querying for node %s interface map", o.TargetSwitchId), err.Error())
		return
	}
	if len(queryResponse.Items) != 1 {
		diags.AddError(fmt.Sprintf("query found %d results, expected 1", len(queryResponse.Items)), query.String())
		return
	}

	for _, iMapInterface := range queryResponse.Items[0].InterfaceMap.Interfaces {
		if iMapInterface.Name != o.TargetSwitchIfName.ValueString() {
			continue
		}

		o.TargetSwitchIfTransformId = types.Int64Value(int64(iMapInterface.Mapping[1]))
		return
	}

	diags.AddError(
		fmt.Sprintf("failed to find in-use transform ID for interface %s", o.TargetSwitchIfName),
		fmt.Sprintf("interface map %q has %d interfaces, but none are named %s",
			queryResponse.Items[0].InterfaceMap.Id,
			len(queryResponse.Items[0].InterfaceMap.Interfaces),
			o.TargetSwitchIfName,
		),
	)
}

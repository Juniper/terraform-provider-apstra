package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualNetworkBinding struct {
	LeafId customtypes.StringWithAltValues `tfsdk:"leaf_id"`
	VlanId types.Int64                     `tfsdk:"vlan_id"`
	//AccessIds types.Set                       `tfsdk:"access_ids"`
	AccessIds customtypes.SetWithSemanticEquals `tfsdk:"access_ids"`
}

func (o VirtualNetworkBinding) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"leaf_id": customtypes.StringWithAltValuesType{},
		"vlan_id": types.Int64Type,
		//"access_ids": types.SetType{ElemType: customtypes.StringWithAltValuesType{}},
		"access_ids": customtypes.NewSetWithSemanticEqualsType(customtypes.StringWithAltValuesType{}),
		//"access_ids": customtypes.NewSetWithSemanticEqualsType(types.StringType),
	}
}

func (o VirtualNetworkBinding) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"leaf_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Leaf Switch ID",
			Required:            true,
			CustomType:          customtypes.StringWithAltValuesType{},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"vlan_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "VLAN ID",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(design.VlanMin, design.VlanMax)},
		},
		"access_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Access Switch IDs associated with this Leaf Switch",
			CustomType:          customtypes.NewSetWithSemanticEqualsType(customtypes.StringWithAltValuesType{}),
			Optional:            true,
			ElementType:         customtypes.StringWithAltValuesType{},
			//ElementType: types.StringType,
			Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

func (o VirtualNetworkBinding) Request(ctx context.Context, rgInfo map[string]*apstra.RedundancyGroupInfo, diags *diag.Diagnostics) *apstra.VnBinding {
	var vlanId *apstra.Vlan
	if !o.VlanId.IsNull() {
		vlanId = utils.ToPtr(apstra.Vlan(o.VlanId.ValueInt64()))
	}

	//var accessSwitchNodeIds []apstra.ObjectId
	//diags.Append(o.AccessIds.ElementsAs(ctx, &accessSwitchNodeIds, false)...)
	//for i, id := range accessSwitchNodeIds {
	//	// This access switch may be half of a pair...
	//	if rgi, ok := rgInfo[id.String()]; ok {
	//		// ESI/MLAG switch pair member. Use the group ID instead.
	//		accessSwitchNodeIds[i] = rgi.Id
	//	}
	//}
	//utils.Uniq(accessSwitchNodeIds) // if the user specified both switches, we created dups. Clean em up.

	return &apstra.VnBinding{
		//AccessSwitchNodeIds: accessSwitchNodeIds,
		SystemId: apstra.ObjectId(o.LeafId.ValueString()),
		VlanId:   vlanId,
	}
}

func (o *VirtualNetworkBinding) LoadApiData(ctx context.Context, in apstra.VnBinding, rgiMap map[string]*apstra.RedundancyGroupInfo, cfgVlanMap map[string]int64, diags *diag.Diagnostics) {
	// find the redundancy group info for this leaf switch (if any)
	rgi := rgiMap[o.LeafId.ValueString()]

	// set leaf ID
	if rgi != nil {
		o.LeafId = customtypes.NewStringWithAltValuesValue(rgi.Id.String(), rgi.SystemIds[0].String(), rgi.SystemIds[1].String())
	} else {
		o.LeafId = customtypes.NewStringWithAltValuesValue(in.SystemId.String())
	}

	// Set VLAN id. Maybe.
	if in.VlanId != nil {
		if rgi == nil { // The leaf might be part of a redundancy group.
			// Only set VLAN id if the leaf previously had one assigned.
			if cfgVlanMap[o.LeafId.ValueString()] > 0 {
				o.VlanId = types.Int64Value(int64(*in.VlanId))
			}
		} else {
			// Only set VLAN id if any ID associated with the redundancy group previously had one assigned.
			if cfgVlanMap[rgi.Id.String()] > 0 || cfgVlanMap[rgi.SystemIds[0].String()] > 0 || cfgVlanMap[rgi.SystemIds[1].String()] > 0 {
				o.VlanId = types.Int64Value(int64(*in.VlanId))
			}
		}
	}

	//// Set access ids.
	//accessIds := make([]customtypes.StringWithAltValues, len(in.AccessSwitchNodeIds))
	//for i, accessSwitchNodeId := range in.AccessSwitchNodeIds {
	//	if rgi, ok := rgiMap[accessSwitchNodeId.String()]; ok {
	//		accessIds[i] = customtypes.NewStringWithAltValuesValue(rgi.Id.String(), rgi.SystemIds[0].String(), rgi.SystemIds[1].String())
	//	} else {
	//		accessIds[i] = customtypes.NewStringWithAltValuesValue(accessSwitchNodeId.String())
	//	}
	//}
	//o.AccessIds = utils.SetValueOrNull(ctx, customtypes.StringWithAltValuesType{}, accessIds, diags)
}

//func (o *VirtualNetworkBinding) makeRedundant(_ context.Context, rgi apstra.RedundancyGroupInfo, diags *diag.Diagnostics) map[apstra.ObjectId]VirtualNetworkBinding {
//	if o.LeafId.ValueString() != rgi.SystemIds[0].String() &&
//		o.LeafId.ValueString() != rgi.SystemIds[1].String() {
//		diags.AddError(constants.ErrProviderBug, "attempting to makeRedundant but leaf ID not found in redundancy group info")
//		return nil
//	}
//
//	result := make(map[apstra.ObjectId]VirtualNetworkBinding, 2)
//	for _, id := range rgi.SystemIds {
//		result[]
//	}
//}

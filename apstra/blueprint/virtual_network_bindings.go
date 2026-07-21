package blueprint

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualNetworkBindings struct {
	BlueprintId        types.String `tfsdk:"blueprint_id"`
	VirtualNetworkId   types.String `tfsdk:"virtual_network_id"`
	Bindings           types.Set    `tfsdk:"bindings"`
	DhcpServiceEnabled types.Bool   `tfsdk:"dhcp_service_enabled"`
}

func (o VirtualNetworkBindings) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"blueprint_id":         types.StringType,
		"virtual_network_id":   types.StringType,
		"bindings":             types.SetType{ElemType: types.ObjectType{AttrTypes: VirtualNetworkBinding{}.AttrTypes()}},
		"dhcp_service_enabled": types.BoolType,
	}
}

func (o VirtualNetworkBindings) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"virtual_network_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Virtual Network ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"bindings": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Assignment info for each Leaf Switch and any downstream Access Switches. " +
				"Leaf switch IDs must not appear more than once in this set.",
			Required: true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: VirtualNetworkBinding{}.ResourceAttributes(),
			},
			Validators: []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"dhcp_service_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether DHCP relaying is enabled. To avoid state churn, all VN binding " +
				"resources must agree about this setting. Default value: `false`.",
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
	}
}

func (o VirtualNetworkBindings) Request(ctx context.Context, rgMap map[string]*apstra.RedundancyGroupInfo, ps private.State, diags *diag.Diagnostics) *apstra.VirtualNetworkBindingsRequest {
	// private state enumerates previously-created bindings which we may need to delete
	var p private.ResourceDatacenterVirtualNetworkBindings
	if ps != nil {
		p.LoadPrivateState(ctx, ps, diags)
		if diags.HasError() {
			return nil
		}
	}

	var vnBindingSlice []VirtualNetworkBinding
	diags.Append(o.Bindings.ElementsAs(ctx, &vnBindingSlice, false)...)
	if diags.HasError() {
		return nil
	}

	// Build a map of bindings we'll send to the API. Because of possible redundancy
	// group IDs, we don't know the actual size of this map yet.
	vnBindings := make(map[apstra.ObjectId]*apstra.VnBinding)
	for _, vnBinding := range vnBindingSlice {
		// Determine if the leaf binding should be treated as an ESI/MLAG binding
		if rgi, ok := rgMap[vnBinding.LeafId.ValueString()]; ok {
			// This leaf switch ID is half of a pair. Swap in the RG ID in its place.
			vnBinding.LeafId = customtypes.NewStringWithAltValuesValue(rgi.Id.String())
		}

		vnBindings[apstra.ObjectId(vnBinding.LeafId.ValueString())] = vnBinding.Request(ctx, rgMap, diags)
		delete(p.SystemIdToVlan, vnBinding.LeafId.ValueString()) // remove this from the to-be-deleted list
	}
	for deleteMe := range p.SystemIdToVlan {
		vnBindings[apstra.ObjectId(deleteMe)] = nil
	}

	return &apstra.VirtualNetworkBindingsRequest{
		VnId:               apstra.ObjectId(o.VirtualNetworkId.ValueString()),
		VnBindings:         vnBindings,
		SviIps:             nil, // todo
		DhcpServiceEnabled: (*apstra.DhcpServiceEnabled)(o.DhcpServiceEnabled.ValueBoolPointer()),
	}
}

func (o *VirtualNetworkBindings) LoadApiData(ctx context.Context, in *apstra.VirtualNetworkData, rgiMap map[string]*apstra.RedundancyGroupInfo, ps private.State, diags *diag.Diagnostics) {
	var p private.ResourceDatacenterVirtualNetworkBindings
	p.LoadPrivateState(ctx, ps, diags)
	if diags.HasError() {
		return
	}

	var bindings []VirtualNetworkBinding
	for _, b := range in.VnBindings {
		if _, ok := p.SystemIdToVlan[b.SystemId.String()]; !ok {
			continue // ignore leaf bindings not previously created by this resource
		}

		var binding VirtualNetworkBinding
		binding.LoadApiData(ctx, b, rgiMap, p.SystemIdToVlan, diags)
		bindings = append(bindings, binding)
	}
	if diags.HasError() {
		return
	}

	o.Bindings = utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: VirtualNetworkBinding{}.AttrTypes()}, bindings, diags)
	o.DhcpServiceEnabled = types.BoolValue(bool(in.DhcpService))
}

func (o VirtualNetworkBindings) SetPrivateState(ctx context.Context, rgiMap map[string]*apstra.RedundancyGroupInfo, ps private.State, diags *diag.Diagnostics) {
	// extract bindings
	var ourBindings []VirtualNetworkBinding
	diags.Append(o.Bindings.ElementsAs(ctx, &ourBindings, false)...)
	if diags.HasError() {
		return
	}

	// convert bindings to SDK type
	sdkBindings := make([]apstra.VnBinding, len(ourBindings))
	for i, ourBinding := range ourBindings {
		sdkBindings[i] = *ourBinding.Request(ctx, rgiMap, diags)
	}
	if diags.HasError() {
		return
	}

	// extract slice of leaf IDs
	leafIds := make([]string, len(ourBindings))
	for i, ourBinding := range ourBindings {
		leafIds[i] = ourBinding.LeafId.ValueString()
	}

	// load private state object
	var p private.ResourceDatacenterVirtualNetworkBindings
	p.LoadSystemIdToVlanApiData(ctx, sdkBindings, diags)
	//p.LoadRedundancyGroupIdToSystemIDsApiData(ctx, rgiMap, leafIds, diags)
	if diags.HasError() {
		return
	}

	// set private state
	p.SetPrivateState(ctx, ps, diags)
}

//func (o VirtualNetworkBindings) GetRedundancyGroupMemebership(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) map[string][2]string {
//	query := new(apstra.PathQuery).
//		SetClient(client).
//		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeRedundancyGroup.QEEAttribute(),
//			{Key: "name", Value: apstra.QEStringVal("n_redundancy_group")},
//		}).
//		Out([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOfSystems.QEEAttribute()}).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeSystem.QEEAttribute(),
//			{Key: "system_type", Value: apstra.QEStringVal("switch")},
//			{Key: "name", Value: apstra.QEStringVal("n_system")},
//		})
//
//	var queryResult struct {
//		Items []struct {
//			RedundancyGroup struct {
//				ID string `json:"id"`
//			} `json:"n_redundancy_group"`
//			System struct {
//				ID string `json:"id"`
//			} `json:"n_system"`
//		} `json:"items"`
//	}
//
//	err := query.Do(ctx, &queryResult)
//	if err != nil {
//		diags.AddError("Failed querying for redundancy groups", err.Error())
//		return nil
//	}
//
//	result := make(map[string][2]string, len(queryResult.Items))
//	for _, item := range queryResult.Items {
//		resultItem, ok := result[item.RedundancyGroup.ID]
//		if ok {
//			resultItem[1] = item.RedundancyGroup.ID // resultItem already existed. Add the item at index 1.
//		} else {
//			resultItem[0] = item.RedundancyGroup.ID // resultItem is the zero value. Add the first item.
//		}
//		result[item.RedundancyGroup.ID] = resultItem // Add the updated array to the map.
//	}
//
//	return result
//}

//func redundantVnBindings(binding VirtualNetworkBinding, rgi apstra.RedundancyGroupInfo) map[apstra.ObjectId]*apstra.VnBinding {
//	result := make(map[apstra.ObjectId]*apstra.VnBinding, 2)
//	for _, sysId := range rgi.SystemIds {
//		result[sysId] =VirtualNetworkBinding{
//			LeafId:    customtypes.NewStringWithAltValuesValue(rgi.SystemIds[0].String()),
//			VlanId:    vnBinding.VlanId,
//			AccessIds: types.Set{},
//		}.Request(ctx, diags)
//	}
//	}
//}

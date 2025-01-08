package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"virtual_network_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Virtual Network ID.",
			Required:            true,
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

func (o VirtualNetworkBindings) Request(ctx context.Context, ps private.State, diags *diag.Diagnostics) *apstra.VirtualNetworkBindingsRequest {
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

	vnBindings := make(map[apstra.ObjectId]*apstra.VnBinding, len(o.Bindings.Elements()))
	for _, vnBinding := range vnBindingSlice {
		vnBindings[apstra.ObjectId(vnBinding.LeafId.ValueString())] = vnBinding.Request(ctx, diags)
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

func (o VirtualNetworkBindings) SetPrivateState(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
	var bindings []VirtualNetworkBinding
	diags.Append(o.Bindings.ElementsAs(ctx, &bindings, false)...)
	if diags.HasError() {
		return
	}

	var p private.ResourceDatacenterVirtualNetworkBindings
	p.SystemIdToVlan = make(map[string]int64, len(bindings))
	for _, binding := range bindings {
		p.SystemIdToVlan[binding.LeafId.ValueString()] = binding.VlanId.ValueInt64() // 0 value for null case is okay
	}

	p.SetPrivateState(ctx, ps, diags)
}

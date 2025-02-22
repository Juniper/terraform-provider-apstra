package primitives

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualNetworkSingle struct {
	Id                       types.String `tfsdk:"id"`
	BatchId                  types.String `tfsdk:"batch_id"`
	PipelineId               types.String `tfsdk:"pipeline_id"`
	VirtualNetworkId         types.String `tfsdk:"virtual_network_id"`
	Tagged                   types.Bool   `tfsdk:"tagged"`
	BgpPeeringGenericSystems types.Map    `tfsdk:"bgp_peering_generic_systems"`
	StaticRoutes             types.Map    `tfsdk:"static_routes"`
}

func (o VirtualNetworkSingle) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                          types.StringType,
		"batch_id":                    types.StringType,
		"pipeline_id":                 types.StringType,
		"virtual_network_id":          types.StringType,
		"tagged":                      types.BoolType,
		"bgp_peering_generic_systems": types.MapType{ElemType: types.ObjectType{AttrTypes: BgpPeeringGenericSystem{}.AttrTypes()}},
		"static_routes":               types.MapType{ElemType: types.ObjectType{AttrTypes: StaticRoute{}.AttrTypes()}},
	}
}

func (o VirtualNetworkSingle) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive element",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"batch_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's downstream collection",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{virtualNetworkSingleBatchIdPlanModifier{}},
		},
		"pipeline_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this CT Primitive Element's upstream pipeline",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"virtual_network_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the desired Virtual Network",
			Required:            true,
		},
		"tagged": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the selected Virtual Network should be presented with an 802.1Q tag",
			Required:            true,
		},
		"bgp_peering_generic_systems": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of BGP Peering (Generic System) primitives",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: BgpPeeringGenericSystem{}.ResourceAttributes(),
			},
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
			Optional:   true,
		},
		"static_routes": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Static Route primitives",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: StaticRoute{}.ResourceAttributes(),
			},
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
			Optional:   true,
		},
	}
}

func (o VirtualNetworkSingle) attributes(_ context.Context, _ *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan {
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan{
		Tagged:   o.Tagged.ValueBool(),
		VnNodeId: (*apstra.ObjectId)(o.VirtualNetworkId.ValueStringPointer()),
	}
}

func (o VirtualNetworkSingle) primitive(ctx context.Context, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	result := apstra.ConnectivityTemplatePrimitive{Attributes: o.attributes(ctx, diags)}

	if !o.PipelineId.IsUnknown() {
		result.PipelineId = (*apstra.ObjectId)(o.PipelineId.ValueStringPointer()) // nil when null
	}
	if !o.Id.IsUnknown() {
		result.Id = (*apstra.ObjectId)(o.Id.ValueStringPointer()) // nil when null
	}
	if !o.BatchId.IsUnknown() {
		result.BatchId = (*apstra.ObjectId)(o.BatchId.ValueStringPointer()) // nil when null
	}

	result.Subpolicies = append(result.Subpolicies, BgpPeeringGenericSystemSubpolicies(ctx, o.BgpPeeringGenericSystems, diags)...)
	result.Subpolicies = append(result.Subpolicies, StaticRouteSubpolicies(ctx, o.StaticRoutes, diags)...)

	return &result
}

func VirtualNetworkSingleSubpolicies(ctx context.Context, virtualNetworkSingleMap types.Map, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	var VirtualNetworkSingles map[string]VirtualNetworkSingle
	diags.Append(virtualNetworkSingleMap.ElementsAs(ctx, &VirtualNetworkSingles, false)...)
	if diags.HasError() {
		return nil
	}

	subpolicies := make([]*apstra.ConnectivityTemplatePrimitive, len(VirtualNetworkSingles))
	i := 0
	for k, v := range VirtualNetworkSingles {
		subpolicies[i] = v.primitive(ctx, diags)
		if diags.HasError() {
			return nil
		}
		subpolicies[i].Label = k
		i++
	}

	return subpolicies
}

func newVirtualNetworkSingle(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan, _ *diag.Diagnostics) VirtualNetworkSingle {
	return VirtualNetworkSingle{
		// Name:          // handled by caller
		VirtualNetworkId: types.StringPointerValue((*string)(in.VnNodeId)),
		Tagged:           types.BoolValue(in.Tagged),
	}
}

func VirtualNetworkSinglePrimitivesFromSubpolicies(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) types.Map {
	result := make(map[string]VirtualNetworkSingle)

	for i, subpolicy := range subpolicies {
		if subpolicy == nil {
			diags.AddError(constants.ErrProviderBug, fmt.Sprintf("subpolicy %d in API response is nil", i))
			continue
		}

		if p, ok := (subpolicy.Attributes).(*apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan); ok {
			if p == nil {
				diags.AddError(
					"API response contains nil subpolicy",
					"While extracting RoutingPolicy primitives, encountered nil subpolicy at index "+strconv.Itoa(i),
				)
				continue
			}

			newPrimitive := newVirtualNetworkSingle(ctx, p, diags)
			newPrimitive.PipelineId = types.StringPointerValue((*string)(subpolicy.PipelineId))
			newPrimitive.Id = types.StringPointerValue((*string)(subpolicy.Id))
			newPrimitive.BatchId = types.StringPointerValue((*string)(subpolicy.BatchId))
			newPrimitive.BgpPeeringGenericSystems = BgpPeeringGenericSystemPrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			newPrimitive.StaticRoutes = StaticRoutePrimitivesFromSubpolicies(ctx, subpolicy.Subpolicies, diags)
			result[subpolicy.Label] = newPrimitive
		}
	}
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: VirtualNetworkSingle{}.AttrTypes()})
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: VirtualNetworkSingle{}.AttrTypes()}, result, diags)
}

func LoadIDsIntoVirtualNetworkSingleMap(ctx context.Context, subpolicies []*apstra.ConnectivityTemplatePrimitive, inMap types.Map, diags *diag.Diagnostics) types.Map {
	result := make(map[string]VirtualNetworkSingle, len(inMap.Elements()))
	inMap.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: VirtualNetworkSingle{}.AttrTypes()})
	}

	for _, p := range subpolicies {
		if _, ok := p.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan); !ok {
			continue // wrong type and nil value both wind up getting skipped
		}

		if v, ok := result[p.Label]; ok {
			v.PipelineId = types.StringPointerValue((*string)(p.PipelineId))
			v.Id = types.StringPointerValue((*string)(p.Id))
			v.BatchId = types.StringPointerValue((*string)(p.BatchId))
			v.BgpPeeringGenericSystems = LoadIDsIntoBgpPeeringGenericSystemMap(ctx, p.Subpolicies, v.BgpPeeringGenericSystems, diags)
			v.StaticRoutes = LoadIDsIntoStaticRouteMap(ctx, p.Subpolicies, v.StaticRoutes, diags)
			result[p.Label] = v
		}
	}

	return utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: VirtualNetworkSingle{}.AttrTypes()}, result, diags)
}

var _ planmodifier.String = (*virtualNetworkSingleBatchIdPlanModifier)(nil)

type virtualNetworkSingleBatchIdPlanModifier struct{}

func (o virtualNetworkSingleBatchIdPlanModifier) Description(_ context.Context) string {
	return "preserves the the state value unless we're transitioning between zero and non-zero child primitives, in which case null or unknown is planned"
}

func (o virtualNetworkSingleBatchIdPlanModifier) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o virtualNetworkSingleBatchIdPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var plan, state VirtualNetworkSingle

	// unpacking the parent object's planned value should always work
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, req.Path.ParentPath(), &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// attempting to unpack the parent object's state indicates whether state *exists*
	d := req.State.GetAttribute(ctx, req.Path.ParentPath(), &state)
	stateDoesNotExist := d.HasError()

	planHasChildren := len(plan.BgpPeeringGenericSystems.Elements())+
		len(plan.StaticRoutes.Elements()) > 0

	planChildrenUnknown := plan.BgpPeeringGenericSystems.IsUnknown() ||
		plan.StaticRoutes.IsUnknown()

	// are we a new object?
	if stateDoesNotExist {
		if planHasChildren || planChildrenUnknown {
			resp.PlanValue = types.StringUnknown()
		} else {
			resp.PlanValue = types.StringNull()
		}
		return
	}

	stateHasChildren := len(state.BgpPeeringGenericSystems.Elements())+
		len(state.StaticRoutes.Elements()) > 0

	if (planHasChildren || planChildrenUnknown) == stateHasChildren {
		// state and plan agree about whether a batch ID is required. Reuse the old value.
		resp.PlanValue = req.StateValue
		return
	}

	// We've either gained our first, or lost our last child primitive. Set the plan value accordingly.
	if planHasChildren || planChildrenUnknown {
		resp.PlanValue = types.StringUnknown()
	} else {
		resp.PlanValue = types.StringNull()
	}
}

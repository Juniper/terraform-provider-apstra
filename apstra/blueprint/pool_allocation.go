package blueprint

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PoolAllocation struct {
	BlueprintId   types.String `tfsdk:"blueprint_id"`
	Role          types.String `tfsdk:"role"`
	PoolIds       types.Set    `tfsdk:"pool_ids"`
	RoutingZoneId types.String `tfsdk:"routing_zone_id"`
}

func (o PoolAllocation) ResourceAttributes() map[string]resourceSchema.Attribute {
	// roleSemanticEqualityFunc is a quick-and-dirty string plan modifier function intended to facilitate
	// migration from the following API strings to the "better" strings we'd like to present to terraform
	// users:
	// - ipv6_spine_leaf_link_ips       -> spine_leaf_link_ips_ipv6
	// - ipv6_spine_superspine_link_ips -> spine_superspine_link_ips_ipv6
	// - ipv6_to_generic_link_ips       -> to_generic_link_ips_ipv6
	roleSemanticEqualityFunc := func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		if req.PlanValue.Equal(req.StateValue) {
			// plan and state are equal -- no problem!
			resp.RequiresReplace = false
			return
		}

		var err error
		var plan, state apstra.ResourceGroupName

		// use two strategies when parsing the state value to apstra.ResourceGroupName
		err = state.FromString(req.StateValue.ValueString())
		if err != nil {
			err = utils.ApiStringerFromFriendlyString(&state, req.StateValue.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("failed to parse state value %q", req.StateValue.ValueString()), err.Error())
				return
			}
		}

		// use two strategies when parsing the plan value to apstra.ResourceGroupName
		err = plan.FromString(req.PlanValue.ValueString())
		if err != nil {
			err = utils.ApiStringerFromFriendlyString(&plan, req.PlanValue.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("failed to parse plan value %q", req.PlanValue.ValueString()), err.Error())
				return
			}
		}

		if plan != state {
			// plan and state have different IOTA values! This is a major configuration change. Rip-and-replace the resource.
			resp.RequiresReplace = true
			return
		}

		resp.RequiresReplace = false
	}

	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Blueprint to which the Resource Pool should be allocated.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"pool_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Apstra IDs of the Resource Pools to be allocated to the given Blueprint role.",
			Required:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"role": resourceSchema.StringAttribute{
			MarkdownDescription: "Fabric Role (Apstra Resource Group Name) must be one of:\n  - " +
				strings.Join(utils.ResourceGroupNameStrings(), "\n  - ") + "\n",
			Required: true,
			PlanModifiers: []planmodifier.String{
				//stringplanmodifier.RequiresReplace(),
				// RequiresReplace has been swapped for RequiresReplaceIf to support migration from old
				// role strings to new:
				// - ipv6_spine_leaf_link_ips       -> spine_leaf_link_ips_ipv6
				// - ipv6_spine_superspine_link_ips -> spine_superspine_link_ips_ipv6
				// - ipv6_to_generic_link_ips       -> to_generic_link_ips_ipv6
				//
				// This migration ability will have gone into effect around v0.63.0 in early August 2024.
				// We should probably keep it around for at least a year because needlessly removing/replacing
				// resource allocations is pretty disruptive.
				stringplanmodifier.RequiresReplaceIf(
					roleSemanticEqualityFunc,
					"permit nondisruptive migration from old API strings to new terraform strings",
					"permit nondisruptive migration from old API strings to new terraform strings",
				),
			},
			Validators: []validator.String{stringvalidator.OneOf(utils.ResourceGroupNameStrings()...)},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Used to allocate a Resource Pool to a "+
				"`role` associated with specific Routing Zone within a Blueprint, rather than "+
				"to a fabric-wide `role`. `%s` and `%s` are examples of roles which can be "+
				"allocaated to a specific Routing Zone. When omitted, the specified Resource "+
				"Pools are allocated to a fabric-wide `role`.",
				apstra.ResourceGroupNameLeafIp4, apstra.ResourceGroupNameVirtualNetworkSviIpv4),
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o *PoolAllocation) LoadApiData(ctx context.Context, in *apstra.ResourceGroupAllocation, diags *diag.Diagnostics) {
	o.PoolIds = utils.SetValueOrNull(ctx, types.StringType, in.PoolIds, diags)
}

func (o *PoolAllocation) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ResourceGroupAllocation {
	// Parse 'role' into a ResourceGroupName
	var rgName apstra.ResourceGroupName
	err := utils.ApiStringerFromFriendlyString(&rgName, o.Role.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing role %q", o.Role.ValueString()), err.Error())
		return nil
	}

	// extract pool IDs
	poolIds := make([]apstra.ObjectId, len(o.PoolIds.Elements()))
	diags.Append(o.PoolIds.ElementsAs(ctx, &poolIds, false)...)
	if diags.HasError() {
		return nil
	}

	rg := apstra.ResourceGroup{
		Type: rgName.Type(),
		Name: rgName,
	}

	if !o.RoutingZoneId.IsNull() {
		szId := apstra.ObjectId(o.RoutingZoneId.ValueString())
		rg.SecurityZoneId = &szId
	}

	return &apstra.ResourceGroupAllocation{
		ResourceGroup: rg,
		PoolIds:       poolIds,
	}
}

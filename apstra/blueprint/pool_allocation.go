package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/utils"
)

type PoolAllocation struct {
	BlueprintId   types.String `tfsdk:"blueprint_id"`
	Role          types.String `tfsdk:"role"`
	PoolIds       types.Set    `tfsdk:"pool_ids"`
	RoutingZoneId types.String `tfsdk:"routing_zone_id"`
	//PoolAllocationId types.String `tfsdk:"pool_allocation_id"` // placeholder for freeform
}

func (o PoolAllocation) ResourceAttributes() map[string]resourceSchema.Attribute {
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
			MarkdownDescription: "Fabric Role (Apstra Resource Group Name)",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{
				stringvalidator.OneOf(
					utils.AllResourceGroupNameStrings()...,
				),
			},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Used to allocate a Resource Pool to a "+
				"`role` associated with specific Routing Zone within a Blueprint, rather than "+
				"to a fabric-wide `role`. `%s` and `%s` are examples of roles which can be "+
				"allocaated to a specific Routing Zone. When omitted, the specified Resource "+
				"Pools are allocated to a fabric-wide `role`.",
				apstra.ResourceGroupNameLeafIp4, apstra.ResourceGroupNameVirtualNetworkSviIpv4),
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.AtMostNOf(1,
					path.Expressions{
						path.MatchRelative(),
						// other blueprint objects to which pools can be assigned must be listed here
						//path.MatchRoot("pool_allocation_id"), //placeholder for freeform
					}...,
				),
			},
		},
		// placeholder for freeform
		//"pool_allocation_id": resourceSchema.StringAttribute{
		//	MarkdownDescription: "",
		//	Optional:            true,
		//	Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		//},
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

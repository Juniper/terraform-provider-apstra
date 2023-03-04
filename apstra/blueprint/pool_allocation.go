package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-apstra/apstra/utils"
)

type PoolAllocation struct {
	BlueprintId types.String `tfsdk:"blueprint_id""`
	Role        types.String `tfsdk:"role"`
	PoolIds     types.Set    `tfsdk:"pool_ids"`
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
	}
}

func (o *PoolAllocation) LoadApiData(ctx context.Context, in *goapstra.ResourceGroupAllocation, diags *diag.Diagnostics) {
	o.PoolIds = utils.SetValueOrNull(ctx, types.StringType, in.PoolIds, diags)
}

func (o *PoolAllocation) Validate(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	// Ensure the configured blueprint ID exists on Apstra
	if !o.BlueprintId.IsUnknown() {
		_, err := client.GetBlueprintStatus(ctx, goapstra.ObjectId(o.BlueprintId.ValueString()))
		if err != nil {
			diags.AddError(
				fmt.Sprintf("Error retrieving blueprint %q", o.BlueprintId.ValueString()),
				err.Error())
		}
	}

	if !o.Role.IsUnknown() {
		// Extract role to ResourceGroupName
		var rgName goapstra.ResourceGroupName
		err := rgName.FromString(o.Role.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing role %q", o.Role.ValueString()),
				err.Error())
			return
		}

		// Get list of poolIds from Apstra
		var apiPoolIds []goapstra.ObjectId
		switch rgName.Type() {
		case goapstra.ResourceTypeAsnPool:
			apiPoolIds, err = client.ListAsnPoolIds(ctx)
		case goapstra.ResourceTypeIp4Pool:
			apiPoolIds, err = client.ListIp4PoolIds(ctx)
		case goapstra.ResourceTypeIp6Pool:
			apiPoolIds, err = client.ListIp6PoolIds(ctx)
		case goapstra.ResourceTypeVniPool:
			apiPoolIds, err = client.ListVniPoolIds(ctx)
		default:
			diags.AddError("error determining Resource Group Type by Name",
				fmt.Sprintf("Resource Group %q not recognized", o.Role.ValueString()))
		}
		if err != nil {
			diags.AddError("error listing pool IDs", err.Error())
		}
		if diags.HasError() {
			return
		}

		// Quick function to check for 'id' among 'ids'
		contains := func(ids []goapstra.ObjectId, id goapstra.ObjectId) bool {
			for i := range ids {
				if ids[i] == id {
					return true
				}
			}
			return false
		}

		// Check that each PoolId configuration element appears in the API results
		for _, elem := range o.PoolIds.Elements() {
			id := elem.(basetypes.StringValue).ValueString()
			if !contains(apiPoolIds, goapstra.ObjectId(id)) {
				diags.AddError(
					"pool not found",
					fmt.Sprintf("pool id %q of type %q not found", id, rgName.Type().String()))
				return
			}
		}
	}
}

func (o *PoolAllocation) Request(ctx context.Context, diags *diag.Diagnostics) *goapstra.ResourceGroupAllocation {
	// Parse 'role' into a ResourceGroupName
	var rgName goapstra.ResourceGroupName
	err := rgName.FromString(o.Role.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing role %q", o.Role.ValueString()), err.Error())
		return nil
	}

	// extract pool IDs
	poolIds := make([]goapstra.ObjectId, len(o.PoolIds.Elements()))
	diags.Append(o.PoolIds.ElementsAs(ctx, &poolIds, false)...)
	if diags.HasError() {
		return nil
	}

	return &goapstra.ResourceGroupAllocation{
		ResourceGroup: goapstra.ResourceGroup{
			Type: rgName.Type(),
			Name: rgName,
		},
		PoolIds: poolIds,
	}
}

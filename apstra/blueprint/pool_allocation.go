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
	"terraform-provider-apstra/apstra/utils"
)

type PoolAllocation struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
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

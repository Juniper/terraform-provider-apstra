package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

type PoolAllocation struct {
	BlueprintId types.String `tfsdk:"blueprint_id""`
	Role        types.String `tfsdk:"role"`
	PoolId      types.String `tfsdk:"pool_id"`
}

func (o PoolAllocation) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Blueprint to which the Resource Pool should be allocated.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"pool_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Resource Pool to be allocated.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"role": resourceSchema.StringAttribute{
			MarkdownDescription: "",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					utils.AllResourceGroupNameStrings()...,
				),
			},
		},
	}
}

func (o *PoolAllocation) Validate(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	_, err := client.GetBlueprintStatus(ctx, goapstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error retrieving blueprint %q", o.BlueprintId.ValueString()),
			err.Error())
	}

	var rgName goapstra.ResourceGroupName
	err = rgName.FromString(o.Role.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing role %q", o.Role.ValueString()),
			err.Error())
		return
	}

	switch rgName.Type() {
	case goapstra.ResourceTypeAsnPool:
		_, err = client.GetAsnPool(ctx, goapstra.ObjectId(o.PoolId.ValueString()))
	case goapstra.ResourceTypeIp4Pool:
		_, err = client.GetIp4Pool(ctx, goapstra.ObjectId(o.PoolId.ValueString()))
	case goapstra.ResourceTypeIp6Pool:
		_, err = client.GetIp6Pool(ctx, goapstra.ObjectId(o.PoolId.ValueString()))
	case goapstra.ResourceTypeVniPool:
		_, err = client.GetVniPool(ctx, goapstra.ObjectId(o.PoolId.ValueString()))
	}
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error retrieving %q pool %q", rgName.Type().String(), o.PoolId.ValueString()),
			err.Error())
	}
}

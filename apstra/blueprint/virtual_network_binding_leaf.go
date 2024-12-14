package blueprint

import (
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualNetworkBindingLeaf struct {
	LeafId    customtypes.StringWithAltValues `tfsdk:"leaf_id"`
	VlanId    types.Int64                     `tfsdk:"vlan_id"`
	AccessIds customtypes.StringWithAltValues `tfsdk:"access_ids"`
}

func (o VirtualNetworkBindingLeaf) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"leaf_id":    customtypes.StringWithAltValuesType{},
		"vlan_id":    types.Int64Type,
		"access_ids": types.SetType{ElemType: customtypes.StringWithAltValuesType{}},
	}
}

func (o VirtualNetworkBindingLeaf) ResourceAttributes() map[string]resourceSchema.Attribute {
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
			Optional:            true,
			ElementType:         customtypes.StringWithAltValuesType{},
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
	}
}

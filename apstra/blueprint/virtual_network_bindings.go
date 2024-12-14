package blueprint

import (
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualNetworkBinding struct {
	BlueprintId      types.String `tfsdk:"blueprint_id"`
	VirtualNetworkId types.String `tfsdk:"virtual_network_id"`
	Bindings         types.Set    `tfsdk:"bindings"`
}

func (o VirtualNetworkBinding) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"blueprint_id":       types.StringType,
		"virtual_network_id": types.StringType,
		"bindings":           types.SetType{ElemType: types.ObjectType{AttrTypes: VirtualNetworkBindingLeaf{}.AttrTypes()}},
	}
}

func (o VirtualNetworkBinding) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"virtual_network_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Virtual Network ID.",
			CustomType:          customtypes.StringWithAltValuesType{},
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"bindings": resourceSchema.SetAttribute{
			MarkdownDescription: "Assignment info for each Leaf Switch and any downstream Access Switches",
			Optional:            true,
			ElementType:         customtypes.StringWithAltValuesType{},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

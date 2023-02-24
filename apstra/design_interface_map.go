package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type interfaceMap struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	LogicalDevice types.String `tfsdk:"logical_device_id"`
	DeviceProfile types.String `tfsdk:"device_profile_id"`
	Interfaces    types.Set    `tfsdk:"interfaces"`
}

func (o interfaceMap) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Interface Map ID.  Required when the Interface Map name is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Interface Map name displayed in the Apstra web UI.  Required when Interface Map ID is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of Logical Device referenced by this interface map.",
			Computed:            true,
		},
		"device_profile_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of Device Profile referenced by this interface map.",
			Computed:            true,
		},
		"interfaces": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Detailed mapping of each physical interface to its role in the logical device",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: interfaceMapInterface{}.dataSourceSchema(),
			},
		},
	}
}

func (o *interfaceMap) loadApiData(ctx context.Context, in *goapstra.InterfaceMapData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	o.LogicalDevice = types.StringValue(string(in.LogicalDeviceId))
	o.DeviceProfile = types.StringValue(string(in.DeviceProfileId))
	o.Interfaces = newInterfaceMapInterfaceSet(ctx, in.Interfaces, diags)
}

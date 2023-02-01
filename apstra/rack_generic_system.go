package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func genericSystemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Generic name, must be unique within the rack-type.",
			Computed:            true,
		},
		"count": schema.Int64Attribute{
			MarkdownDescription: "Number of Generic Systems of this type.",
			Computed:            true,
		},
		"port_channel_id_min": schema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Computed:            true,
		},
		"port_channel_id_max": schema.Int64Attribute{
			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
			Computed:            true,
		},
		"logical_device": logicalDeviceData{}.schema(),
		"tag_data":       tagsDataAttributeSchema(),
		"links":          dLinksAttributeSchema(),
	}
}

func validateGenericSystem(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	gs := rt.Data.GenericSystems[i]
	if gs.LogicalDevice == nil {
		diags.AddError("generic system logical device info missing",
			fmt.Sprintf("rack type '%s', generic system '%s' logical device is nil",
				rt.Id, gs.Label))
	}
}

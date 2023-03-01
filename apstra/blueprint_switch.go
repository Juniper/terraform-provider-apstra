package apstra

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type blueprintSwitch struct {
	DeviceKey       types.String `tfsdk:"device_key"`
	InterfaceMapId  types.String `tfsdk:"interface_map_id"`
	DeviceProfileId types.String `tfsdk:"device_profile_id"`
	SystemNodeId    types.String `tfsdk:"system_node_id"`
}

func (o blueprintSwitch) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"device_key": schema.StringAttribute{
			MarkdownDescription: "Unique ID for a device, generally the serial number.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"interface_map_id": schema.StringAttribute{
			MarkdownDescription: "links a Logical Device (design element) to a Device Profile which" +
				"describes a hardware model. Optional when only a single interface map references the " +
				"logical device underpinning the node in question.",
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"device_profile_id": schema.StringAttribute{
			MarkdownDescription: "Device Profile is selected by the Interface Map associated with this switch.",
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"system_node_id": schema.StringAttribute{
			MarkdownDescription: "ID number of the blueprint graphdb node representing this system.",
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o blueprintSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"device_key":        types.StringType,
		"interface_map_id":  types.StringType,
		"device_profile_id": types.StringType,
		"system_node_id":    types.StringType,
	}
}

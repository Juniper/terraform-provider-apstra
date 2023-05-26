package systemAgents

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SystemAck struct {
	AgentId   types.String `tfsdk:"agent_id"`
	DeviceKey types.String `tfsdk:"device_key"`
	SystemId  types.String `tfsdk:"system_id"`
}

func (o SystemAck) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"agent_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID for the System Agent responsible for the Managed Device.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"device_key": resourceSchema.StringAttribute{
			MarkdownDescription: "Key which uniquely identifies a System asset, probably the serial number.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"system_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID for the System discovered by the System Agent.",
			Computed:            true,
		},
	}
}

func (o *SystemAck) Acknowledge(ctx context.Context, si *apstra.ManagedSystemInfo, client *apstra.Client, diags *diag.Diagnostics) {
	// update with new SystemUserConfig
	err := client.UpdateSystem(ctx, apstra.SystemId(o.SystemId.ValueString()), &apstra.SystemUserConfig{
		AosHclModel: si.Facts.AosHclModel,
		AdminState:  apstra.SystemAdminStateNormal,
	})
	if err != nil {
		diags.AddError(
			"error updating managed device",
			fmt.Sprintf("unexpected error while updating user config: %s", err.Error()),
		)
	}
}

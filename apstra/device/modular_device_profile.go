package device

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

type ModularDeviceProfile struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ChassisProfileId   types.String `tfsdk:"chassis_profile_id"`
	LineCardProfileIds types.Map    `tfsdk:"line_card_profile_ids"`
}

func (o ModularDeviceProfile) ResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Apstra Object ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name displayed in web UI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"chassis_profile_id": schema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of the Chassis Device Profile.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"line_card_profile_ids": schema.MapAttribute{
			MarkdownDescription: "A map of Line Card Device Profile IDs, keyed by slot number",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			ElementType:         types.StringType,
		},
	}
}

func (o *ModularDeviceProfile) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.ModularDeviceProfile {
	var slotIds map[string]apstra.ObjectId
	diags.Append(o.LineCardProfileIds.ElementsAs(ctx, &slotIds, false)...)
	if diags.HasError() {
		return nil
	}

	slotConfigurations := make(map[uint64]apstra.ModularDeviceSlotConfiguration, len(slotIds))
	for k, v := range slotIds {
		slotNum, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			panic(err) // the config should have already been validated - this should never happen
		}

		slotConfigurations[slotNum] = apstra.ModularDeviceSlotConfiguration{LinecardProfileId: v}
	}

	return &apstra.ModularDeviceProfile{
		Label:              o.Name.ValueString(),
		ChassisProfileId:   apstra.ObjectId(o.ChassisProfileId.ValueString()),
		SlotConfigurations: slotConfigurations,
	}
}

func (o *ModularDeviceProfile) LoadApiData(_ context.Context, in *apstra.ModularDeviceProfile, _ *diag.Diagnostics) {
	lineCardProfileIds := make(map[string]attr.Value, len(in.SlotConfigurations))
	for k, v := range in.SlotConfigurations {
		lineCardProfileIds[strconv.FormatUint(k, 10)] = types.StringValue(v.LinecardProfileId.String())
	}

	o.Name = types.StringValue(in.Label)
	o.ChassisProfileId = types.StringValue(in.ChassisProfileId.String())
	o.LineCardProfileIds = types.MapValueMust(types.StringType, lineCardProfileIds)
}

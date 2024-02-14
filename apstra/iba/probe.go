package iba

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Probe struct {
	BlueprintId       types.String `tfsdk:"blueprint_id"`
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	PredefinedProbeId types.String `tfsdk:"predefined_probe_id"`
	ProbeConfig       types.String `tfsdk:"probe_config"`
	Stages            types.Set    `tfsdk:"stages"`
	ProbeJson         types.String `tfsdk:"probe_json"`
}

func (o Probe) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify the Blueprint that the IBA Probe belongs to.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "IBA Probe ID.",
			Computed:            true,
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "IBA Probe Name.",
			Computed:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Description of the IBA Probe",
			Computed:            true,
		},
		"stages": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of names of stages in the IBA Probe",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"predefined_probe_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Id of predefined IBA Probe",
			Optional:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"probe_config": resourceSchema.StringAttribute{
			MarkdownDescription: "Configuration elements for the IBA Probe",
			Optional:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{stringvalidator.AlsoRequires(path.MatchRelative().AtParent().
				AtName("predefined_probe_id"))},
		},
		"probe_json": resourceSchema.StringAttribute{
			MarkdownDescription: "Define the probe as json. If this is present, there can be no predefined probe.",
			Optional:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{stringvalidator.ConflictsWith(path.MatchRelative().AtParent().
				AtName("predefined_probe_id"))},
		},
	}
}

func (o *Probe) LoadApiData(ctx context.Context, in *apstra.IbaProbe, diag *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Label)
	o.Description = types.StringValue(in.Description)
	s := make([]string, len(in.Stages))
	for i, j := range in.Stages {
		s[i] = j["name"].(string)
	}
	o.Stages = utils.SetValueOrNull(ctx, types.StringType, s, diag)
}

func (o *Probe) PredefinedProbeRequest(ctx context.Context, d *diag.Diagnostics) *apstra.IbaPredefinedProbeRequest {

	return &apstra.IbaPredefinedProbeRequest{
		Name: o.PredefinedProbeId.ValueString(),
		Data: []byte(o.ProbeConfig.ValueString()),
	}
}

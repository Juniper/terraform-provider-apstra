package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

type Deploy struct {
	BlueprintId           types.String `tfsdk:"blueprint_id"`
	Comment               types.String `tfsdk:"comment"`
	HasUncommittedChanges types.Bool   `tfsdk:"has_uncommitted_changes"`
	ActiveRevision        types.Int64  `tfsdk:"revision_active"`
	StageDRevision        types.Int64  `tfsdk:"revision_staged"`
}

func (o Deploy) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "", // todo
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"comment": dataSourceSchema.StringAttribute{
			MarkdownDescription: "", // todo
			Computed:            true,
		},
		"has_uncommitted_changes": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "", // todo
			Computed:            true,
		},
		"revision_active": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "", // todo
			Computed:            true,
		},
		"revision_staged": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "", // todo
			Computed:            true,
		},
	}
}

func (o Deploy) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "", // todo
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"comment": resourceSchema.StringAttribute{
			MarkdownDescription: "", // todo
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"has_uncommitted_changes": resourceSchema.BoolAttribute{
			MarkdownDescription: "", // todo
			Computed:            true,
		},
		"revision_active": resourceSchema.Int64Attribute{
			MarkdownDescription: "", // todo
			Computed:            true,
		},
		"revision_staged": resourceSchema.Int64Attribute{
			MarkdownDescription: "", // todo
			Computed:            true,
		},
	}
}

func (o *Deploy) Deploy(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	status, err := client.GetBlueprintStatus(ctx, goapstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		diags.AddError("error getting Blueprint status", err.Error())
		return
	}

	if status.BuildErrorsCount > 0 {
		diags.AddError("Blueprint has build errors",
			fmt.Sprintf("%d build errors must be resolved", status.BuildErrorsCount))
		return
	}

	if status.BuildWarningsCount > 0 {
		diags.AddWarning("Blueprint has build warnings",
			fmt.Sprintf("%d build warnings must be resolved", status.BuildWarningsCount))

	}

	if !status.HasUncommittedChanges {
		diags.AddWarning("no uncommitted changes",
			fmt.Sprintf(
				"deploy of Blueprint %q requested but current revision %d has no uncommitted changes",
				o.BlueprintId.ValueString(), status.Version))
		return
	}

	response, err := client.DeployBlueprint(ctx, &goapstra.BlueprintDeployRequest{
		Id:          goapstra.ObjectId(o.BlueprintId.ValueString()),
		Description: o.Comment.ValueString(),
		Version:     status.Version,
	})
	if err != nil {
		diags.AddError("error deploying Blueprint", err.Error())
		return
	}
	if response.Error != nil {
		diags.AddError(
			fmt.Sprintf("blueprint deployment: status %q", response.Status.String()),
			*response.Error,
		)
		return
	}
	if response.Status != goapstra.DeployStatusSuccess {
		diags.AddError(
			"blueprint deploy status",
			fmt.Sprintf("status: %q", response.Status.String()),
		)
		return
	}

	o.ActiveRevision = types.Int64Value(int64(response.Version))
	o.StageDRevision = types.Int64Value(int64(response.Version))
	o.HasUncommittedChanges = types.BoolValue(false)
}

func (o *Deploy) Read(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	bpId := goapstra.ObjectId(o.BlueprintId.ValueString())
	status, err := client.GetBlueprintStatus(ctx, bpId)
	if err != nil {
		diags.AddError("error getting Blueprint status", err.Error())
		return
	}

	if status.BuildErrorsCount > 0 {
		diags.AddWarning("Blueprint has build errors",
			fmt.Sprintf("%d build errors must be resolved", status.BuildErrorsCount))
	}

	if status.BuildWarningsCount > 0 {
		diags.AddWarning("Blueprint has build warnings",
			fmt.Sprintf("%d build warnings must be resolved", status.BuildWarningsCount))

	}

	o.HasUncommittedChanges = types.BoolValue(status.HasUncommittedChanges)

	//if o.Comment.IsUnknown() {
	revision, err := client.GetLastDeployedRevision(ctx, bpId)
	if err != nil {
		diags.AddWarning(
			fmt.Sprintf("error reading blueprint %q revision %d", bpId, status.Version),
			err.Error(),
		)
	}
	if revision == nil {
		return
	}
	//if revision == nil {
	//	o.Comment = types.StringNull()
	//	o.Revision = types.Int64Null()
	//} else {
	o.Comment = utils.StringValueOrNull(ctx, revision.Description, diags)
	o.ActiveRevision = types.Int64Value(int64(revision.RevisionId))
	o.StageDRevision = types.Int64Value(int64(status.Version))
	//}
	//}
}

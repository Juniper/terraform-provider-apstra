package blueprint

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"terraform-provider-apstra/apstra/utils"
	"text/template"
)

type Deploy struct {
	BlueprintId           types.String `tfsdk:"blueprint_id"`
	Comment               types.String `tfsdk:"comment"`
	HasUncommittedChanges types.Bool   `tfsdk:"has_uncommitted_changes"`
	ActiveRevision        types.Int64  `tfsdk:"revision_active"`
	StagedRevision        types.Int64  `tfsdk:"revision_staged"`
}

func (o Deploy) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the blueprint.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"comment": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Comment associated with the Deployment/Commit.",
			Computed:            true,
		},
		"has_uncommitted_changes": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "True when there are uncommitted changes in the staging Blueprint.",
			Computed:            true,
		},
		"revision_active": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Revision numbers increment with each Blueprint change. This is " +
				"the currently deployed revision number.",
			Computed: true,
		},
		"revision_staged": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Revision numbers increment with each Blueprint change. This is " +
				"the revision number currently in staging.",
			Computed: true,
		},
	}
}

func (o Deploy) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the blueprint.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"comment": resourceSchema.StringAttribute{
			MarkdownDescription: "Comment associated with the Deployment/Commit. This field supports templating " +
				"using the `text/template` library (currently supported replacements: ['Version']) and " +
				"environment variable expansion using `os.ExpandEnv` to include contextual information like the " +
				"Terraform username, CI system job ID, etc...",
			Computed:   true,
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			Default:    stringdefault.StaticString("Terraform {{.TerraformVersion}}, Apstra provider {{.ProviderVersion}}, User $USER."),
		},
		"has_uncommitted_changes": resourceSchema.BoolAttribute{
			MarkdownDescription: "True when there are uncommited changes in the staging Blueprint.",
			Computed:            true,
		},
		"revision_active": resourceSchema.Int64Attribute{
			MarkdownDescription: "Revision numbers increment with each Blueprint change. This is " +
				"the currently deployed revision number.",
			Computed: true,
		},
		"revision_staged": resourceSchema.Int64Attribute{
			MarkdownDescription: "Revision numbers increment with each Blueprint change. This is " +
				"the revision number currently in staging.",
			Computed: true,
		},
	}
}

func (o *Deploy) Deploy(ctx context.Context, commentTemplate *CommentTemplate, client *apstra.Client, diags *diag.Diagnostics) {
	status, err := client.GetBlueprintStatus(ctx, apstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		diags.AddError("error getting Blueprint status", err.Error())
		return
	}

	if status.BuildErrorsCount > 0 {
		diags.AddError("Blueprint has build errors",
			fmt.Sprintf("Blueprint has %d build errors which must be resolved prior to deployment", status.BuildErrorsCount))
		return
	}

	if status.BuildWarningsCount > 0 {
		diags.AddWarning("Blueprint has build warnings",
			fmt.Sprintf("Blueprint has %d build warnings, but deployment may proceed", status.BuildWarningsCount))

	}

	if !status.HasUncommittedChanges {
		diags.AddWarning("no uncommitted changes",
			fmt.Sprintf(
				"deploy of Blueprint %q requested but current revision %d has no uncommitted changes",
				o.BlueprintId.ValueString(), status.Version))
		o.HasUncommittedChanges = types.BoolValue(status.HasUncommittedChanges)
		o.ActiveRevision = types.Int64Value(int64(status.Version))
		o.StagedRevision = types.Int64Value(int64(status.Version))
		return
	}

	t, err := new(template.Template).Parse(o.Comment.ValueString())
	if err != nil {
		diags.AddWarning(
			fmt.Sprintf("error creating deployment comment template from string %q", o.Comment.ValueString()),
			err.Error())
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, commentTemplate)
	if err != nil {
		diags.AddWarning("error executing deployment comment template", err.Error())
	}

	response, err := client.DeployBlueprint(ctx, &apstra.BlueprintDeployRequest{
		Id:          apstra.ObjectId(o.BlueprintId.ValueString()),
		Description: os.ExpandEnv(buf.String()),
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
	if response.Status != apstra.DeployStatusSuccess {
		diags.AddError(
			"blueprint deploy status",
			fmt.Sprintf("status: %q", response.Status.String()),
		)
		return
	}

	o.ActiveRevision = types.Int64Value(int64(response.Version))
	o.StagedRevision = types.Int64Value(int64(response.Version))
	o.HasUncommittedChanges = types.BoolValue(false)
}

func (o *Deploy) Read(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	bpId := apstra.ObjectId(o.BlueprintId.ValueString())
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

	revision, err := client.GetLastDeployedRevision(ctx, bpId)
	if err != nil {
		var ace apstra.ApstraClientErr
		if !(errors.As(err, &ace) && ace.Type() == apstra.ErrUncommitted) {
			diags.AddError(
				fmt.Sprintf("failed reading blueprint %q revision", bpId),
				err.Error(),
			)
			return
		}

		// instantiate bogus revision because the API doesn't have one.
		// zero values okay.
		revision = &apstra.BlueprintRevision{}
	}

	o.Comment = utils.StringValueOrNull(ctx, revision.Description, diags)
	o.ActiveRevision = types.Int64Value(int64(revision.RevisionId))
	o.StagedRevision = types.Int64Value(int64(status.Version))
}

type CommentTemplate struct {
	ProviderVersion  string
	TerraformVersion string
}

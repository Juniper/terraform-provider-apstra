package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"terraform-provider-apstra/apstra/blueprint"
)

var _ resource.ResourceWithConfigure = &resourceBlueprintCommit{}

type resourceBlueprintCommit struct {
	client *goapstra.Client
}

func (o *resourceBlueprintCommit) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_commit"
}

func (o *resourceBlueprintCommit) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceBlueprintCommit) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource commits a staging Blueprint after checking for build errors.",
		Attributes:          blueprint.Commit{}.ResourceAttributes(),
	}
}

func (o *resourceBlueprintCommit) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.Commit
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := o.client.GetBlueprintStatus(ctx, goapstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error getting Blueprint status", err.Error())
	}

	if status.BuildErrorsCount > 0 {
		resp.Diagnostics.AddError("cannot commit Blueprint",
			fmt.Sprintf("Blueprint %q has %d build errors which must be resolved",
				plan.BlueprintId.ValueString(), status.BuildErrorsCount))
		return
	}

	//if status.BuildWarningsCount > 0 {
	//	resp.Diagnostics.AddWarning("cannot commit Blueprint",
	//		fmt.Sprintf("Blueprint %q has %d build errors which must be resolved",
	//			plan.BlueprintId.ValueString(), status.BuildErrorsCount))
	//	return
	//}

	dump, _ := json.MarshalIndent(&status, "", "  ")
	resp.Diagnostics.AddWarning("status", string(dump))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceBlueprintCommit) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//TODO implement me
}

func (o *resourceBlueprintCommit) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//TODO implement me
}

func (o *resourceBlueprintCommit) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//TODO implement me
}

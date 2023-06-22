package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"terraform-provider-apstra/apstra/blueprint"
	"terraform-provider-apstra/apstra/utils"
)

var _ resource.ResourceWithConfigure = &resourceDatacenterPropertySet{}

type resourceDatacenterPropertySet struct {
	client   *apstra.Client
	lockFunc func(context.Context, string) error
}

func (o *resourceDatacenterPropertySet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_property_set"
}

func (o *resourceDatacenterPropertySet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
	o.lockFunc = ResourceGetBlueprintLockFunc(ctx, req, resp)
}

func (o *resourceDatacenterPropertySet) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource imports a property set into a Blueprint.",
		Attributes:          blueprint.DatacenterPropertySet{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterPropertySet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.DatacenterPropertySet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// extract the keys to be imported
	var keysToImport []string
	resp.Diagnostics.Append(plan.Keys.ElementsAs(ctx, &keysToImport, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := bpClient.ImportPropertySet(ctx, apstra.ObjectId(plan.Id.ValueString()), keysToImport...)
	if err != nil {
		resp.Diagnostics.AddError("Error importing DatacenterPropertySet", err.Error())
		return
	}

	// check our assumption that the ID returned from an import call matches the
	// ID of the imported Property Set because this feels like something which
	// might change.
	if id.String() != plan.Id.ValueString() {
		resp.Diagnostics.AddWarning("provider bug",
			fmt.Sprintf("when importing Property Set %s imported into Blueprint %s, API returned unexpected ID %q",
				plan.Id, plan.BlueprintId, id))
		// we probably don't need to return here
	}

	// Read it back
	api, err := bpClient.GetPropertySet(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed reading just-imported Property Set %s from Blueprint %s",
			plan.Id, plan.BlueprintId), err.Error())
		return
	}

	// create new state object
	var state blueprint.DatacenterPropertySet
	state.BlueprintId = plan.BlueprintId
	state.LoadApiData(ctx, api, &resp.Diagnostics) // this

	// extract keys which actually got imported
	var importedKeys []string
	resp.Diagnostics.Append(plan.Keys.ElementsAs(ctx, &importedKeys, false)...)
	if resp.Diagnostics.HasError() {
		// set the state prior to returning because the PS has been imported
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	// keysToImport and importedKeys should match...
	extraImportedKeys, failedImportedKeys := utils.DiffSliceSets(keysToImport, importedKeys)
	if len(extraImportedKeys) != 0 {
		resp.Diagnostics.AddWarning(
			fmt.Sprintf("import of PropertySet %s produced unexpected Keys", plan.Id),
			fmt.Sprintf("extra Keys: %v", extraImportedKeys))
		// do not return without setting state
	}

	if len(failedImportedKeys) != 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("keys"),
			fmt.Sprintf("failed to import all desired Keys from PropertySet %s", plan.Id),
			fmt.Sprintf("the following Keys could not be imported: %v", failedImportedKeys),
		)
		// do not return without setting state
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterPropertySet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	// Retrieve values from state
	var state blueprint.DatacenterPropertySet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to crreate blueprint client", err.Error())
	}

	api, err := bpClient.GetPropertySet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddAttributeError(path.Root("name"),
			fmt.Sprintf("Failed to read imported PropertySet %s", state.Id), err.Error())
		return
	}

	// create new state object
	var newState blueprint.DatacenterPropertySet
	newState.LoadApiData(ctx, api, &resp.Diagnostics)
	newState.BlueprintId = state.BlueprintId
	// If the user uses a blank set of keys, we are importing everything, so we do not want to update the list.
	newState.Keys = state.Keys

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceDatacenterPropertySet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve values from plan
	var plan blueprint.DatacenterPropertySet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the blueprint mutex.
	err := o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error locking blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(plan.BlueprintId.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error creating blueprint client", err.Error())
		return
	}

	keys := make([]string, len(plan.Keys.Elements()))
	resp.Diagnostics.Append(plan.Keys.ElementsAs(ctx, &keys, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Property Set
	err = bpClient.UpdatePropertySet(ctx, apstra.ObjectId(plan.Id.ValueString()), keys...)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error updating Blueprint %s Property Set %s", plan.BlueprintId, plan.Id),
			err.Error())
		return
	}

	api, err := bpClient.GetPropertySet(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failure reading just-updated Property Set %s in Blueprint %s",
				plan.Id, plan.BlueprintId),
			err.Error())
		return
	}

	// create new state object
	var state blueprint.DatacenterPropertySet
	state.LoadApiData(ctx, api, &resp.Diagnostics)
	state.BlueprintId = plan.BlueprintId
	state.Keys = plan.Keys

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterPropertySet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	// Retrieve values from state
	var state blueprint.DatacenterPropertySet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bpClient, err := o.client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(state.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("unable to get blueprint client", err.Error())
		return
	}

	// Delete Property Set by calling API
	err = bpClient.DeletePropertySet(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("unable to delete Property Set %s from Blueprint %s", state.Id, state.BlueprintId),
			err.Error())
	}
}

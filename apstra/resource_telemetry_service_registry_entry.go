package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/analytics"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceTelemetryServiceRegistryEntry{}
var _ resourceWithSetClient = &resourceTelemetryServiceRegistryEntry{}

type resourceTelemetryServiceRegistryEntry struct {
	client *apstra.Client
}

func (o *resourceTelemetryServiceRegistryEntry) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_telemetry_service_registry_entry"
}

func (o *resourceTelemetryServiceRegistryEntry) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceTelemetryServiceRegistryEntry) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This resource creates a Property Set in the Apstra Design tab.",
		Attributes:          analytics.TelemetryServiceRegistryEntry{}.ResourceAttributes(),
	}
}

func (o *resourceTelemetryServiceRegistryEntry) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan analytics.TelemetryServiceRegistryEntry
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the plan into an API Request
	tsRequest := plan.Request(ctx, &resp.Diagnostics)
	// Created Registry Entry will not be builtin
	plan.Builtin = types.BoolValue(false)
	if plan.Version.IsUnknown() {
		plan.Version = types.StringNull()
	}
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := o.client.CreateTelemetryServiceRegistryEntry(ctx, tsRequest)
	if err != nil {
		resp.Diagnostics.AddError("error creating new TelemetryServiceRegistryEntry", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceTelemetryServiceRegistryEntry) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state analytics.TelemetryServiceRegistryEntry
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *apstra.TelemetryServiceRegistryEntry

	api, err = o.client.GetTelemetryServiceRegistryEntry(ctx, state.ServiceName.ValueString())
	if utils.IsApstra404(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving TelemetryServiceRegistryEntry", err.Error())
		return
	}

	// create new state object
	var newstate analytics.TelemetryServiceRegistryEntry
	newstate.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if utils.JSONEqual(newstate.ApplicationSchema, state.ApplicationSchema, &resp.Diagnostics) {
		newstate.ApplicationSchema = state.ApplicationSchema
	}
	if resp.Diagnostics.HasError() {
		return
	}
	// Set state
	//
	resp.Diagnostics.Append(resp.State.Set(ctx, &newstate)...)
}

// Update resource
func (o *resourceTelemetryServiceRegistryEntry) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var cfg, plan analytics.TelemetryServiceRegistryEntry
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.ServiceName != plan.ServiceName {
		resp.Diagnostics.AddError("Cannot Change Service Name", fmt.Sprintf("Was %s New Value %s", cfg.ServiceName.ValueString(), plan.ServiceName.ValueString()))
		return
	}
	tsReq := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Property Set
	err := o.client.UpdateTelemetryServiceRegistryEntry(ctx, plan.ServiceName.ValueString(), tsReq)
	if err != nil {
		resp.Diagnostics.AddError("error updating Property Set", err.Error())
		return
	}

	// read the state fromm the API
	api, err := o.client.GetTelemetryServiceRegistryEntry(ctx, plan.ServiceName.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"TelemetryServiceRegistryEntry not found",
				fmt.Sprintf("just-updated TelemetryServiceRegistryEntry with Service Name %q not found.", plan.ServiceName.ValueString()))
			return
		}
		resp.Diagnostics.AddError("error  updating Telemetry Service Registry Entry set", err.Error())
		return
	}

	// Do the Get dance so that in case the application schema is stored differently the next
	// Read() does not confuse terraform
	as := plan.ApplicationSchema
	plan.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if utils.JSONEqual(plan.ApplicationSchema, as, &resp.Diagnostics) {
		plan.ApplicationSchema = as
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceTelemetryServiceRegistryEntry) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state analytics.TelemetryServiceRegistryEntry
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Property Set by calling API
	err := o.client.DeleteTelemetryServiceRegistryEntry(ctx, state.ServiceName.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Property Set", err.Error())
		return
	}
}

func (o *resourceTelemetryServiceRegistryEntry) setClient(client *apstra.Client) {
	o.client = client
}

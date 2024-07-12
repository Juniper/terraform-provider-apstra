package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/analytics"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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
		MarkdownDescription: docCategoryDesign + "This resource creates a Telemetry Service Registry Entry.",
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
	if resp.Diagnostics.HasError() {
		return
	}

	// Created Registry Entry will not be builtin
	plan.Builtin = types.BoolValue(false)
	if plan.Version.IsUnknown() {
		plan.Version = types.StringValue("Version_0")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := o.client.CreateTelemetryServiceRegistryEntry(ctx, tsRequest)
	if err != nil {
		resp.Diagnostics.AddError("error creating new TelemetryServiceRegistryEntry", err.Error())
		return
	}

	api, err := o.client.GetTelemetryServiceRegistryEntry(ctx, plan.ServiceName.ValueString())
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving TelemetryServiceRegistryEntry", err.Error())
		return
	}

	plan.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
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

	api, err := o.client.GetTelemetryServiceRegistryEntry(ctx, state.ServiceName.ValueString())
	if utils.IsApstra404(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving TelemetryServiceRegistryEntry", err.Error())
		return
	}

	state.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceTelemetryServiceRegistryEntry) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan analytics.TelemetryServiceRegistryEntry
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tsReq := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.UpdateTelemetryServiceRegistryEntry(ctx, plan.ServiceName.ValueString(), tsReq)
	if err != nil {
		resp.Diagnostics.AddError("error updating Telemetry Service Registry Entry", err.Error())
		return
	}

	api, err := o.client.GetTelemetryServiceRegistryEntry(ctx, plan.ServiceName.ValueString())
	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving TelemetryServiceRegistryEntry", err.Error())
		return
	}

	plan.LoadApiData(ctx, api, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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

	err := o.client.DeleteTelemetryServiceRegistryEntry(ctx, state.ServiceName.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Telemetry Service Registry Entry", err.Error())
		return
	}
}

func (o *resourceTelemetryServiceRegistryEntry) setClient(client *apstra.Client) {
	o.client = client
}

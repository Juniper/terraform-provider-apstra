package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceRackType{}
var _ resource.ResourceWithValidateConfig = &resourceRackType{}

type resourceRackType struct {
	client *apstra.Client
}

func (o *resourceRackType) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack_type"
}

func (o *resourceRackType) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceRackType) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a Rack Type in the Apstra Design tab.",
		Attributes:          design.RackType{}.ResourceAttributes(),
	}
}

func (o *resourceRackType) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Retrieve values from config
	var config design.RackType
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// access switches must have a value
	if config.AccessSwitches.IsUnknown() {
		return // cannot proceed
	}

	// each access switch must have a value
	for _, accessSwitch := range config.AccessSwitches.Elements() {
		if accessSwitch.IsUnknown() {
			return // cannot proceed
		}
	}

	// extract access switches
	accessSwitches := make(map[string]design.AccessSwitch)
	resp.Diagnostics.Append(config.AccessSwitches.ElementsAs(ctx, &accessSwitches, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check each access switch
	for name, accessSwitch := range accessSwitches {
		// links must have a value
		if accessSwitch.Links.IsUnknown() {
			return // cannot proceed
		}

		// each link must have a value
		for _, link := range accessSwitch.Links.Elements() {
			if link.IsUnknown() {
				return // cannot proceed
			}
		}

		// extract links from access switches
		var links []design.RackLink
		resp.Diagnostics.Append(accessSwitch.Links.ElementsAs(ctx, &links, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// map keyed by link name to ensure names are unique
		linkNameMap := make(map[string]bool, len(links))

		// check each link
		for _, link := range links {
			// link name must have a value
			if link.Name.IsNull() {
				return // cannot proceed
			}

			if linkNameMap[link.Name.ValueString()] {
				// the name has been seen before!
				resp.Diagnostics.AddAttributeError(
					path.Root("access_switches"), "Link names must be unique",
					fmt.Sprintf("Access Switch with name %q has multiple links with name %q",
						name, link.Name.ValueString()))
			} else {
				// save name in the map
				linkNameMap[link.Name.ValueString()] = true
			}
		}
	}

	// generic systems must have a value
	if config.GenericSystems.IsUnknown() {
		return // cannot proceed
	}

	// each generic system must have a value
	for _, genericSystem := range config.GenericSystems.Elements() {
		if genericSystem.IsUnknown() {
			return // cannot proceed
		}
	}

	// extract generic systems
	genericSystems := make(map[string]design.GenericSystem)
	resp.Diagnostics.Append(config.GenericSystems.ElementsAs(ctx, &genericSystems, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// check each generic system
	for name, genericSystem := range genericSystems {
		// links must have a value
		if genericSystem.Links.IsUnknown() {
			return // cannot proceed
		}

		// each link must have a value
		for _, link := range genericSystem.Links.Elements() {
			if link.IsUnknown() {
				return // cannot proceed
			}
		}

		// extract links from generic system
		var links []design.RackLink
		resp.Diagnostics.Append(genericSystem.Links.ElementsAs(ctx, &links, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// map keyed by link name to ensure names are unique
		linkNameMap := make(map[string]bool, len(links))

		// check each link
		for _, link := range links {
			// link name must have a value
			if link.Name.IsNull() {
				return // cannot proceed
			}

			if linkNameMap[link.Name.ValueString()] {
				// the name has been seen before!
				resp.Diagnostics.AddAttributeError(
					path.Root("generic_systems"), "Link names must be unique",
					fmt.Sprintf("Generic System with name %q has multiple links with name %q",
						name, link.Name.ValueString()))
			} else {
				// save name in the map
				linkNameMap[link.Name.ValueString()] = true
			}
		}
	}
}

func (o *resourceRackType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan design.RackType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a RackTypeRequest
	rtRequest := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the RackType object (nested objects are referenced by ID)
	id, err := o.client.CreateRackType(ctx, rtRequest)
	if err != nil {
		resp.Diagnostics.AddError("error creating rack type", err.Error())
		return
	}

	// retrieve the RackType object with fully-enumerated embedded objects
	rt, err := o.client.GetRackType(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving rack type info after creation", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	design.ValidateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a state object
	state := design.RackType{}
	state.Id = types.StringValue(string(rt.Id))
	state.LoadApiData(ctx, rt.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the plan into the state
	state.CopyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceRackType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state design.RackType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// fetch the rack type detail from the API
	rt, err := o.client.GetRackType(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error reading rack type", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	design.ValidateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a new state object
	var newState design.RackType
	newState.LoadApiData(ctx, rt.Data, &resp.Diagnostics)
	newState.Id = types.StringValue(string(rt.Id))
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the previous state into the new state
	newState.CopyWriteOnlyElements(ctx, &state, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (o *resourceRackType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve plan
	var plan design.RackType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a RackTypeRequest
	rtRequest := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// send the Request to Apstra
	err := o.client.UpdateRackType(ctx, apstra.ObjectId(plan.Id.ValueString()), rtRequest)
	if err != nil {
		resp.Diagnostics.AddError("error while updating Rack Type", err.Error())
		return
	}

	// retrieve the RackType object with fully-enumerated embedded objects
	rt, err := o.client.GetRackType(ctx, apstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error retrieving rack type info after creation", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	design.ValidateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a state object
	var state design.RackType
	state.LoadApiData(ctx, rt.Data, &resp.Diagnostics)
	state.Id = types.StringValue(string(rt.Id))
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the (old) into state
	state.CopyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceRackType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state design.RackType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.DeleteRackType(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay in Delete()
		}
		resp.Diagnostics.AddError("error deleting Rack Type", err.Error())
		return
	}
}

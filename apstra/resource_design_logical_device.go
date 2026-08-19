package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceDesignLogicalDevice{}
var _ resource.ResourceWithValidateConfig = &resourceDesignLogicalDevice{}
var _ resourceWithSetClient = &resourceDesignLogicalDevice{}

type resourceDesignLogicalDevice struct {
	client *apstra.Client
}

func (dld *resourceDesignLogicalDevice) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_design_logical_device"
}

func (dld *resourceDesignLogicalDevice) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, dld, req, resp)
}

func (dld *resourceDesignLogicalDevice) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This resource creates a Logical Device in the Apstra _Design_ tab.",
		Attributes:          design.LogicalDevice{}.ResourceAttributes(),
	}
}

func (dld *resourceDesignLogicalDevice) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config design.LogicalDevice
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Validate(ctx, &resp.Diagnostics)
}

func (dld *resourceDesignLogicalDevice) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan design.LogicalDevice
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the Logical Device
	id, err := dld.client.CreateLogicalDevice2(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Logical Device", err.Error())
		return
	}

	// Save computed values and set the state
	plan.ID = types.StringValue(id)
	plan.Definition = plan.DefinitionAsObject(ctx, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (dld *resourceDesignLogicalDevice) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state design.LogicalDevice
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Logical Device from API and then update what is in state from what the API returns
	ld, err := dld.client.GetLogicalDevice2(ctx, state.ID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"error reading Logical Device",
			fmt.Sprintf("Could not Read %q - %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Update state with values from API
	state.LoadAPIData(ctx, ld, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set thes state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (dld *resourceDesignLogicalDevice) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan design.LogicalDevice
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Logical Device
	err := dld.client.UpdateLogicalDevice2(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error updating Logical Device", err.Error())
		return
	}

	// Update computed value and set the state
	plan.Definition = plan.DefinitionAsObject(ctx, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (dld *resourceDesignLogicalDevice) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state design.LogicalDevice
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Logical Device by calling API
	err := dld.client.DeleteLogicalDevice2(ctx, state.ID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting Logical Device", err.Error())
		return
	}
}

func (dld *resourceDesignLogicalDevice) setClient(client *apstra.Client) {
	dld.client = client
}

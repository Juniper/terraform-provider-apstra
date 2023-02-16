package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

const (
	vlanMin = 1
	vlanMax = 4094

	poIdMin = 0
	poIdMax = 4096
)

var _ resource.ResourceWithConfigure = &resourceRackType{}
var _ resource.ResourceWithValidateConfig = &resourceRackType{}

type resourceRackType struct {
	client *goapstra.Client
}

func (o *resourceRackType) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack_type"
}

func (o *resourceRackType) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errResourceConfigureProviderDataDetail,
			fmt.Sprintf(errResourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *resourceRackType) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates a Rack Type in the Apstra Design tab.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Object ID for the Rack Type, assigned by Apstra.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Rack Type name, displayed in the Apstra web UI.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Rack Type description, displayed in the Apstra web UI.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"fabric_connectivity_design": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Must be one of '%s'.", strings.Join(fcdModes(), "', '")),
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf(fcdModes()...)},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"leaf_switches": schema.MapNestedAttribute{
				MarkdownDescription: "Each Rack Type is required to have at least one Leaf Switch.",
				Required:            true,
				Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: rRackTypeLeafSwitch{}.attributes(),
				},
			},
			"access_switches": schema.MapNestedAttribute{
				MarkdownDescription: "Access Switches are optional, link to Leaf Switches in the same rack",
				Optional:            true,
				Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: rRackTypeAccessSwitch{}.attributes(),
				},
			},
			"generic_systems": schema.MapNestedAttribute{
				MarkdownDescription: "Generic Systems are rack elements not" +
					"managed by Apstra: Servers, routers, firewalls, etc...",
				Optional:   true,
				Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: rRackTypeGenericSystem{}.attributes(),
				},
			},
		},
	}
}

func (o *resourceRackType) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if o.client == nil { // cannot proceed without a client
		return
	}

	var config rRackType
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//config.validateConfigAccessSwitches(ctx, path.Root("access_switches"), &resp.Diagnostics)
	//config.validateConfigGenericSystems(ctx, path.Root("generic_systems"), &resp.Diagnostics)
}

func (o *resourceRackType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rRackType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a RackTypeRequest
	rtRequest := plan.request(ctx, &resp.Diagnostics)
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
	validateRackType(ctx, rt, &resp.Diagnostics) // todo: chase this down for places HasError() should be checked
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a state object
	state := rRackType{}
	state.loadApiResponse(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the plan into the state
	state.copyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// todo: errpath with AtListIndex() are probably mostly wrong
func (o *resourceRackType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve values from state
	var state rRackType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// fetch the rack type detail from the API
	rt, err := o.client.GetRackType(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error reading rack type", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	validateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a new state object
	newState := rRackType{}
	newState.loadApiResponse(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the previous state into the new state
	newState.copyWriteOnlyElements(ctx, &state, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// todo: bug: copyWriteOnlyElements needs to check whether the destination is known, not overwrite when, e.g. logical device ID changes
func (o *resourceRackType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve plan
	var plan rRackType
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a RackTypeRequest
	rtRequest := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// send the request to Apstra
	err := o.client.UpdateRackType(ctx, goapstra.ObjectId(plan.Id.ValueString()), rtRequest)
	if err != nil {
		resp.Diagnostics.AddError("error while updating Rack Type", err.Error())
		return
	}

	// retrieve the RackType object with fully-enumerated embedded objects
	rt, err := o.client.GetRackType(ctx, goapstra.ObjectId(plan.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("error retrieving rack type info after creation", err.Error())
		return
	}

	// validate API response to catch problems which might crash the provider
	validateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a state object
	state := &rRackType{}
	state.loadApiResponse(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the (old) into state
	state.copyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceRackType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	// Retrieve values from state
	var state rRackType
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.DeleteRackType(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			return // 404 is okay in Delete()
		}
		resp.Diagnostics.AddError("error deleting Rack Type", err.Error())
	}
}

type rRackType struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Map    `tfsdk:"leaf_switches"`
	AccessSwitches           types.Map    `tfsdk:"access_switches"`
	GenericSystems           types.Map    `tfsdk:"generic_systems"`
}

func (o *rRackType) fabricConnectivityDesign(_ context.Context, diags *diag.Diagnostics) goapstra.FabricConnectivityDesign {
	var fcd goapstra.FabricConnectivityDesign
	err := fcd.FromString(o.FabricConnectivityDesign.ValueString())
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing FCD '%s' - %s",
				o.FabricConnectivityDesign.ValueString(), err.Error()))
	}
	return fcd
}

func (o *rRackType) getSwitchRedundancyProtocolByName(ctx context.Context, name string, path path.Path, diags *diag.Diagnostics) fmt.Stringer {
	leaf := o.leafSwitchByName(ctx, name, diags)
	access := o.accessSwitchByName(ctx, name, diags)
	if leaf == nil && access == nil {
		diags.AddAttributeError(path, errInvalidConfig,
			fmt.Sprintf("target switch %q not found in rack type %q", name, o.Id))
		return nil
	}
	if leaf != nil && access != nil {
		diags.AddError(errProviderBug, "link seems to be attached to both leaf and access switches")
		return nil
	}

	var leafRedundancyProtocol goapstra.LeafRedundancyProtocol
	if leaf != nil {
		if leaf.RedundancyProtocol.IsNull() {
			return goapstra.LeafRedundancyProtocolNone
		}
		err := leafRedundancyProtocol.FromString(leaf.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing leaf switch redundancy protocol", err.Error())
			return nil
		}
		return leafRedundancyProtocol
	}

	var accessRedundancyProtocol goapstra.AccessRedundancyProtocol
	if access != nil {
		if !access.EsiLagInfo.IsNull() {
			return goapstra.AccessRedundancyProtocolEsi
		}
		if access.RedundancyProtocol.IsNull() {
			return goapstra.AccessRedundancyProtocolNone
		}
		err := accessRedundancyProtocol.FromString(access.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing access switch redundancy protocol", err.Error())
			return nil
		}
		return accessRedundancyProtocol
	}
	diags.AddError(errProviderBug, "somehow we've reached the end of getSwitchRedundancyProtocolByName without finding a solution")
	return nil
}

func (o *rRackType) loadApiResponse(ctx context.Context, in *goapstra.RackType, diags *diag.Diagnostics) {
	leafSwitches := make(map[string]rRackTypeLeafSwitch, len(in.Data.LeafSwitches))
	for _, leafIn := range in.Data.LeafSwitches {
		var leafSwitch rRackTypeLeafSwitch
		leafSwitch.loadApiResponse(ctx, &leafIn, in.Data.FabricConnectivityDesign, diags)
		leafSwitches[leafIn.Label] = leafSwitch
		if diags.HasError() {
			return
		}
	}

	accessSwitches := make(map[string]rRackTypeAccessSwitch, len(in.Data.AccessSwitches))
	for _, accessIn := range in.Data.AccessSwitches {
		var accessSwitch rRackTypeAccessSwitch
		accessSwitch.loadApiResponse(ctx, &accessIn, diags)
		accessSwitches[accessIn.Label] = accessSwitch
		if diags.HasError() {
			return
		}
	}

	genericSystems := make(map[string]rRackTypeGenericSystem, len(in.Data.GenericSystems))
	for _, genericIn := range in.Data.GenericSystems {
		var genericSystem rRackTypeGenericSystem
		genericSystem.loadApiResponse(ctx, &genericIn, diags)
		genericSystems[genericIn.Label] = genericSystem
		if diags.HasError() {
			return
		}
	}

	var description types.String
	if in.Data.Description == "" {
		description = types.StringNull()
	} else {
		description = types.StringValue(in.Data.Description)
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)
	o.Description = description
	o.FabricConnectivityDesign = types.StringValue(in.Data.FabricConnectivityDesign.String())
	o.LeafSwitches = mapValueOrNull(ctx, rRackTypeLeafSwitch{}.attrType(), leafSwitches, diags)
	o.AccessSwitches = mapValueOrNull(ctx, rRackTypeAccessSwitch{}.attrType(), accessSwitches, diags)
	o.GenericSystems = mapValueOrNull(ctx, rRackTypeGenericSystem{}.attrType(), genericSystems, diags)
}

// copyWriteOnlyElements copies elements (IDs of nested design API objects)
// from 'src' (plan or state - something which knows these facts) into 'o' a
// rRackType to be used as state.
func (o *rRackType) copyWriteOnlyElements(ctx context.Context, src *rRackType, diags *diag.Diagnostics) {
	// first extract native go structs from the TF set of objects
	dstLeafSwitches := o.leafSwitches(ctx, diags)
	dstAccessSwitches := o.accessSwitches(ctx, diags)
	dstGenericSystems := o.genericSystems(ctx, diags)

	// invoke the copyWriteOnlyElements on every leaf switch object
	for name, dstLeafSwitch := range dstLeafSwitches {
		srcLeafSwitch, ok := src.leafSwitches(ctx, diags)[name]
		if !ok {
			continue
		}
		if diags.HasError() {
			return
		}

		dstLeafSwitch.copyWriteOnlyElements(ctx, &srcLeafSwitch, diags)
		if diags.HasError() {
			return
		}
		dstLeafSwitches[name] = dstLeafSwitch
	}

	// invoke the copyWriteOnlyElements on every access switch object
	for name, dstAccessSwitch := range dstAccessSwitches {
		srcAccessSwitch, ok := src.accessSwitches(ctx, diags)[name]
		if !ok {
			continue
		}
		if diags.HasError() {
			return
		}

		dstAccessSwitch.copyWriteOnlyElements(ctx, &srcAccessSwitch, diags)
		if diags.HasError() {
			return
		}
		dstAccessSwitches[name] = dstAccessSwitch
	}

	// invoke the copyWriteOnlyElements on every generic system object
	for name, dstGenericSystem := range dstGenericSystems {
		srcGenericSystem, ok := src.genericSystems(ctx, diags)[name]
		if !ok {
			continue
		}
		if diags.HasError() {
			return
		}

		dstGenericSystem.copyWriteOnlyElements(ctx, &srcGenericSystem, diags)
		if diags.HasError() {
			return
		}
		dstGenericSystems[name] = dstGenericSystem
	}

	// transform the native go objects (with copied object IDs) back to TF set
	leafSwitchMap := mapValueOrNull(ctx, rRackTypeLeafSwitch{}.attrType(), dstLeafSwitches, diags)
	accessSwitchMap := mapValueOrNull(ctx, rRackTypeAccessSwitch{}.attrType(), dstAccessSwitches, diags)
	genericSystemMap := mapValueOrNull(ctx, rRackTypeGenericSystem{}.attrType(), dstGenericSystems, diags)
	if diags.HasError() {
		return
	}

	// save the TF sets into rRackType
	o.LeafSwitches = leafSwitchMap
	o.AccessSwitches = accessSwitchMap
	o.GenericSystems = genericSystemMap
}

func (o *rRackLink) linkAttachmentType(upstreamRedundancyMode fmt.Stringer) goapstra.RackLinkAttachmentType {
	switch upstreamRedundancyMode.String() {
	case goapstra.LeafRedundancyProtocolNone.String():
		return goapstra.RackLinkAttachmentTypeSingle
	case goapstra.AccessRedundancyProtocolNone.String():
		return goapstra.RackLinkAttachmentTypeSingle
	}

	if o.LagMode.IsNull() {
		return goapstra.RackLinkAttachmentTypeSingle
	}

	if o.SwitchPeer.IsNull() {
		return goapstra.RackLinkAttachmentTypeSingle
	}

	switch o.LagMode.ValueString() {
	case goapstra.RackLinkLagModeActive.String():
		return goapstra.RackLinkAttachmentTypeDual
	case goapstra.RackLinkLagModePassive.String():
		return goapstra.RackLinkAttachmentTypeDual
	case goapstra.RackLinkLagModeStatic.String():
		return goapstra.RackLinkAttachmentTypeDual
	}
	return goapstra.RackLinkAttachmentTypeSingle
}

func (o *rRackType) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.RackTypeRequest {
	fcd := o.fabricConnectivityDesign(ctx, diags)
	if diags.HasError() {
		return nil
	}

	leafSwitches := o.leafSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	accessSwitches := o.accessSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	genericSystems := o.genericSystems(ctx, diags)
	if diags.HasError() {
		return nil
	}

	var i int

	leafSwitchRequests := make([]goapstra.RackElementLeafSwitchRequest, len(leafSwitches))
	i = 0
	for name, leafSwitch := range leafSwitches {
		lsr := leafSwitch.request(ctx, path.Root("leaf_switches").AtMapKey(name), fcd, diags)
		if diags.HasError() {
			return nil
		}
		leafSwitchRequests[i] = *lsr
		i++
	}

	accessSwitchRequests := make([]goapstra.RackElementAccessSwitchRequest, len(accessSwitches))
	i = 0
	for name, accessSwitch := range accessSwitches {
		asr := accessSwitch.request(ctx, path.Root("access_switches").AtMapKey(name), o, diags)
		if diags.HasError() {
			return nil
		}
		accessSwitchRequests[i] = *asr
		i++
	}

	genericSystemsRequests := make([]goapstra.RackElementGenericSystemRequest, len(genericSystems))
	i = 0
	for name, genericSystem := range genericSystems {
		gsr := genericSystem.request(ctx, path.Root("generic_systems").AtMapKey(name), o, diags)
		if diags.HasError() {
			return nil
		}
		genericSystemsRequests[i] = *gsr
		i++
	}

	return &goapstra.RackTypeRequest{
		DisplayName:              o.Name.ValueString(),
		Description:              o.Description.ValueString(),
		FabricConnectivityDesign: fcd,
		LeafSwitches:             leafSwitchRequests,
		AccessSwitches:           accessSwitchRequests,
		GenericSystems:           genericSystemsRequests,
	}
}

func (o *rRackType) leafSwitches(ctx context.Context, diags *diag.Diagnostics) map[string]rRackTypeLeafSwitch {
	var leafSwitches map[string]rRackTypeLeafSwitch
	d := o.LeafSwitches.ElementsAs(ctx, &leafSwitches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	// copy the leaf switch name from the map key into the object's Name field
	for name, leafSwitch := range leafSwitches {
		leafSwitch.Name = types.StringValue(name)
		leafSwitches[name] = leafSwitch
	}
	return leafSwitches
}

func (o *rRackType) leafSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackTypeLeafSwitch {
	leafSwitches := o.leafSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if leafSwitch, ok := leafSwitches[requested]; ok {
		return &leafSwitch
	}

	return nil
}

func (o *rRackType) accessSwitches(ctx context.Context, diags *diag.Diagnostics) map[string]rRackTypeAccessSwitch {
	var accessSwitches map[string]rRackTypeAccessSwitch
	d := o.AccessSwitches.ElementsAs(ctx, &accessSwitches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	// copy the access switch name from the map key into the object's Name field
	for name, accessSwitch := range accessSwitches {
		accessSwitch.Name = types.StringValue(name)
		accessSwitches[name] = accessSwitch
	}
	return accessSwitches
}

func (o *rRackType) accessSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackTypeAccessSwitch {
	accessSwitches := o.accessSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if accessSwitch, ok := accessSwitches[requested]; ok {
		return &accessSwitch
	}

	return nil
}

func (o *rRackType) genericSystems(ctx context.Context, diags *diag.Diagnostics) map[string]rRackTypeGenericSystem {
	var genericSystems map[string]rRackTypeGenericSystem
	d := o.GenericSystems.ElementsAs(ctx, &genericSystems, true)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	// copy the generic system name from the map key into the object's Name field
	for name, genericSystem := range genericSystems {
		genericSystem.Name = types.StringValue(name)
		genericSystems[name] = genericSystem
	}
	return genericSystems
}

func (o *rRackType) genericSystemByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackTypeGenericSystem {
	genericSystems := o.genericSystems(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if genericSystem, ok := genericSystems[requested]; ok {
		return &genericSystem
	}

	return nil
}

// fcdModes returns permitted fabric_connectivity_design mode strings
func fcdModes() []string {
	return []string{
		goapstra.FabricConnectivityDesignL3Clos.String(),
		goapstra.FabricConnectivityDesignL3Collapsed.String()}
}

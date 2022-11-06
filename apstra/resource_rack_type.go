package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"strings"
)

const (
	vlanMin = 1
	vlanMax = 4094

	poIdMin = 1
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

func (o *resourceRackType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource creates a Rack Type in the Apstra Design tab.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Object ID for the Rack Type, assigned by Apstra.",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				MarkdownDescription: "Rack Type name, displayed in the Apstra web UI.",
				Type:                types.StringType,
				Required:            true,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"description": {
				MarkdownDescription: "Rack Type description, displayed in the Apstra web UI.",
				Type:                types.StringType,
				Optional:            true,
			},
			"fabric_connectivity_design": {
				MarkdownDescription: fmt.Sprintf("Must be one of '%s'.", strings.Join(fcdModes(), "', '")),
				Type:                types.StringType,
				Required:            true,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.OneOf(fcdModes()...)},
			},
			"leaf_switches": {
				MarkdownDescription: "Each Rack Type is required to have at least one Leaf Switch.",
				Required:            true,
				Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
						Type:                types.StringType,
						Required:            true,
						Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
					},
					"logical_device_id": {
						MarkdownDescription: "Apstra Object ID of the Logical Device used to model this switch.",
						Type:                types.StringType,
						Required:            true,
					},
					"spine_link_count": {
						MarkdownDescription: "Links per spine.",
						Type:                types.Int64Type,
						Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
						Optional:            true,
					},
					"spine_link_speed": {
						MarkdownDescription: "Speed of spine-facing links, something like '10G'",
						Type:                types.StringType,
						Optional:            true,
					},
					"redundancy_protocol": {
						MarkdownDescription: fmt.Sprintf("Enabling a redundancy protocol converts a single "+
							"Leaf Switch into a LAG-capable switch pair. Must be one of '%s'.",
							strings.Join(leafRedundancyModes(), "', '")),
						Type:       types.StringType,
						Optional:   true,
						Validators: []tfsdk.AttributeValidator{stringvalidator.OneOf(leafRedundancyModes()...)},
					},
					"logical_device": logicalDeviceDataAttributeSchema(),
					"tag_ids":        tagIdsAttributeSchema(),
					"tag_data":       tagsDataAttributeSchema(),
					"mlag_info": {
						MarkdownDescription: fmt.Sprintf("Required when `redundancy_protocol` set to `%s`, "+
							"defines the connectivity between MLAG peers.", goapstra.LeafRedundancyProtocolMlag.String()),
						Optional: true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"mlag_keepalive_vlan": {
								MarkdownDescription: "MLAG keepalive VLAN ID.",
								Required:            true,
								Type:                types.Int64Type,
								Validators: []tfsdk.AttributeValidator{
									int64validator.Between(vlanMin, vlanMax),
								},
							},
							"peer_link_count": {
								MarkdownDescription: "Number of links between MLAG devices.",
								Required:            true,
								Type:                types.Int64Type,
								Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
							},
							"peer_link_speed": {
								MarkdownDescription: "Speed of links between MLAG devices.",
								Required:            true,
								Type:                types.StringType,
							},
							"peer_link_port_channel_id": {
								MarkdownDescription: "Port channel number used for L2 Peer Link. Omit to allow Apstra to choose.",
								Optional:            true,
								Type:                types.Int64Type,
								Validators: []tfsdk.AttributeValidator{
									int64validator.Between(poIdMin, poIdMax),
								},
							},
							"l3_peer_link_count": {
								MarkdownDescription: "Number of L3 links between MLAG devices.",
								Optional:            true,
								Type:                types.Int64Type,
								Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
							},
							"l3_peer_link_speed": {
								MarkdownDescription: "Speed of l3 links between MLAG devices.",
								Optional:            true,
								Type:                types.StringType,
							},
							"l3_peer_link_port_channel_id": {
								MarkdownDescription: "Port channel number used for L3 Peer Link. Omit to allow Apstra to choose.",
								Optional:            true,
								Type:                types.Int64Type,
								Validators: []tfsdk.AttributeValidator{
									int64validator.Between(poIdMin, poIdMax),
								},
							},
						}),
					},
				}),
			},
			"access_switches": {
				MarkdownDescription: "Access switches provide fan-out connectivity from Leaf Switches.",
				Optional:            true,
				Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
						Type:                types.StringType,
						Required:            true,
						Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
					},
					"count": {
						MarkdownDescription: "Number of Access Switches of this type.",
						Type:                types.Int64Type,
						Required:            true,
						Validators: []tfsdk.AttributeValidator{
							int64validator.AtLeast(1),
						},
					},
					"redundancy_protocol": {
						MarkdownDescription: "Indicates whether the switch is a redundant pair.",
						Type:                types.StringType,
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{useStateForUnknownNull()},
					},
					"logical_device_id": {
						MarkdownDescription: "Apstra Object ID of the Logical Device used to model this switch.",
						Type:                types.StringType,
						Required:            true,
					},
					"logical_device": logicalDeviceDataAttributeSchema(),
					"links":          rRackLinkAttributeSchema(),
					"tag_ids":        tagIdsAttributeSchema(),
					"tag_data":       tagsDataAttributeSchema(),
					"esi_lag_info": {
						MarkdownDescription: "Including this stanza converts the Access Switch into a redundant pair.",
						Optional:            true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"l3_peer_link_count": {
								MarkdownDescription: "Number of L3 links between ESI-LAG devices.",
								Required:            true,
								Type:                types.Int64Type,
								Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
							},
							"l3_peer_link_speed": {
								MarkdownDescription: "Speed of l3 links between ESI-LAG devices.",
								Required:            true,
								Type:                types.StringType,
							},
						}),
					},
				}),
			},
			"generic_systems": {
				MarkdownDescription: "Generic Systems are rack elements not" +
					"managed by Apstra: Servers, routers, firewalls, etc...",
				Optional:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Generic name, must be unique within the rack-type.",
						Type:                types.StringType,
						Required:            true,
						Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
					},
					"count": {
						MarkdownDescription: "Number of Generic Systems of this type.",
						Type:                types.Int64Type,
						Required:            true,
						Validators: []tfsdk.AttributeValidator{
							int64validator.AtLeast(1),
						},
					},
					"port_channel_id_min": {
						MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
						Optional:            true,
						Computed:            true,
						Type:                types.Int64Type,
					},
					"port_channel_id_max": {
						MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
						Optional:            true,
						Computed:            true,
						Type:                types.Int64Type,
					},
					"logical_device_id": {
						MarkdownDescription: "Apstra Object ID of the Logical Device used to model this switch.",
						Type:                types.StringType,
						Required:            true,
					},
					"logical_device": logicalDeviceDataAttributeSchema(),
					"links":          rRackLinkAttributeSchema(),
					"tag_ids":        tagIdsAttributeSchema(),
					"tag_data":       tagsDataAttributeSchema(),
				}),
			},
		},
	}, nil
}

func (o *resourceRackType) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if o.client == nil { // cannot proceed without a client
		return
	}

	var config rRackType
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// todo: >1 links per switch means lag is required
	config.validateConfigLeafSwitches(ctx, path.Root("leaf_switches"), &resp.Diagnostics)
	config.validateConfigAccessSwitches(ctx, path.Root("access_switches"), &resp.Diagnostics)
	config.validateConfigGenericSystems(ctx, path.Root("generic_systems"), &resp.Diagnostics)
}

func (o *resourceRackType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rRackType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
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
	validateRackType(rt, &resp.Diagnostics) // todo: chase this down for places HasError() should be checked
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a state object
	state := &rRackType{}
	state.parseApi(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the plan into the state
	state.copyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// todo: tag lists should be come sets
// todo: errpath with AtListIndex() are probably mostly wrong
func (o *resourceRackType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve values from state
	var state rRackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// fetch the rack type detail from the API
	rt, err := o.client.GetRackType(ctx, goapstra.ObjectId(state.Id.Value))
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
	validateRackType(rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// parse the API response into a new state object
	newState := &rRackType{}
	newState.parseApi(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// copy nested object IDs (those not available from the API) from the previous state into the new state
	newState.copyWriteOnlyElements(ctx, &state, &resp.Diagnostics)

	old, _ := json.MarshalIndent(&state, "", "  ")
	new, _ := json.MarshalIndent(newState, "", "  ")

	os.WriteFile("/tmp/old", old, 0644)
	os.WriteFile("/tmp/new", new, 0644)

	// Set state
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

func (o *resourceRackType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Retrieve state
	var state rRackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve plan
	var plan rRackType
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//// force values as needed
	//plan.forceValues(&resp.Diagnostics)
	//if diags.HasError() {
	//	return
	//}

	//// populate rack elements (leaf/access/generic) from global catalog
	//plan.populateDataFromGlobalCatalog(ctx, o.client, &resp.Diagnostics)
	//if resp.Diagnostics.HasError() {
	//	return
	//}

	//// Prepare a goapstra.RackTypeRequest
	//rtReq := plan.goapstraRequest(&resp.Diagnostics)
	//if diags.HasError() {
	//	return
	//}

	//err := o.client.UpdateRackType(ctx, goapstra.ObjectId(state.Id.Value), rtReq)
	//if err != nil {
	//	resp.Diagnostics.AddError("error while updating Rack Type", err.Error())
	//}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (o *resourceRackType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	// Retrieve values from state
	var state rRackType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	LeafSwitches             types.Set    `tfsdk:"leaf_switches"`
	AccessSwitches           types.Set    `tfsdk:"access_switches"`
	GenericSystems           types.Set    `tfsdk:"generic_systems"`
}

func (o *rRackType) fabricConnectivityDesign() (*goapstra.FabricConnectivityDesign, error) {
	var fcd goapstra.FabricConnectivityDesign
	return &fcd, fcd.FromString(o.FabricConnectivityDesign.ValueString())
}

func (o *rRackType) getLeafSwitchByName(ctx context.Context, name string, diags *diag.Diagnostics) *rRackTypeLeafSwitch {
	var leafSwitches []rRackTypeLeafSwitch
	d := o.LeafSwitches.ElementsAs(ctx, &leafSwitches, true)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	for _, leafSwitch := range leafSwitches {
		if leafSwitch.Name == name {
			return &leafSwitch
		}
	}
	return nil
}

func (o *rRackType) getAccessSwitchByName(ctx context.Context, name string, diags *diag.Diagnostics) *rRackTypeAccessSwitch {
	var accessSwitches []rRackTypeAccessSwitch
	d := o.AccessSwitches.ElementsAs(ctx, &accessSwitches, true)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	for _, accessSwitch := range accessSwitches {
		if accessSwitch.Name == name {
			return &accessSwitch
		}
	}
	return nil
}

func (o *rRackType) getSwitchRedundancyProtocolByName(ctx context.Context, name string, path path.Path, diags *diag.Diagnostics) fmt.Stringer {
	leaf := o.getLeafSwitchByName(ctx, name, diags)
	access := o.getAccessSwitchByName(ctx, name, diags)
	if leaf == nil && access == nil {
		diags.AddAttributeError(path, errInvalidConfig,
			fmt.Sprintf("target switch '%s' not found in rack type '%s'", name, o.Id))
		return nil
	}
	if leaf != nil && access != nil {
		diags.AddError(errProviderBug, "link seems to be attached to both leaf and access switches")
		return nil
	}

	var leafRedundancyProtocol goapstra.LeafRedundancyProtocol
	if leaf != nil {
		if leaf.RedundancyProtocol == nil {
			return goapstra.LeafRedundancyProtocolNone
		}
		err := leafRedundancyProtocol.FromString(*leaf.RedundancyProtocol)
		if err != nil {
			diags.AddAttributeError(path, "error parsing leaf switch redundancy protocol", err.Error())
			return nil
		}
		return leafRedundancyProtocol
	}

	var accessRedundancyProtocol goapstra.AccessRedundancyProtocol
	if access != nil {
		if access.RedundancyProtocol == nil {
			return goapstra.AccessRedundancyProtocolNone
		}
		err := accessRedundancyProtocol.FromString(*access.RedundancyProtocol)
		if err != nil {
			diags.AddAttributeError(path, "error parsing access switch redundancy protocol", err.Error())
			return nil
		}
		return accessRedundancyProtocol
	}
	diags.AddError(errProviderBug, "somehow we've reached the end of getSwitchRedundancyProtocolByName without finding a solution")
	return nil
}

func (o *rRackType) parseApi(ctx context.Context, in *goapstra.RackType, diags *diag.Diagnostics) {
	var d diag.Diagnostics

	var leafSwitchSet types.Set
	if len(in.Data.LeafSwitches) > 0 {
		leafSwitches := make([]rRackTypeLeafSwitch, len(in.Data.LeafSwitches))
		for i := range in.Data.LeafSwitches {
			leafSwitches[i].parseApi(&in.Data.LeafSwitches[i], in.Data.FabricConnectivityDesign)
		}
		leafSwitchSet, d = types.SetValueFrom(ctx, rRackTypeLeafSwitch{}.attrType(), leafSwitches)
		diags.Append(d...)
	} else {
		leafSwitchSet = types.SetNull(rRackTypeLeafSwitch{}.attrType())
	}

	var accessSwitchSet types.Set
	if len(in.Data.AccessSwitches) > 0 {
		accessSwitches := make([]rRackTypeAccessSwitch, len(in.Data.AccessSwitches))
		for i := range in.Data.AccessSwitches {
			accessSwitches[i].parseApi(&in.Data.AccessSwitches[i])
		}
		accessSwitchSet, d = types.SetValueFrom(ctx, rRackTypeLeafSwitch{}.attrType(), accessSwitches)
		diags.Append(d...)
	} else {
		accessSwitchSet = types.SetNull(rRackTypeAccessSwitch{}.attrType())
	}

	var genericSystemSet types.Set
	if len(in.Data.GenericSystems) > 0 {
		genericSystems := make([]rRackTypeGenericSystem, len(in.Data.GenericSystems))
		for i := range in.Data.GenericSystems {
			genericSystems[i].parseApi(&in.Data.GenericSystems[i])
		}
		genericSystemSet, d = types.SetValueFrom(ctx, rRackTypeGenericSystem{}.attrType(), genericSystems)
		diags.Append(d...)
	} else {
		genericSystemSet = types.SetNull(rRackTypeGenericSystem{}.attrType())
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)
	o.Description = types.StringValue(in.Data.Description)
	o.FabricConnectivityDesign = types.StringValue(in.Data.FabricConnectivityDesign.String())
	o.LeafSwitches = leafSwitchSet
	o.AccessSwitches = accessSwitchSet
	o.GenericSystems = genericSystemSet
}

// copyWriteOnlyElements copies elements (IDs of nested design API objects)
// from 'src' (plan or state - something which knows these facts) into 'o' a
// rRackType to be used as state.
func (o *rRackType) copyWriteOnlyElements(ctx context.Context, src *rRackType, diags *diag.Diagnostics) {
	// first extract native go structs from the TF set of objects
	leafSwitches := o.leafSwitches(ctx, diags)
	accessSwitches := o.accessSwitches(ctx, diags)
	genericSystems := o.genericSystems(ctx, diags)

	// invoke the copyWriteOnlyElements on every leaf switch object
	for i, leafSwitch := range leafSwitches {
		srcLeafSwitch := src.leafSwitchByName(ctx, leafSwitch.Name, diags)
		if diags.HasError() {
			return
		}
		leafSwitches[i].copyWriteOnlyElements(srcLeafSwitch, diags)
		if diags.HasError() {
			return
		}
	}

	// invoke the copyWriteOnlyElements on every access switch object
	for _, accessSwitch := range accessSwitches {
		srcAccessSwitch := src.accessSwitchByName(ctx, accessSwitch.Name, diags)
		if diags.HasError() {
			return
		}
		accessSwitch.copyWriteOnlyElements(srcAccessSwitch, diags)
		if diags.HasError() {
			return
		}
	}

	// invoke the copyWriteOnlyElements on every generic system object
	for _, genericSystem := range genericSystems {
		srcGenericSystem := src.genericSystemByName(ctx, genericSystem.Name, diags)
		if diags.HasError() {
			return
		}
		genericSystem.copyWriteOnlyElements(srcGenericSystem, diags)
		if diags.HasError() {
			return
		}
	}

	var d diag.Diagnostics
	var leafSwitchSet, accessSwitchSet, genericSystemSet types.Set

	// transform the native go objects (with copied object IDs) back to TF set
	if len(leafSwitches) > 0 {
		leafSwitchSet, d = types.SetValueFrom(ctx, rRackTypeLeafSwitch{}.attrType(), leafSwitches)
		diags.Append(d...)
	} else {
		leafSwitchSet = types.SetNull(rRackTypeLeafSwitch{}.attrType())
	}

	// transform the native go objects (with copied object IDs) back to TF set
	if len(accessSwitches) > 0 {
		accessSwitchSet, d = types.SetValueFrom(ctx, rRackTypeAccessSwitch{}.attrType(), accessSwitches)
		diags.Append(d...)
	} else {
		accessSwitchSet = types.SetNull(rRackTypeAccessSwitch{}.attrType())
	}

	// transform the native go objects (with copied object IDs) back to TF set
	if len(genericSystems) > 0 {
		genericSystemSet, d = types.SetValueFrom(ctx, rRackTypeGenericSystem{}.attrType(), genericSystems)
		diags.Append(d...)
	} else {
		genericSystemSet = types.SetNull(rRackTypeGenericSystem{}.attrType())
	}

	// save the TF sets into rRackType
	o.LeafSwitches = leafSwitchSet
	o.AccessSwitches = accessSwitchSet
	o.GenericSystems = genericSystemSet
}

type rRackTypeLeafSwitch struct {
	Name               string            `tfsdk:"name"`
	LogicalDeviceId    string            `tfsdk:"logical_device_id"`
	SpineLinkCount     *int64            `tfsdk:"spine_link_count"`
	SpineLinkSpeed     *string           `tfsdk:"spine_link_speed"`
	RedundancyProtocol *string           `tfsdk:"redundancy_protocol"`
	LogicalDevice      logicalDeviceData `tfsdk:"logical_device"`
	TagIds             []string          `tfsdk:"tag_ids"`
	TagData            []tagData         `tfsdk:"tag_data"`
	MlagInfo           *mlagInfo         `tfsdk:"mlag_info""`
}

func (o rRackTypeLeafSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"logical_device_id":   types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"redundancy_protocol": types.StringType,
		"logical_device":      logicalDeviceData{}.attrType(),
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
		"mlag_info":           mlagInfo{}.attrType()}
}

func (o rRackTypeLeafSwitch) attrType() attr.Type {
	return types.ObjectType{AttrTypes: o.attrTypes()}
}

func (o *rRackTypeLeafSwitch) request(path path.Path, diags *diag.Diagnostics) *goapstra.RackElementLeafSwitchRequest {
	var linkPerSpineCount int
	if o.SpineLinkCount != nil {
		linkPerSpineCount = int(*o.SpineLinkCount)
	}

	var linkPerSpineSpeed goapstra.LogicalDevicePortSpeed
	if o.SpineLinkSpeed != nil {
		linkPerSpineSpeed = goapstra.LogicalDevicePortSpeed(*o.SpineLinkSpeed)
	}

	redundancyProtocol := goapstra.LeafRedundancyProtocolNone
	if o.RedundancyProtocol != nil {
		err := redundancyProtocol.FromString(*o.RedundancyProtocol)
		if err != nil {
			diags.AddAttributeError(path.AtMapKey("redundancy_protocol"),
				"error parsing redundancy_protocol", err.Error())
			return nil
		}
	}

	var tagIds []goapstra.ObjectId
	if o.TagIds != nil {
		tagIds = make([]goapstra.ObjectId, len(o.TagIds))
		for i, tagId := range o.TagIds {
			tagIds[i] = goapstra.ObjectId(tagId)
		}
	}

	return &goapstra.RackElementLeafSwitchRequest{
		Label:              o.Name,
		MlagInfo:           o.MlagInfo.request(),
		LinkPerSpineCount:  linkPerSpineCount,
		LinkPerSpineSpeed:  linkPerSpineSpeed,
		RedundancyProtocol: redundancyProtocol,
		Tags:               tagIds,
		LogicalDeviceId:    goapstra.ObjectId(o.LogicalDeviceId),
	}
}

func (o *rRackTypeLeafSwitch) validateConfig(errPath path.Path, rack *rRackType, diags *diag.Diagnostics) {
	fcd, err := rack.fabricConnectivityDesign()
	if err != nil {
		diags.AddAttributeError(errPath.AtMapKey("fabric_connectivity_design"), "parse error", err.Error())
		return
	}

	switch *fcd {
	case goapstra.FabricConnectivityDesignL3Clos:
		o.validateForL3Clos(errPath, diags) // todo: figure out how to use AtSetValue()
	case goapstra.FabricConnectivityDesignL3Collapsed:
		o.validateForL3Collapsed(errPath, diags)
	default:
		diags.AddAttributeError(errPath, errProviderBug, fmt.Sprintf("unknown fabric connectivity design '%s'", fcd.String()))
	}

	if o.RedundancyProtocol != nil && *o.RedundancyProtocol == goapstra.LeafRedundancyProtocolMlag.String() {
		o.validateMlagInfo(errPath, diags)
	}
}

func (o *rRackTypeLeafSwitch) validateForL3Clos(errPath path.Path, diags *diag.Diagnostics) {
	if o.SpineLinkCount == nil {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'spine_link_count' must be specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Clos))
	}

	if o.SpineLinkSpeed == nil {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'spine_link_speed' must be specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Clos))
	}
}

func (o *rRackTypeLeafSwitch) validateForL3Collapsed(errPath path.Path, diags *diag.Diagnostics) {
	if o.SpineLinkCount != nil {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'spine_link_count' must not be specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Collapsed))
	}

	if o.SpineLinkSpeed != nil {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'spine_link_speed' must bnot e specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Collapsed))
	}

	if o.RedundancyProtocol != nil {
		var redundancyProtocol goapstra.LeafRedundancyProtocol
		err := redundancyProtocol.FromString(*o.RedundancyProtocol)
		if err != nil {
			diags.AddAttributeError(errPath.AtMapKey("redundancy_protocol"), "parse_error", err.Error())
			return
		}
		if redundancyProtocol == goapstra.LeafRedundancyProtocolMlag {
			diags.AddAttributeError(errPath, errInvalidConfig,
				fmt.Sprintf("'redundancy_protocol' = '%s' is not allowed when 'fabric_connectivity_design' = '%s'",
					goapstra.LeafRedundancyProtocolMlag, goapstra.FabricConnectivityDesignL3Collapsed))
		}
	}
}

func (o *rRackTypeLeafSwitch) validateMlagInfo(errPath path.Path, diags *diag.Diagnostics) {
	var redundancyProtocol goapstra.LeafRedundancyProtocol
	err := redundancyProtocol.FromString(*o.RedundancyProtocol)
	if err != nil {
		diags.AddAttributeError(errPath.AtMapKey("redundancy_protocol"), "parse_error", err.Error())
		return
	}

	if o.MlagInfo == nil && redundancyProtocol == goapstra.LeafRedundancyProtocolMlag {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'mlag_info' required with 'redundancy_protocol' = '%s'", redundancyProtocol.String()))
	}

	if o.MlagInfo == nil {
		return
	}

	if redundancyProtocol != goapstra.LeafRedundancyProtocolMlag {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'mlag_info' incompatible with 'redundancy_protocol of '%s'", redundancyProtocol.String()))
	}

	if o.MlagInfo.PeerLinkPortChannelId != nil &&
		o.MlagInfo.L3PeerLinkPortChannelId != nil &&
		*o.MlagInfo.PeerLinkPortChannelId == *o.MlagInfo.L3PeerLinkPortChannelId {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'peer_link_port_channel_id' and 'l3_peer_link_port_channel_id' cannot both use value %d",
				*o.MlagInfo.PeerLinkPortChannelId))
	}

	if o.MlagInfo.L3PeerLinkCount != nil && o.MlagInfo.L3PeerLinkSpeed == nil {
		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_count' requires 'l3_peer_link_speed'")
	}
	if o.MlagInfo.L3PeerLinkSpeed != nil && o.MlagInfo.L3PeerLinkCount == nil {
		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_speed' requires 'l3_peer_link_count'")
	}

	if o.MlagInfo.L3PeerLinkPortChannelId != nil && o.MlagInfo.L3PeerLinkCount == nil {
		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_port_channel_id' requires 'l3_peer_link_count'")
	}
	if o.MlagInfo.L3PeerLinkCount != nil && o.MlagInfo.L3PeerLinkPortChannelId == nil {
		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_count' requires 'l3_peer_link_port_channel_id'")
	}
}

func (o *rRackTypeLeafSwitch) isRedundant() bool {
	if o.RedundancyProtocol == nil {
		return false
	}
	return true
}

func (o *rRackTypeLeafSwitch) parseApi(in *goapstra.RackElementLeafSwitch, fcd goapstra.FabricConnectivityDesign) {
	o.Name = in.Label
	if fcd != goapstra.FabricConnectivityDesignL3Collapsed {
		count := int64(in.LinkPerSpineCount)
		speed := string(in.LinkPerSpineSpeed)
		o.SpineLinkCount = &count
		o.SpineLinkSpeed = &speed
	}

	if in.RedundancyProtocol != goapstra.LeafRedundancyProtocolNone {
		redundancyProtocol := in.RedundancyProtocol.String()
		o.RedundancyProtocol = &redundancyProtocol
	}

	if in.MlagInfo != nil && in.MlagInfo.LeafLeafLinkCount > 0 {
		o.MlagInfo = &mlagInfo{}
		o.MlagInfo.parseApi(in.MlagInfo)
	}

	o.LogicalDevice.parseApi(in.LogicalDevice)

	if len(in.Tags) > 0 {
		o.TagData = make([]tagData, len(in.Tags)) // populated below
		for i := range in.Tags {
			o.TagData[i].parseApi(&in.Tags[i])
		}
	}
}

func (o *rRackTypeLeafSwitch) copyWriteOnlyElements(src *rRackTypeLeafSwitch, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddWarning(errProviderBug, "rRackTypeLeafSwitch.copyWriteOnlyElements: attempt to copy from nil source")
	}
	o.LogicalDeviceId = src.LogicalDeviceId
	o.TagIds = src.TagIds
}

type rRackTypeAccessSwitch struct {
	Name               string            `tfsdk:"name"`
	Count              int64             `tfsdk:"count"`
	RedundancyProtocol *string           `tfsdk:"redundancy_protocol"`
	LogicalDeviceId    string            `tfsdk:"logical_device_id"`
	LogicalDevice      logicalDeviceData `tfsdk:"logical_device"`
	Links              []rRackLink       `tfsdk:"links"`
	TagIds             []string          `tfsdk:"tag_ids"`
	TagData            []tagData         `tfsdk:"tag_data"`
	EsiLagInfo         *esiLagInfo       `tfsdk:"esi_lag_info""`
}

func (o rRackTypeAccessSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"count":               types.Int64Type,
		"redundancy_protocol": types.StringType,
		"logical_device_id":   types.StringType,
		"logical_device":      logicalDeviceData{}.attrType(),
		"links":               types.SetType{ElemType: rRackLink{}.attrType()},
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
		"esi_lag_info":        esiLagInfo{}.attrType()}
}

func (o rRackTypeAccessSwitch) attrType() attr.Type {
	return types.ObjectType{AttrTypes: o.attrTypes()}
}

func (o *rRackTypeAccessSwitch) request(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) *goapstra.RackElementAccessSwitchRequest {
	redundancyProtocol := goapstra.AccessRedundancyProtocolNone
	if o.RedundancyProtocol != nil {
		err := redundancyProtocol.FromString(*o.RedundancyProtocol)
		if err != nil {
			diags.AddAttributeError(path.AtMapKey("redundancy_protocol"),
				"error parsing redundancy_protocol", err.Error())
			return nil
		}
	}

	lacpActive := goapstra.RackLinkLagModeActive.String()
	links := make([]goapstra.RackLinkRequest, len(o.Links))
	for i, link := range o.Links {
		link.LagMode = &lacpActive
		links[i] = *link.request(ctx, path.AtListIndex(i), rack, diags)
	}

	var tagIds []goapstra.ObjectId
	if o.TagIds != nil {
		tagIds = make([]goapstra.ObjectId, len(o.TagIds))
		for i, tagId := range o.TagIds {
			tagIds[i] = goapstra.ObjectId(tagId)
		}
	}

	var esiLagInfo *goapstra.EsiLagInfo
	if o.EsiLagInfo != nil {
		esiLagInfo.AccessAccessLinkCount = int(o.EsiLagInfo.L3PeerLinkCount)
		esiLagInfo.AccessAccessLinkSpeed = goapstra.LogicalDevicePortSpeed(o.EsiLagInfo.L3PeerLinkSpeed)
	}

	return &goapstra.RackElementAccessSwitchRequest{
		InstanceCount:      int(o.Count),
		RedundancyProtocol: redundancyProtocol,
		Links:              links,
		Label:              o.Name,
		LogicalDeviceId:    goapstra.ObjectId(o.LogicalDeviceId),
		Tags:               tagIds,
		EsiLagInfo:         esiLagInfo,
	}
}

func (o *rRackTypeAccessSwitch) validateConfig(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) {
	arp := goapstra.AccessRedundancyProtocolNone
	if o.RedundancyProtocol != nil {
		err := arp.FromString(*o.RedundancyProtocol)
		if err != nil {
			diags.AddAttributeError(path, "error parsing redundancy protocol", err.Error())
		}
	}

	for i, link := range o.Links {
		link.validateConfigForAccessSwitch(ctx, arp, rack, path.AtListIndex(i), diags)
	}
}

func (o *rRackTypeAccessSwitch) isRedundant() bool {
	if o.RedundancyProtocol == nil {
		return false
	}
	return true
}

func (o *rRackTypeAccessSwitch) parseApi(in *goapstra.RackElementAccessSwitch) {
	o.Name = in.Label
	o.Count = int64(in.InstanceCount)
	if in.RedundancyProtocol != goapstra.AccessRedundancyProtocolNone {
		redundancyProtocol := in.RedundancyProtocol.String()
		o.RedundancyProtocol = &redundancyProtocol
	}
	if in.EsiLagInfo != nil {
		o.EsiLagInfo = &esiLagInfo{}
		o.EsiLagInfo.parseApi(in.EsiLagInfo)
	}
	o.LogicalDevice.parseApi(in.LogicalDevice)

	if len(in.Tags) > 0 {
		o.TagData = make([]tagData, len(in.Tags)) // populated below
		for i := range in.Tags {
			o.TagData[i].parseApi(&in.Tags[i])
		}
	}

	o.Links = make([]rRackLink, len(in.Links))
	for i := range in.Links {
		o.Links[i].parseApi(&in.Links[i])
	}
}

func (o *rRackTypeAccessSwitch) getLinkByName(desired string) *rRackLink {
	for _, link := range o.Links {
		if link.Name == desired {
			return &link
		}
	}
	return nil
}

func (o *rRackTypeAccessSwitch) copyWriteOnlyElements(src *rRackTypeAccessSwitch, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddWarning(errProviderBug, "rRackTypeAccessSwitch.copyWriteOnlyElements: attempt to copy from nil source")
	}
	o.LogicalDeviceId = src.LogicalDeviceId
	o.TagIds = src.TagIds

	for i, link := range o.Links {
		o.Links[i].copyWriteOnlyElements(src.getLinkByName(link.Name), diags)
		if diags.HasError() {
			return
		}
	}
}

type rRackTypeGenericSystem struct {
	Name             string            `tfsdk:"name"`
	Count            int64             `tfsdk:"count"`
	PortChannelIdMin *int64            `tfsdk:"port_channel_id_min"`
	PortChannelIdMax *int64            `tfsdk:"port_channel_id_max"`
	LogicalDeviceId  string            `tfsdk:"logical_device_id"`
	LogicalDevice    logicalDeviceData `tfsdk:"logical_device"`
	Links            []rRackLink       `tfsdk:"links"`
	TagIds           []string          `tfsdk:"tag_ids"`
	TagData          []tagData         `tfsdk:"tag_data"`
}

func (o rRackTypeGenericSystem) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"count":               types.Int64Type,
		"port_channel_id_min": types.Int64Type,
		"port_channel_id_max": types.Int64Type,
		"logical_device_id":   types.StringType,
		"logical_device":      logicalDeviceData{}.attrType(),
		"links":               types.SetType{ElemType: rRackLink{}.attrType()},
		"tag_ids":             types.SetType{ElemType: types.StringType},
		"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
	}
}

func (o rRackTypeGenericSystem) attrType() attr.Type {
	return types.ObjectType{AttrTypes: o.attrTypes()}
}

func (o *rRackTypeGenericSystem) request(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) *goapstra.RackElementGenericSystemRequest {
	var poIdMinVal, poIdMaxVal int
	if o.PortChannelIdMin != nil {
		poIdMinVal = int(*o.PortChannelIdMin)
	}
	if o.PortChannelIdMax != nil {
		poIdMaxVal = int(*o.PortChannelIdMax)
	}

	linkRequests := make([]goapstra.RackLinkRequest, len(o.Links))
	for i, link := range o.Links {
		lagMode := goapstra.RackLinkLagModeActive.String()
		link.LagMode = &lagMode
		linkRequests[i] = *link.request(ctx, path.AtListIndex(i), rack, diags)
	}

	var tagIds []goapstra.ObjectId
	if o.TagIds != nil {
		tagIds = make([]goapstra.ObjectId, len(o.TagIds))
		for i, tagId := range o.TagIds {
			tagIds[i] = goapstra.ObjectId(tagId)
		}
	}

	return &goapstra.RackElementGenericSystemRequest{
		Count:            int(o.Count),
		AsnDomain:        goapstra.FeatureSwitchDisabled,
		ManagementLevel:  goapstra.GenericSystemUnmanaged,
		PortChannelIdMin: poIdMinVal,
		PortChannelIdMax: poIdMaxVal,
		Loopback:         goapstra.FeatureSwitchDisabled,
		Tags:             tagIds,
		Label:            o.Name,
		Links:            linkRequests,
		LogicalDeviceId:  goapstra.ObjectId(o.LogicalDeviceId),
	}
}

func (o *rRackTypeGenericSystem) validateConfig(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) {
	if o.PortChannelIdMin != nil && o.PortChannelIdMax == nil {
		diags.AddAttributeError(path, errInvalidConfig, "'port_channel_id_min' requires 'port_channel_id_max'")
	}
	if o.PortChannelIdMax != nil && o.PortChannelIdMin == nil {
		diags.AddAttributeError(path, errInvalidConfig, "'port_channel_id_max' requires 'port_channel_id_min'")
	}

	if o.PortChannelIdMin != nil && o.PortChannelIdMax != nil && *o.PortChannelIdMin > *o.PortChannelIdMax {
		diags.AddAttributeError(path, errInvalidConfig, "port_channel_id_min > port_channel_id_max")
	}

	for i, link := range o.Links {
		link.validateConfigForGenericSystem(ctx, rack, path.AtListIndex(i), diags)
	}
}

func (o *rRackTypeGenericSystem) parseApi(in *goapstra.RackElementGenericSystem) {
	o.Name = in.Label
	o.Count = int64(in.Count)
	portChannelIdMin := int64(in.PortChannelIdMin)
	portChannelIdMax := int64(in.PortChannelIdMax)
	o.PortChannelIdMin = &portChannelIdMin
	o.PortChannelIdMax = &portChannelIdMax
	o.LogicalDevice.parseApi(in.LogicalDevice)
	o.Links = make([]rRackLink, len(in.Links))

	if len(in.Tags) > 0 {
		o.TagData = make([]tagData, len(in.Tags)) // populated below
		for i := range in.Tags {
			o.TagData[i].parseApi(&in.Tags[i])
		}
	}

	for i := range in.Links {
		o.Links[i].parseApi(&in.Links[i])
	}
}

func (o *rRackTypeGenericSystem) getLinkByName(desired string) *rRackLink {
	for _, link := range o.Links {
		if link.Name == desired {
			return &link
		}
	}
	return nil
}

func (o *rRackTypeGenericSystem) copyWriteOnlyElements(src *rRackTypeGenericSystem, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddWarning(errProviderBug, "rRackTypeGenericSystem.copyWriteOnlyElements: attempt to copy from nil source")
	}
	o.LogicalDeviceId = src.LogicalDeviceId
	o.TagIds = src.TagIds

	for i, link := range o.Links {
		o.Links[i].copyWriteOnlyElements(src.getLinkByName(link.Name), diags)
		if diags.HasError() {
			return
		}
	}
}

type rRackLink struct {
	Name             string    `tfsdk:"name"`
	TargetSwitchName string    `tfsdk:"target_switch_name"`
	LagMode          *string   `tfsdk:"lag_mode"`
	LinksPerSwitch   *int64    `tfsdk:"links_per_switch"`
	Speed            string    `tfsdk:"speed"`
	SwitchPeer       *string   `tfsdk:"switch_peer"`
	TagIds           []string  `tfsdk:"tag_ids"`
	TagData          []tagData `tfsdk:"tag_data"`
}

func (o rRackLink) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":               types.StringType,
			"target_switch_name": types.StringType,
			"lag_mode":           types.StringType,
			"links_per_switch":   types.Int64Type,
			"speed":              types.StringType,
			"switch_peer":        types.StringType,
			"tag_ids":            types.SetType{ElemType: types.StringType},
			"tag_data":           types.SetType{ElemType: tagData{}.attrType()}}}
}

func (o *rRackLink) request(ctx context.Context, path path.Path, rack *rRackType, diags *diag.Diagnostics) *goapstra.RackLinkRequest {
	var err error

	tags := make([]goapstra.ObjectId, len(o.TagIds))
	for i, tag := range o.TagIds {
		tags[i] = goapstra.ObjectId(tag)
	}

	lagMode := goapstra.RackLinkLagModeNone
	if o.LagMode != nil {
		err = lagMode.FromString(*o.LagMode)
		if err != nil {
			diags.AddAttributeError(path, "error parsing lag_mode", err.Error())
			return nil
		}
	}

	switchPeer := goapstra.RackLinkSwitchPeerNone
	if o.SwitchPeer != nil {
		err = switchPeer.FromString(*o.SwitchPeer)
		if err != nil {
			diags.AddAttributeError(path, "error parsing switch_peer", err.Error())
			return nil
		}
	}

	leaf := rack.getLeafSwitchByName(ctx, o.TargetSwitchName, diags)
	access := rack.getAccessSwitchByName(ctx, o.TargetSwitchName, diags)
	if leaf == nil && access == nil {
		diags.AddAttributeError(path, errInvalidConfig,
			fmt.Sprintf("target switch '%s' not found in rack type '%s'", o.TargetSwitchName, rack.Id))
		return nil
	}
	if leaf != nil && access != nil {
		diags.AddError(errProviderBug, "link seems to be attached to both leaf and access switches")
		return nil
	}

	upstreamRedundancyProtocol := rack.getSwitchRedundancyProtocolByName(ctx, o.TargetSwitchName, path, diags)
	if diags.HasError() {
		return nil
	}

	linksPerSwitch := 1
	if o.LinksPerSwitch != nil {
		linksPerSwitch = int(*o.LinksPerSwitch)
	}

	return &goapstra.RackLinkRequest{
		Label:              o.Name,
		Tags:               tags,
		LinkPerSwitchCount: linksPerSwitch,
		LinkSpeed:          goapstra.LogicalDevicePortSpeed(o.Speed),
		TargetSwitchLabel:  o.TargetSwitchName,
		AttachmentType:     o.linkAttachmentType(upstreamRedundancyProtocol),
		LagMode:            lagMode,
		SwitchPeer:         switchPeer,
	}
}

func (o *rRackLink) linkAttachmentType(upstreamRedundancyMode fmt.Stringer) goapstra.RackLinkAttachmentType {
	switch upstreamRedundancyMode.String() {
	case goapstra.LeafRedundancyProtocolNone.String():
		return goapstra.RackLinkAttachmentTypeSingle
	case goapstra.AccessRedundancyProtocolNone.String():
		return goapstra.RackLinkAttachmentTypeSingle
	}

	if o.LagMode == nil {
		return goapstra.RackLinkAttachmentTypeSingle
	}

	switch *o.LagMode {
	case goapstra.RackLinkLagModeActive.String():
		return goapstra.RackLinkAttachmentTypeDual
	case goapstra.RackLinkLagModePassive.String():
		return goapstra.RackLinkAttachmentTypeDual
	case goapstra.RackLinkLagModeStatic.String():
		return goapstra.RackLinkAttachmentTypeDual
	}
	return goapstra.RackLinkAttachmentTypeSingle
}

func (o *rRackLink) validateConfigForAccessSwitch(ctx context.Context, arp goapstra.AccessRedundancyProtocol, rack *rRackType, errPath path.Path, diags *diag.Diagnostics) {
	if o.LagMode != nil {
		diags.AddAttributeError(errPath, errInvalidConfig, "'lag_mode' not permitted on Access Switch links")
		return
	}

	leaf := rack.getLeafSwitchByName(ctx, o.TargetSwitchName, diags)
	if leaf == nil {
		diags.AddAttributeError(errPath, "leaf switch not found",
			fmt.Sprintf("leaf switch '%s' not found in rack type '%s'", o.TargetSwitchName, rack.Id))
		return
	}
	if diags.HasError() {
		return
	}

	lrp := goapstra.LeafRedundancyProtocolNone
	if leaf.RedundancyProtocol != nil {
		err := lrp.FromString(*leaf.RedundancyProtocol)
		if err != nil {
			diags.AddAttributeError(errPath,
				fmt.Sprintf("error parsing leaf switch redundancy protocol '%s'", *leaf.RedundancyProtocol),
				err.Error())
		}
	}

	if arp == goapstra.AccessRedundancyProtocolEsi &&
		lrp != goapstra.LeafRedundancyProtocolEsi {
		diags.AddAttributeError(errPath, errInvalidConfig,
			"ESI access switches only support connection to ESI leafs")
		return
	}

	if o.SwitchPeer != nil && // primary/secondary has been selected ...and...
		lrp == goapstra.LeafRedundancyProtocolNone { // upstream is not ESI/MLAG
		diags.AddAttributeError(errPath, errInvalidConfig,
			"'switch_peer' must not be set when upstream switch is non-redundant")
	}
}

func (o *rRackLink) validateConfigForGenericSystem(ctx context.Context, rack *rRackType, errPath path.Path, diags *diag.Diagnostics) {
	lagMode := goapstra.RackLinkLagModeNone
	if o.LagMode != nil {
		err := lagMode.FromString(*o.LagMode)
		if err != nil {
			diags.AddAttributeError(errPath, "error parsing lag mode", err.Error())
		}
	}

	linksPerSwitch := int64(1)
	if o.LinksPerSwitch != nil {
		linksPerSwitch = *o.LinksPerSwitch
	}
	if lagMode == goapstra.RackLinkLagModeNone && linksPerSwitch > 1 {
		diags.AddAttributeError(errPath, errInvalidConfig, "'lag_mode' must be set when 'links_per_switch' is set")
	}

	leaf := rack.getLeafSwitchByName(ctx, o.TargetSwitchName, diags)
	access := rack.getAccessSwitchByName(ctx, o.TargetSwitchName, diags)
	if leaf == nil && access == nil {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("target switch '%s' not found in rack type '%s'", o.TargetSwitchName, rack.Id))
		return
	}
	if leaf != nil && access != nil {
		diags.AddError(errProviderBug, "link seems to be attached to both leaf and access switches")
		return
	}

	var targetSwitchIsRedundant bool
	if leaf != nil {
		targetSwitchIsRedundant = leaf.isRedundant()
	}
	if access != nil {
		targetSwitchIsRedundant = access.isRedundant()
	}

	if !targetSwitchIsRedundant && o.SwitchPeer != nil {
		diags.AddAttributeError(errPath.AtMapKey("switch_peer"), errInvalidConfig,
			"links to non-redundant switches must not specify 'switch_peer'")
	}

	if targetSwitchIsRedundant && (o.SwitchPeer == nil && o.LagMode == nil) {
		diags.AddAttributeError(errPath.AtMapKey("switch_peer"), errInvalidConfig,
			"links to redundant switches must specify 'switch_peer' or 'lag_mode'")
	}
}

func (o *rRackLink) parseApi(in *goapstra.RackLink) {
	o.Name = in.Label
	o.TargetSwitchName = in.TargetSwitchLabel
	if in.LagMode != goapstra.RackLinkLagModeNone {
		lagMode := in.LagMode.String()
		o.LagMode = &lagMode
	}
	linksPerSwitchCount := int64(in.LinkPerSwitchCount)
	o.LinksPerSwitch = &linksPerSwitchCount
	o.Speed = string(in.LinkSpeed)
	if in.SwitchPeer != goapstra.RackLinkSwitchPeerNone {
		switchPeer := in.SwitchPeer.String()
		o.SwitchPeer = &switchPeer
	}

	if len(in.Tags) > 0 {
		o.TagData = make([]tagData, len(in.Tags)) // populated below
		for i := range in.Tags {
			o.TagData[i].parseApi(&in.Tags[i])
		}
	}
}

func (o *rRackLink) copyWriteOnlyElements(src *rRackLink, diags *diag.Diagnostics) {
	if src == nil {
		diags.AddWarning(errProviderBug, "rRackLink.copyWriteOnlyElements: attempt to copy from nil source")
	}
	o.TagIds = src.TagIds
}

func (o *rRackType) validateConfigLeafSwitches(ctx context.Context, errPath path.Path, diags *diag.Diagnostics) {
	var leafSwitches []rRackTypeLeafSwitch
	d := o.LeafSwitches.ElementsAs(ctx, &leafSwitches, true)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	for _, leafSwitch := range leafSwitches {
		setVal, d := types.ObjectValueFrom(ctx, leafSwitch.attrTypes(), &leafSwitch)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		leafSwitch.validateConfig(errPath.AtSetValue(setVal), o, diags)
	}
}

func (o *rRackType) validateConfigAccessSwitches(ctx context.Context, errPath path.Path, diags *diag.Diagnostics) {
	var accessSwitches []rRackTypeAccessSwitch
	d := o.AccessSwitches.ElementsAs(ctx, &accessSwitches, true)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	for _, accessSwitch := range accessSwitches {
		setVal, d := types.ObjectValueFrom(ctx, accessSwitch.attrTypes(), &accessSwitch)
		diags.Append(d...)
		if diags.HasError() {
			return
		}

		accessSwitch.validateConfig(ctx, errPath.AtSetValue(setVal), o, diags)
	}
}

func (o *rRackType) validateConfigGenericSystems(ctx context.Context, errPath path.Path, diags *diag.Diagnostics) {
	var genericSystems []rRackTypeGenericSystem
	d := o.GenericSystems.ElementsAs(ctx, &genericSystems, true)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	for _, genericSystem := range genericSystems {
		setVal, d := types.ObjectValueFrom(ctx, genericSystem.attrTypes(), &genericSystem)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		genericSystem.validateConfig(ctx, errPath.AtSetValue(setVal), o, diags)
	}
}

func (o *rRackType) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.RackTypeRequest {
	var fcd goapstra.FabricConnectivityDesign
	err := fcd.FromString(o.FabricConnectivityDesign.ValueString())
	if err != nil {
		diags.AddAttributeError(path.Root("fabric_connectivity_design"),
			"error parsing fabric_connectivity_design", err.Error())
		return nil
	}

	leafSwitches := o.leafSwitches(ctx, diags)
	accessSwitches := o.accessSwitches(ctx, diags)
	genericSystems := o.genericSystems(ctx, diags)
	if diags.HasError() {
		return nil
	}

	leafSwitchRequests := make([]goapstra.RackElementLeafSwitchRequest, len(leafSwitches))
	for i, leafSwitch := range leafSwitches {
		setVal, d := types.ObjectValueFrom(ctx, leafSwitch.attrTypes(), &leafSwitch)
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}
		leafSwitchRequests[i] = *leafSwitch.request(path.Root("leaf_switches").AtSetValue(setVal), diags)
	}

	accessSwitchRequests := make([]goapstra.RackElementAccessSwitchRequest, len(accessSwitches))
	for i, accessSwitch := range accessSwitches {
		setVal, d := types.ObjectValueFrom(ctx, accessSwitch.attrTypes(), &accessSwitch)
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}
		accessSwitchRequests[i] = *accessSwitch.request(ctx, path.Root("access_switches").AtSetValue(setVal), o, diags)
	}

	genericSystemsRequests := make([]goapstra.RackElementGenericSystemRequest, len(genericSystems))
	for i, genericSystem := range genericSystems {
		setVal, d := types.ObjectValueFrom(ctx, genericSystem.attrTypes(), &genericSystem)
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}
		genericSystemsRequests[i] = *genericSystem.request(ctx, path.Root("generic_systems").AtSetValue(setVal), o, diags)
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

func (o *rRackType) leafSwitches(ctx context.Context, diags *diag.Diagnostics) []rRackTypeLeafSwitch {
	var leafSwitches []rRackTypeLeafSwitch
	d := o.LeafSwitches.ElementsAs(ctx, &leafSwitches, true)
	diags.Append(d...)
	return leafSwitches
}

func (o *rRackType) leafSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackTypeLeafSwitch {
	leafSwitches := o.leafSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}
	for _, leafSwitch := range leafSwitches {
		if leafSwitch.Name == requested {
			return &leafSwitch
		}
	}
	return nil
}

func (o *rRackType) accessSwitches(ctx context.Context, diags *diag.Diagnostics) []rRackTypeAccessSwitch {
	var accessSwitches []rRackTypeAccessSwitch
	d := o.AccessSwitches.ElementsAs(ctx, &accessSwitches, true)
	diags.Append(d...)
	return accessSwitches
}

func (o *rRackType) accessSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackTypeAccessSwitch {
	accessSwitches := o.accessSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}
	for _, accessSwitch := range accessSwitches {
		if accessSwitch.Name == requested {
			return &accessSwitch
		}
	}
	return nil
}

func (o *rRackType) genericSystems(ctx context.Context, diags *diag.Diagnostics) []rRackTypeGenericSystem {
	var genericSystems []rRackTypeGenericSystem
	d := o.AccessSwitches.ElementsAs(ctx, &genericSystems, true)
	diags.Append(d...)
	return genericSystems
}

func (o *rRackType) genericSystemByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackTypeGenericSystem {
	genericSystems := o.genericSystems(ctx, diags)
	if diags.HasError() {
		return nil
	}
	for _, genericSystem := range genericSystems {
		if genericSystem.Name == requested {
			return &genericSystem
		}
	}
	return nil
}

// fcdModes returns permitted fabric_connectivity_design mode strings
func fcdModes() []string {
	return []string{
		goapstra.FabricConnectivityDesignL3Clos.String(),
		goapstra.FabricConnectivityDesignL3Collapsed.String()}
}

// leafRedundancyModes returns permitted fabric_connectivity_design mode strings
func leafRedundancyModes() []string {
	return []string{
		goapstra.LeafRedundancyProtocolEsi.String(),
		goapstra.LeafRedundancyProtocolMlag.String()}
}

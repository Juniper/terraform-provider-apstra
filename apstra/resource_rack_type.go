package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Validators:          []tfsdk.AttributeValidator{listvalidator.SizeAtLeast(1)},
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Switch name, used when creating intra-rack links.",
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
					"tag_names":      tagLabelsAttributeSchema(),
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
								MarkdownDescription: "Peer link port-channel ID.",
								Required:            true,
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
								MarkdownDescription: "L3 peer link port-channel ID.",
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

	config.validateConfigLeafSwitches(ctx, path.Root("leaf_switches"), &resp.Diagnostics)
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

	// force values as needed
	plan.forceValues(&resp.Diagnostics)
	if diags.HasError() {
		return
	}

	// populate rack elements (leaf/access/generic) from global catalog
	plan.populateDataFromGlobalCatalog(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare a goapstra.RackTypeRequest
	rtReq := plan.goapstraRequest(&resp.Diagnostics)
	if diags.HasError() {
		return
	}

	//d, _ := json.Marshal(rtReq)
	//resp.Diagnostics.AddWarning("rtReq", string(d))
	//
	// send the request to Apstra
	id, err := o.client.CreateRackType(ctx, rtReq)
	if err != nil {
		resp.Diagnostics.AddError("error creating rack type", err.Error())
		return
	}

	plan.Id = types.String{Value: string(id)}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

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

	validateRackType(rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState rRackType
	newState.parseApiResponse(rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// force values as needed
	newState.forceValues(&resp.Diagnostics)
	if diags.HasError() {
		return
	}

	newState.copyWriteOnlyElements(&state, &resp.Diagnostics)

	oldJson, _ := json.Marshal(&state)
	newJson, _ := json.Marshal(&newState)
	resp.Diagnostics.AddWarning("old", string(oldJson))
	resp.Diagnostics.AddWarning("new", string(newJson))

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (o *resourceRackType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	rt, err := o.client.GetRackType(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error retrieving Rack Type", err.Error())
		return
	}

	_ = rt
	panic("implement me")
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

	err := o.client.DeleteRackType(ctx, goapstra.ObjectId(state.Id.Value))
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
	LeafSwitches             types.List   `tfsdk:"leaf_switches"`
}

func (o *rRackType) populateDataFromGlobalCatalog(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	o.populateLeafSwitchesDataFromGlobalCatalog(ctx, client, diags)
}

func (o *rRackType) populateLeafSwitchesDataFromGlobalCatalog(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	for i := range o.LeafSwitches.Elems {
		o.populateLeafSwitchDataFromGlobalCatalog(ctx, i, client, diags)
	}
}

func (o *rRackType) populateLeafSwitchDataFromGlobalCatalog(ctx context.Context, idx int, client *goapstra.Client, diags *diag.Diagnostics) {
	o.populateLeafSwitchLogicalDeviceFromGlobalCatalog(ctx, idx, client, diags)
	o.populateLeafSwitchTagsDataFromGlobalCatalog(ctx, idx, client, diags)
}

func (o *rRackType) populateLeafSwitchLogicalDeviceFromGlobalCatalog(ctx context.Context, idx int, client *goapstra.Client, diags *diag.Diagnostics) {
	id := o.LeafSwitches.Elems[idx].(types.Object).Attrs["logical_device_id"].(types.String).Value
	errPath := path.Root("leaf_switches").AtListIndex(idx)
	o.LeafSwitches.Elems[idx].(types.Object).Attrs["logical_device"] = getLogicalDeviceObj(ctx, client, id, errPath, diags)
	if diags.HasError() {
		return
	}
}

func (o *rRackType) populateLeafSwitchTagsDataFromGlobalCatalog(ctx context.Context, idx int, client *goapstra.Client, diags *diag.Diagnostics) {
	tagNamesSet := o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_names"].(types.Set)

	// if tagNamesSet is Null, then tag data will also be null
	if tagNamesSet.IsNull() {
		o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_data"] = newTagDataSet(0)
		return
	}

	// extract a slice of tag names (labels) used by leaf[idx]
	var tagNameStrings []string
	//tagNamesSet.ElementsAs(ctx, &tagNames, false) // todo: can this replace the for loop below?
	for _, elem := range tagNamesSet.Elems {
		if elem.IsNull() {
			continue
		}
		tagNameStrings = append(tagNameStrings, elem.(types.String).Value)
	}

	// fetch the relevant tags from Apstra
	tags, err := client.GetTagsByLabels(ctx, tagNameStrings)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			diags.AddAttributeError(
				path.Root("leaf_switches").AtListIndex(idx),
				"tag not found",
				fmt.Sprintf("at least one of the requested tags does not exist: '%s'",
					strings.Join(tagNameStrings, "', '")),
			)
			return
		}
		diags.AddError("error requesting tag data", err.Error())
	}

	// assign the tag data to 'o'
	o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_data"] = sliceTagToSetObject(tags)
}

func (o *rRackType) validateConfigLeafSwitches(ctx context.Context, errPath path.Path, diags *diag.Diagnostics) {
	for i := range o.LeafSwitches.Elems {
		o.validateConfigLeafSwitch(ctx, i, errPath.AtListIndex(i), diags)
	}
}

func (o *rRackType) validateConfigLeafSwitch(ctx context.Context, idx int, errPath path.Path, diags *diag.Diagnostics) {
	o.validateLeafForFabricConnectivityDesign(ctx, idx, errPath, diags)
	if diags.HasError() {
		return
	}

	o.validateLeafMlagInfo(ctx, idx, errPath.AtMapKey("mlag_info"), diags)
	if diags.HasError() {
		return
	}
}

func (o *rRackType) validateLeafForFabricConnectivityDesign(_ context.Context, idx int, errPath path.Path, diags *diag.Diagnostics) {
	// check leaf switch for compatibility with fabric connectivity design
	switch o.FabricConnectivityDesign.Value {
	case goapstra.FabricConnectivityDesignL3Clos.String():
		o.validateLeafForL3Clos(idx, errPath, diags)
	case goapstra.FabricConnectivityDesignL3Collapsed.String():
		o.validateLeafForL3Collapsed(idx, errPath, diags)
	default:
		diags.AddAttributeError(errPath, errProviderBug,
			fmt.Sprintf("unknown fabric connectivity design '%s'", o.FabricConnectivityDesign))
	}
}

func (o *rRackType) validateLeafForL3Clos(idx int, errPath path.Path, diags *diag.Diagnostics) {
	leafObj := o.LeafSwitches.Elems[idx].(types.Object)

	if leafObj.Attrs["spine_link_count"].IsNull() {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'spine_link_count' must be specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Clos))
	}

	if leafObj.Attrs["spine_link_speed"].IsNull() {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'spine_link_speed' must be specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Clos))
	}
}

func (o *rRackType) validateLeafForL3Collapsed(idx int, errPath path.Path, diags *diag.Diagnostics) {
	leafObj := o.LeafSwitches.Elems[idx].(types.Object)

	if !leafObj.Attrs["spine_link_count"].IsNull() {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'spine_link_count' must not be specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Collapsed))
	}

	if !leafObj.Attrs["spine_link_speed"].IsNull() {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'spine_link_speed' must bnot e specified when 'fabric_connectivity_design' is '%s'",
				goapstra.FabricConnectivityDesignL3Collapsed))
	}

	if leafObj.Attrs["redundancy_protocol"].(types.String).Value == goapstra.LeafRedundancyProtocolMlag.String() {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'redundancy_protocol' = '%s' is not allowed when 'fabric_connectivity_design' = '%s'",
				goapstra.LeafRedundancyProtocolMlag, goapstra.FabricConnectivityDesignL3Collapsed))
	}
}

func (o *rRackType) validateLeafMlagInfo(_ context.Context, idx int, errPath path.Path, diags *diag.Diagnostics) {
	leafObj := o.LeafSwitches.Elems[idx].(types.Object)
	mlagInfo := leafObj.Attrs["mlag_info"].(types.Object)
	redundancyProtocol := leafObj.Attrs["redundancy_protocol"].(types.String)

	if mlagInfo.IsNull() &&
		redundancyProtocol.Value == goapstra.LeafRedundancyProtocolMlag.String() {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'mlag_info' required with 'redundancy_protocol' = '%s'", redundancyProtocol.Value))
	}

	if mlagInfo.IsNull() {
		return
	}

	if redundancyProtocol.Value != goapstra.LeafRedundancyProtocolMlag.String() {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'mlag_info' incompatible with 'redundancy_protocol of '%s'", redundancyProtocol.Value))
	}

	l2LinkPoId := mlagInfo.Attrs["peer_link_port_channel_id"].(types.Int64)
	l3LinkPoId := mlagInfo.Attrs["l3_peer_link_port_channel_id"].(types.Int64)

	if !l2LinkPoId.IsNull() && !l3LinkPoId.IsNull() &&
		l2LinkPoId.Value == l3LinkPoId.Value &&
		l2LinkPoId.Value != 0 {
		diags.AddAttributeError(errPath, errInvalidConfig,
			fmt.Sprintf("'peer_link_port_channel_id' and 'l3_peer_link_port_channel_id' cannot both use value %d",
				l2LinkPoId.Value))
	}

	l3LinkCount := mlagInfo.Attrs["l3_peer_link_count"].(types.Int64)
	l3LinkSpeed := mlagInfo.Attrs["l3_peer_link_speed"].(types.String)

	if l3LinkCount.IsNull() && !l3LinkSpeed.IsNull() {
		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_speed' requires 'l3_peer_link_count'")
	}

	if l3LinkCount.IsNull() && !l3LinkPoId.IsNull() {
		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_port_channel_id' requires 'l3_peer_link_count'")
	}

	if !l3LinkCount.IsNull() && l3LinkSpeed.IsNull() {
		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_count' requires 'l3_peer_link_speed'")
	}

	if !l3LinkCount.IsNull() && l3LinkPoId.IsNull() {
		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_count' requires 'l3_peer_link_port_channel_id'")
	}
}

func (o *rRackType) parseApiResponse(rt *goapstra.RackType, diags *diag.Diagnostics) {
	o.Id = types.String{Value: string(rt.Id)}
	o.Name = types.String{Value: rt.Data.DisplayName}
	o.Description = types.String{Value: rt.Data.Description}
	o.FabricConnectivityDesign = types.String{Value: rt.Data.FabricConnectivityDesign.String()}
	o.parseApiResponseLeafSwitches(rt.Data.LeafSwitches, diags)
	//o.AccessSwitches =           parseRackTypeAccessSwitches(rt.Data.AccessSwitches, diags)
	//o.GenericSystems =           parseRackTypeGenericSystems(rt.Data.GenericSystems, diags)
}

func (o *rRackType) parseApiResponseLeafSwitches(in []goapstra.RackElementLeafSwitch, diags *diag.Diagnostics) {
	o.LeafSwitches = newLeafSwitchList(len(in)) // this uses data source code?
	for i, ls := range in {
		o.parseApiResponseLeafSwitch(&ls, i, diags)
	}
}

func (o *rRackType) parseApiResponseLeafSwitch(in *goapstra.RackElementLeafSwitch, idx int, _ *diag.Diagnostics) {
	o.LeafSwitches.Elems[idx] = types.Object{
		AttrTypes: rLeafSwitchAttrTypes(),
		Attrs: map[string]attr.Value{
			"name":                types.String{Value: in.Label},
			"spine_link_count":    parseApiLeafSwitchLinkPerSpineCountToTypesInt64(in),
			"spine_link_speed":    parseApiLeafSwitchLinkPerSpineSpeedToTypesString(in),
			"redundancy_protocol": parseApiLeafRedundancyProtocolToTypesString(in),
			"logical_device":      parseApiLogicalDeviceToTypesObject(in.LogicalDevice),
			"mlag_info":           parseApiLeafMlagInfoToTypesObject(in.MlagInfo),
			"tag_names":           parseSliceApiTagDataToTypesSetString(in.Tags),
			"tag_data":            parseApiSliceTagDataToTypesSetObject(in.Tags),
		},
	}
}

func (o *rRackType) copyWriteOnlyElements(orig *rRackType, diags *diag.Diagnostics) {
	o.copyWriteOnlyElementsLeafSwitches(orig, diags)
}

func (o *rRackType) copyWriteOnlyElementsLeafSwitches(orig *rRackType, diags *diag.Diagnostics) {
	for i, ls := range orig.LeafSwitches.Elems {
		o.copyWriteOnlyElementsLeafSwitch(ls.(types.Object), i, diags)
	}
}

func (o *rRackType) copyWriteOnlyElementsLeafSwitch(orig types.Object, idx int, _ *diag.Diagnostics) {
	logicalDeviceId := orig.Attrs["logical_device_id"].(types.String).Value
	o.LeafSwitches.Elems[idx].(types.Object).Attrs["logical_device_id"] = types.String{Value: logicalDeviceId}
}

func getLogicalDeviceObj(ctx context.Context, client *goapstra.Client, id string, errPath path.Path, diags *diag.Diagnostics) types.Object {
	logicalDevice, err := client.GetLogicalDevice(ctx, goapstra.ObjectId(id))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if ace.Type() == goapstra.ErrNotfound {
			diags.AddAttributeError(errPath, "logical device not found",
				fmt.Sprintf("logical device '%s' does not exist", id))
		}
		diags.AddError(fmt.Sprintf("error retrieving logical device '%s'", id), err.Error())
	}
	return parseApiLogicalDeviceToTypesObject(logicalDevice.Data)
}

// forceValues handles user-optional values and values which are required by
// Apstra, but which we can predict so we don't want to bother the user.
func (o *rRackType) forceValues(diags *diag.Diagnostics) {
	// handle "description" omitted from config
	if o.Description.Unknown {
		o.Description = types.String{Null: true}
	}

	// handle empty "description" from API
	if !o.Description.IsUnknown() && !o.Description.IsNull() && o.Description.Value == "" {
		o.Description = types.String{Null: true}
	}

	// force leaf switch values as needed
	o.forceLeafSwitchesValues(diags)
}

func (o *rRackType) forceLeafSwitchesValues(diags *diag.Diagnostics) {
	for i := range o.LeafSwitches.Elems {
		o.forceLeafSwitchValues(i, diags)
	}
}

func (o *rRackType) forceLeafSwitchValues(idx int, _ *diag.Diagnostics) {
	leafSwitchObj := o.LeafSwitches.Elems[idx].(types.Object)
	//tagNames := leafSwitchObj.Attrs["tag_names"].(types.Set)
	//tagData := leafSwitchObj.Attrs["tag_data"].(types.Set)

	// handle "tag_names" omitted from config
	if leafSwitchObj.Attrs["tag_names"].IsUnknown() {
		o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_names"] = types.Set{
			ElemType: types.StringType,
			Null:     true,
		}
		o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_data"] = types.Set{
			Null:     true,
			ElemType: types.ObjectType{AttrTypes: tagDataAttrTypes()},
		}
	}

	////handle empty "tag_data" from API
	//if !tagData.Unknown && !tagData.Null && len(tagData.Elems) == 0 {
	//	o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_data"] = types.Set{
	//		Null: true,
	//		ElemType: types.ObjectType{AttrTypes: tagDataAttrTypes()},
	//	}
	//}

	//// handle always-empty "tag_names" from API
	//if o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_data"].IsNull() {
	//	o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_names"] = types.Set{
	//		ElemType: types.StringType,
	//		Null:     true,
	//	}
	//} else {
	//	o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_names"] = types.Set{
	//		ElemType: types.StringType,
	//		Elems:
	//	}
	//}

	//d, _ := json.Marshal(leafSwitchObj.Attrs["tag_names"])
	//diags.AddWarning("tag_names", string(d))

	switch o.FabricConnectivityDesign.Value {
	case goapstra.FabricConnectivityDesignL3Clos.String():
		// nothing yet
	case goapstra.FabricConnectivityDesignL3Collapsed.String():
		// spine link info must be null with collapsed fabric
		o.LeafSwitches.Elems[idx].(types.Object).Attrs["spine_link_count"] = types.Int64{Null: true}
		o.LeafSwitches.Elems[idx].(types.Object).Attrs["spine_link_speed"] = types.String{Null: true}
	}
}

func (o *rRackType) renderFabricConnectivityDesign() goapstra.FabricConnectivityDesign {
	switch o.FabricConnectivityDesign.Value {
	case goapstra.FabricConnectivityDesignL3Collapsed.String():
		return goapstra.FabricConnectivityDesignL3Collapsed
	default:
		return goapstra.FabricConnectivityDesignL3Clos
	}
}

func (o *rRackType) goapstraRequest(diags *diag.Diagnostics) *goapstra.RackTypeRequest {
	return &goapstra.RackTypeRequest{
		DisplayName:              o.Name.Value,
		Description:              o.Description.Value,
		FabricConnectivityDesign: o.renderFabricConnectivityDesign(),
		LeafSwitches:             o.leafSwitchRequests(diags),
		//AccessSwitches:           o.accessSwitchRequests(diags),
		//GenericSystems:           o.genericSystemRequests(diags),
	}
}

// fcdModes returns permitted fabric_connectivity_design mode strings
func fcdModes() []string {
	return []string{
		goapstra.FabricConnectivityDesignL3Clos.String(),
		goapstra.FabricConnectivityDesignL3Collapsed.String()}
}

func (o *rRackType) leafSwitchRequests(_ *diag.Diagnostics) []goapstra.RackElementLeafSwitchRequest {
	result := make([]goapstra.RackElementLeafSwitchRequest, len(o.LeafSwitches.Elems))
	for i, leafSwitchListElem := range o.LeafSwitches.Elems {
		leafSwitchObj := leafSwitchListElem.(types.Object)
		result[i] = goapstra.RackElementLeafSwitchRequest{
			Label:              leafSwitchObj.Attrs["name"].(types.String).Value,
			LogicalDeviceId:    goapstra.ObjectId(leafSwitchObj.Attrs["logical_device_id"].(types.String).Value),
			LinkPerSpineCount:  int(leafSwitchObj.Attrs["spine_link_count"].(types.Int64).Value),
			LinkPerSpineSpeed:  goapstra.LogicalDevicePortSpeed(leafSwitchObj.Attrs["spine_link_speed"].(types.String).Value),
			RedundancyProtocol: renderLeafRedundancyProtocol(leafSwitchObj.Attrs["redundancy_protocol"].(types.String)),
			Tags:               renderSliceStringsFromSetStrings(leafSwitchObj.Attrs["tag_names"].(types.Set)),
			MlagInfo:           renderLeafMlagInfo(leafSwitchObj.Attrs["mlag_info"].(types.Object)),
		}
	}
	return result
}

// leafRedundancyModes returns permitted fabric_connectivity_design mode strings
func leafRedundancyModes() []string {
	return []string{
		goapstra.LeafRedundancyProtocolEsi.String(),
		goapstra.LeafRedundancyProtocolMlag.String()}
}

func renderLeafRedundancyProtocol(in types.String) goapstra.LeafRedundancyProtocol {
	if in.IsNull() {
		return goapstra.LeafRedundancyProtocolNone
	}
	switch in.Value {
	case goapstra.LeafRedundancyProtocolEsi.String():
		return goapstra.LeafRedundancyProtocolEsi
	case goapstra.LeafRedundancyProtocolMlag.String():
		return goapstra.LeafRedundancyProtocolMlag
	}
	return goapstra.LeafRedundancyProtocolNone
}

func renderLeafMlagInfo(mlagInfo types.Object) *goapstra.LeafMlagInfo {
	if mlagInfo.IsNull() {
		return nil
	}

	return &goapstra.LeafMlagInfo{
		MlagVlanId:                  int(mlagInfo.Attrs["mlag_keepalive_vlan"].(types.Int64).Value),
		LeafLeafLinkCount:           int(mlagInfo.Attrs["peer_link_count"].(types.Int64).Value),
		LeafLeafLinkSpeed:           goapstra.LogicalDevicePortSpeed(mlagInfo.Attrs["peer_link_speed"].(types.String).Value),
		LeafLeafLinkPortChannelId:   int(mlagInfo.Attrs["peer_link_port_channel_id"].(types.Int64).Value),
		LeafLeafL3LinkCount:         int(mlagInfo.Attrs["l3_peer_link_count"].(types.Int64).Value),
		LeafLeafL3LinkSpeed:         goapstra.LogicalDevicePortSpeed(mlagInfo.Attrs["l3_peer_link_speed"].(types.String).Value),
		LeafLeafL3LinkPortChannelId: int(mlagInfo.Attrs["l3_peer_link_port_channel_id"].(types.Int64).Value),
	}
}

func rLeafSwitchAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"redundancy_protocol": types.StringType,
		"logical_device_id":   types.StringType,
		"logical_device": types.ObjectType{
			AttrTypes: logicalDeviceDataElementAttrTypes()},
		"tag_names": types.SetType{ElemType: types.StringType},
		"tag_data": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: tagDataAttrTypes()}},
		"mlag_info": types.ObjectType{
			AttrTypes: mlagInfoAttrTypes()},
	}
}

func newLeafSwitchList(size int) types.List {
	return types.List{
		Elems:    make([]attr.Value, size),
		ElemType: types.ObjectType{AttrTypes: rLeafSwitchAttrTypes()},
	}
}

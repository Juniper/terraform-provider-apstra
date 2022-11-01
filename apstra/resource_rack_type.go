package apstra

//
//import (
//	"bitbucket.org/apstrktr/goapstra"
//	"context"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
//	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
//	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
//	"github.com/hashicorp/terraform-plugin-framework/attr"
//	"github.com/hashicorp/terraform-plugin-framework/diag"
//	"github.com/hashicorp/terraform-plugin-framework/path"
//	"github.com/hashicorp/terraform-plugin-framework/resource"
//	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
//	"github.com/hashicorp/terraform-plugin-framework/types"
//	"os"
//	"strings"
//)
//
//const (
//	vlanMin = 1
//	vlanMax = 4094
//
//	poIdMin = 1
//	poIdMax = 4096
//)
//
//var _ resource.ResourceWithConfigure = &resourceRackType{}
//var _ resource.ResourceWithValidateConfig = &resourceRackType{}
//
//type resourceRackType struct {
//	client *goapstra.Client
//}
//
//func (o *resourceRackType) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
//	resp.TypeName = req.ProviderTypeName + "_rack_type"
//}
//
//func (o *resourceRackType) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
//	if req.ProviderData == nil {
//		return
//	}
//
//	if pd, ok := req.ProviderData.(*providerData); ok {
//		o.client = pd.client
//	} else {
//		resp.Diagnostics.AddError(
//			errResourceConfigureProviderDataDetail,
//			fmt.Sprintf(errResourceConfigureProviderDataDetail, pd, req.ProviderData),
//		)
//	}
//}
//
//func (o *resourceRackType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
//	return tfsdk.Schema{
//		MarkdownDescription: "This resource creates a Rack Type in the Apstra Design tab.",
//		Attributes: map[string]tfsdk.Attribute{
//			"id": {
//				MarkdownDescription: "Object ID for the Rack Type, assigned by Apstra.",
//				Type:                types.StringType,
//				Computed:            true,
//				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
//			},
//			"name": {
//				MarkdownDescription: "Rack Type name, displayed in the Apstra web UI.",
//				Type:                types.StringType,
//				Required:            true,
//				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
//			},
//			"description": {
//				MarkdownDescription: "Rack Type description, displayed in the Apstra web UI.",
//				Type:                types.StringType,
//				Optional:            true,
//			},
//			"fabric_connectivity_design": {
//				MarkdownDescription: fmt.Sprintf("Must be one of '%s'.", strings.Join(fcdModes(), "', '")),
//				Type:                types.StringType,
//				Required:            true,
//				Validators:          []tfsdk.AttributeValidator{stringvalidator.OneOf(fcdModes()...)},
//			},
//			"leaf_switches": {
//				MarkdownDescription: "Each Rack Type is required to have at least one Leaf Switch.",
//				Required:            true,
//				Validators:          []tfsdk.AttributeValidator{listvalidator.SizeAtLeast(1)},
//				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
//					"name": {
//						MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
//						Type:                types.StringType,
//						Required:            true,
//						Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
//					},
//					"logical_device_id": {
//						MarkdownDescription: "Apstra Object ID of the Logical Device used to model this switch.",
//						Type:                types.StringType,
//						Required:            true,
//					},
//					"spine_link_count": {
//						MarkdownDescription: "Links per spine.",
//						Type:                types.Int64Type,
//						Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
//						Optional:            true,
//					},
//					"spine_link_speed": {
//						MarkdownDescription: "Speed of spine-facing links, something like '10G'",
//						Type:                types.StringType,
//						Optional:            true,
//					},
//					"redundancy_protocol": {
//						MarkdownDescription: fmt.Sprintf("Enabling a redundancy protocol converts a single "+
//							"Leaf Switch into a LAG-capable switch pair. Must be one of '%s'.",
//							strings.Join(leafRedundancyModes(), "', '")),
//						Type:       types.StringType,
//						Optional:   true,
//						Validators: []tfsdk.AttributeValidator{stringvalidator.OneOf(leafRedundancyModes()...)},
//					},
//					"logical_device": logicalDeviceDataAttributeSchema(),
//					"tag_ids":        tagIdsAttributeSchema(),
//					"tag_data":       tagsDataAttributeSchema(),
//					"mlag_info": {
//						MarkdownDescription: fmt.Sprintf("Required when `redundancy_protocol` set to `%s`, "+
//							"defines the connectivity between MLAG peers.", goapstra.LeafRedundancyProtocolMlag.String()),
//						Optional: true,
//						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
//							"mlag_keepalive_vlan": {
//								MarkdownDescription: "MLAG keepalive VLAN ID.",
//								Required:            true,
//								Type:                types.Int64Type,
//								Validators: []tfsdk.AttributeValidator{
//									int64validator.Between(vlanMin, vlanMax),
//								},
//							},
//							"peer_link_count": {
//								MarkdownDescription: "Number of links between MLAG devices.",
//								Required:            true,
//								Type:                types.Int64Type,
//								Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
//							},
//							"peer_link_speed": {
//								MarkdownDescription: "Speed of links between MLAG devices.",
//								Required:            true,
//								Type:                types.StringType,
//							},
//							"peer_link_port_channel_id": {
//								MarkdownDescription: "Port channel number used for L2 Peer Link. Omit to allow Apstra to choose.",
//								Optional:            true,
//								Type:                types.Int64Type,
//								Validators: []tfsdk.AttributeValidator{
//									int64validator.Between(poIdMin, poIdMax),
//								},
//							},
//							"l3_peer_link_count": {
//								MarkdownDescription: "Number of L3 links between MLAG devices.",
//								Optional:            true,
//								Type:                types.Int64Type,
//								Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
//							},
//							"l3_peer_link_speed": {
//								MarkdownDescription: "Speed of l3 links between MLAG devices.",
//								Optional:            true,
//								Type:                types.StringType,
//							},
//							"l3_peer_link_port_channel_id": {
//								MarkdownDescription: "Port channel number used for L3 Peer Link. Omit to allow Apstra to choose.",
//								Optional:            true,
//								Type:                types.Int64Type,
//								Validators: []tfsdk.AttributeValidator{
//									int64validator.Between(poIdMin, poIdMax),
//								},
//							},
//						}),
//					},
//				}),
//			},
//			"access_switches": {
//				MarkdownDescription: "Access switches provide fan-out connectivity from Leaf Switches.",
//				Optional:            true,
//				Validators:          []tfsdk.AttributeValidator{listvalidator.SizeAtLeast(1)},
//				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
//					"name": {
//						MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
//						Type:                types.StringType,
//						Required:            true,
//						Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
//					},
//					"count": {
//						MarkdownDescription: "Number of Access Switches of this type.",
//						Type:                types.Int64Type,
//						Required:            true,
//						Validators: []tfsdk.AttributeValidator{
//							int64validator.AtLeast(1),
//						},
//					},
//					"redundancy_protocol": {
//						MarkdownDescription: "Indicates whether the switch is a redundant pair.",
//						Type:                types.StringType,
//						Computed:            true,
//						PlanModifiers:       tfsdk.AttributePlanModifiers{useStateForUnknownNull()},
//					},
//					"logical_device_id": {
//						MarkdownDescription: "Apstra Object ID of the Logical Device used to model this switch.",
//						Type:                types.StringType,
//						Required:            true,
//					},
//					"logical_device": logicalDeviceDataAttributeSchema(),
//					"links":          rRackLinkAttributeSchema(),
//					"tag_ids":        tagIdsAttributeSchema(),
//					"tag_data":       tagsDataAttributeSchema(),
//					"esi_lag_info": {
//						MarkdownDescription: "Including this stanza converts the Access Switch into a redundant pair.",
//						Optional:            true,
//						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
//							"l3_peer_link_count": {
//								MarkdownDescription: "Number of L3 links between ESI-LAG devices.",
//								Required:            true,
//								Type:                types.Int64Type,
//								Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
//							},
//							"l3_peer_link_speed": {
//								MarkdownDescription: "Speed of l3 links between ESI-LAG devices.",
//								Required:            true,
//								Type:                types.StringType,
//							},
//						}),
//					},
//				}),
//			},
//		},
//	}, nil
//}
//
//func (o *resourceRackType) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
//	if o.client == nil { // cannot proceed without a client
//		return
//	}
//
//	var config rRackType
//	diags := req.Config.Get(ctx, &config)
//	resp.Diagnostics.Append(diags...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	config.validateConfigLeafSwitches(ctx, path.Root("leaf_switches"), &resp.Diagnostics)
//	config.validateConfigAccessSwitches(ctx, path.Root("access_switches"), &resp.Diagnostics)
//}
//
//func (o *resourceRackType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
//	if o.client == nil {
//		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
//		return
//	}
//
//	// Retrieve values from plan
//	var plan rRackType
//	diags := req.Plan.Get(ctx, &plan)
//	resp.Diagnostics.Append(diags...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	// force values as needed
//	plan.forceValues(&resp.Diagnostics)
//	if diags.HasError() {
//		return
//	}
//
//	// populate rack elements (leaf/access/generic) from global catalog
//	plan.populateDataFromGlobalCatalog(ctx, o.client, &resp.Diagnostics)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	// Prepare a goapstra.RackTypeRequest
//	rtReq := plan.goapstraRequest(&resp.Diagnostics)
//	if diags.HasError() {
//		return
//	}
//
//	// send the request to Apstra
//	id, err := o.client.CreateRackType(ctx, rtReq)
//	if err != nil {
//		resp.Diagnostics.AddError("error creating rack type", err.Error())
//		return
//	}
//
//	plan.Id = types.String{Value: string(id)}
//	diags = resp.State.Set(ctx, &plan)
//	resp.Diagnostics.Append(diags...)
//}
//
//func (o *resourceRackType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
//	if o.client == nil {
//		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
//		return
//	}
//
//	// Retrieve values from state
//	var state rRackType
//	diags := req.State.Get(ctx, &state)
//	resp.Diagnostics.Append(diags...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	rt, err := o.client.GetRackType(ctx, goapstra.ObjectId(state.Id.Value))
//	if err != nil {
//		var ace goapstra.ApstraClientErr
//		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
//			resp.State.RemoveResource(ctx)
//			return
//		}
//		resp.Diagnostics.AddError("error reading rack type", err.Error())
//		return
//	}
//
//	validateRackType(rt, &resp.Diagnostics)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	var newState rRackType
//	newState.parseApiResponse(ctx, rt, &resp.Diagnostics)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	// force values as needed
//	newState.forceValues(&resp.Diagnostics)
//	if diags.HasError() {
//		return
//	}
//
//	newState.copyWriteOnlyElements(&state, &resp.Diagnostics)
//
//	diags = resp.State.Set(ctx, &newState)
//	resp.Diagnostics.Append(diags...)
//}
//
//func (o *resourceRackType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
//	if o.client == nil {
//		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
//		return
//	}
//
//	// Retrieve state
//	var state rRackType
//	diags := req.State.Get(ctx, &state)
//	resp.Diagnostics.Append(diags...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	// Retrieve plan
//	var plan rRackType
//	diags = req.Plan.Get(ctx, &plan)
//	resp.Diagnostics.Append(diags...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	planJson, _ := json.MarshalIndent(&plan, "", "  ")
//	os.WriteFile("/tmp/plan", planJson, 0644)
//
//	// force values as needed
//	plan.forceValues(&resp.Diagnostics)
//	if diags.HasError() {
//		return
//	}
//
//	// populate rack elements (leaf/access/generic) from global catalog
//	plan.populateDataFromGlobalCatalog(ctx, o.client, &resp.Diagnostics)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	// Prepare a goapstra.RackTypeRequest
//	rtReq := plan.goapstraRequest(&resp.Diagnostics)
//	if diags.HasError() {
//		return
//	}
//
//	err := o.client.UpdateRackType(ctx, goapstra.ObjectId(state.Id.Value), rtReq)
//	if err != nil {
//		resp.Diagnostics.AddError("error while updating Rack Type", err.Error())
//	}
//
//	diags = resp.State.Set(ctx, &plan)
//	resp.Diagnostics.Append(diags...)
//}
//
//func (o *resourceRackType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
//	if o.client == nil {
//		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
//		return
//	}
//
//	// Retrieve values from state
//	var state rRackType
//	diags := req.State.Get(ctx, &state)
//	resp.Diagnostics.Append(diags...)
//	if resp.Diagnostics.HasError() {
//		return
//	}
//
//	err := o.client.DeleteRackType(ctx, goapstra.ObjectId(state.Id.Value))
//	if err != nil {
//		var ace goapstra.ApstraClientErr
//		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
//			return // 404 is okay in Delete()
//		}
//		resp.Diagnostics.AddError("error deleting Rack Type", err.Error())
//	}
//}
//
//type rRackType struct {
//	Id                       types.String `tfsdk:"id"`
//	Name                     types.String `tfsdk:"name"`
//	Description              types.String `tfsdk:"description"`
//	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
//	LeafSwitches             types.List   `tfsdk:"leaf_switches"`
//	AccessSwitches           types.List   `tfsdk:"access_switches"`
//}
//
//type rRackTypeLeafSwitch struct {
//	Name string `tfsdk:"name"`
//	LogicalDeviceId string `tfsdk:"logical_device_id"`
//	SpineLinkCount *int64 `tfsdk:"spine_link_count"`
//	SpineLinkSpeed *string `tfsdk:"spine_link_speed"`
//	RedundancyProtocol *string `tfsdk:"redundancy_protocol"`
//	LogicalDevice logicalDevice `tfsdk:"logical_device"`
//	TagIds *[]string `tfsdk:"tag_ids"`
//	TagData []tagData `tfsdk:"tag_data"`
//
//	"tag_ids":        tagIdsAttributeSchema(),
//	"tag_data":       tagsDataAttributeSchema(),
//	"mlag_info": {
//	MarkdownDescription: fmt.Sprintf("Required when `redundancy_protocol` set to `%s`, "+
//	"defines the connectivity between MLAG peers.", goapstra.LeafRedundancyProtocolMlag.String()),
//	Optional: true,
//	Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
//	"mlag_keepalive_vlan": {
//	MarkdownDescription: "MLAG keepalive VLAN ID.",
//	Required:            true,
//	Type:                types.Int64Type,
//	Validators: []tfsdk.AttributeValidator{
//	int64validator.Between(vlanMin, vlanMax),
//},
//},
//	"peer_link_count": {
//	MarkdownDescription: "Number of links between MLAG devices.",
//	Required:            true,
//	Type:                types.Int64Type,
//	Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
//},
//	"peer_link_speed": {
//	MarkdownDescription: "Speed of links between MLAG devices.",
//	Required:            true,
//	Type:                types.StringType,
//},
//	"peer_link_port_channel_id": {
//	MarkdownDescription: "Port channel number used for L2 Peer Link. Omit to allow Apstra to choose.",
//	Optional:            true,
//	Type:                types.Int64Type,
//	Validators: []tfsdk.AttributeValidator{
//	int64validator.Between(poIdMin, poIdMax),
//},
//},
//	"l3_peer_link_count": {
//	MarkdownDescription: "Number of L3 links between MLAG devices.",
//	Optional:            true,
//	Type:                types.Int64Type,
//	Validators:          []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
//},
//	"l3_peer_link_speed": {
//	MarkdownDescription: "Speed of l3 links between MLAG devices.",
//	Optional:            true,
//	Type:                types.StringType,
//},
//	"l3_peer_link_port_channel_id": {
//	MarkdownDescription: "Port channel number used for L3 Peer Link. Omit to allow Apstra to choose.",
//	Optional:            true,
//	Type:                types.Int64Type,
//	Validators: []tfsdk.AttributeValidator{
//	int64validator.Between(poIdMin, poIdMax),
//},
//},
//}),
//},
//}),
//},
//
//}
//
//func (o *rRackType) populateDataFromGlobalCatalog(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
//	o.populateLeafSwitchesDataFromGlobalCatalog(ctx, client, path.Root("leaf_switches"), diags)
//	o.populateAccessSwitchesDataFromGlobalCatalog(ctx, client, path.Root("access_switches"), diags)
//}
//
//func (o *rRackType) populateLeafSwitchesDataFromGlobalCatalog(ctx context.Context, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) {
//	for i := range o.LeafSwitches.Elems {
//		o.populateLeafSwitchDataFromGlobalCatalog(ctx, i, client, errPath.AtListIndex(i), diags)
//	}
//}
//
//func (o *rRackType) populateLeafSwitchDataFromGlobalCatalog(ctx context.Context, idx int, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) {
//	o.populateLeafSwitchLogicalDeviceFromGlobalCatalog(ctx, idx, client, errPath.AtMapKey("logical_device"), diags)
//	o.populateLeafSwitchTagsDataFromGlobalCatalog(ctx, idx, client, errPath.AtMapKey("tag_ids"), diags)
//}
//
//func (o *rRackType) populateLeafSwitchLogicalDeviceFromGlobalCatalog(ctx context.Context, idx int, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) {
//	id := o.LeafSwitches.Elems[idx].(types.Object).Attrs["logical_device_id"].(types.String).Value
//	o.LeafSwitches.Elems[idx].(types.Object).Attrs["logical_device"] = getLogicalDeviceObj(ctx, client, id, errPath, diags)
//	if diags.HasError() {
//		return
//	}
//}
//
//func (o *rRackType) populateLeafSwitchTagsDataFromGlobalCatalog(ctx context.Context, idx int, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) {
//	tagIdSet := o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_ids"].(types.Set)
//	tagData := setTagIdStringToSetTagDataObj(ctx, tagIdSet, client, errPath.AtMapKey("tag_ids"), diags)
//	if diags.HasError() {
//		return
//	}
//
//	o.LeafSwitches.Elems[idx].(types.Object).Attrs["tag_data"] = tagData
//}
//
//func (o *rRackType) populateAccessSwitchesDataFromGlobalCatalog(ctx context.Context, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) {
//	for i := range o.AccessSwitches.Elems {
//		o.populateAccessSwitchDataFromGlobalCatalog(ctx, i, client, errPath.AtListIndex(i), diags)
//	}
//}
//
//func (o *rRackType) populateAccessSwitchDataFromGlobalCatalog(ctx context.Context, idx int, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) {
//	o.populateAccessSwitchLogicalDeviceFromGlobalCatalog(ctx, idx, client, errPath, diags)
//	o.populateAccessSwitchTagsDataFromGlobalCatalog(ctx, idx, client, errPath, diags)
//}
//
//func (o *rRackType) populateAccessSwitchLogicalDeviceFromGlobalCatalog(ctx context.Context, idx int, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) {
//	id := o.AccessSwitches.Elems[idx].(types.Object).Attrs["logical_device_id"].(types.String).Value
//	o.AccessSwitches.Elems[idx].(types.Object).Attrs["logical_device"] = getLogicalDeviceObj(ctx, client, id, errPath, diags)
//	if diags.HasError() {
//		return
//	}
//}
//
//func (o *rRackType) populateAccessSwitchTagsDataFromGlobalCatalog(ctx context.Context, idx int, client *goapstra.Client, errPath path.Path, diags *diag.Diagnostics) {
//	tagIdSet := o.AccessSwitches.Elems[idx].(types.Object).Attrs["tag_ids"].(types.Set)
//	tagData := setTagIdStringToSetTagDataObj(ctx, tagIdSet, client, errPath.AtMapKey("tag_ids"), diags)
//	if diags.HasError() {
//		return
//	}
//
//	o.AccessSwitches.Elems[idx].(types.Object).Attrs["tag_data"] = tagData
//}
//
//func (o *rRackType) validateConfigLeafSwitches(ctx context.Context, errPath path.Path, diags *diag.Diagnostics) {
//	for i := range o.LeafSwitches.Elems {
//		o.validateConfigLeafSwitch(ctx, i, errPath.AtListIndex(i), diags)
//	}
//}
//
//func (o *rRackType) validateConfigLeafSwitch(ctx context.Context, idx int, errPath path.Path, diags *diag.Diagnostics) {
//	o.validateLeafForFabricConnectivityDesign(ctx, idx, errPath, diags)
//	if diags.HasError() {
//		return
//	}
//
//	o.validateLeafMlagInfo(ctx, idx, errPath.AtMapKey("mlag_info"), diags)
//	if diags.HasError() {
//		return
//	}
//}
//
//func (o *rRackType) validateLeafForFabricConnectivityDesign(_ context.Context, idx int, errPath path.Path, diags *diag.Diagnostics) {
//	// check leaf switch for compatibility with fabric connectivity design
//	switch o.FabricConnectivityDesign.Value {
//	case goapstra.FabricConnectivityDesignL3Clos.String():
//		o.validateLeafForL3Clos(idx, errPath, diags)
//	case goapstra.FabricConnectivityDesignL3Collapsed.String():
//		o.validateLeafForL3Collapsed(idx, errPath, diags)
//	default:
//		diags.AddAttributeError(errPath, errProviderBug,
//			fmt.Sprintf("unknown fabric connectivity design '%s'", o.FabricConnectivityDesign))
//	}
//}
//
//func (o *rRackType) validateLeafForL3Clos(idx int, errPath path.Path, diags *diag.Diagnostics) {
//	leafObj := o.LeafSwitches.Elems[idx].(types.Object)
//
//	if leafObj.Attrs["spine_link_count"].IsNull() {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			fmt.Sprintf("'spine_link_count' must be specified when 'fabric_connectivity_design' is '%s'",
//				goapstra.FabricConnectivityDesignL3Clos))
//	}
//
//	if leafObj.Attrs["spine_link_speed"].IsNull() {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			fmt.Sprintf("'spine_link_speed' must be specified when 'fabric_connectivity_design' is '%s'",
//				goapstra.FabricConnectivityDesignL3Clos))
//	}
//}
//
//func (o *rRackType) validateLeafForL3Collapsed(idx int, errPath path.Path, diags *diag.Diagnostics) {
//	leafObj := o.LeafSwitches.Elems[idx].(types.Object)
//
//	if !leafObj.Attrs["spine_link_count"].IsNull() {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			fmt.Sprintf("'spine_link_count' must not be specified when 'fabric_connectivity_design' is '%s'",
//				goapstra.FabricConnectivityDesignL3Collapsed))
//	}
//
//	if !leafObj.Attrs["spine_link_speed"].IsNull() {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			fmt.Sprintf("'spine_link_speed' must bnot e specified when 'fabric_connectivity_design' is '%s'",
//				goapstra.FabricConnectivityDesignL3Collapsed))
//	}
//
//	if leafObj.Attrs["redundancy_protocol"].(types.String).Value == goapstra.LeafRedundancyProtocolMlag.String() {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			fmt.Sprintf("'redundancy_protocol' = '%s' is not allowed when 'fabric_connectivity_design' = '%s'",
//				goapstra.LeafRedundancyProtocolMlag, goapstra.FabricConnectivityDesignL3Collapsed))
//	}
//}
//
//func (o *rRackType) validateLeafMlagInfo(_ context.Context, idx int, errPath path.Path, diags *diag.Diagnostics) {
//	leafObj := o.LeafSwitches.Elems[idx].(types.Object)
//	mlagInfo := leafObj.Attrs["mlag_info"].(types.Object)
//	redundancyProtocol := leafObj.Attrs["redundancy_protocol"].(types.String)
//
//	if mlagInfo.IsNull() &&
//		redundancyProtocol.Value == goapstra.LeafRedundancyProtocolMlag.String() {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			fmt.Sprintf("'mlag_info' required with 'redundancy_protocol' = '%s'", redundancyProtocol.Value))
//	}
//
//	if mlagInfo.IsNull() {
//		return
//	}
//
//	if redundancyProtocol.Value != goapstra.LeafRedundancyProtocolMlag.String() {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			fmt.Sprintf("'mlag_info' incompatible with 'redundancy_protocol of '%s'", redundancyProtocol.Value))
//	}
//
//	l2LinkPoId := mlagInfo.Attrs["peer_link_port_channel_id"].(types.Int64)
//	l3LinkPoId := mlagInfo.Attrs["l3_peer_link_port_channel_id"].(types.Int64)
//
//	if !l2LinkPoId.IsNull() && !l3LinkPoId.IsNull() &&
//		l2LinkPoId.Value == l3LinkPoId.Value {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			fmt.Sprintf("'peer_link_port_channel_id' and 'l3_peer_link_port_channel_id' cannot both use value %d",
//				l2LinkPoId.Value))
//	}
//
//	l3LinkCount := mlagInfo.Attrs["l3_peer_link_count"].(types.Int64)
//	l3LinkSpeed := mlagInfo.Attrs["l3_peer_link_speed"].(types.String)
//
//	if l3LinkCount.IsNull() && !l3LinkSpeed.IsNull() {
//		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_speed' requires 'l3_peer_link_count'")
//	}
//
//	if l3LinkCount.IsNull() && !l3LinkPoId.IsNull() {
//		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_port_channel_id' requires 'l3_peer_link_count'")
//	}
//
//	if !l3LinkCount.IsNull() && l3LinkSpeed.IsNull() {
//		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_count' requires 'l3_peer_link_speed'")
//	}
//
//	if !l3LinkCount.IsNull() && l3LinkPoId.IsNull() {
//		diags.AddAttributeError(errPath, errInvalidConfig, "'l3_peer_link_count' requires 'l3_peer_link_port_channel_id'")
//	}
//}
//
//func (o *rRackType) validateConfigAccessSwitches(ctx context.Context, errPath path.Path, diags *diag.Diagnostics) {
//	for i := range o.AccessSwitches.Elems {
//		o.validateConfigAccessSwitch(ctx, i, errPath.AtListIndex(i), diags)
//	}
//}
//
//func (o *rRackType) validateConfigAccessSwitch(ctx context.Context, idx int, errPath path.Path, diags *diag.Diagnostics) {
//	accessSwitchObj := o.AccessSwitches.Elems[idx].(types.Object)
//	linkSet := accessSwitchObj.Attrs["links"].(types.Set)
//	for i := range linkSet.Elems {
//		o.validateConfigAccessSwitchLink(ctx, idx, i, errPath.AtMapKey("links").AtListIndex(i), diags)
//	}
//}
//
//func (o *rRackType) validateConfigAccessSwitchLink(ctx context.Context, accessIdx int, linkIdx int, errPath path.Path, diags *diag.Diagnostics) {
//	accessSwitch := o.AccessSwitches.Elems[accessIdx].(types.Object)
//	links := accessSwitch.Attrs["links"].(types.Set)
//	link := links.Elems[linkIdx].(types.Object)
//
//	if !link.Attrs["lag_mode"].IsNull() {
//		diags.AddAttributeError(errPath, errInvalidConfig, "'lag_mode' not permitted on Access Switch links")
//		return
//	}
//
//	linkTargetName := link.Attrs["target_switch_name"].(types.String).Value
//	linkTarget := o.leafSwitchByName(linkTargetName)
//	if linkTarget == nil {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			fmt.Sprintf("no leaf switch named '%s'", linkTarget))
//		return
//	}
//
//	leafRedundancyProtocol := renderLeafRedundancyProtocol(*linkTarget)
//	accessRedundancyProtocol := renderAccessRedundancyProtocol(accessSwitch)
//
//	if accessRedundancyProtocol == goapstra.AccessRedundancyProtocolEsi &&
//		leafRedundancyProtocol != goapstra.LeafRedundancyProtocolEsi {
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			"ESI access switches only support connection to ESI leafs")
//		return
//	}
//
//	switchPeer := link.Attrs["switch_peer"].(types.String)
//	if !switchPeer.IsNull() && // primary/secondary has been selected ...and...
//		leafRedundancyProtocol == goapstra.LeafRedundancyProtocolNone { // upstream is not ESI/MLAG
//		diags.AddAttributeError(errPath, errInvalidConfig,
//			"'switch_peer' must not be set when upstream switch is non-redundant")
//	}
//}
//
//func (o *rRackType) switchByName(name string) *types.Object {
//	leaf := o.leafSwitchByName(name)
//	if leaf != nil {
//		return leaf
//	}
//	return nil
//}
//
//func (o *rRackType) leafSwitchByName(name string) *types.Object {
//	for _, leafSwitchAttrValue := range o.LeafSwitches.Elems {
//		if leafSwitchAttrValue.IsNull() || leafSwitchAttrValue.IsUnknown() {
//			continue
//		}
//		leafSwitchObj := leafSwitchAttrValue.(types.Object)
//		nameString := leafSwitchObj.Attrs["name"].(types.String)
//		if nameString.IsNull() || nameString.IsUnknown() {
//			continue
//		}
//		if nameString.Value == name {
//			return &leafSwitchObj
//		}
//	}
//	return nil
//}
//
//func (o *rRackType) parseApiResponse(ctx context.Context, rt *goapstra.RackType, diags *diag.Diagnostics) {
//	o.Id = types.String{Value: string(rt.Id)}
//	o.Name = types.String{Value: rt.Data.DisplayName}
//	o.Description = types.String{Value: rt.Data.Description}
//	o.FabricConnectivityDesign = types.String{Value: rt.Data.FabricConnectivityDesign.String()}
//	o.parseApiResponseLeafSwitches(ctx, rt.Data.LeafSwitches, diags)
//	o.parseApiResponseAccessSwitches(ctx, rt.Data.AccessSwitches, diags)
//	//o.GenericSystems =           parseRackTypeGenericSystems(rt.Data.GenericSystems, diags)
//}
//
//func (o *rRackType) parseApiResponseLeafSwitches(ctx context.Context, in []goapstra.RackElementLeafSwitch, diags *diag.Diagnostics) {
//	o.LeafSwitches = newRLeafSwitchList(len(in))
//	for i, ls := range in {
//		o.parseApiResponseLeafSwitch(ctx, &ls, i, diags)
//	}
//}
//
//func (o *rRackType) parseApiResponseLeafSwitch(ctx context.Context, in *goapstra.RackElementLeafSwitch, idx int, diags *diag.Diagnostics) {
//	o.LeafSwitches.Elems[idx] = types.Object{
//		AttrTypes: rLeafSwitchAttrTypes(),
//		Attrs: map[string]attr.Value{
//			"name":                types.String{Value: in.Label},
//			"spine_link_count":    parseApiLeafSwitchLinkPerSpineCountToTypesInt64(in),
//			"spine_link_speed":    parseApiLeafSwitchLinkPerSpineSpeedToTypesString(in),
//			"redundancy_protocol": parseApiLeafRedundancyProtocolToTypesString(in),
//			"logical_device":      parseApiLogicalDeviceToTypesObject(ctx, in.LogicalDevice, diags),
//			"mlag_info":           parseApiLeafMlagInfoToTypesObject(in.MlagInfo),
//			"tag_ids":             parseApiSliceTagDataToTypesSetString(in.Tags),
//			"tag_data":            parseApiSliceTagDataToTypesSetObject(in.Tags),
//		},
//	}
//}
//
//func (o *rRackType) parseApiResponseAccessSwitches(ctx context.Context, in []goapstra.RackElementAccessSwitch, diags *diag.Diagnostics) {
//	o.AccessSwitches = newRAccessSwitchList(len(in))
//	for i, as := range in {
//		o.parseApiResponseAccessSwitch(ctx, &as, i, diags)
//	}
//}
//
//func (o *rRackType) parseApiResponseAccessSwitch(ctx context.Context, in *goapstra.RackElementAccessSwitch, idx int, diags *diag.Diagnostics) {
//	o.AccessSwitches.Elems[idx] = types.Object{
//		AttrTypes: rAccessSwitchAttrTypes(),
//		Attrs: map[string]attr.Value{
//			"name":                types.String{Value: in.Label},
//			"count":               types.Int64{Value: int64(in.InstanceCount)},
//			"redundancy_protocol": parseApiAccessRedundancyProtocolToTypesString(in),
//			"logical_device":      parseApiLogicalDeviceToTypesObject(ctx, in.LogicalDevice, diags),
//			"tag_ids":             parseApiSliceTagDataToTypesSetString(in.Tags),
//			"tag_data":            parseApiSliceTagDataToTypesSetObject(in.Tags),
//			"esi_lag_info":        parseApiAccessEsiLagInfoToTypesObject(in.EsiLagInfo),
//			"links":               parseApiSliceRackLinkToTypesSetObject(in.Links),
//		},
//	}
//}
//
//func (o *rRackType) copyWriteOnlyElements(src *rRackType, diags *diag.Diagnostics) {
//	o.copyWriteOnlyElementsLeafSwitches(src, diags)
//	o.copyWriteOnlyElementsAccessSwitches(src, diags)
//}
//
//func (o *rRackType) copyWriteOnlyElementsLeafSwitches(src *rRackType, diags *diag.Diagnostics) {
//	for i, srcLeafSwitch := range src.LeafSwitches.Elems {
//		o.copyWriteOnlyElementsLeafSwitch(srcLeafSwitch.(types.Object), i, diags)
//	}
//}
//
//func (o *rRackType) copyWriteOnlyElementsLeafSwitch(src types.Object, idx int, _ *diag.Diagnostics) {
//	logicalDeviceId := src.Attrs["logical_device_id"].(types.String)
//	o.LeafSwitches.Elems[idx].(types.Object).Attrs["logical_device_id"] = logicalDeviceId
//}
//
//func (o *rRackType) copyWriteOnlyElementsAccessSwitches(src *rRackType, diags *diag.Diagnostics) {
//	for i, srcAccessSwitch := range src.AccessSwitches.Elems {
//		o.copyWriteOnlyElementsAccessSwitch(srcAccessSwitch.(types.Object), i, diags)
//	}
//}
//
//func (o *rRackType) copyWriteOnlyElementsAccessSwitch(src types.Object, idx int, _ *diag.Diagnostics) {
//	logicalDeviceId := src.Attrs["logical_device_id"].(types.String)
//	o.AccessSwitches.Elems[idx].(types.Object).Attrs["logical_device_id"] = logicalDeviceId
//}
//
//func getLogicalDeviceObj(ctx context.Context, client *goapstra.Client, id string, errPath path.Path, diags *diag.Diagnostics) types.Object {
//	logicalDevice, err := client.GetLogicalDevice(ctx, goapstra.ObjectId(id))
//	if err != nil {
//		var ace goapstra.ApstraClientErr
//		if ace.Type() == goapstra.ErrNotfound {
//			diags.AddAttributeError(errPath, "logical device not found",
//				fmt.Sprintf("logical device '%s' does not exist", id))
//		}
//		diags.AddError(fmt.Sprintf("error retrieving logical device '%s'", id), err.Error())
//	}
//	return parseApiLogicalDeviceToTypesObject(ctx, logicalDevice.Data, diags)
//}
//
//func (o *rRackType) request(ctx context.Context, path path.Path, client *goapstra.Client, diags *diag.Diagnostics) *goapstra.RackTypeRequest {
//	var fcd goapstra.FabricConnectivityDesign
//	err := fcd.FromString(o.FabricConnectivityDesign.ValueString())
//	if err != nil {
//		diags.AddAttributeError(path.AtMapKey("fabric_connectivity_design"),
//			"error parsing fabric_connectivity_design", err.Error())
//		return nil
//	}
//
//	leafSwitches := make([]goapstra.RackElementLeafSwitchRequest, len(o.LeafSwitches.))
//
//	return &goapstra.RackTypeRequest{
//		DisplayName:              o.Name.ValueString(),
//		Description:              o.Description.ValueString(),
//		FabricConnectivityDesign: fcd,
//		LeafSwitches:             nil,
//		AccessSwitches:           nil,
//		GenericSystems:           nil,
//	}
//}
//
//// forceValues handles user-optional values and values which are required by
//// Apstra, but which we can predict so we don't want to bother the user.
//func (o *rRackType) forceValues(diags *diag.Diagnostics) {
//	//// handle "description" omitted from config
//	//if o.Description.Unknown {
//	//	o.Description = types.String{Null: true}
//	//}
//
//	// handle empty "description" from API
//	if !o.Description.IsUnknown() && !o.Description.IsNull() && o.Description.Value == "" {
//		o.Description = types.String{Null: true}
//	}
//
//	// force leaf switch values as needed
//	o.forceValuesLeafSwitches(diags)
//	o.forceValuesAccessSwitches(diags)
//}
//
//func (o *rRackType) forceValuesLeafSwitches(diags *diag.Diagnostics) {
//	for i := range o.LeafSwitches.Elems {
//		o.forceValuesLeafSwitch(i, diags)
//	}
//}
//
//func (o *rRackType) forceValuesLeafSwitch(idx int, diags *diag.Diagnostics) {
//	//leafObj := o.LeafSwitches.Elems[idx].(types.Object)
//	//if leafObj.Attrs["redundancy_protocol"].IsNull() {
//	//	leafObj.Attrs["redundancy_protocol"] = types.String{}
//	//}
//
//	forceValuesTagIdsAndTagDataOnResourceRackElement(o.LeafSwitches.Elems[idx]) // todo try using leafObj here
//
//	switch o.FabricConnectivityDesign.Value {
//	case goapstra.FabricConnectivityDesignL3Clos.String():
//		// nothing yet
//	case goapstra.FabricConnectivityDesignL3Collapsed.String():
//		// spine link info must be null with collapsed fabric
//		o.LeafSwitches.Elems[idx].(types.Object).Attrs["spine_link_count"] = types.Int64{Null: true}
//		o.LeafSwitches.Elems[idx].(types.Object).Attrs["spine_link_speed"] = types.String{Null: true}
//	}
//}
//
//func (o *rRackType) forceValuesAccessSwitches(diags *diag.Diagnostics) {
//	for i := range o.AccessSwitches.Elems {
//		o.forceValuesAccessSwitch(i, diags)
//	}
//}
//
//func (o *rRackType) forceValuesAccessSwitch(idx int, diags *diag.Diagnostics) {
//	forceValuesTagIdsAndTagDataOnResourceRackElement(o.AccessSwitches.Elems[idx])
//	accessSwitchObj := o.AccessSwitches.Elems[idx].(types.Object)
//
//	if !accessSwitchObj.Attrs["esi_lag_info"].IsNull() {
//		redundancyProtocol := types.String{Value: goapstra.AccessRedundancyProtocolEsi.String()}
//		o.AccessSwitches.Elems[idx].(types.Object).Attrs["redundancy_protocol"] = redundancyProtocol
//	} else {
//		redundancyProtocol := types.String{Null: true}
//		o.AccessSwitches.Elems[idx].(types.Object).Attrs["redundancy_protocol"] = redundancyProtocol
//	}
//
//	// todo forcevalues link (tag names and lacp mode) -- done?
//	for i := range accessSwitchObj.Attrs["links"].(types.Set).Elems {
//		o.forceValuesAccessSwitchLink(idx, i, diags)
//	}
//}
//
//func (o *rRackType) forceValuesAccessSwitchLink(switchIdx, linkIdx int, diags *diag.Diagnostics) {
//	accessSwitchObj := o.AccessSwitches.Elems[switchIdx].(types.Object)
//	linkSet := accessSwitchObj.Attrs["links"].(types.Set)
//	linkObj := linkSet.Elems[linkIdx].(types.Object)
//	linksPerSwitch := linkObj.Attrs["links_per_switch"].(types.Int64)
//	switchPeer := linkObj.Attrs["switch_peer"].(types.String)
//
//	// access switches always use lacp active mode
//	lagMode := types.String{Value: goapstra.RackLinkLagModeActive.String()}
//	o.AccessSwitches.Elems[switchIdx].(types.Object).Attrs["links"].(types.Set).Elems[linkIdx].(types.Object).Attrs["lag_mode"] = lagMode
//
//	// unknown link per switch count gets set to 1
//	if linksPerSwitch.IsUnknown() {
//		linksPerSwitch = types.Int64{Value: 1}
//		o.AccessSwitches.Elems[switchIdx].(types.Object).Attrs["links"].(types.Set).Elems[linkIdx].(types.Object).Attrs["links_per_switch"] = linksPerSwitch
//	}
//
//	// unknown switch peer gets set to none
//	if switchPeer.IsUnknown() {
//		switchPeer = types.String{Value: goapstra.RackLinkSwitchPeerNone.String()}
//		o.AccessSwitches.Elems[switchIdx].(types.Object).Attrs["links"].(types.Set).Elems[linkIdx].(types.Object).Attrs["switch_peer"] = switchPeer
//	}
//
//	// handle "tag_ids" omitted from config
//	forceValuesTagIdsAndTagDataOnResourceRackElement(o.AccessSwitches.Elems[switchIdx].(types.Object).Attrs["links"].(types.Set).Elems[linkIdx])
//}
//
//func forceValuesTagIdsAndTagDataOnResourceRackElement(in attr.Value) {
//	if in.(types.Object).Attrs["tag_ids"].IsNull() {
//		in.(types.Object).Attrs["tag_ids"] = types.Set{
//			Null:     true,
//			ElemType: types.StringType,
//		}
//		in.(types.Object).Attrs["tag_data"] = types.Set{
//			Null:     true,
//			ElemType: types.ObjectType{AttrTypes: tagDataAttrTypes()},
//		}
//	}
//}
//
//func (o *rRackType) renderFabricConnectivityDesign() goapstra.FabricConnectivityDesign {
//	switch o.FabricConnectivityDesign.Value {
//	case goapstra.FabricConnectivityDesignL3Collapsed.String():
//		return goapstra.FabricConnectivityDesignL3Collapsed
//	default:
//		return goapstra.FabricConnectivityDesignL3Clos
//	}
//}
//
//func (o *rRackType) goapstraRequest(diags *diag.Diagnostics) *goapstra.RackTypeRequest {
//	return &goapstra.RackTypeRequest{
//		DisplayName:              o.Name.Value,
//		Description:              o.Description.Value,
//		FabricConnectivityDesign: o.renderFabricConnectivityDesign(),
//		LeafSwitches:             o.leafSwitchRequests(diags),
//		AccessSwitches:           o.accessSwitchRequests(diags),
//		//GenericSystems:           o.genericSystemRequests(diags),
//	}
//}
//
//// fcdModes returns permitted fabric_connectivity_design mode strings
//func fcdModes() []string {
//	return []string{
//		goapstra.FabricConnectivityDesignL3Clos.String(),
//		goapstra.FabricConnectivityDesignL3Collapsed.String()}
//}
//
//func (o *rRackType) leafSwitchRequests(_ *diag.Diagnostics) []goapstra.RackElementLeafSwitchRequest {
//	result := make([]goapstra.RackElementLeafSwitchRequest, len(o.LeafSwitches.Elems))
//	for i, leafSwitchListElem := range o.LeafSwitches.Elems {
//		leafSwitchObj := leafSwitchListElem.(types.Object)
//		result[i] = goapstra.RackElementLeafSwitchRequest{
//			Label:              leafSwitchObj.Attrs["name"].(types.String).Value,
//			LogicalDeviceId:    goapstra.ObjectId(leafSwitchObj.Attrs["logical_device_id"].(types.String).Value),
//			LinkPerSpineCount:  int(leafSwitchObj.Attrs["spine_link_count"].(types.Int64).Value),
//			LinkPerSpineSpeed:  goapstra.LogicalDevicePortSpeed(leafSwitchObj.Attrs["spine_link_speed"].(types.String).Value),
//			RedundancyProtocol: renderLeafRedundancyProtocol(leafSwitchObj),
//			Tags:               renderTagIdsToSliceStringsFromRackElement(leafSwitchObj),
//			MlagInfo:           renderLeafMlagInfo(leafSwitchObj),
//		}
//	}
//	return result
//}
//
//func (o *rRackType) accessSwitchRequests(_ *diag.Diagnostics) []goapstra.RackElementAccessSwitchRequest {
//	result := make([]goapstra.RackElementAccessSwitchRequest, len(o.AccessSwitches.Elems))
//	for i, accessSwitchListElem := range o.AccessSwitches.Elems {
//		accessSwitchObj := accessSwitchListElem.(types.Object)
//		result[i] = goapstra.RackElementAccessSwitchRequest{
//			Label:              accessSwitchObj.Attrs["name"].(types.String).Value,
//			InstanceCount:      int(accessSwitchObj.Attrs["count"].(types.Int64).Value),
//			LogicalDeviceId:    goapstra.ObjectId(accessSwitchObj.Attrs["logical_device_id"].(types.String).Value),
//			RedundancyProtocol: renderAccessRedundancyProtocol(accessSwitchObj),
//			Tags:               renderTagIdsToSliceStringsFromRackElement(accessSwitchObj),
//			EsiLagInfo:         renderAccessEsiLagInfo(accessSwitchObj),
//			Links:              o.renderLinkRequests(accessSwitchObj),
//		}
//	}
//	return result
//}
//
//// leafRedundancyModes returns permitted fabric_connectivity_design mode strings
//func leafRedundancyModes() []string {
//	return []string{
//		goapstra.LeafRedundancyProtocolEsi.String(),
//		goapstra.LeafRedundancyProtocolMlag.String()}
//}
//
//func renderLeafRedundancyProtocol(leafSwitch types.Object) goapstra.LeafRedundancyProtocol {
//	redundancyProtocol := leafSwitch.Attrs["redundancy_protocol"].(types.String)
//
//	if redundancyProtocol.IsNull() {
//		return goapstra.LeafRedundancyProtocolNone
//	}
//
//	switch redundancyProtocol.Value {
//	case goapstra.LeafRedundancyProtocolEsi.String():
//		return goapstra.LeafRedundancyProtocolEsi
//	case goapstra.LeafRedundancyProtocolMlag.String():
//		return goapstra.LeafRedundancyProtocolMlag
//	}
//	return goapstra.LeafRedundancyProtocolNone
//}
//
//func renderLeafMlagInfo(leafSwitch types.Object) *goapstra.LeafMlagInfo {
//	mlagInfo := leafSwitch.Attrs["mlag_info"].(types.Object)
//	if mlagInfo.IsNull() {
//		return nil
//	}
//
//	var LeafLeafLinkPortChannelId int
//	if !mlagInfo.Attrs["peer_link_port_channel_id"].IsNull() {
//		value := mlagInfo.Attrs["peer_link_port_channel_id"].(types.Int64).Value
//		LeafLeafLinkPortChannelId = int(value)
//	}
//
//	var LeafLeafL3LinkPortChannelId int
//	if !mlagInfo.Attrs["l3_peer_link_port_channel_id"].IsNull() {
//		value := mlagInfo.Attrs["l3_peer_link_port_channel_id"].(types.Int64).Value
//		LeafLeafL3LinkPortChannelId = int(value)
//	}
//
//	return &goapstra.LeafMlagInfo{
//		MlagVlanId:                  int(mlagInfo.Attrs["mlag_keepalive_vlan"].(types.Int64).Value),
//		LeafLeafLinkCount:           int(mlagInfo.Attrs["peer_link_count"].(types.Int64).Value),
//		LeafLeafLinkSpeed:           goapstra.LogicalDevicePortSpeed(mlagInfo.Attrs["peer_link_speed"].(types.String).Value),
//		LeafLeafLinkPortChannelId:   LeafLeafLinkPortChannelId,
//		LeafLeafL3LinkCount:         int(mlagInfo.Attrs["l3_peer_link_count"].(types.Int64).Value),
//		LeafLeafL3LinkSpeed:         goapstra.LogicalDevicePortSpeed(mlagInfo.Attrs["l3_peer_link_speed"].(types.String).Value),
//		LeafLeafL3LinkPortChannelId: LeafLeafL3LinkPortChannelId,
//	}
//}
//
//func renderAccessRedundancyProtocol(accessSwitch types.Object) goapstra.AccessRedundancyProtocol {
//	if accessSwitch.Attrs["esi_lag_info"].IsNull() {
//		return goapstra.AccessRedundancyProtocolNone
//	}
//	return goapstra.AccessRedundancyProtocolEsi
//}
//
//func renderLinkAttachmentType(link types.Object, targetSwitch types.Object) goapstra.RackLinkAttachmentType {
//	targetSwitchRedundancyProtocol := targetSwitch.Attrs["redundancy_protocol"].(types.String)
//	if targetSwitchRedundancyProtocol.IsNull() {
//		return goapstra.RackLinkAttachmentTypeSingle
//	}
//
//	lagMode := link.Attrs["lag_mode"].(types.String)
//	switch lagMode.Value {
//	case goapstra.RackLinkLagModeActive.String():
//		return goapstra.RackLinkAttachmentTypeDual
//	case goapstra.RackLinkLagModePassive.String():
//		return goapstra.RackLinkAttachmentTypeDual
//	case goapstra.RackLinkLagModeStatic.String():
//		return goapstra.RackLinkAttachmentTypeDual
//	}
//	return goapstra.RackLinkAttachmentTypeSingle
//}
//
//func renderLinkLagMode(link types.Object) goapstra.RackLinkLagMode {
//	lagMode := link.Attrs["lag_mode"].(types.String)
//	switch lagMode.Value {
//	case goapstra.RackLinkLagModeActive.String():
//		return goapstra.RackLinkLagModeActive
//	case goapstra.RackLinkLagModePassive.String():
//		return goapstra.RackLinkLagModePassive
//	case goapstra.RackLinkLagModeStatic.String():
//		return goapstra.RackLinkLagModeStatic
//	}
//	return goapstra.RackLinkLagModeNone
//}
//
//func renderLinkSwitchPeer(link types.Object) goapstra.RackLinkSwitchPeer {
//	switchPeer := link.Attrs["switch_peer"].(types.String)
//	switch switchPeer.Value {
//	case goapstra.RackLinkSwitchPeerFirst.String():
//		return goapstra.RackLinkSwitchPeerFirst
//	case goapstra.RackLinkSwitchPeerSecond.String():
//		return goapstra.RackLinkSwitchPeerSecond
//	}
//	return goapstra.RackLinkSwitchPeerNone
//}
//
//func renderAccessEsiLagInfo(accessSwitch types.Object) *goapstra.EsiLagInfo {
//	esiLagInfo := accessSwitch.Attrs["esi_lag_info"].(types.Object)
//	if esiLagInfo.IsNull() {
//		return nil
//	}
//
//	return &goapstra.EsiLagInfo{
//		AccessAccessLinkCount: int(esiLagInfo.Attrs["l3_peer_link_count"].(types.Int64).Value),
//		AccessAccessLinkSpeed: goapstra.LogicalDevicePortSpeed(esiLagInfo.Attrs["l3_peer_link_speed"].(types.String).Value),
//	}
//}
//
//func (o *rRackType) renderLinkRequests(rackElement types.Object) []goapstra.RackLinkRequest {
//	links := rackElement.Attrs["links"].(types.Set)
//	result := make([]goapstra.RackLinkRequest, len(links.Elems))
//	for i, linkAttrValue := range links.Elems {
//		result[i] = *o.renderLinkRequest(linkAttrValue.(types.Object))
//	}
//	return result
//}
//
//func (o *rRackType) renderLinkRequest(link types.Object) *goapstra.RackLinkRequest {
//	targetSwitchName := link.Attrs["target_switch_name"].(types.String).Value
//	targetSwitchObj := o.switchByName(targetSwitchName)
//
//	return &goapstra.RackLinkRequest{
//		Label:              link.Attrs["name"].(types.String).Value,
//		Tags:               renderTagIdsToSliceStringsFromRackElement(link),
//		LinkPerSwitchCount: int(link.Attrs["links_per_switch"].(types.Int64).Value),
//		LinkSpeed:          goapstra.LogicalDevicePortSpeed(link.Attrs["speed"].(types.String).Value),
//		TargetSwitchLabel:  targetSwitchName,
//		AttachmentType:     renderLinkAttachmentType(link, *targetSwitchObj),
//		LagMode:            renderLinkLagMode(link),
//		SwitchPeer:         renderLinkSwitchPeer(link),
//	}
//}
//
//func rLeafSwitchAttrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"name":                types.StringType,
//		"spine_link_count":    types.Int64Type,
//		"spine_link_speed":    types.StringType,
//		"redundancy_protocol": types.StringType,
//		"logical_device_id":   types.StringType,
//		"logical_device":      logicalDeviceAttrType(),
//		"tag_ids":             tagIdsAttrType(),
//		"tag_data":            tagDataAttrType(),
//		"mlag_info":           mlagInfoAttrType(),
//	}
//}
//
//func rAccessSwitchAttrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"name":                types.StringType,
//		"count":               types.Int64Type,
//		"redundancy_protocol": types.StringType,
//		"logical_device_id":   types.StringType,
//		"logical_device":      logicalDeviceAttrType(),
//		"tag_ids":             tagIdsAttrType(),
//		"tag_data":            tagDataAttrType(),
//		"links":               rLinksAttrType(),
//		"esi_lag_info":        esiLagInfoAttrType(),
//	}
//}
//
//func rLinksAttrType() attr.Type {
//	return types.SetType{
//		ElemType: types.ObjectType{
//			AttrTypes: rLinksAttrTypes()}}
//}
//
//func rLinksAttrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"name":               types.StringType,
//		"target_switch_name": types.StringType,
//		"lag_mode":           types.StringType,
//		"links_per_switch":   types.Int64Type,
//		"speed":              types.StringType,
//		"switch_peer":        types.StringType,
//		"tag_data":           tagDataAttrType(),
//		"tag_ids":            tagIdsAttrType(),
//	}
//}
//
//func newRLeafSwitchList(size int) types.List {
//	return types.List{
//		Null:     size == 0,
//		Elems:    make([]attr.Value, size),
//		ElemType: types.ObjectType{AttrTypes: rLeafSwitchAttrTypes()},
//	}
//}
//
//func newRAccessSwitchList(size int) types.List {
//	return types.List{
//		Null:     size == 0,
//		Elems:    make([]attr.Value, size),
//		ElemType: types.ObjectType{AttrTypes: rAccessSwitchAttrTypes()},
//	}
//}
//
//func parseApiLeafSwitchLinkPerSpineCountToTypesInt64(in *goapstra.RackElementLeafSwitch) types.Int64 {
//	if in.LinkPerSpineCount == 0 {
//		return types.Int64{Null: true}
//	}
//	return types.Int64{Value: int64(in.LinkPerSpineCount)}
//}
//
//func parseApiLeafSwitchLinkPerSpineSpeedToTypesString(in *goapstra.RackElementLeafSwitch) types.String {
//	if in.LinkPerSpineCount == 0 {
//		return types.String{Null: true}
//	}
//	return types.String{Value: string(in.LinkPerSpineSpeed)}
//}
//
//func parseApiLeafRedundancyProtocolToTypesString(in *goapstra.RackElementLeafSwitch) types.String {
//	if in.RedundancyProtocol == goapstra.LeafRedundancyProtocolNone {
//		return types.String{Null: true}
//	}
//	return types.String{Value: in.RedundancyProtocol.String()}
//}
//
//func parseApiLeafMlagInfoToTypesObject(in *goapstra.LeafMlagInfo) types.Object {
//	if in == nil || (in.LeafLeafLinkCount == 0 && in.LeafLeafL3LinkCount == 0) {
//		return types.Object{
//			Null:      true,
//			AttrTypes: mlagInfoAttrTypes(),
//		}
//	}
//
//	var l3PeerLinkCount, l3PeerLinkPortChannelId types.Int64
//	var l3PeerLinkSPeed types.String
//	if in.LeafLeafL3LinkCount == 0 {
//		// link count of zero means all L3 link descriptors should be null
//		l3PeerLinkCount.Null = true
//		l3PeerLinkSPeed.Null = true
//		l3PeerLinkPortChannelId.Null = true
//	} else {
//		// we have links, so populate attributes from API response
//		l3PeerLinkCount.Value = int64(in.LeafLeafL3LinkCount)
//		l3PeerLinkSPeed.Value = string(in.LeafLeafL3LinkSpeed)
//		if in.LeafLeafL3LinkPortChannelId == 0 {
//			// Don't save PoId /0/ - use /null/ instead
//			l3PeerLinkPortChannelId.Null = true
//		} else {
//			l3PeerLinkPortChannelId.Value = int64(in.LeafLeafL3LinkPortChannelId)
//		}
//	}
//
//	var peerLinkPortChannelId types.Int64
//	if in.LeafLeafLinkPortChannelId == 0 {
//		// Don't save PoId /0/ - use /null/ instead
//		peerLinkPortChannelId.Null = true
//	} else {
//		peerLinkPortChannelId.Value = int64(in.LeafLeafLinkPortChannelId)
//	}
//
//	return types.Object{
//		AttrTypes: mlagInfoAttrTypes(),
//		Attrs: map[string]attr.Value{
//			"mlag_keepalive_vlan":          types.Int64{Value: int64(in.MlagVlanId)},
//			"peer_link_count":              types.Int64{Value: int64(in.LeafLeafLinkCount)},
//			"peer_link_speed":              types.String{Value: string(in.LeafLeafLinkSpeed)},
//			"peer_link_port_channel_id":    peerLinkPortChannelId,
//			"l3_peer_link_count":           l3PeerLinkCount,
//			"l3_peer_link_speed":           l3PeerLinkSPeed,
//			"l3_peer_link_port_channel_id": l3PeerLinkPortChannelId,
//		},
//	}
//}
//
//func parseApiAccessRedundancyProtocolToTypesString(in *goapstra.RackElementAccessSwitch) types.String {
//	if in.RedundancyProtocol == goapstra.AccessRedundancyProtocolNone {
//		return types.String{Null: true}
//	} else {
//		return types.String{Value: in.RedundancyProtocol.String()}
//	}
//}
//
//func parseApiAccessEsiLagInfoToTypesObject(in *goapstra.EsiLagInfo) types.Object {
//	if in == nil || in.AccessAccessLinkCount == 0 {
//		return types.Object{
//			Null:      true,
//			AttrTypes: esiLagInfoAttrTypes(),
//		}
//	}
//
//	return types.Object{
//		AttrTypes: esiLagInfoAttrTypes(),
//		Attrs: map[string]attr.Value{
//			"l3_peer_link_count": types.Int64{Value: int64(in.AccessAccessLinkCount)},
//			"l3_peer_link_speed": types.String{Value: string(in.AccessAccessLinkSpeed)},
//		},
//	}
//}
//
//func parseApiSliceRackLinkToTypesSetObject(links []goapstra.RackLink) types.Set {
//	result := newLinkSet(len(links))
//	for i, link := range links {
//		var switchPeer types.String
//		if link.SwitchPeer == goapstra.RackLinkSwitchPeerNone {
//			switchPeer = types.String{Null: true}
//		} else {
//			switchPeer = types.String{Value: link.SwitchPeer.String()}
//		}
//		result.Elems[i] = types.Object{
//			AttrTypes: dLinksAttrTypes(),
//			Attrs: map[string]attr.Value{
//				"name":               types.String{Value: link.Label},
//				"target_switch_name": types.String{Value: link.TargetSwitchLabel},
//				"lag_mode":           types.String{Value: link.LagMode.String()},
//				"links_per_switch":   types.Int64{Value: int64(link.LinkPerSwitchCount)},
//				"speed":              types.String{Value: string(link.LinkSpeed)},
//				"switch_peer":        switchPeer,
//				"tag_data":           parseApiSliceTagDataToTypesSetObject(link.Tags),
//			},
//		}
//	}
//	return result
//}
//
//func mlagInfoAttrType() attr.Type {
//	return types.ObjectType{
//		AttrTypes: mlagInfoAttrTypes()}
//}
//
//func mlagInfoAttrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"mlag_keepalive_vlan":          types.Int64Type,
//		"peer_link_count":              types.Int64Type,
//		"peer_link_speed":              types.StringType,
//		"peer_link_port_channel_id":    types.Int64Type,
//		"l3_peer_link_count":           types.Int64Type,
//		"l3_peer_link_speed":           types.StringType,
//		"l3_peer_link_port_channel_id": types.Int64Type,
//	}
//}
//
//func esiLagInfoAttrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"l3_peer_link_count": types.Int64Type,
//		"l3_peer_link_speed": types.StringType,
//	}
//}
//
//func esiLagInfoAttrType() attr.Type {
//	return types.ObjectType{
//		AttrTypes: esiLagInfoAttrTypes()}
//}
//
//func dLinksAttrTypes() map[string]attr.Type {
//	return map[string]attr.Type{
//		"name":               types.StringType,
//		"target_switch_name": types.StringType,
//		"lag_mode":           types.StringType,
//		"links_per_switch":   types.Int64Type,
//		"speed":              types.StringType,
//		"switch_peer":        types.StringType,
//		"tag_data":           tagDataAttrType(),
//	}
//}
//
//func newLinkSet(size int) types.Set {
//	return types.Set{
//		Elems: make([]attr.Value, size),
//		ElemType: types.ObjectType{
//			AttrTypes: dLinksAttrTypes()},
//	}
//}

package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
	"terraform-provider-apstra/apstra/utils"
)

type DeviceAllocation struct {
	BlueprintId           types.String `tfsdk:"blueprint_id"`           // required
	NodeName              types.String `tfsdk:"node_name"`              // required
	DeviceKey             types.String `tfsdk:"device_key"`             // optional
	InterfaceMapCatalogId types.String `tfsdk:"interface_map_id"`       // computed + optional
	NodeId                types.String `tfsdk:"node_id"`                // computed
	DeviceProfileNodeId   types.String `tfsdk:"device_profile_node_id"` // computed
	DeployMode            types.String `tfsdk:"deploy_mode"`            // optional
}

func (o DeviceAllocation) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"node_name": resourceSchema.StringAttribute{
			MarkdownDescription: "GraphDB node 'label which identifies the switch. Strings like 'spine1' " +
				"and 'rack_2_leaf_1' are appropriate here.",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"device_key": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique ID for a Managed Device, generally the serial number, used to. " +
				"assign a Managed Device to a fabric role.",
			Optional:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.AtLeastOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("interface_map_id"),
				}...),
			},
		},
		"interface_map_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Interface Maps link a Logical Device (fabric design element) to a " +
				"Device Profile which describes a hardware model. This field is required when `device_key` " +
				"is omitted, or when `device_key` is supplied, but does not provide enough information to`. " +
				"select an Interface Map. This field represents the Blueprint graphDB node ID, which is " +
				"the same string as the global ID used in the design API global catalog.",
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"node_id": resourceSchema.StringAttribute{
			MarkdownDescription: "GraphDB Node ID of the fabric node to which we're allocating an Interface Map " +
				"and Managed Device.",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"device_profile_node_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Device Profiles specify attributes of specific hardware models.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"deploy_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "Set *Deploy Mode* of the associated fabric node.", // todo
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("device_key")),
				stringvalidator.OneOf(utils.AllNodeDeployModes()...),
			},
		},
		"deploy_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "", // todo
			Optional:            true,
			Default:             stringdefault.StaticString(apstra.NodeDeployModeDeploy.String()),
			Validators:          []validator.String{stringvalidator.OneOf(utils.AllNodeDeployModes()...)},
		},
	}
}

func (o *DeviceAllocation) validateInterfaceMapId(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			InterfaceMap struct {
				Id string `json:"id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}

	query := client.NewQuery(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetType(apstra.BlueprintTypeStaging).
		SetContext(ctx).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("system")},
			{"id", apstra.QEStringVal(o.NodeId.ValueString())},
		}).
		Out([]apstra.QEEAttribute{{"type", apstra.QEStringVal("logical_device")}}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("logical_device")},
		}).
		In([]apstra.QEEAttribute{{"type", apstra.QEStringVal("logical_device")}}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("interface_map")},
			{"name", apstra.QEStringVal("n_interface_map")},
			{"id", apstra.QEStringVal(o.InterfaceMapCatalogId.ValueString())},
		}).
		Out([]apstra.QEEAttribute{{"type", apstra.QEStringVal("device_profile")}}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("device_profile")},
			{"id", apstra.QEStringVal(o.DeviceProfileNodeId.ValueString())},
		})

	err := query.Do(&result)
	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
			return
		}
		diags.AddError("error running interface map query", err.Error())
		return
	}

	if len(result.Items) != 1 {
		diags.AddError("error validating interface_map_id",
			fmt.Sprintf(
				"expected 1 path linking system %q, logical device (any), interface map %q and device profile %q, got %d\nquery: %q",
				o.NodeName.ValueString(), o.InterfaceMapCatalogId.ValueString(),
				o.DeviceProfileNodeId.ValueString(), len(result.Items), query.String()))
	}
}

func (o *DeviceAllocation) populateInterfaceMapIdFromNodeIdAndDeviceProfileNodeId(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			InterfaceMap struct {
				Id string `json:"id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}

	query := client.NewQuery(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		SetType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("system")},
			{"id", apstra.QEStringVal(o.NodeId.ValueString())},
		}).
		Out([]apstra.QEEAttribute{{"type", apstra.QEStringVal("logical_device")}}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("logical_device")},
		}).
		In([]apstra.QEEAttribute{{"type", apstra.QEStringVal("logical_device")}}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("interface_map")},
			{"name", apstra.QEStringVal("n_interface_map")},
		}).
		Out([]apstra.QEEAttribute{{"type", apstra.QEStringVal("device_profile")}}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("device_profile")},
			{"id", apstra.QEStringVal(o.DeviceProfileNodeId.ValueString())},
		})

	err := query.Do(&result)
	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
			return
		}
		diags.AddError("error running interface map query", err.Error())
		return
	}

	switch len(result.Items) {
	case 0:
		diags.AddError("unable to assign interface_map",
			fmt.Sprintf("no interface_map links system '%s' to device profile '%s'\nquery: %q",
				o.NodeName.ValueString(), o.DeviceProfileNodeId.ValueString(), query.String()))
	case 1:
		o.InterfaceMapCatalogId = types.StringValue(result.Items[0].InterfaceMap.Id)
	default:
		candidates := make([]string, len(result.Items))
		for i := range result.Items {
			candidates[i] = result.Items[i].InterfaceMap.Id
		}
		diags.AddError("multiple Interface Map candidates - `interface_map` configuration attribute required",
			fmt.Sprintf("enter one of the following into `interface_map` to use Device Profile %q at Node %q - [%s]",
				o.DeviceProfileNodeId.ValueString(), o.NodeName.ValueString(), strings.Join(candidates, ", ")))
	}
}

func (o *DeviceAllocation) nodeIdFromNodeName(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	query := client.NewQuery(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetType(apstra.BlueprintTypeStaging).
		SetContext(ctx).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("system")},
			{"label", apstra.QEStringVal(o.NodeName.ValueString())},
			{"name", apstra.QEStringVal("n_system")},
		})

	err := query.Do(&result)
	if err != nil {
		diags.AddError("error running system node query", err.Error())
		return
	}

	switch len(result.Items) {
	case 0:
		diags.AddError("switch node not found in blueprint",
			fmt.Sprintf("switch/system node with label %q: not found with query %q",
				o.NodeName.ValueString(), query.String()))
	case 1:
		// no error case
		o.NodeId = types.StringValue(result.Items[0].System.Id)
	default:
		diags.AddError("multiple matches found in blueprint",
			fmt.Sprintf("switch/system node with label %q: %d matches found using query %q",
				o.NodeName.ValueString(), len(result.Items), query.String()),
		)
	}
}

// PopulateDataFromGraphDb attempts to set
//   - NodeId (from node_name)
//   - InterfaceMapCatalogId (when not set)
//   - DeviceProfileNodeId
//   - from DeviceKey when set
//   - from InterfaceMapCatalogId when DeviceKey not set
func (o *DeviceAllocation) PopulateDataFromGraphDb(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	if o.NodeId.IsUnknown() {
		// this should only be true once, in Create()
		o.nodeIdFromNodeName(ctx, client, diags)
	}
	if diags.HasError() {
		return
	}

	switch {
	case (!o.InterfaceMapCatalogId.IsUnknown() && !o.InterfaceMapCatalogId.IsNull()) && o.DeviceKey.IsNull():
		// interface_map_id known, device_key not supplied
		o.deviceProfileNodeIdFromInterfaceMapCatalogId(ctx, client, diags) // this will clear BlueprintId on 404
	case !o.DeviceKey.IsNull() && o.InterfaceMapCatalogId.IsUnknown():
		// device_key known, interface_map_id not supplied
		o.deviceProfileNodeIdFromDeviceKey(ctx, client, diags) // this will clear BlueprintId on 404
	case !o.InterfaceMapCatalogId.IsNull() && !o.InterfaceMapCatalogId.IsUnknown() && !o.DeviceKey.IsNull():
		// device_key and interface_map_id both supplied
		o.deviceProfileNodeIdFromDeviceKey(ctx, client, diags) // this will clear BlueprintId on 404
		if o.BlueprintId.IsNull() {
			return
		}
		if !diags.HasError() {
			o.validateInterfaceMapId(ctx, client, diags) // this will clear BlueprintId on 404
		}
	default:
		// config validation should not have allowed this
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("cannot proceed\n"+
				"  device_key null: %t, device_key unknown: %t\n"+
				"  interface_map_id null: %t, interface_map_id known: %t",
				o.DeviceKey.IsNull(), o.DeviceKey.IsUnknown(),
				o.InterfaceMapCatalogId.IsNull(), o.InterfaceMapCatalogId.IsUnknown(),
			),
		)
	}
	if diags.HasError() || o.BlueprintId.IsNull() {
		return
	}

	if o.InterfaceMapCatalogId.IsUnknown() {
		o.populateInterfaceMapIdFromNodeIdAndDeviceProfileNodeId(ctx, client, diags)
	}
	if diags.HasError() || o.BlueprintId.IsNull() {
		return
	}

	o.validateInterfaceMapId(ctx, client, diags) // this will clear BlueprintId on 404
	if diags.HasError() || o.BlueprintId.IsNull() {
		return
	}
}

// SetInterfaceMap creates or deletes the graph db relationship between a switch
// 'system' node and its interface map.
func (o *DeviceAllocation) SetInterfaceMap(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	assignments := make(apstra.SystemIdToInterfaceMapAssignment, 1)
	if o.InterfaceMapCatalogId.IsNull() {
		assignments[o.NodeId.ValueString()] = nil
	} else {
		assignments[o.NodeId.ValueString()] = o.InterfaceMapCatalogId.ValueString()
	}
	bpClient, err := client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
			return
		}
		diags.AddError("error creating blueprint client", err.Error())
		return
	}

	err = bpClient.SetInterfaceMapAssignments(ctx, assignments)
	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
			return
		}
		diags.AddError(fmt.Sprintf("error (re)setting interface map for node %q", o.NodeId.ValueString()), err.Error())
	}
	return
}

// SetNodeSystemId assigns a managed device 'device_key' (serial number) to a
// switch 'system' node in the blueprint graphdb. Returns false when Apstra
// returns a 404 to the blueprint operation, indicating the blueprint doesn't
// exist and resources depending on the blueprint's existence should be removed.
func (o *DeviceAllocation) SetNodeSystemId(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	if o.DeviceKey.IsNull() {
		return
	}

	patch := &struct {
		SystemId string `json:"system_id"`
	}{
		SystemId: o.DeviceKey.ValueString(),
	}

	nodeId := apstra.ObjectId(o.NodeId.ValueString())
	blueprintId := apstra.ObjectId(o.BlueprintId.ValueString())
	err := client.PatchNode(ctx, blueprintId, nodeId, &patch, nil)
	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
			return
		}
		diags.AddError(fmt.Sprintf("failed to (re)assign system_id for node '%s'", nodeId), err.Error())
	}
	return
}

// ReadSystemNode uses the BlueprintId and NodeId to determine the current
// DeviceKey value in the blueprint.
func (o *DeviceAllocation) ReadSystemNode(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	if o.NodeId.IsNull() {
		diags.AddError(errProviderBug, "ReadSystemNode invoked with null NodeId")
		return
	}

	var result struct {
		Items []struct {
			System struct {
				SystemId string `json:"system_id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	query := client.NewQuery(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		SetType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("system")},
			{"id", apstra.QEStringVal(o.NodeId.ValueString())},
			{"name", apstra.QEStringVal("n_system")},
		})

	err := query.Do(&result)
	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
		} else {
			diags.AddError("error querying system_id", err.Error())
		}
		return
	}

	switch len(result.Items) {
	case 0:
		// node doesn't exist!?!?
		o.NodeId = types.StringNull()
		o.DeviceKey = types.StringNull()
	case 1:
		// expected result
		o.DeviceKey = utils.StringValueOrNull(ctx, result.Items[0].System.SystemId, diags)
	default:
		ids := make([]string, len(result.Items))
		for i := range result.Items {
			ids[i] = result.Items[i].System.SystemId
		}
		diags.AddError("cannot proceed: multiple graphdb nodes share common id",
			fmt.Sprintf("label %q used by %q", o.NodeName.ValueString(), strings.Join(ids, "\", \"")))
	}
}

func (o *DeviceAllocation) GetCurrentInterfaceMapId(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			InterfaceMap struct {
				Id string `json:"id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}

	err := client.NewQuery(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("system")},
			{"id", apstra.QEStringVal(o.NodeId.ValueString())},
		}).
		Out([]apstra.QEEAttribute{{"type", apstra.QEStringVal("interface_map")}}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("interface_map")},
			{"name", apstra.QEStringVal("n_interface_map")},
		}).
		Do(&result)
	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
			return
		}
		diags.AddError("error querying interface map assignment", err.Error())
		return
	}

	switch len(result.Items) {
	case 0:
		o.InterfaceMapCatalogId = types.StringNull()
	case 1:
		o.InterfaceMapCatalogId = types.StringValue(result.Items[0].InterfaceMap.Id)
	default:
		iMaps := make([]string, len(result.Items))
		for i := range result.Items {
			iMaps[i] = result.Items[i].InterfaceMap.Id
		}
		diags.AddError("cannot proceed: graphdb links system node to multiple interface maps",
			fmt.Sprintf("%q matches %q", o.NodeName.ValueString(), strings.Join(iMaps, "\", \"")))
	}
}

func (o *DeviceAllocation) GetCurrentDeviceProfileId(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			DeviceProfile struct {
				Id string `json:"id"`
			} `json:"n_device_profile"`
		} `json:"items"`
	}

	query := client.NewQuery(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("system")},
			{"id", apstra.QEStringVal(o.NodeId.ValueString())},
		}).
		Out([]apstra.QEEAttribute{{"type", apstra.QEStringVal("interface_map")}}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("interface_map")},
		}).
		Out([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("device_profile")},
		}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("device_profile")},
			{"name", apstra.QEStringVal("n_device_profile")},
		})

	err := query.Do(&result)
	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
		} else {
			diags.AddError("error querying device profile", err.Error())
		}
		return
	}

	switch len(result.Items) {
	case 0:
		o.DeviceProfileNodeId = types.StringNull()
	case 1:
		o.DeviceProfileNodeId = types.StringValue(result.Items[0].DeviceProfile.Id)
	default:
		ids := make([]string, len(result.Items))
		for i := range result.Items {
			ids[i] = result.Items[i].DeviceProfile.Id
		}
		diags.AddError("cannot proceed: graphdb links system node to multiple device profiles",
			fmt.Sprintf("%q matches %q", o.NodeName.ValueString(), strings.Join(ids, "\", \"")))
	}
}

func (o *DeviceAllocation) deviceProfileNodeIdFromInterfaceMapCatalogId(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	var response struct {
		Items []struct {
			DeviceProfile struct {
				Id string `json:"id"`
			} `json:"n_device_profile"`
		} `json:"items"`
	}

	query := client.NewQuery(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetType(apstra.BlueprintTypeStaging).
		SetContext(ctx).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("interface_map")},
			{"id", apstra.QEStringVal(o.InterfaceMapCatalogId.ValueString())},
		}).
		Out([]apstra.QEEAttribute{{"type", apstra.QEStringVal("device_profile")}}).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("device_profile")},
			{"name", apstra.QEStringVal("n_device_profile")},
		})

	err := query.Do(&response)

	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
			return
		}
		diags.AddError("error querying graphDB for device profile", err.Error())
		return
	}

	switch len(response.Items) {
	case 0:
		diags.AddError("no results when querying for Device Profile", fmt.Sprintf("query string %q", query.String()))
	case 1:
		o.DeviceProfileNodeId = types.StringValue(response.Items[0].DeviceProfile.Id)
	default:
		diags.AddError("multiple matches when querying for Device Profile", fmt.Sprintf("query string %q", query.String()))
	}
}

func (o *DeviceAllocation) deviceProfileNodeIdFromDeviceKey(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	gasi := utils.GetAllSystemsInfo(ctx, client, diags)
	if diags.HasError() {
		return
	}

	si, ok := gasi[o.DeviceKey.ValueString()]
	if !ok {
		diags.AddAttributeError(
			path.Root("device_key"),
			"Device Key not found",
			fmt.Sprintf("Device Key %q not found", o.DeviceKey.ValueString()),
		)
		return
	}

	query := client.NewQuery(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		SetType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{"type", apstra.QEStringVal("device_profile")},
			{"device_profile_id", apstra.QEStringVal(si.Facts.AosHclModel.String())},
			{"name", apstra.QEStringVal("n_device_profile")},
		})

	var result struct {
		Items []struct {
			DeviceProfile struct {
				Id string `json:"id"`
			} `json:"n_device_profile"`
		} `json:"items"`
	}

	err := query.Do(&result)
	if err != nil {
		if utils.IsApstra404(err) {
			o.BlueprintId = types.StringNull()
			return
		}
		diags.AddError("error querying graphDB for device profile", err.Error())
		return
	}

	if len(result.Items) != 1 {
		diags.AddError(fmt.Sprintf(
			"expected 1 graphDB query result, got %d", len(result.Items)),
			fmt.Sprintf("query: %q", query.String()))
		return
	}

	o.DeviceProfileNodeId = types.StringValue(result.Items[0].DeviceProfile.Id)
}

func (o *DeviceAllocation) SetNodeDeployMode(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	if o.DeployMode.IsNull() {
		return
	}

	if o.DeployMode.IsUnknown() {
		o.GetNodeDeployMode(ctx, client, diags)
		return
	}

	var deployMode apstra.NodeDeployMode
	err := utils.ApiStringerFromFriendlyString(&deployMode, o.DeployMode.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing deploy mode %q", o.DeployMode.ValueString()), err.Error())
		return
	}

	type patch struct {
		Id         string `json:"id"`
		DeployMode string `json:"deploy_mode,omitempty"`
	}

	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing deploy mode %q", o.DeployMode.ValueString()), err.Error())
		return
	}

	setDeployMode := patch{
		Id:         o.NodeId.ValueString(),
		DeployMode: deployMode.String(),
	}
	deployModeResponse := patch{}

	err = client.PatchNode(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), apstra.ObjectId(o.NodeId.ValueString()), &setDeployMode, &deployModeResponse)
	if err != nil {
		diags.AddError("error setting deploy mode", err.Error())
		return
	}

	if setDeployMode.Id != deployModeResponse.Id {
		diags.AddError("inconsistent response while setting deploy mode",
			fmt.Sprintf("request ID: %q, response ID: %q", setDeployMode.Id, deployModeResponse.Id))
	}

	if setDeployMode.DeployMode != deployModeResponse.DeployMode {
		diags.AddError("inconsistent response while setting deploy mode",
			fmt.Sprintf("request DeployMode: %q, response DeployMode: %q", setDeployMode.DeployMode, deployModeResponse.DeployMode))
	}
}

func (o *DeviceAllocation) GetNodeDeployMode(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	bpClient, err := client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		diags.AddError("error creating blueprint client", err.Error())
		return
	}

	var nodesResponse struct {
		Nodes map[string]struct {
			Id         string `json:"id"`
			Type       string `json:"type"`
			DeployMode string `json:"deploy_mode"`
		} `json:"nodes"`
	}
	err = bpClient.GetNodes(ctx, apstra.NodeTypeSystem, &nodesResponse)
	if err != nil {
		diags.AddError("error fetching blueprint nodes", err.Error())
		return
	}

	node, ok := nodesResponse.Nodes[o.NodeId.ValueString()]
	if !ok {
		o.DeployMode = types.StringNull()
		return
	}

	var deployMode apstra.NodeDeployMode
	err = deployMode.FromString(node.DeployMode)
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing deploy mode %q", node.DeployMode), err.Error())
		return
	}

	o.DeployMode = types.StringValue(utils.StringersToFriendlyString(deployMode))
}

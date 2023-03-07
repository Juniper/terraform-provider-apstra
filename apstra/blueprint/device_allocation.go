package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
	"terraform-provider-apstra/apstra/utils"
)

type DeviceAllocation struct {
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	NodeName        types.String `tfsdk:"node_name"`
	DeviceKey       types.String `tfsdk:"device_key"`
	InterfaceMapId  types.String `tfsdk:"interface_map_id"`
	DeviceProfileId types.String `tfsdk:"device_profile_id"`
	SystemNodeId    types.String `tfsdk:"system_node_id"`
}

func (o DeviceAllocation) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Blueprint to which the Resource Pool should be allocated.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"node_name": resourceSchema.StringAttribute{
			MarkdownDescription: "", // todo
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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
				"is omitted, or when `device_key` is supplied, but multiple Interface Maps link the system " +
				"node Logical Device to the specific Device Profile (hardware model) indicated by `device_key`.",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"device_profile_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Device Profiles specify attributes of specific hardware models.", //todo
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"system_node_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID number of the Blueprint graphdb node representing this system.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
	}
}

func (o *DeviceAllocation) validateInterfaceMapId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			InterfaceMap struct {
				Id string `json:"id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}

	err := client.NewQuery(goapstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(o.NodeName.ValueString())},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("logical_device")},
		}).
		In([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"name", goapstra.QEStringVal("n_interface_map")},
			{"id", goapstra.QEStringVal(o.InterfaceMapId.ValueString())},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("device_profile")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("device_profile")},
			{"device_profile_id", goapstra.QEStringVal(o.DeviceProfileId.ValueString())},
		}).Do(&result)
	if err != nil {
		diags.AddError("error running interface map query", err.Error())
		return
	}

	if len(result.Items) != 1 {
		diags.AddError("error validating interface_map_id",
			fmt.Sprintf("expected 1 path linking system %q, interface map %q and device profile %q, got %d",
				o.NodeName.ValueString(), o.InterfaceMapId.ValueString(), o.DeviceProfileId.ValueString(), len(result.Items)))
	}
}

func (o *DeviceAllocation) populateInterfaceMapId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			InterfaceMap struct {
				Id string `json:"id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}

	err := client.NewQuery(goapstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(o.NodeName.ValueString())},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("logical_device")},
		}).
		In([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"name", goapstra.QEStringVal("n_interface_map")},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("device_profile")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("device_profile")},
			{"device_profile_id", goapstra.QEStringVal(o.DeviceProfileId.ValueString())},
		}).
		Do(&result)
	if err != nil {
		diags.AddError("error running interface map query", err.Error())
		return
	}

	switch len(result.Items) {
	case 0:
		diags.AddError("unable to assign interface_map",
			fmt.Sprintf("no interface_map links system '%s' to device profile '%s'",
				o.NodeName.ValueString(), o.DeviceProfileId.ValueString()))
	case 1:
		o.InterfaceMapId = types.StringValue(result.Items[0].InterfaceMap.Id)
	default:
		candidates := make([]string, len(result.Items))
		for i := range result.Items {
			candidates[i] = result.Items[i].InterfaceMap.Id
		}
		diags.AddError("multiple Interface Map candidates - additional configuration required",
			fmt.Sprintf("enter one of the following into interface_map to use Device Profile %q at Node %q - [%s]",
				o.DeviceProfileId.ValueString(), o.NodeName.ValueString(), strings.Join(candidates, ", ")))
	}
}

func (o *DeviceAllocation) populateSystemNodeId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	err := client.NewQuery(goapstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(o.NodeName.ValueString())},
			{"name", goapstra.QEStringVal("n_system")},
		}).
		Do(&result)
	if err != nil {
		diags.AddError("error running system node query", err.Error())
		return
	}

	switch len(result.Items) {
	case 0:
		diags.AddError("switch node not found in blueprint",
			fmt.Sprintf("switch/system node with label '%s' not found in blueprint", o.NodeName.ValueString()))
	case 1:
		// no error case
		o.SystemNodeId = types.StringValue(result.Items[0].System.Id)
	default:
		diags.AddError("multiple switches found in blueprint",
			fmt.Sprintf("switch/system node with label '%s': %d matches found in blueprint",
				o.NodeName.ValueString(), len(result.Items)))
	}
}

func (o *DeviceAllocation) PopulateDataFromGraphDb(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	deviceProfile := utils.DeviceProfileFromDeviceKey(ctx, o.DeviceKey.ValueString(), client, diags)
	if diags.HasError() {
		return
	}
	o.DeviceProfileId = types.StringValue(deviceProfile.String())

	if o.InterfaceMapId.IsUnknown() {
		o.populateInterfaceMapId(ctx, client, diags)
	} else {
		o.validateInterfaceMapId(ctx, client, diags)
	}
	if diags.HasError() {
		return
	}

	// Determine the graph db node ID of the desired system
	o.populateSystemNodeId(ctx, client, diags)
	if diags.HasError() {
		return
	}
}

func (o *DeviceAllocation) SetInterfaceMap(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	assignments := make(goapstra.SystemIdToInterfaceMapAssignment, 1)
	if o.InterfaceMapId.IsNull() {
		assignments[o.SystemNodeId.ValueString()] = nil
	} else {
		assignments[o.SystemNodeId.ValueString()] = o.InterfaceMapId.ValueString()
	}
	bpClient, err := client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(o.BlueprintId.ValueString()))
	if err != nil {
		diags.AddError("error creating blueprint client", err.Error())
		return
	}

	err = bpClient.SetInterfaceMapAssignments(ctx, assignments)
	if err != nil {
		diags.AddError(fmt.Sprintf("error (re)setting interface map for node %q", o.SystemNodeId.ValueString()), err.Error())
		return
	}
}

func (o *DeviceAllocation) SetNodeSystemId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	patch := &struct {
		SystemId string `json:"system_id"`
	}{
		SystemId: o.DeviceKey.ValueString(),
	}

	nodeId := goapstra.ObjectId(o.SystemNodeId.ValueString())
	blueprintId := goapstra.ObjectId(o.BlueprintId.ValueString())
	err := client.PatchNode(ctx, blueprintId, nodeId, &patch, nil)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to (re)assign system_id for node '%s'", nodeId), err.Error())
		return
	}
}

func (o *DeviceAllocation) GetCurrentSystemId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	err := client.NewQuery(goapstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(o.NodeName.ValueString())},
			{"name", goapstra.QEStringVal("n_system")},
		}).
		Do(&result)
	if err != nil {
		diags.AddError("error querying system_id", err.Error())
		return
	}
	switch len(result.Items) {
	case 0:
		o.SystemNodeId = types.StringNull()
	case 1:
		o.SystemNodeId = types.StringValue(result.Items[0].System.Id)
	default:
		ids := make([]string, len(result.Items))
		for i := range result.Items {
			ids[i] = result.Items[i].System.Id
		}
		diags.AddError("cannot proceed: multiple graphdb nodes share common label",
			fmt.Sprintf("label %q used by %q", o.NodeName.ValueString(), strings.Join(ids, "\", \"")))
	}
}

func (o *DeviceAllocation) GetCurrentInterfaceMapId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			InterfaceMap struct {
				Id string `json:"id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}

	err := client.NewQuery(goapstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(o.NodeName.ValueString())},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("interface_map")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"name", goapstra.QEStringVal("n_interface_map")},
		}).
		Do(&result)
	if err != nil {
		diags.AddError("error querying interface map assignment", err.Error())
		return
	}

	switch len(result.Items) {
	case 0:
		o.InterfaceMapId = types.StringNull()
	case 1:
		o.InterfaceMapId = types.StringValue(result.Items[0].InterfaceMap.Id)
	default:
		iMaps := make([]string, len(result.Items))
		for i := range result.Items {
			iMaps[i] = result.Items[i].InterfaceMap.Id
		}
		diags.AddError("cannot proceed: graphdb links system node to multiple interface maps",
			fmt.Sprintf("%q matches %q", o.NodeName.ValueString(), strings.Join(iMaps, "\", \"")))
	}
}

func (o *DeviceAllocation) GetCurrentDeviceProfileId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	var result struct {
		Items []struct {
			DeviceProfile struct {
				Id string `json:"id"`
			} `json:"n_device_profile"`
		} `json:"items"`
	}

	err := client.NewQuery(goapstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(o.NodeName.ValueString())},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("interface_map")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
		}).
		Out([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("device_profile")},
		}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("device_profile")},
			{"name", goapstra.QEStringVal("n_device_profile")},
		}).
		Do(&result)
	if err != nil {
		diags.AddError("error querying device profile", err.Error())
	}

	switch len(result.Items) {
	case 0:
		o.DeviceProfileId = types.StringNull()
	case 1:
		o.DeviceProfileId = types.StringValue(result.Items[0].DeviceProfile.Id)
	default:
		ids := make([]string, len(result.Items))
		for i := range result.Items {
			ids[i] = result.Items[i].DeviceProfile.Id
		}
		diags.AddError("cannot proceed: graphdb links system node to multiple device profiles",
			fmt.Sprintf("%q matches %q", o.NodeName.ValueString(), strings.Join(ids, "\", \"")))
	}
}

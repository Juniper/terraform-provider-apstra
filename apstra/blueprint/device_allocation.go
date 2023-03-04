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
			MarkdownDescription: "Unique ID for a device, generally the serial number.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"interface_map_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Interface Maps link a Logical Device (fabric design element) " +
				"to a Device Profile which describes a hardware model. Optional when only a single " +
				"interface map references the Logical Device describing the specified `node_name`.",
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

//func (o *DeviceAllocation) Validate(ctx context.Context, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	if !o.Role.IsUnknown() {
//		// Extract role to ResourceGroupName
//		var rgName goapstra.ResourceGroupName
//		err := rgName.FromString(o.Role.ValueString())
//		if err != nil {
//			diags.AddError(fmt.Sprintf("error parsing role %q", o.Role.ValueString()),
//				err.Error())
//			return
//		}
//
//		// Get list of poolIds from Apstra
//		var apiPoolIds []goapstra.ObjectId
//		switch rgName.Type() {
//		case goapstra.ResourceTypeAsnPool:
//			apiPoolIds, err = client.ListAsnPoolIds(ctx)
//		case goapstra.ResourceTypeIp4Pool:
//			apiPoolIds, err = client.ListIp4PoolIds(ctx)
//		case goapstra.ResourceTypeIp6Pool:
//			apiPoolIds, err = client.ListIp6PoolIds(ctx)
//		case goapstra.ResourceTypeVniPool:
//			apiPoolIds, err = client.ListVniPoolIds(ctx)
//		default:
//			diags.AddError("error determining Resource Group Type by Name",
//				fmt.Sprintf("Resource Group %q not recognized", o.Role.ValueString()))
//		}
//		if err != nil {
//			diags.AddError("error listing pool IDs", err.Error())
//		}
//		if diags.HasError() {
//			return
//		}
//
//		// Quick function to check for 'id' among 'ids'
//		contains := func(ids []goapstra.ObjectId, id goapstra.ObjectId) bool {
//			for i := range ids {
//				if ids[i] == id {
//					return true
//				}
//			}
//			return false
//		}
//
//		// Check that each PoolId configuration element appears in the API results
//		for _, elem := range o.PoolIds.Elements() {
//			id := elem.(basetypes.StringValue).ValueString()
//			if !contains(apiPoolIds, goapstra.ObjectId(id)) {
//				diags.AddError(
//					"pool not found",
//					fmt.Sprintf("pool id %q of type %q not found", id, rgName.Type().String()))
//				return
//			}
//		}
//	}
//}

//func (o *DeviceAllocation) LoadApiData(ctx context.Context, in *goapstra.ResourceGroupAllocation, diags *diag.Diagnostics) {
//	o.PoolIds = utils.SetValueOrNull(ctx, types.StringType, in.PoolIds, diags)
//}

func (o *DeviceAllocation) populateDeviceProfileId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
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

	o.DeviceProfileId = types.StringValue(si.Facts.AosHclModel.String())
}

func (o *DeviceAllocation) populateInterfaceMapId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	// Common start to the query:
	//  - our 'system' node
	//  - with an outbound 'logical_device' relationship
	//  - to a 'logical_device' node
	//  - with an inbound 'logical_device' relationship
	//  ... continued in the conditional below ...
	query := client.NewQuery(goapstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(o.NodeName.ValueString())},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("logical_device")},
		}).
		In([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}})

	if o.InterfaceMapId.IsUnknown() {
		query = query.
			// ... continue building the query ...
			// - to any 'interface_map' node (we are hoping for exactly one candidate)
			// ... more query building below when we run it ...
			Node([]goapstra.QEEAttribute{
				{"type", goapstra.QEStringVal("interface_map")},
				{"name", goapstra.QEStringVal("n_interface_map")},
			})
	} else {
		query = query.
			// ... continue building the query ...
			// - to an 'interface_map' node matching the configured ID
			// ... more query building below when we run it ...
			Node([]goapstra.QEEAttribute{
				{"type", goapstra.QEStringVal("interface_map")},
				{"name", goapstra.QEStringVal("n_interface_map")},
				{"id", goapstra.QEStringVal(o.InterfaceMapId.ValueString())},
			})
	}

	query = query.
		// ... continue building the query ...
		// - with an outbound 'device_profile' relationship
		// -- to a 'device_profile' node matching our device.
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("device_profile")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("device_profile")},
			{"device_profile_id", goapstra.QEStringVal(o.DeviceProfileId.ValueString())},
		})

	var result struct {
		Items []struct {
			InterfaceMap struct {
				Id string `json:"id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}

	err := query.Do(&result)
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
	query := client.NewQuery(goapstra.ObjectId(o.BlueprintId.ValueString())).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(o.NodeName.ValueString())},
			{"name", goapstra.QEStringVal("n_system")},
		})

	var result struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	err := query.Do(&result)
	if err != nil {
		diags.AddError("error running system node query", err.Error())
		return
	}

	switch len(result.Items) {
	case 0:
		diags.AddError("switch node not found in blueprint",
			fmt.Sprintf("switch/system node with label '%s' not found in blueprint", o.NodeName.ValueString()))
		return
	case 1:
		// no error case
		o.SystemNodeId = types.StringValue(result.Items[0].System.Id)
	default:
		diags.AddError("multiple switches found in blueprint",
			fmt.Sprintf("switch/system node with label '%s': %d matches found in blueprint",
				o.NodeName.ValueString(), len(result.Items)))
		return
	}
}

func (o *DeviceAllocation) PopulateDataFromGraphDb(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	// Determine the Device Profile by examining the blueprint graph db
	o.populateDeviceProfileId(ctx, client, diags)
	if diags.HasError() {
		return
	}

	// Determine the Device Profile by examining the config and blueprint graph db
	o.populateInterfaceMapId(ctx, client, diags)
	if diags.HasError() {
		return
	}

	// Determine the graph db node ID of the desired system
	o.populateSystemNodeId(ctx, client, diags)
	if diags.HasError() {
		return
	}
}

func (o *DeviceAllocation) AllocateDevice(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	patch := &struct {
		SystemId string `json:"system_id"`
	}{
		SystemId: o.DeviceKey.ValueString(),
	}

	nodeId := goapstra.ObjectId(o.SystemNodeId.ValueString())
	blueprintId := goapstra.ObjectId(o.BlueprintId.ValueString())
	err := client.PatchNode(ctx, blueprintId, nodeId, &patch, nil)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to assign switch device for node '%s'", nodeId), err.Error())
		return
	}
}

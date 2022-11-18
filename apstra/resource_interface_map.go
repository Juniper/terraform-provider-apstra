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
	"regexp"
	"strconv"
	"strings"
)

const (
	ldInterfaceSep     = "/"
	ldInterfaceExample = "1" + ldInterfaceSep + "2"
	ldInterfaceSynax   = "<panel>" + ldInterfaceSep + "<port>"
)

var _ resource.ResourceWithConfigure = &resourceInterfaceMap{}
var _ resource.ResourceWithValidateConfig = &resourceInterfaceMap{}

type resourceInterfaceMap struct {
	client *goapstra.Client
}

func (o *resourceInterfaceMap) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface_map"
}

func (o *resourceInterfaceMap) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (o *resourceInterfaceMap) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	var diags diag.Diagnostics
	ldpValidator, err := regexp.Compile("^[1-9][0-9]*" + ldInterfaceSep + "[1-9][0-9]*$")
	if err != nil {
		diags.AddError(
			errProviderBug,
			"error compiling regular expression for resource_interface_map logical_device_port string validation")
		return tfsdk.Schema{}, diags
	}

	return tfsdk.Schema{
		MarkdownDescription: "This resource creates an Interface Map",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Apstra ID number of the Interface Map",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				MarkdownDescription: "Interface Map name as displayed in the web UI",
				Type:                types.StringType,
				Required:            true,
			},
			"device_profile_id": {
				MarkdownDescription: "ID of Device Profile to be mapped.",
				Type:                types.StringType,
				Required:            true,
			},
			"logical_device_id": {
				MarkdownDescription: "ID of Logical Device to be mapped.",
				Type:                types.StringType,
				Required:            true,
			},
			"interfaces": {
				MarkdownDescription: "Ordered list of interface mapping info.",
				Required:            true,
				Validators:          []tfsdk.AttributeValidator{listvalidator.SizeAtLeast(1)},
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"physical_interface_name": {
						MarkdownDescription: "Interface name found in the Device Profile, e.g. \"et-0/0/1:2\"",
						Type:                types.StringType,
						Required:            true,
						Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
					},
					"logical_device_port": {
						MarkdownDescription: "Panel and Port number of logical device expressed in the form \"" +
							ldInterfaceSynax + "\". Both numbers are 1-indexed, so the 2nd port on the 1st panel " +
							"would be \"" + ldInterfaceExample + "\".",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{stringvalidator.RegexMatches(ldpValidator,
							"must be of the form \""+ldInterfaceSynax+"\", where both values are 1-indexed. "+
								"2nd port on 1st panel would be: \""+ldInterfaceExample+"\".")},
					},
					"transformation_id": {
						MarkdownDescription: "Transformation ID number identifying the desired port behavior, as found " +
							"in the Device Profile. Required only when multiple transformation candidates are found for " +
							"a given physical_interface_name and speed (as determined by the Logical Device and logical_device_port.",
						Type:       types.Int64Type,
						Optional:   true,
						Computed:   true,
						Validators: []tfsdk.AttributeValidator{int64validator.AtLeast(1)},
					},
				}),
			},
			"unused_interfaces": {
				MarkdownDescription: "Ordered list of interface mapping info for unused interfaces.",
				Computed:            true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"physical_interface_name": {
						MarkdownDescription: "Interface name found in the Device Profile, e.g. \"et-0/0/1:2\"",
						Type:                types.StringType,
						Computed:            true,
					},
					"logical_device_port": {
						MarkdownDescription: "Panel and Port number of logical device expressed in the form \"" +
							ldInterfaceSynax + "\". Both numbers are 1-indexed, so the 2nd port on the 1st panel " +
							"would be \"" + ldInterfaceExample + "\".",
						Type:     types.StringType,
						Computed: true,
					},
					"transformation_id": {
						MarkdownDescription: "Transformation ID number identifying the desired port behavior, as found " +
							"in the Device Profile.",
						Type:     types.Int64Type,
						Computed: true,
					},
				}),
			},
		},
	}, nil
}

func (o *resourceInterfaceMap) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if o.client == nil { // cannot proceed without a client
		return
	}

	var config rInterfaceMap
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configInterfaces []rInterfaceMapInterface
	diags = config.Interfaces.ElementsAs(ctx, &configInterfaces, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// maps to ensure uniqueness
	physicalInterfaceNames := make(map[string]struct{})
	logicalDevicePorts := make(map[string]struct{})

	for i, configInterface := range configInterfaces {
		if _, found := physicalInterfaceNames[configInterface.PhysicalInterfaceName]; found {
			resp.Diagnostics.AddAttributeError(path.Root("interfaces").AtListIndex(i),
				errInvalidConfig, "duplicate physical_interface_name detected")
			return
		}
		if _, found := logicalDevicePorts[configInterface.LogicalDevicePort]; found {
			resp.Diagnostics.AddAttributeError(path.Root("interfaces").AtListIndex(i),
				errInvalidConfig, "duplicate logical_device_port detected")
			return
		}
		physicalInterfaceNames[configInterface.PhysicalInterfaceName] = struct{}{}
		logicalDevicePorts[configInterface.LogicalDevicePort] = struct{}{}
	}
}

func (o *resourceInterfaceMap) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rInterfaceMap
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// fetch the Logical Device and Device Profile
	ld, dp := plan.fetchEmbeddedObjects(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.validatePortSelections(ctx, ld, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.request(ctx, ld, dp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := o.client.CreateInterfaceMap(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating Interface Map", err.Error())
		return
	}

	var state rInterfaceMap

	// id is not in the goapstra.InterfaceMapData object we're using, so set it directly
	state.Id = types.StringValue(string(id))

	state.parseApi(ctx, request, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (o *resourceInterfaceMap) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state rInterfaceMap
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Interface Map from API and then update what is in state from what the API returns
	iMap, err := o.client.GetInterfaceMap(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error reading Interface Map", err.Error())
			return
		}
	}

	var newState rInterfaceMap
	newState.Id = types.StringValue(string(iMap.Id))
	newState.parseApi(ctx, iMap.Data, &resp.Diagnostics)

	// Set state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (o *resourceInterfaceMap) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan rInterfaceMap
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state rInterfaceMap
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := goapstra.ObjectId(state.Id.ValueString())
	ld, dp := plan.fetchEmbeddedObjects(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.request(ctx, ld, dp, &resp.Diagnostics)
	err := o.client.UpdateInterfaceMap(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError("error updating Interface Map", err.Error())
		return
	}

	var newState rInterfaceMap

	// id is not in the goapstra.InterfaceMapData object we're using, so set it directly
	newState.Id = types.StringValue(string(id))

	newState.parseApi(ctx, request, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (o *resourceInterfaceMap) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rInterfaceMap
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Interface Map by calling API
	err := o.client.DeleteInterfaceMap(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound {
			resp.Diagnostics.AddError(
				"error deleting Interface Map", err.Error())
		}
		return
	}
}

type rInterfaceMap struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DeviceProfileId  types.String `tfsdk:"device_profile_id"`
	LogicalDeviceId  types.String `tfsdk:"logical_device_id"`
	Interfaces       types.List   `tfsdk:"interfaces"`
	UnusedInterfaces types.List   `tfsdk:"unused_interfaces"`
}

func (o *rInterfaceMap) fetchEmbeddedObjects(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) (*goapstra.LogicalDevice, *goapstra.DeviceProfile) {
	var ace goapstra.ApstraClientErr
	// fetch the logical device
	ld, err := client.GetLogicalDevice(ctx, goapstra.ObjectId(o.LogicalDeviceId.ValueString()))
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			diags.AddAttributeError(path.Root("logical_device_id"), errInvalidConfig,
				fmt.Sprintf(fmt.Sprintf("logical device'%s' not found", o.DeviceProfileId.ValueString())))
		}
		diags.AddError("error while fetching logical device", err.Error())
	}

	// fetch the device profile specified by the user
	dp, err := client.GetDeviceProfile(ctx, goapstra.ObjectId(o.DeviceProfileId.ValueString()))
	if err != nil {
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			diags.AddAttributeError(path.Root("device_profile_id"), errInvalidConfig,
				fmt.Sprintf(fmt.Sprintf("device profile '%s' not found", o.DeviceProfileId.ValueString())))
		}
		diags.AddError("error while fetching device profile", err.Error())
	}

	return ld, dp
}

func (o *rInterfaceMap) interfaces(ctx context.Context, diags *diag.Diagnostics) []rInterfaceMapInterface {
	var result []rInterfaceMapInterface
	diags.Append(o.Interfaces.ElementsAs(ctx, &result, true)...)
	return result
}

func (o *rInterfaceMap) ldPortNames(ctx context.Context, diags *diag.Diagnostics) []string {
	interfaces := o.interfaces(ctx, diags)
	if diags.HasError() {
		return nil
	}

	result := make([]string, len(interfaces))
	for i, planIntf := range interfaces {
		result[i] = planIntf.LogicalDevicePort
	}
	return result
}

// validatePortSelections ensures that (a) all logical device ports in o appear in ld
// and (b) no ports defined in ld are missing from o
func (o *rInterfaceMap) validatePortSelections(ctx context.Context, ld *goapstra.LogicalDevice, diags *diag.Diagnostics) {
	plannedPortNames := o.ldPortNames(ctx, diags)
	if diags.HasError() {
		return
	}

	ldii := getLogicalDevicePortInfo(ld)
	requiredPortNames := make([]string, len(ldii))
	var i int
	for k := range ldii {
		requiredPortNames[i] = k
		i++
	}

	var bogusPortNames []string
	var n int
	for i = range plannedPortNames {
		requiredPortNames, n = sliceWithoutString(requiredPortNames, plannedPortNames[i])
		if n == 0 {
			bogusPortNames = append(bogusPortNames, plannedPortNames[i])
		}
		if n > 1 {
			diags.AddError(errProviderBug,
				fmt.Sprintf("logical device '%s' has two instances of port '%s'",
					ld.Id, plannedPortNames[i]))
		}
	}

	if len(bogusPortNames) > 0 {
		diags.AddError(errInvalidConfig,
			fmt.Sprintf("ports '%s' not defined by logical device '%s'",
				strings.Join(bogusPortNames, ","), ld.Id))
	}

	if len(requiredPortNames) != 0 {
		diags.AddError(errInvalidConfig,
			fmt.Sprintf("ports '%s' required by logical device '%s' must be specified",
				strings.Join(requiredPortNames, ","), ld.Id))
	}
}

func (o *rInterfaceMap) iMapInterfaces(ctx context.Context, ld *goapstra.LogicalDevice, dp *goapstra.DeviceProfile, diags *diag.Diagnostics) []goapstra.InterfaceMapInterface {
	// extract interface list from plan
	var planInterfaces []rInterfaceMapInterface
	diags.Append(o.Interfaces.ElementsAs(ctx, &planInterfaces, true)...)
	if diags.HasError() {
		return nil
	}

	ldpiMap := getLogicalDevicePortInfo(ld)
	if len(ldpiMap) != len(planInterfaces) {
		diags.AddError(errProviderBug,
			fmt.Sprintf("%d planned interfaces and %d logical device interfaces - "+
				"we should have caught this earlier", len(ldpiMap), len(planInterfaces)))
		return nil
	}

	result := make([]goapstra.InterfaceMapInterface, len(planInterfaces))
	portIdToSelectedTransformId := make(map[int]int) // to ensure we don't double-dip transform IDs

	for i, planInterface := range planInterfaces {
		// extract the logical device panel and port number (1 indexed)
		ldPanel, ldPort := ldPanelAndPortFromString(planInterface.LogicalDevicePort, diags)
		if diags.HasError() {
			return nil
		}

		// ldpi (logical device port info) is the ldPortInfo (speed and roles)
		// associated with this interface
		ldpi := ldpiMap[planInterface.LogicalDevicePort]

		// extract candidate transformations from the DP PortInfo based on the
		// configured physical interface name, and the speed indicated by the LD
		portId, transformations := getPortIdAndTransformations(dp, ldpi.Speed, planInterface.PhysicalInterfaceName, diags)
		if diags.HasError() {
			return nil
		}

		var transformId int
		//var transformation goapstra.Transformation
		//var ok bool
		if planInterface.TransformationId != nil { // plan includes a transform #
			if transformation, ok := transformations[int(*planInterface.TransformationId)]; ok {
				transformId = transformation.TransformationId
			} else {
				diags.AddError(errInvalidConfig,
					fmt.Sprintf("planned transform %d for logical device interface "+
						"'%s' not available using device profile '%s' and interface '%s'",
						planInterface.TransformationId,
						planInterface.LogicalDevicePort,
						o.LogicalDeviceId.ValueString(),
						planInterface.PhysicalInterfaceName))
				return nil
			}
		} else { // plan does not include a transform #
			if len(transformations) == 1 { // we got exactly one candidate -- use it!
				for k, _ := range transformations { // loop runs once, copies the only map key
					transformId = k
				}
			} else {
				// we got multiple transform candidates - tell the user to specify one
				dump, err := json.MarshalIndent(&transformations, "", "  ")
				if err != nil {
					diags.AddError("error marshaling transformation candidates", err.Error())
				}
				diags.AddAttributeError(path.Root("interfaces").AtListIndex(i),
					"selected physical port supports multiple transformations - indicate selection with 'transform_id'",
					"\n"+string(dump))
				return nil
			}
		}

		if previousTransformId, ok := portIdToSelectedTransformId[portId]; !ok {
			// no transform ID previously selected for this port id.
			// save the transform ID for future checks.
			portIdToSelectedTransformId[portId] = transformId
		} else {
			// transform ID previously selected for this port id.
			// it better match...
			if previousTransformId != transformId {
				diags.AddError(errInvalidConfig,
					fmt.Sprintf("configuration selects both transformation %d and %d for device profile port id %d",
						previousTransformId, transformId, portId))
				return nil
			}
		}

		transformation := transformations[transformId]
		interfaceIdx := -1
		for j, intf := range transformation.Interfaces {
			if planInterface.PhysicalInterfaceName == intf.Name {
				interfaceIdx = j
				break
			}
		}
		if interfaceIdx == -1 {
			diags.AddError(errProviderBug, "failed to set interfaceIdx")
			return nil
		}

		transformInterface := transformation.Interfaces[interfaceIdx]

		result[i] = goapstra.InterfaceMapInterface{
			Name:  planInterface.PhysicalInterfaceName,
			Roles: ldpiMap[planInterface.LogicalDevicePort].Roles,
			Mapping: goapstra.InterfaceMapMapping{
				DPPortId:      portId,
				DPTransformId: transformation.TransformationId,
				DPInterfaceId: transformInterface.InterfaceId,
				LDPanel:       ldPanel,
				LDPort:        ldPort,
			},
			ActiveState: transformInterface.State == "active",
			Position:    i + 1,
			Speed:       transformInterface.Speed,
			Setting: goapstra.InterfaceMapInterfaceSetting{
				Param: transformation.Interfaces[interfaceIdx].Setting,
			},
		}
	}
	return result
}

func (o *rInterfaceMap) request(ctx context.Context, ld *goapstra.LogicalDevice, dp *goapstra.DeviceProfile, diags *diag.Diagnostics) *goapstra.InterfaceMapData {
	allocatedInterfaces := o.iMapInterfaces(ctx, ld, dp, diags)
	if diags.HasError() {
		return nil
	}

	unallocatedInterfaces := iMapUnallocaedInterfaces(allocatedInterfaces, dp, diags)
	if diags.HasError() {
		return nil
	}

	return &goapstra.InterfaceMapData{
		LogicalDeviceId: ld.Id,
		DeviceProfileId: dp.Id,
		Label:           o.Name.ValueString(),
		Interfaces:      append(allocatedInterfaces, unallocatedInterfaces...),
	}
}

func (o *rInterfaceMap) parseApi(ctx context.Context, in *goapstra.InterfaceMapData, diags *diag.Diagnostics) {
	// create two slices. Data from elements of in.Interfaces will filter into one of these depending
	// on whether the element represents an "in use" interface. both receiving slices are oversize.
	a := make([]rInterfaceMapInterface, len(in.Interfaces))
	b := make([]rInterfaceMapInterface, len(in.Interfaces))

	var aIdx, bIdx int              // aIdx and bIdx keep track of our location in the "a" and "b" slices...
	var intf rInterfaceMapInterface // used to parse each element of in.Interfaces

	for i := range in.Interfaces { // i keeps track of our location in the in.Interfaces slice...
		// parse the interface object
		intf.parseApi(&in.Interfaces[i])

		// logical device port -1/-1 indicates bogus assignment of physical port (unused)
		if intf.LogicalDevicePort != "-1/-1" {
			a[aIdx] = intf
			aIdx++
		} else {
			b[bIdx] = intf
			bIdx++
		}
	}

	a = a[:aIdx] // trim the slice of allocated ports to size
	b = b[:bIdx] // trim the slice of unallocated ports to size

	aList, d := types.ListValueFrom(ctx, rInterfaceMapInterface{}.attrType(), a)
	diags.Append(d...)

	bList, d := types.ListValueFrom(ctx, rInterfaceMapInterface{}.attrType(), b)
	diags.Append(d...)

	o.Name = types.StringValue(in.Label)
	o.DeviceProfileId = types.StringValue(string(in.DeviceProfileId))
	o.LogicalDeviceId = types.StringValue(string(in.LogicalDeviceId))
	o.Interfaces = aList
	o.UnusedInterfaces = bList
}

type rInterfaceMapInterface struct {
	PhysicalInterfaceName string `tfsdk:"physical_interface_name"`
	LogicalDevicePort     string `tfsdk:"logical_device_port"`
	TransformationId      *int64 `tfsdk:"transformation_id"`
}

func (o rInterfaceMapInterface) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"physical_interface_name": types.StringType,
			"logical_device_port":     types.StringType,
			"transformation_id":       types.Int64Type}}
}

func (o *rInterfaceMapInterface) parseApi(in *goapstra.InterfaceMapInterface) {
	transformId := int64(in.Mapping.DPTransformId)
	o.TransformationId = &transformId
	o.LogicalDevicePort = fmt.Sprintf("%d%s%d", in.Mapping.LDPanel, ldInterfaceSep, in.Mapping.LDPort)
	o.PhysicalInterfaceName = in.Name
}

func ldPanelAndPortFromString(in string, diags *diag.Diagnostics) (int, int) {
	panelAndPort := strings.Split(in, ldInterfaceSep)
	if len(panelAndPort) != 2 {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error splitting interface name '%s'", in))
		return 0, 0
	}
	ldPanel, err := strconv.Atoi(panelAndPort[0])
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing logical device panel id from '%s'", panelAndPort[0])+err.Error())
	}
	ldPort, err := strconv.Atoi(panelAndPort[1])
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing logical device port id from '%s'", panelAndPort[1])+err.Error())
	}
	return ldPanel, ldPort
}

type ldPortInfo struct {
	Speed goapstra.LogicalDevicePortSpeed
	Roles goapstra.LogicalDevicePortRoleFlags
}

func getLogicalDevicePortInfo(ld *goapstra.LogicalDevice) map[string]ldPortInfo {
	result := make(map[string]ldPortInfo)
	for panelIdx, panel := range ld.Data.Panels {
		panelNum := panelIdx + 1
		nextPort := 1
		for _, portGrp := range panel.PortGroups {
			for i := 0; i < portGrp.Count; i++ {
				result[fmt.Sprintf("%d%s%d", panelNum, ldInterfaceSep, nextPort)] = ldPortInfo{
					Speed: portGrp.Speed,
					Roles: portGrp.Roles,
				}
				nextPort++
			}
		}
	}
	return result
}

// getPortIdAndTransformations takes a device profile, speed and physical
// interface name. It is expected that exactly one port in the device profile
// should match the supplied physical interface name. It returns the matching
// port ID and a map of "active" transformations keyed by transformtion ID.
// Only transformations matching the specified physical interface name and speed
// are returned.
func getPortIdAndTransformations(dp *goapstra.DeviceProfile, speed goapstra.LogicalDevicePortSpeed, phyIntfName string, diags *diag.Diagnostics) (int, map[int]goapstra.Transformation) {
	// find the device profile "port info" by physical port name (expecting exactly one match from DP)
	dpPort, err := dp.PortByInterfaceName(phyIntfName)
	if err != nil {
		diags.AddError(errInvalidConfig,
			fmt.Sprintf("device profile '%s' has no ports which use name '%s'",
				dp.Id, phyIntfName))
		return 0, nil
	}

	transformations := dpPort.TransformationCandidates(phyIntfName, speed)
	if len(transformations) == 0 {
		diags.AddError(errInvalidConfig,
			fmt.Sprintf("no active port in device profile '%s' matches interface name '%s' with the logical "+
				"device port speed '%s'", dp.Id,
				phyIntfName, speed))
		return 0, nil
	}
	return dpPort.PortId, transformations
}

func iMapUnallocaedInterfaces(allocatedPorts []goapstra.InterfaceMapInterface, dp *goapstra.DeviceProfile, diags *diag.Diagnostics) []goapstra.InterfaceMapInterface {
	// make a map[portId]struct{} so we can quickly determine whether
	// a port ID has been previously allocated.
	allocatedPortCount := len(allocatedPorts)
	allocatedPortIds := make(map[int]struct{}, allocatedPortCount)
	for _, ap := range allocatedPorts {
		allocatedPortIds[ap.Mapping.DPPortId] = struct{}{}
	}

	missingAllocationCount := len(dp.Data.Ports) - len(allocatedPortIds) // device profile ports - used port IDs (ignore breakout ports)

	result := make([]goapstra.InterfaceMapInterface, missingAllocationCount)
	var i int
	for _, dpPort := range dp.Data.Ports {
		if _, ok := allocatedPortIds[dpPort.PortId]; ok {
			continue
		}

		transformation := dpPort.DefaultTransform()
		if transformation == nil {
			diags.AddError(errProviderBug, "port has no default transformation")
		}

		result[i] = goapstra.InterfaceMapInterface{
			Name:  transformation.Interfaces[0].Name,
			Roles: goapstra.LogicalDevicePortRoleUnused,
			Mapping: goapstra.InterfaceMapMapping{
				DPPortId:      dpPort.PortId,
				DPTransformId: transformation.TransformationId,
				DPInterfaceId: transformation.Interfaces[0].InterfaceId, // blindly use the first interface - UI seems to do this and testing shows there's always at least 1
				LDPanel:       -1,
				LDPort:        -1,
			},
			ActiveState: true, // unclear what this is, UI sets "active"
			Position:    allocatedPortCount + i + 1,
			Speed:       transformation.Interfaces[0].Speed,
			Setting:     goapstra.InterfaceMapInterfaceSetting{Param: transformation.Interfaces[0].Setting},
		}

		i++
	}
	return result
}

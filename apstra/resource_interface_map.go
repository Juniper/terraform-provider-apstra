package tfapstra

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"sort"
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
var _ resourceWithSetClient = &resourceInterfaceMap{}

type resourceInterfaceMap struct {
	client *apstra.Client
}

func (o *resourceInterfaceMap) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface_map"
}

func (o *resourceInterfaceMap) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceInterfaceMap) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This resource creates an Interface Map",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Apstra ID number of the Interface Map",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Interface Map name as displayed in the web UI",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"device_profile_id": schema.StringAttribute{
				MarkdownDescription: "ID of Device Profile to be mapped.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"logical_device_id": schema.StringAttribute{
				MarkdownDescription: "ID of Logical Device to be mapped.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"interfaces": schema.SetNestedAttribute{
				MarkdownDescription: "Ordered list of interface mapping info.",
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: rInterfaceMapInterface{}.attributes(),
				},
			},
			"unused_interfaces": schema.SetNestedAttribute{
				MarkdownDescription: "Ordered list of interface mapping info for unused interfaces.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: rInterfaceMapInterface{}.unusedAttributes(),
				},
			},
		},
	}
}

func (o *resourceInterfaceMap) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if o.client == nil { // cannot proceed without a client
		return
	}

	var config rInterfaceMap
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configInterfaces []rInterfaceMapInterface
	resp.Diagnostics.Append(config.Interfaces.ElementsAs(ctx, &configInterfaces, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// maps to ensure uniqueness
	physicalInterfaceNames := make(map[string]struct{})
	logicalDevicePorts := make(map[string]struct{})

	for _, configInterface := range configInterfaces {
		physicalInterfaceName := configInterface.PhysicalInterfaceName.ValueString()
		logicalDevicePortName := configInterface.LogicalDevicePort.ValueString()

		if _, found := physicalInterfaceNames[physicalInterfaceName]; found {
			resp.Diagnostics.AddAttributeError(path.Root("interfaces"),
				errInvalidConfig, fmt.Sprintf("duplicate physical_interface_name '%s' detected", physicalInterfaceName))
			return
		}
		if _, found := logicalDevicePorts[logicalDevicePortName]; found {
			resp.Diagnostics.AddAttributeError(path.Root("interfaces"),
				errInvalidConfig, fmt.Sprintf("duplicate logical_device_port '%s' detected", logicalDevicePortName))
			return
		}

		physicalInterfaceNames[physicalInterfaceName] = struct{}{}
		logicalDevicePorts[logicalDevicePortName] = struct{}{}
	}
}

func (o *resourceInterfaceMap) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan rInterfaceMap
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	// create new state object
	var state rInterfaceMap

	// id is not in the apstra.InterfaceMapData object we're using, so set it directly
	state.Id = types.StringValue(string(id))

	state.loadApiData(ctx, request, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceInterfaceMap) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state rInterfaceMap
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Interface Map from API and then update what is in state from what the API returns
	iMap, err := o.client.GetInterfaceMap(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
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
	newState.loadApiData(ctx, iMap.Data, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceInterfaceMap) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan rInterfaceMap
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	err := o.client.UpdateInterfaceMap(ctx, apstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		resp.Diagnostics.AddError("error updating Interface Map", err.Error())
		return
	}

	// create new state object
	var newState rInterfaceMap

	// id is not in the apstra.InterfaceMapData object we're using, so set it directly
	newState.Id = plan.Id

	newState.loadApiData(ctx, request, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Delete resource
func (o *resourceInterfaceMap) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state rInterfaceMap
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Interface Map by calling API
	err := o.client.DeleteInterfaceMap(ctx, apstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError(
			"error deleting Interface Map", err.Error())
		return
	}
}

func (o *resourceInterfaceMap) setClient(client *apstra.Client) {
	o.client = client
}

type rInterfaceMap struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DeviceProfileId  types.String `tfsdk:"device_profile_id"`
	LogicalDeviceId  types.String `tfsdk:"logical_device_id"`
	Interfaces       types.Set    `tfsdk:"interfaces"`
	UnusedInterfaces types.Set    `tfsdk:"unused_interfaces"`
}

func (o *rInterfaceMap) fetchEmbeddedObjects(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) (*apstra.LogicalDevice, *apstra.DeviceProfile) {
	// fetch the logical device
	ld, err := client.GetLogicalDevice(ctx, apstra.ObjectId(o.LogicalDeviceId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			diags.AddAttributeError(path.Root("logical_device_id"), errInvalidConfig,
				fmt.Sprintf("logical device %q not found", o.DeviceProfileId))
		} else {
			diags.AddError("failed to fetch logical device", err.Error())
		}
	}

	// fetch the device profile specified by the user
	dp, err := client.GetDeviceProfile(ctx, apstra.ObjectId(o.DeviceProfileId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			diags.AddAttributeError(path.Root("device_profile_id"), errInvalidConfig,
				fmt.Sprintf("device profile %q not found", o.DeviceProfileId))
		} else {
			diags.AddError("failed to fetch device profile", err.Error())
		}
	}

	return ld, dp
}

func (o *rInterfaceMap) interfaces(ctx context.Context, diags *diag.Diagnostics) []rInterfaceMapInterface {
	result := make([]rInterfaceMapInterface, len(o.Interfaces.Elements()))
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
		result[i] = planIntf.LogicalDevicePort.ValueString()
	}

	return result
}

// validatePortSelections ensures that (a) all logical device ports in o appear in ld
// and (b) no ports defined in ld are missing from o
func (o *rInterfaceMap) validatePortSelections(ctx context.Context, ld *apstra.LogicalDevice, diags *diag.Diagnostics) {
	plannedPortNames := o.ldPortNames(ctx, diags)
	if diags.HasError() {
		return
	}

	// prepare a slice of port names required by the apstra.LogicalDeviceData []string{"1/1", "1/2", ...etc...}
	ldii := getLogicalDevicePortInfo(ld)
	requiredPortNames := make([]string, len(ldii))
	var i int
	for k := range ldii {
		requiredPortNames[i] = k
		i++
	}

	// collect names of ports which appear in the logical device, but are not part of the interface map
	var bogusPortNames []string
	var n int
	for i = range plannedPortNames {
		requiredPortNames, n = sliceWithoutElement(requiredPortNames, plannedPortNames[i])
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
		sort.Strings(bogusPortNames)
		diags.AddError(errInvalidConfig,
			fmt.Sprintf("ports '%s' not defined by logical device '%s'",
				strings.Join(bogusPortNames, ","), ld.Id))
	}

	if len(requiredPortNames) != 0 {
		sort.Strings(requiredPortNames)
		diags.AddError(errInvalidConfig,
			fmt.Sprintf("ports '%s' required by logical device '%s' must be specified",
				strings.Join(requiredPortNames, ","), ld.Id))
	}
}

// iMapInterfaces returns a []InterfaceMapInterface representing interfaces
// allocated from dp to ld according to the rules specified in o. The returned
// []InterfaceMapInterface also includes unallocated interfaces (marked as
// unused) belonging to the same transformation as any allocated interfaces.
// This satisfies Apstra's requirement that all interfaces belonging to a
// transformation be mapped together.
func (o *rInterfaceMap) iMapInterfaces(ctx context.Context, ld *apstra.LogicalDevice, dp *apstra.DeviceProfile, diags *diag.Diagnostics) []apstra.InterfaceMapInterface {
	// extract interface list from plan
	var planInterfaces []rInterfaceMapInterface
	diags.Append(o.Interfaces.ElementsAs(ctx, &planInterfaces, true)...)
	if diags.HasError() {
		return nil
	}

	// portIdToUnusedInterfaces is a map of per-transform interface IDs
	// keyed by port ID (physical switch interface). We use it to track unused
	// interfaces within a transformation because the Apstra API requires that
	// all interfaces within a transformation be mapped together, even those
	// which are not used.
	portIdToUnusedInterfaces := make(map[int]unusedInterfaces)

	ldpiMap := getLogicalDevicePortInfo(ld)
	if len(ldpiMap) != len(planInterfaces) {
		diags.AddError(errProviderBug,
			fmt.Sprintf("%d planned interfaces and %d logical device interfaces - "+
				"we should have caught this earlier", len(ldpiMap), len(planInterfaces)))
		return nil
	}

	result := make([]apstra.InterfaceMapInterface, len(planInterfaces))
	portIdToSelectedTransformId := make(map[int]int) // to ensure we don't double-dip transform IDs

	for i, planInterface := range planInterfaces {
		// extract the logical device panel and port number (1 indexed)
		ldPanel, ldPort := ldPanelAndPortFromString(planInterface.LogicalDevicePort.ValueString(), diags)
		if diags.HasError() {
			return nil
		}

		// ldpi (logical device port info) is the ldPortInfo (speed and roles)
		// associated with this interface
		ldpi, ok := ldpiMap[planInterface.LogicalDevicePort.ValueString()]
		if !ok {
			av, d := types.ObjectValueFrom(ctx, rInterfaceMapInterface{}.attrTypes(), &planInterface)
			diags.Append(d...)
			diags.AddAttributeError(
				path.Root("interfaces").AtSetValue(av),
				errInvalidConfig,
				fmt.Sprintf("Specified interface %s does not exist in logical device %q.\n"+
					"In addition to being a configuration bug, there may be a provider bug as well."+
					"This condition should have been caught by earlier validation. Please report this"+
					"error to the provider developers.", planInterface.LogicalDevicePort, ld.Id))
			return nil
		}

		// extract candidate transformations from the DP PortInfo based on the
		// configured physical interface name, and the speed indicated by the LD
		portId, transformations := getPortIdAndTransformations(dp, ldpi.Speed, planInterface.PhysicalInterfaceName.ValueString(), diags)
		if diags.HasError() {
			return nil
		}

		var transformId int
		if planInterface.TransformationId.IsUnknown() { // plan does not include a transform id
			if len(transformations) == 1 { // we got exactly one candidate -- use it!
				for k := range transformations { // loop runs once, copies the only map key
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
		} else { // plan includes a transform id
			if transformation, ok := transformations[int(planInterface.TransformationId.ValueInt64())]; ok {
				transformId = transformation.TransformationId
			} else {
				diags.AddError(errInvalidConfig,
					fmt.Sprintf("planned transform %d for logical device"+
						" interface at index %d ('%s') not available using device"+
						" profile '%s' and interface '%s'",
						planInterface.TransformationId.ValueInt64(),
						i,
						planInterface.LogicalDevicePort.ValueString(),
						o.LogicalDeviceId.ValueString(),
						planInterface.PhysicalInterfaceName.ValueString()))
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
			if planInterface.PhysicalInterfaceName.ValueString() == intf.Name {
				interfaceIdx = j
				break
			}
		}
		if interfaceIdx == -1 {
			diags.AddError(errProviderBug, "failed to set interfaceIdx")
			return nil
		}

		transformInterface := transformation.Interfaces[interfaceIdx]

		// at this point we have selected a transformation and an interface within it.

		// the portIdToUnusedInterfaces map's entry for this port must
		// have our selected per-transform interface ID removed. Either that's
		// an edit to the current map entry, or it's a new map entry.
		if transformInterfacesUnused, found := portIdToUnusedInterfaces[portId]; found {
			// we've already been tracking this port+transform interfaces
			// remove this interface ID from the per-transform list of unused interfaces
			unused, _ := sliceWithoutElement(transformInterfacesUnused.interfaces, transformInterface.InterfaceId)
			portIdToUnusedInterfaces[portId] = unusedInterfaces{
				transformId: transformId,
				interfaces:  unused,
			}
			//sliceWithoutInt(transformInterfacesUnused, transformInterface.InterfaceId)
		} else {
			// New port+transform.
			// Add it to the tracking map with this interface ID removed from the list
			unused, _ := sliceWithoutElement(transformation.InterfaceIds(), transformInterface.InterfaceId)
			portIdToUnusedInterfaces[portId] = unusedInterfaces{
				transformId: transformId,
				interfaces:  unused,
			}
		}

		result[i] = apstra.InterfaceMapInterface{
			Name:  planInterface.PhysicalInterfaceName.ValueString(),
			Roles: ldpiMap[planInterface.LogicalDevicePort.ValueString()].Roles,
			Mapping: apstra.InterfaceMapMapping{
				DPPortId:      portId,
				DPTransformId: transformation.TransformationId,
				DPInterfaceId: transformInterface.InterfaceId,
				LDPanel:       ldPanel,
				LDPort:        ldPort,
			},
			ActiveState: transformInterface.State == "active",
			Position:    portId,
			Speed:       transformInterface.Speed,
			Setting: apstra.InterfaceMapInterfaceSetting{
				Param: transformation.Interfaces[interfaceIdx].Setting,
			},
		}
	}

	// now loop over any interfaces languishing in portIdToUnusedInterfaces.
	// These likely unused members of breakout ports. Apstra requires them to be
	// included in the interface map.
	positionIdx := len(planInterfaces) + 1
	for portId, unused := range portIdToUnusedInterfaces {
		for _, unusedInterfaceId := range unused.interfaces {
			portInfo, err := dp.Data.PortById(portId)
			if err != nil {
				diags.AddError("error getting unused port by ID", err.Error())
				return nil
			}
			transformation, err := portInfo.Transformation(unused.transformId)
			if err != nil {
				diags.AddError("error getting transformation by ID", err.Error())
				return nil
			}
			intf, err := transformation.Interface(unusedInterfaceId)
			if err != nil {
				diags.AddError("error getting transformation interface by ID", err.Error())
				return nil
			}
			result = append(result, apstra.InterfaceMapInterface{
				Name:  intf.Name,
				Roles: apstra.LogicalDevicePortRoleUnused,
				Mapping: apstra.InterfaceMapMapping{
					DPPortId:      portId,
					DPTransformId: unused.transformId,
					DPInterfaceId: unusedInterfaceId,
					LDPanel:       -1,
					LDPort:        -1,
				},
				ActiveState: true,
				Position:    positionIdx,
				Speed:       intf.Speed,
				Setting: apstra.InterfaceMapInterfaceSetting{
					Param: intf.Setting,
				},
			})
			positionIdx++
		}
	}
	return result
}

type unusedInterfaces struct {
	transformId int
	interfaces  []int
}

func (o *rInterfaceMap) request(ctx context.Context, ld *apstra.LogicalDevice, dp *apstra.DeviceProfile, diags *diag.Diagnostics) *apstra.InterfaceMapData {
	allocatedInterfaces := o.iMapInterfaces(ctx, ld, dp, diags)
	if diags.HasError() {
		return nil
	}

	unallocatedInterfaces := iMapUnallocaedInterfaces(allocatedInterfaces, dp, diags)
	if diags.HasError() {
		return nil
	}

	return &apstra.InterfaceMapData{
		LogicalDeviceId: ld.Id,
		DeviceProfileId: dp.Id,
		Label:           o.Name.ValueString(),
		Interfaces:      append(allocatedInterfaces, unallocatedInterfaces...),
	}
}

func (o *rInterfaceMap) loadApiData(ctx context.Context, in *apstra.InterfaceMapData, diags *diag.Diagnostics) {
	// create two slices. Data from elements of in.Interfaces will filter into one of these depending
	// on whether the element represents an "in use" interface. both receiving slices are oversize.
	a := make([]rInterfaceMapInterface, len(in.Interfaces)) // allocated / in use interfaces
	b := make([]rInterfaceMapInterface, len(in.Interfaces)) // un-allocated / not in use interfaces

	var aIdx, bIdx int              // aIdx and bIdx keep track of our location in the "a" and "b" slices...
	var intf rInterfaceMapInterface // used to parse each element of in.Interfaces

	for i := range in.Interfaces { // i keeps track of our location in the in.Interfaces slice...
		// parse the interface object
		intf.loadApiData(ctx, &in.Interfaces[i], diags)

		// add interface to the used or un-used map according to whether the logical device port ID is null
		if intf.LogicalDevicePort.IsNull() {
			b[bIdx] = intf
			bIdx++
		} else {
			a[aIdx] = intf
			aIdx++
		}
	}

	a = a[:aIdx] // trim the slice of allocated ports to size
	b = b[:bIdx] // trim the slice of unallocated ports to size

	aList, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: rInterfaceMapInterface{}.attrTypes()}, a)
	diags.Append(d...)

	bList, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: rInterfaceMapInterface{}.attrTypes()}, b)
	diags.Append(d...)

	o.Name = types.StringValue(in.Label)
	o.DeviceProfileId = types.StringValue(string(in.DeviceProfileId))
	o.LogicalDeviceId = types.StringValue(string(in.LogicalDeviceId))
	o.Interfaces = aList
	o.UnusedInterfaces = bList
}

type rInterfaceMapInterface struct {
	PhysicalInterfaceName types.String `tfsdk:"physical_interface_name"`
	LogicalDevicePort     types.String `tfsdk:"logical_device_port"`
	TransformationId      types.Int64  `tfsdk:"transformation_id"`
}

func (o rInterfaceMapInterface) attributes() map[string]schema.Attribute {
	ldpValidator := regexp.MustCompile("^[1-9][0-9]*" + ldInterfaceSep + "[1-9][0-9]*$")

	return map[string]schema.Attribute{
		"physical_interface_name": schema.StringAttribute{
			MarkdownDescription: "Interface name found in the Device Profile, e.g. \"et-0/0/1:2\"",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"logical_device_port": schema.StringAttribute{
			MarkdownDescription: "Panel and Port number of logical device expressed in the form \"" +
				ldInterfaceSynax + "\". Both numbers are 1-indexed, so the 2nd port on the 1st panel " +
				"would be \"" + ldInterfaceExample + "\".",
			Required: true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(ldpValidator, "must be of the form \""+
					ldInterfaceSynax+"\", where both values are 1-indexed. "+
					"2nd port on 1st panel would be: \""+ldInterfaceExample+"\"."),
			},
		},
		"transformation_id": schema.Int64Attribute{
			MarkdownDescription: "Transformation ID number identifying the desired port behavior, detailed " +
				"in the Device Profile. Required only when multiple transformation candidates are found for " +
				"a given physical_interface_name and speed as determined by definitions found the Logical " +
				"Device definition and logical_device_port field.",
			Optional:      true,
			Computed:      true,
			Validators:    []validator.Int64{int64validator.AtLeast(1)},
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
	}
}

func (o rInterfaceMapInterface) unusedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"physical_interface_name": schema.StringAttribute{
			MarkdownDescription: "Interface name found in the Device Profile, e.g. \"et-0/0/1:2\"",
			Computed:            true,
		},
		"logical_device_port": schema.StringAttribute{
			MarkdownDescription: "Panel and Port number of logical device expressed in the form \"" +
				ldInterfaceSynax + "\". Both numbers are 1-indexed, so the 2nd port on the 1st panel " +
				"would be \"" + ldInterfaceExample + "\".",
			Computed: true,
		},
		"transformation_id": schema.Int64Attribute{
			MarkdownDescription: "Transformation ID number identifying the desired port behavior, as found " +
				"in the Device Profile.",
			Computed: true,
		},
	}
}

func (o rInterfaceMapInterface) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"physical_interface_name": types.StringType,
		"logical_device_port":     types.StringType,
		"transformation_id":       types.Int64Type,
	}
}

func (o *rInterfaceMapInterface) loadApiData(_ context.Context, in *apstra.InterfaceMapInterface, _ *diag.Diagnostics) {
	o.PhysicalInterfaceName = types.StringValue(in.Name)
	o.TransformationId = types.Int64Value(int64(in.Mapping.DPTransformId))

	if in.Mapping.LDPanel == -1 || in.Mapping.LDPort == -1 { // "-1/-1" is the sign of an unallocated interface
		o.LogicalDevicePort = types.StringNull()
	} else {
		o.LogicalDevicePort = types.StringValue(fmt.Sprintf("%d%s%d", in.Mapping.LDPanel, ldInterfaceSep, in.Mapping.LDPort))
	}
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
	Speed apstra.LogicalDevicePortSpeed
	Roles apstra.LogicalDevicePortRoleFlags
}

// getLogicalDevicePortInfo extracts a map[string]ldPortInfo keyed by logical
// device panel/port number, e.g. "1/3"
func getLogicalDevicePortInfo(ld *apstra.LogicalDevice) map[string]ldPortInfo {
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
func getPortIdAndTransformations(dp *apstra.DeviceProfile, speed apstra.LogicalDevicePortSpeed, phyIntfName string, diags *diag.Diagnostics) (int, map[int]apstra.Transformation) {
	// find the device profile "port info" by physical port name (expecting exactly one match from DP)
	dpPort, err := dp.Data.PortByInterfaceName(phyIntfName)
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

// iMapUnallocatedInterfaces takes []apstra.InterfaceMapInterface and
// *apstra.DeviceProfile, returns []apstra.InterfaceMapInterface
// representing all interfaces from the *apstra.DeviceProfile which did
// not appear in the supplied slice.
func iMapUnallocaedInterfaces(allocatedPorts []apstra.InterfaceMapInterface, dp *apstra.DeviceProfile, diags *diag.Diagnostics) []apstra.InterfaceMapInterface {
	// make a map[portId]struct{} so we can quickly determine whether
	// a port ID has been previously allocated.
	allocatedPortCount := len(allocatedPorts)
	allocatedPortIds := make(map[int]struct{}, allocatedPortCount)
	for _, ap := range allocatedPorts {
		allocatedPortIds[ap.Mapping.DPPortId] = struct{}{}
	}

	missingAllocationCount := len(dp.Data.Ports) - len(allocatedPortIds) // device profile ports - used port IDs (ignore breakout ports)

	result := make([]apstra.InterfaceMapInterface, missingAllocationCount)
	var i int
	for _, dpPort := range dp.Data.Ports {
		if _, ok := allocatedPortIds[dpPort.PortId]; ok {
			continue
		}

		transformation := dpPort.DefaultTransform()
		if transformation == nil {
			diags.AddError(errProviderBug, "port has no default transformation")
			return nil
		}

		result[i] = apstra.InterfaceMapInterface{
			Name:  transformation.Interfaces[0].Name,
			Roles: apstra.LogicalDevicePortRoleUnused,
			Mapping: apstra.InterfaceMapMapping{
				DPPortId:      dpPort.PortId,
				DPTransformId: transformation.TransformationId,
				DPInterfaceId: transformation.Interfaces[0].InterfaceId, // blindly use the first interface - UI seems to do this and testing shows there's always at least 1
				LDPanel:       -1,
				LDPort:        -1,
			},
			ActiveState: true, // unclear what this is, UI sets "active"
			Position:    dpPort.PortId,
			Speed:       transformation.Interfaces[0].Speed,
			Setting:     apstra.InterfaceMapInterfaceSetting{Param: transformation.Interfaces[0].Setting},
		}

		i++
	}
	return result
}

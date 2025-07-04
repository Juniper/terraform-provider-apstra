package blueprint

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterGenericSystem struct {
	Id                    types.String         `tfsdk:"id"`
	BlueprintId           types.String         `tfsdk:"blueprint_id"`
	Name                  types.String         `tfsdk:"name"`
	Hostname              types.String         `tfsdk:"hostname"`
	Tags                  types.Set            `tfsdk:"tags"`
	Links                 types.Set            `tfsdk:"links"`
	Asn                   types.Int64          `tfsdk:"asn"`
	LoopbackIpv4          cidrtypes.IPv4Prefix `tfsdk:"loopback_ipv4"`
	LoopbackIpv6          cidrtypes.IPv6Prefix `tfsdk:"loopback_ipv6"`
	PortChannelIdMin      types.Int64          `tfsdk:"port_channel_id_min"`
	PortChannelIdMax      types.Int64          `tfsdk:"port_channel_id_max"`
	External              types.Bool           `tfsdk:"external"`
	DeployMode            types.String         `tfsdk:"deploy_mode"`
	ClearCtsOnDestroy     types.Bool           `tfsdk:"clear_cts_on_destroy"`
	AppPointsByGroupLabel types.Map            `tfsdk:"link_application_point_ids_by_group_label"`
	AppPointsByTag        types.Map            `tfsdk:"link_application_point_ids_by_tag"`
}

func (o DatacenterGenericSystem) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Validators:          []validator.String{stringvalidator.LengthBetween(1, 64)},
		},
		"hostname": resourceSchema.StringAttribute{
			MarkdownDescription: "System hostname.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Validators: []validator.String{
				stringvalidator.RegexMatches(apstraregexp.HostNameConstraint, apstraregexp.HostNameConstraintMsg),
				stringvalidator.LengthBetween(1, 32),
			},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Tag labels to be applied to this Generic System. If a Tag doesn't exist " +
				"in the Blueprint it will be created automatically.",
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"links": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Generic System link details.",
			Required:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: DatacenterGenericSystemLink{}.ResourceAttributes(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				genericSystemLinkSetValidator{},
			},
		},
		"asn": resourceSchema.Int64Attribute{
			MarkdownDescription: "AS number of the Generic System. Note that in some circumstances Apstra may assign " +
				"an ASN to the generic system even when none is supplied via this attribute. The automatically" +
				"assigned value will be overwritten by Terraform during a subsequent apply operation.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(1, math.MaxUint32)},
		},
		"loopback_ipv4": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address of loopback interface in CIDR notation",
			CustomType:          cidrtypes.IPv4PrefixType{},
			Optional:            true,
		},
		"loopback_ipv6": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address of loopback interface in CIDR notation",
			CustomType:          cidrtypes.IPv6PrefixType{},
			Optional:            true,
		},
		"port_channel_id_min": resourceSchema.Int64Attribute{
			MarkdownDescription: "Omit this attribute to allow any available port-channel to be used.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(design.PoIdMin),
			Validators: []validator.Int64{
				int64validator.Between(design.PoIdMin, design.PoIdMax),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("port_channel_id_max")),
			},
		},
		"port_channel_id_max": resourceSchema.Int64Attribute{
			MarkdownDescription: "Omit this attribute to allow any available port-channel to be used.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(design.PoIdMin),
			Validators: []validator.Int64{
				int64validator.Between(design.PoIdMin, design.PoIdMax),
				int64validator.AtLeastSumOf(path.MatchRelative().AtParent().AtName("port_channel_id_min")),
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("port_channel_id_min")),
			},
		},
		"external": resourceSchema.BoolAttribute{
			MarkdownDescription: "Set `true` to create an External Generic System",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
		},
		"deploy_mode": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Set the Apstra Deploy Mode for this Generic System. Default: `%s`",
				enum.DeployModeDeploy),
			Optional:   true,
			Computed:   true,
			Default:    stringdefault.StaticString(enum.DeployModeDeploy.String()),
			Validators: []validator.String{stringvalidator.OneOf(utils.AllNodeDeployModes()...)},
		},
		"clear_cts_on_destroy": resourceSchema.BoolAttribute{
			MarkdownDescription: "When `true`, Connectivity Templates associated with this Generic System will be " +
				"automatically cleared in a variety of circumstances where they would ordinarily block Generic System " +
				"changes, including:\n" +
				"  - Deletion of the Generic System\n" +
				"  - Deletion of a Generic System Link or LAG interface\n" +
				"  - Orphaning a LAG interface by reassigning all of its member links to new roles by changing their " +
				"`group_label` attribute\n",
			Optional: true,
		},
		"link_application_point_ids_by_group_label": resourceSchema.MapAttribute{
			MarkdownDescription: "Map of application point ids keyed by `group_label`. The value at each key is a string " +
				"representing the physical or logical (for LAG interfaces) switch port where the server is connected.",
			ElementType: types.StringType,
			Computed:    true,
		},
		"link_application_point_ids_by_tag": resourceSchema.MapAttribute{
			MarkdownDescription: "Map of application point ids keyed by `tag`. The value at each key is a set of strings " +
				"representing the physical or logical (for LAG interfaces) switch ports where server links tagged with " +
				"the map key are connected. Note that some link tag related config drift may not be reflected in this " +
				"attribute until after an `apply` has corrected the drift.",
			ElementType: types.SetType{ElemType: types.StringType},
			Computed:    true,
		},
	}
}

func (o *DatacenterGenericSystem) CreateRequest(ctx context.Context, diags *diag.Diagnostics) *apstra.CreateLinksWithNewSystemRequest {
	bogusLdTemplateUsedInEveryRequest := apstra.LogicalDevice{
		Id: "tf-ld-template",
		Data: &apstra.LogicalDeviceData{
			DisplayName: "TF LD template",
			Panels: []apstra.LogicalDevicePanel{{
				PanelLayout: apstra.LogicalDevicePanelLayout{RowCount: 1, ColumnCount: 1},
				PortIndexing: apstra.LogicalDevicePortIndexing{
					Order:      apstra.PortIndexingHorizontalFirst,
					StartIndex: 1,
					Schema:     apstra.PortIndexingSchemaAbsolute,
				},
				PortGroups: []apstra.LogicalDevicePortGroup{{Count: 1, Speed: "100M", Roles: apstra.LogicalDevicePortRoles{}}},
			}},
		},
	}

	// extract []DatacenterGenericSystemLink from the plan
	planLinks := o.GetLinks(ctx, diags)
	if diags.HasError() {
		return nil
	}

	var systemType apstra.SystemType
	if o.External.ValueBool() {
		systemType = apstra.SystemTypeExternal
	} else {
		systemType = apstra.SystemTypeServer
	}

	// start building the request object
	request := apstra.CreateLinksWithNewSystemRequest{
		Links: make([]apstra.CreateLinkRequest, len(planLinks)),
		System: apstra.CreateLinksWithNewSystemRequestSystem{
			Hostname:         o.Hostname.ValueString(),
			Label:            o.Name.ValueString(),
			LogicalDevice:    &bogusLdTemplateUsedInEveryRequest,
			Type:             systemType,
			PortChannelIdMin: int(o.PortChannelIdMin.ValueInt64()),
			PortChannelIdMax: int(o.PortChannelIdMax.ValueInt64()),
		},
	}

	// populate the tags in the request object without checking diags for errors
	diags.Append(o.Tags.ElementsAs(ctx, &request.System.Tags, false)...)

	// populate each link in the request object
	for i, link := range planLinks {
		request.Links[i] = *link.request(ctx, diags) // collect all errors
	}

	return &request
}

func (o *DatacenterGenericSystem) GetLinks(ctx context.Context, diags *diag.Diagnostics) []DatacenterGenericSystemLink {
	var result []DatacenterGenericSystemLink
	diags.Append(o.Links.ElementsAs(ctx, &result, false)...)
	return result
}

func (o *DatacenterGenericSystem) ReadTags(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	tags, err := bp.GetNodeTags(ctx, apstra.ObjectId(o.Id.ValueString()))
	if err != nil {
		diags.AddError(fmt.Sprintf("couldn't get tags for node %s", o.Id), err.Error())
		return
	}

	o.Tags = utils.SetValueOrNull(ctx, types.StringType, tags, diags)
}

func (o *DatacenterGenericSystem) ReadLinks(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// Extract the prior state into a map (stateLinks) of links keyed by
	// endpoint digest (switch_id:interface_name).
	// We use a map for quick access to the prior data. We're looking at prior
	// state data so that we know whether to populate the `group_label` (an
	// optional field) in our result. If `group_label` isn't found in the
	// prior state, that means the user omitted it, so we should leave it `null`
	// regardless of the value returned by the API.
	stateLinks := o.GetLinks(ctx, diags)
	if diags.HasError() {
		return
	}
	stateLinksMap := make(map[string]*DatacenterGenericSystemLink, len(stateLinks))
	for i, link := range stateLinks {
		stateLinksMap[link.Digest()] = &stateLinks[i]
	}

	// get the list of links from the API and filter out non-Ethernet links
	apiLinks, err := bp.GetCablingMapLinksBySystem(ctx, apstra.ObjectId(o.Id.ValueString()))
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to fetch generic system %s links", o.Id), err.Error())
		return
	}
	for i := len(apiLinks) - 1; i >= 0; i-- { // loop backwards through the slice
		if apiLinks[i].Type != apstra.LinkTypeEthernet { // target non-Ethernet links for deletion
			apiLinks[i] = apiLinks[len(apiLinks)-1] // overwrite unwanted element with last element
			apiLinks = apiLinks[:len(apiLinks)-1]   // shorten the slice to eliminate the newly dup'ed last item.
		}
	}

	oLinks := make([]attr.Value, len(apiLinks)) // oLinks will populate o.Links
	for i, apiLink := range apiLinks {
		var dcgsl DatacenterGenericSystemLink
		// loadApiData handles every detail except for the transform ID
		dcgsl.loadApiData(ctx, &apiLink, apstra.ObjectId(o.Id.ValueString()), diags)
		if diags.HasError() {
			return
		}

		// check the state's record of this link to see if the user previously
		// specified `group_label`. The `group_label` attribute is not
		// "Computed", so we must return `null` to avoid state churn if the
		// user opted for `null` by not setting it.
		if link, ok := stateLinksMap[dcgsl.Digest()]; ok {
			if link.GroupLabel.IsNull() {
				dcgsl.GroupLabel = types.StringNull()
			}
		}

		// read the switch interface transform ID from the API
		dcgsl.getTransformId(ctx, bp, diags)
		if diags.HasError() {
			return
		}

		var d diag.Diagnostics
		oLinks[i], d = types.ObjectValueFrom(ctx, dcgsl.attrTypes(), &dcgsl)
		diags.Append(d...) // collect all errors
	}
	if diags.HasError() {
		return
	}

	// pack the result slice into o.Links
	o.Links = types.SetValueMust(types.ObjectType{AttrTypes: DatacenterGenericSystemLink{}.attrTypes()}, oLinks)
}

// ReadSystemProperties returns an error rather than appending to a
// []diag.Diagnosics because some callers might need to invoke RemoveResource()
// based on the error type.
func (o *DatacenterGenericSystem) ReadSystemProperties(ctx context.Context, bp *apstra.TwoStageL3ClosClient, overwriteKnownValues bool) error {
	nodeInfo, err := bp.GetSystemNodeInfo(ctx, apstra.ObjectId(o.Id.ValueString()))
	if err != nil {
		return err
	}

	if overwriteKnownValues || o.Hostname.IsUnknown() {
		o.Hostname = types.StringValue(nodeInfo.Hostname)
	}

	if overwriteKnownValues || o.Name.IsUnknown() {
		o.Name = types.StringValue(nodeInfo.Label)
	}

	if overwriteKnownValues || o.External.IsUnknown() {
		o.External = types.BoolValue(nodeInfo.External)
	}

	if overwriteKnownValues || o.DeployMode.IsUnknown() {
		deployMode, err := utils.GetNodeDeployMode(ctx, bp, o.Id.ValueString())
		if err != nil {
			return err
		}
		o.DeployMode = types.StringValue(deployMode)
	}

	// asn isn't computed, so will never be unknown
	if overwriteKnownValues && nodeInfo.Asn != nil {
		o.Asn = types.Int64Value(int64(*nodeInfo.Asn))
	}

	// v4 loopback isn't computed, so will never be unknown
	if overwriteKnownValues && nodeInfo.LoopbackIpv4 != nil {
		o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixValue(nodeInfo.LoopbackIpv4.String())
	}

	// v6 loopback isn't computed, so will never be unknown
	if overwriteKnownValues && nodeInfo.LoopbackIpv6 != nil {
		o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixValue(nodeInfo.LoopbackIpv6.String())
	}

	// Port Channel Min & Max
	if overwriteKnownValues {
		o.PortChannelIdMin = types.Int64Value(int64(nodeInfo.PortChannelIdMin))
		o.PortChannelIdMax = types.Int64Value(int64(nodeInfo.PortChannelIdMax))
	}

	return nil
}

// UpdateHostnameAndName uses the node patch API to set the label and
// hostname fields.
func (o *DatacenterGenericSystem) UpdateHostnameAndName(ctx context.Context, bp *apstra.TwoStageL3ClosClient, state *DatacenterGenericSystem, diags *diag.Diagnostics) {
	if o.Hostname.Equal(state.Hostname) && o.Name.Equal(state.Name) {
		// no planned changes to hostname or name attributes
		return
	}

	// node is an apstra-compatible system struct fragment suitable for patching
	node := struct {
		Hostname string `json:"hostname,omitempty"`
		Label    string `json:"label,omitempty"`
	}{
		Hostname: o.Hostname.ValueString(),
		Label:    o.Name.ValueString(),
	}

	// update the system node
	err := bp.Client().PatchNode(ctx, bp.Id(), apstra.ObjectId(o.Id.ValueString()), &node, nil)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error patching blueprint %q node %s with hostname: %s and label %s",
				bp.Id(), o.Id, o.Hostname, o.Name),
			err.Error())
		return
	}

	if !o.Hostname.IsUnknown() && !o.Name.IsUnknown() {
		// no need to retrieve learned values from Apstra
		return
	}

	// retrieve values from Apstra
	err = bp.Client().GetNode(ctx, bp.Id(), apstra.ObjectId(o.Id.ValueString()), &node)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error fetching blueprint %q node %s", bp.Id(), o.Id),
			err.Error())
		return
	}

	o.Hostname = types.StringValue(node.Hostname)
	o.Name = types.StringValue(node.Label)
}

// UpdateTags uses the tagging API to set the tag set
func (o *DatacenterGenericSystem) UpdateTags(ctx context.Context, bp *apstra.TwoStageL3ClosClient, state *DatacenterGenericSystem, diags *diag.Diagnostics) {
	var planTags, stateTags []string
	diags.Append(o.Tags.ElementsAs(ctx, &planTags, false)...)
	diags.Append(state.Tags.ElementsAs(ctx, &stateTags, false)...)
	if diags.HasError() {
		return
	}

	sort.Strings(planTags)
	sort.Strings(stateTags)

	if utils.SlicesMatch(planTags, stateTags) {
		// no planned changes to tag set
		return
	}

	// update node tags
	err := bp.SetNodeTags(ctx, apstra.ObjectId(o.Id.ValueString()), planTags)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to update tags on %s", o.Id), err.Error())
	}
}

// UpdateLinkSet uses the old state to determine which links in the plan should
// be added, removed and kept. Individual links are then added or removed and
// the "keeper" links are updated with the correct tags, modes, etc...
func (o *DatacenterGenericSystem) UpdateLinkSet(ctx context.Context, state *DatacenterGenericSystem, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// extract links from plan (o) and state objects
	var planLinks, stateLinks []DatacenterGenericSystemLink
	diags.Append(o.Links.ElementsAs(ctx, &planLinks, false)...)
	diags.Append(state.Links.ElementsAs(ctx, &stateLinks, false)...)
	if diags.HasError() {
		return
	}

	// transform plan and state links into a map keyed by link digest (device:port)
	planLinksMap := make(map[string]*DatacenterGenericSystemLink, len(planLinks))
	for i, link := range planLinks {
		planLinksMap[link.Digest()] = &planLinks[i]
	}
	stateLinksMap := make(map[string]*DatacenterGenericSystemLink, len(stateLinks))
	for i, link := range stateLinks {
		stateLinksMap[link.Digest()] = &stateLinks[i]
	}

	// compare plan and state, make lists of links to add / check+update / delete
	var addLinks, updateLinksPlan, updateLinksState, delLinks []*DatacenterGenericSystemLink
	var speedChangeLinkDigests []string
	for digest, planLink := range planLinksMap {
		stateLink, ok := stateLinksMap[digest]
		if !ok {
			// target switch:port not found in state - this is a net new link
			addLinks = append(addLinks, planLink)
			continue
		}

		// link exists in plan and state; check if the speed has changed
		if !planLink.TargetSwitchIfTransformId.Equal(stateLink.TargetSwitchIfTransformId) {
			speedChangeLinkDigests = append(speedChangeLinkDigests, digest)
			continue
		}

		// speed remains the same - the link survives, but may need other attributes updated
		//
		// "updateLinks" is two slices: plan and state, so that we can
		// compare and change only required attributes, if any.
		updateLinksPlan = append(updateLinksPlan, planLink)
		updateLinksState = append(updateLinksState, stateLink)
	}

	// any links in state but not in plan must be deleted
	for digest := range stateLinksMap {
		if _, ok := planLinksMap[digest]; !ok {
			delLinks = append(delLinks, stateLinksMap[digest])
		}
	}

	// delete, then add links where the transform id (speed) must be changed
	for _, digest := range speedChangeLinkDigests {
		stateLink := stateLinksMap[digest]
		planLink := planLinksMap[digest]

		var err error
		var ifId apstra.ObjectId
		var ctSlice []apstra.ObjectId

		// if the config calls for it, clear CTs which block re-creation of links
		if o.ClearCtsOnDestroy.ValueBool() {
			// determine the interface ID so we can clear CTs as needed
			ifId, err = IfIdFromSwIdAndIfName(ctx, bp, apstra.ObjectId(stateLink.TargetSwitchId.ValueString()), stateLink.TargetSwitchIfName.ValueString())
			if err != nil {
				diags.AddError("Failed to determine interface ID", err.Error())
				return
			}

			// discover CTs in use
			ctMap, err := bp.GetConnectivityTemplatesByApplicationPoints(ctx, []apstra.ObjectId{ifId})
			if err != nil {
				diags.AddError("Failed to determine connectivity templates in use", err.Error())
				return
			}

			// prepare the CTs to be re-applied
			for ctId, used := range ctMap[ifId] {
				if used {
					ctSlice = append(ctSlice, ctId)
				}
			}
		}

		o.deleteLinksFromSystem(ctx, []*DatacenterGenericSystemLink{stateLink}, bp, diags)
		if diags.HasError() {
			return
		}

		o.addLinksToSystem(ctx, []*DatacenterGenericSystemLink{planLink}, bp, diags)
		if diags.HasError() {
			return
		}

		// if the config called for deletion, restore CTs which blocked re-creation of links
		if o.ClearCtsOnDestroy.ValueBool() {
			// determine the NEW interface ID so we can reattach the CT
			ifId, err = IfIdFromSwIdAndIfName(ctx, bp, apstra.ObjectId(stateLink.TargetSwitchId.ValueString()), stateLink.TargetSwitchIfName.ValueString())
			if err != nil {
				diags.AddError("Failed to determine interface ID", err.Error())
				return
			}

			// reassign the CTs
			if len(ctSlice) > 0 {
				err = bp.SetApplicationPointConnectivityTemplates(ctx, ifId, ctSlice)
				if err != nil {
					diags.AddError("Failed to reassign connectivity templates", err.Error())
				}
			}
		}
	}

	// delete links no longer in the plan
	o.deleteLinksFromSystem(ctx, delLinks, bp, diags)
	if diags.HasError() {
		return
	}

	// add net new links
	o.addLinksToSystem(ctx, addLinks, bp, diags)
	if diags.HasError() {
		return
	}

	// update links in both plan and state (where the transform ID has not changed)
	o.updateLinkParams(ctx, updateLinksPlan, updateLinksState, bp, diags)
	if diags.HasError() {
		return
	}
}

func (o *DatacenterGenericSystem) addLinksToSystem(ctx context.Context, links []*DatacenterGenericSystemLink, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	if len(links) == 0 {
		return
	}

	linkRequests := make([]apstra.CreateLinkRequest, len(links))
	for i, link := range links {
		linkRequests[i] = *link.request(ctx, diags)
		if diags.HasError() {
			return
		}

		linkRequests[i].SystemEndpoint.SystemId = apstra.ObjectId(o.Id.ValueString())
		err := linkRequests[i].LagMode.FromString(link.LagMode.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("failed parsing lag mode %s", link.LagMode), err.Error())
		}
	}
	if diags.HasError() {
		return
	}

	_, err := bp.AddLinksToSystem(ctx, linkRequests)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed adding links to generic system %s", o.Id), err.Error())
	}
}

func (o *DatacenterGenericSystem) deleteLinksFromSystem(ctx context.Context, links []*DatacenterGenericSystemLink, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	if len(links) == 0 {
		return
	}

	linkIdsToDelete := o.linkIds(ctx, links, bp, diags)
	if diags.HasError() {
		return
	}

	// try to delete the links
	err := bp.DeleteLinksFromSystem(ctx, linkIdsToDelete)
	if err == nil {
		return // success!
	}

	// see if we can handle this error...
	var ace apstra.ClientErr
	if !errors.As(err, &ace) || ace.Type() != apstra.ErrCtAssignedToLink || ace.Detail() == nil {
		// cannot handle error
		diags.AddError("failed while deleting Links from Generic System", err.Error())
		return
	}

	// the error detail has to be the correct type...
	detail, ok := ace.Detail().(apstra.ErrCtAssignedToLinkDetail)
	if !ok {
		diags.AddError(
			constants.ErrProviderBug+fmt.Sprintf(" - ErrCtAssignedToLink has unexpected detail type: %T", detail),
			err.Error(),
		)
		return
	}

	// see if the user could have avoided this problem...
	if !o.ClearCtsOnDestroy.ValueBool() {
		diags.AddError(
			fmt.Sprintf("Cannot delete links with Connectivity Templates assigned: %v", detail.LinkIds),
			"You can set 'clear_cts_on_destroy = true' to override this behavior",
		)
		return
	}

	// prep an error diagnostic in case we can't figure this out
	var pendingDiags diag.Diagnostics
	pendingDiags.AddError(
		fmt.Sprintf("failed deleting links %v from generic system %s", linkIdsToDelete, o.Id),
		err.Error())

	// try to clear the connectivity templates from the problem links
	o.ClearConnectivityTemplatesFromLinks(ctx, ace.Detail().(apstra.ErrCtAssignedToLinkDetail).LinkIds, bp, diags)
	if diags.HasError() {
		diags.Append(pendingDiags...) // throw the pending diagnostic on the pile and give up
		return
	}

	// try deleting the links again
	err = bp.DeleteLinksFromSystem(ctx, linkIdsToDelete)
	if err != nil {
		diags.AddError("failed second attempt to delete links after attempting to handle the link deletion error",
			err.Error())
		diags.Append(pendingDiags...) // throw the pending diagnostic on the pile and give up
		return
	}
}

// updateLinkParams is a method on DatacenterGenericSystem (which has links
// embedded), but it does not operate on those links (all of the links). Rather
// it operates only on the links passed as a function argument because only
// those links need to be updated/validated.
func (o *DatacenterGenericSystem) updateLinkParams(ctx context.Context, planLinks, stateLinks []*DatacenterGenericSystemLink, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// create an empty request with room for each link to be modified
	request := make(apstra.SetLinkLagParamsRequest, len(planLinks))
	for i, link := range planLinks {
		// we don't keep the link ID, but we have each link's target switch and
		// interface name. That's enough to uniquely identify the link from the
		// graph data store
		linkId := o.linkId(ctx, link, bp, diags)
		if diags.HasError() {
			return
		}

		linkParams := link.lagParams(ctx, linkId, stateLinks[i], bp, diags)
		if linkParams != nil {
			request[linkId] = *linkParams
		}
	}

	if len(request) != 0 {
		err := bp.SetLinkLagParams(ctx, &request)
		if err != nil {
			// we may be able to figure this out...
			var pendingDiags diag.Diagnostics
			pendingDiags.AddError("failed updating generic system link parameters", err.Error())

			var ace apstra.ClientErr
			if !errors.As(err, &ace) || ace.Type() != apstra.ErrLagHasAssignedStructrues || ace.Detail() == nil {
				diags.Append(pendingDiags...) // cannot handle error
				return
			}

			detail, ok := ace.Detail().(apstra.ErrLagHasAssignedStructuresDetail)
			if !ok || len(detail.GroupLabels) == 0 {
				diags.Append(pendingDiags...) // cannot handle error
				return
			}

			var lagIds []apstra.ObjectId
			for _, groupLabel := range detail.GroupLabels {
				lagId, err := lagLinkIdFromGsIdAndGroupLabel(ctx, bp, apstra.ObjectId(o.Id.ValueString()), groupLabel)
				if err != nil {
					// return both errors
					diags.Append(pendingDiags...)
					diags.AddError("failed to determine upstream switch LAG port ID", err.Error())
					continue
				}

				lagIds = append(lagIds, lagId)
			}

			if !o.ClearCtsOnDestroy.ValueBool() {
				diags.Append(pendingDiags...) // cannot handle error
				diags.AddWarning(
					fmt.Sprintf("Cannot orphan LAGs with Connectivity Templates assigned: %v", lagIds),
					"You can set 'clear_cts_on_destroy = true' to override this behavior",
				)
				return
			}

			o.ClearConnectivityTemplatesFromLinks(ctx, lagIds, bp, diags)

			// try again...
			err = bp.SetLinkLagParams(ctx, &request)
			if err != nil {
				diags.AddError("failed updating generic system LAG parameters after clearing CTs", err.Error()) // cannot handle error
				return
			}
		}
	}

	// one at a time, check/update each link transform ID
	for i, link := range planLinks {
		link.updateTransformId(ctx, stateLinks[i], bp, diags)
	}
}

// LinkGroupLabels returns a sorted []string representing the group labels
// assigned to all of the links belonging to the DatacenterGenericSystem
func (o DatacenterGenericSystem) LinkGroupLabels(ctx context.Context, diags *diag.Diagnostics) []string {
	var links []DatacenterGenericSystemLink
	diags.Append(o.Links.ElementsAs(ctx, &links, false)...)
	if diags.HasError() {
		return nil
	}

	groupLabelMap := make(map[string]struct{}, len(links))
	for _, link := range links {
		groupLabelMap[link.GroupLabel.ValueString()] = struct{}{}
	}

	return utils.MapKeysSorted(groupLabelMap)
}

// LinkTags returns a sorted []string representing all of the tags
// assigned to all of the links belonging to the DatacenterGenericSystem
func (o DatacenterGenericSystem) LinkTags(ctx context.Context, diags *diag.Diagnostics) []string {
	var links []DatacenterGenericSystemLink
	diags.Append(o.Links.ElementsAs(ctx, &links, false)...)
	if diags.HasError() {
		return nil
	}

	tagMap := make(map[string]struct{})
	for _, link := range links {
		var tags []string
		diags.Append(link.Tags.ElementsAs(ctx, &tags, false)...)
		for _, tag := range tags {
			tagMap[tag] = struct{}{}
		}
	}

	return utils.MapKeysSorted(tagMap)
}

func lagLinkIdFromGsIdAndGroupLabel(ctx context.Context, bp *apstra.TwoStageL3ClosClient, gsId apstra.ObjectId, groupLabel string) (apstra.ObjectId, error) {
	query := new(apstra.PathQuery).SetBlueprintId(bp.Id()).SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(gsId.String())}}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_type", Value: apstra.QEStringVal("port_channel")},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeLink.QEEAttribute(),
			{Key: "group_label", Value: apstra.QEStringVal(groupLabel)},
			{Key: "link_type", Value: apstra.QEStringVal("aggregate_link")},
			{Key: "name", Value: apstra.QEStringVal("n_link")},
		})

	var result struct {
		Items []struct {
			Link struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_link"`
		} `json:"items"`
	}

	if err := query.Do(ctx, &result); err != nil {
		return "", err
	}

	switch len(result.Items) {
	case 0:
		return "", fmt.Errorf("query failed to find LAG link ID for system %q group label %q - %s", gsId, groupLabel, query.String())
	case 1:
		return result.Items[0].Link.Id, nil
	default:
		return "", fmt.Errorf("query found multiple find LAG link IDs for system %q group label %q - %s", gsId, groupLabel, query.String())
	}
}

//func switchLagIdFromGsIdAndGroupLabel(ctx context.Context, bp *apstra.TwoStageL3ClosClient, gsId apstra.ObjectId, groupLabel string) (apstra.ObjectId, error) {
//	query := new(apstra.PathQuery).SetBlueprintId(bp.Id()).SetClient(bp.Client()).
//		Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(gsId.String())}}).
//		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeInterface.QEEAttribute(),
//			{Key: "if_type", Value: apstra.QEStringVal("port_channel")},
//		}).
//		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeLink.QEEAttribute(),
//			{Key: "group_label", Value: apstra.QEStringVal(groupLabel)},
//			{Key: "link_type", Value: apstra.QEStringVal("aggregate_link")},
//		}).
//		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeInterface.QEEAttribute(),
//			{Key: "if_type", Value: apstra.QEStringVal("port_channel")},
//			{Key: "name", Value: apstra.QEStringVal("n_application_point")},
//		})
//
//	var result struct {
//		Items []struct {
//			ApplicationPoint struct {
//				Id apstra.ObjectId `json:"id"`
//			} `json:"n_application_point"`
//		} `json:"items"`
//	}
//
//	if err := query.Do(ctx, &result); err != nil {
//		return "", err
//	}
//
//	switch len(result.Items) {
//	case 0:
//		return "", fmt.Errorf("query failed to find upstream interface ID for system %q group label %q - %s", gsId, groupLabel, query.String())
//	case 1:
//		return result.Items[0].ApplicationPoint.Id, nil
//	default:
//		return "", fmt.Errorf("query found multiple find upstream interface IDs for system %q group label %q - %s", gsId, groupLabel, query.String())
//	}
//}

// linkIds performs the graph queries necessary to return the link IDs which
// connect this Generic System (o) to the systems+interfaces specified by links.
func (o *DatacenterGenericSystem) linkIds(ctx context.Context, links []*DatacenterGenericSystemLink, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) []apstra.ObjectId {
	if len(links) == 0 {
		return nil
	}

	// make a map keyed by switch ID to facilitate graph queries:
	//   switchId123 -> []string{"xe-0/0/1", "xe-0/0/5"}
	//   switchId456 -> []string{"xe-0/0/3", "xe-0/0/6"}
	swIdToIfNames := make(map[string][]string)
	for _, link := range links {
		swIdToIfNames[link.TargetSwitchId.ValueString()] = append(
			swIdToIfNames[link.TargetSwitchId.ValueString()], link.TargetSwitchIfName.ValueString())
	}

	var queryResponse struct {
		Items []struct {
			Link struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_link"`
		} `json:"items"`
	}

	// get link IDs from each switch
	var result []apstra.ObjectId
	for switchId, ifNames := range swIdToIfNames {
		query := new(apstra.PathQuery).
			SetBlueprintType(apstra.BlueprintTypeStaging).
			SetBlueprintId(bp.Id()).
			SetClient(bp.Client()).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "id", Value: apstra.QEStringVal(switchId)},
			}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
			Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeLink.QEEAttribute(),
				{Key: "link_type", Value: apstra.QEStringVal("ethernet")},
				{Key: "name", Value: apstra.QEStringVal("n_link")},
			}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeInterface.QEEAttribute(),
				{Key: "if_name", Value: apstra.QEStringValIsIn(ifNames)},
			}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "id", Value: apstra.QEStringVal(switchId)},
			})

		err := query.Do(ctx, &queryResponse)
		if err != nil {
			diags.AddError(
				fmt.Sprintf("failed querying for link IDs from node %s: %q ports %v",
					o.Id, switchId, ifNames),
				err.Error())
			return nil
		}

		resultLinkIds := make([]apstra.ObjectId, len(queryResponse.Items))
		for i, item := range queryResponse.Items {
			resultLinkIds[i] = item.Link.Id
		}

		if len(ifNames) != len(resultLinkIds) {
			diags.AddError(
				fmt.Sprintf("link ID query for node %s interfaces %v returned only %d link IDs", switchId, ifNames, len(resultLinkIds)),
				fmt.Sprintf("got the following link IDs: %v", resultLinkIds))
			return nil
		}

		result = append(result, resultLinkIds...)
	}

	return result
}

// linkId performs the graph queries necessary to return the link IDs which
// connect this Generic System (o) to the systems+interfaces specified by links.
func (o *DatacenterGenericSystem) linkId(ctx context.Context, link *DatacenterGenericSystemLink, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) apstra.ObjectId {
	linkIds := o.linkIds(ctx, []*DatacenterGenericSystemLink{link}, bp, diags)
	if diags.HasError() {
		return ""
	}
	if len(linkIds) != 1 {
		diags.AddError("provider bug", fmt.Sprintf("expected 1 link ID, got %d", len(linkIds)))
		return ""
	}
	return linkIds[0]
}

func IfIdFromSwIdAndIfName(ctx context.Context, bp *apstra.TwoStageL3ClosClient, swId apstra.ObjectId, ifName string) (apstra.ObjectId, error) {
	query := new(apstra.PathQuery).
		SetClient(bp.Client()).
		SetBlueprintId(bp.Id()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(swId)},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_name", Value: apstra.QEStringVal(ifName)},
			{Key: "if_type", Value: apstra.QEStringVal(apstra.InterfaceTypeEthernet.String())},
			{Key: "name", Value: apstra.QEStringVal("n_interface")},
		})

	var response struct {
		Items []struct {
			Interface struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_interface"`
		} `json:"items"`
	}

	err := query.Do(ctx, &response)
	if err != nil {
		return "", fmt.Errorf("failed querying for interface %q ID from node %q: %w", ifName, swId, err)
	}

	switch len(response.Items) {
	case 0:
		return "", fmt.Errorf("interface %q not found on switch with ID %q", ifName, swId)
	case 1:
	default:
		return "", fmt.Errorf("multiple matches (%d) for interface %q on switch with ID %q", len(response.Items), ifName, swId)
	}

	return response.Items[0].Interface.Id, nil
}

// this validator is here because (a) it's just for one attribute of this
// resource and (2) it uses structs from the blueprint package and would cause
// an import cycle if we put it there.
var _ validator.Set = genericSystemLinkSetValidator{}

type genericSystemLinkSetValidator struct{}

func (o genericSystemLinkSetValidator) Description(_ context.Context) string {
	return "ensures that link sets use a valid combination of values"
}

func (o genericSystemLinkSetValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o genericSystemLinkSetValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsUnknown() {
		return // cannot validate an unknown value
	}

	var links []DatacenterGenericSystemLink
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &links, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validate switch endpoint
	digests := make(map[string]bool, len(links))      // track switch interfaces in use
	groupModes := make(map[string]string, len(links)) // track lag modes per group
	for _, link := range links {
		if !link.TargetSwitchId.IsUnknown() || !link.TargetSwitchIfName.IsUnknown() { // skip impossible-to-calculate digests
			digest := link.Digest()
			if digests[digest] {
				resp.Diagnostics.Append(
					validatordiag.InvalidAttributeCombinationDiagnostic(
						req.Path,
						fmt.Sprintf("multiple links claim interface %s on switch %s",
							link.TargetSwitchIfName, link.TargetSwitchId),
					),
				)
				return
			}
		}

		// validate LAG configuration
		if !link.LagMode.IsUnknown() || !link.GroupLabel.IsUnknown() { // skip impossible-to-evaluate LAGs
			lagMode := link.LagMode.ValueString()
			if groupMode, ok := groupModes[link.GroupLabel.ValueString()]; ok && !link.GroupLabel.IsNull() {
				// we have seen this group label before

				if link.LagMode.IsNull() {
					resp.Diagnostics.Append(
						validatordiag.InvalidAttributeCombinationDiagnostic(
							req.Path,
							fmt.Sprintf("because multiple interfaces share group label %q, lag_mode must be set",
								link.GroupLabel.ValueString())))
					return
				}

				if groupMode != lagMode {
					resp.Diagnostics.Append(
						validatordiag.InvalidAttributeCombinationDiagnostic(
							req.Path,
							fmt.Sprintf("interfaces with group label %q have mismatched 'lag_mode': %q and %q",
								link.GroupLabel.ValueString(), groupMode, lagMode)))
					return
				}
			} else {
				groupModes[link.GroupLabel.ValueString()] = lagMode
			}
		}
	}
}

func (o *DatacenterGenericSystem) SetProperties(ctx context.Context, bp *apstra.TwoStageL3ClosClient, state *DatacenterGenericSystem, diags *diag.Diagnostics) {
	// set ASN if we don't have prior state or the ASN needs to be updated
	if state == nil || !o.Asn.Equal(state.Asn) {
		o.setAsn(ctx, bp, diags)
		if diags.HasError() {
			return
		}
	}

	// set loopback v4 if we don't have prior state or the v4 address needs to be updated
	if state == nil || !o.LoopbackIpv4.Equal(state.LoopbackIpv4) {
		o.setLoopbackIPv4(ctx, bp, diags)
		if diags.HasError() {
			return
		}
	}

	// set loopback v6 if we don't have prior state or the v6 address needs to be updated
	if state == nil || !o.LoopbackIpv6.Equal(state.LoopbackIpv6) {
		o.setLoopbackIPv6(ctx, bp, diags)
		if diags.HasError() {
			return
		}
	}

	// Set Port Channel Min and Max if prior state indicates update is needed
	if state != nil && (!o.PortChannelIdMax.Equal(state.PortChannelIdMax) || !o.PortChannelIdMin.Equal(state.PortChannelIdMin)) {
		o.setPortChannelIdMinMax(ctx, bp, diags)
		if diags.HasError() {
			return
		}
	}

	// set deploy mode if we don't have prior state or the deploy mode needs to be updated
	if state == nil || !o.DeployMode.Equal(state.DeployMode) {
		err := utils.SetNodeDeployMode(ctx, bp, o.Id.ValueString(), o.DeployMode.ValueString())
		if err != nil {
			diags.AddError("failed to set node deploy mode", err.Error())
		}
	}
}

// setAsn sets or clears the generic system ASN attribute depending on o.Asn.IsNull()
func (o *DatacenterGenericSystem) setAsn(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	var asn *uint32

	if !o.Asn.IsNull() {
		a := uint32(o.Asn.ValueInt64())
		asn = &a
	}

	err := bp.SetGenericSystemAsn(ctx, apstra.ObjectId(o.Id.ValueString()), asn)
	if err != nil {
		diags.AddError("failed setting generic system ASN", err.Error())
	}
}

// setLoopbackIPv4 sets or clears the generic system loopback IPv4 attribute depending on o.LoopbackIpv4.IsNull()
func (o *DatacenterGenericSystem) setLoopbackIPv4(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	var ipNet *net.IPNet
	var err error

	if !o.LoopbackIpv4.IsNull() {
		var ip net.IP
		ip, ipNet, err = net.ParseCIDR(o.LoopbackIpv4.ValueString())
		if err != nil {
			diags.AddError("failed parsing generic system IPv4 loopback address", err.Error())
			return
		}

		ipNet.IP = ip // overwrite subnet address in the parsed object with host address
	}

	err = bp.SetGenericSystemLoopbackIpv4(ctx, apstra.ObjectId(o.Id.ValueString()), ipNet, 0)
	if err != nil {
		diags.AddError("failed setting generic system IPv4 loopback address", err.Error())
	}
}

// setLoopbackIPv6 sets or clears the generic system loopback IPv6 attribute depending on o.LoopbackIpv6.IsNull()
func (o *DatacenterGenericSystem) setLoopbackIPv6(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	var ipNet *net.IPNet
	var err error

	if !o.LoopbackIpv6.IsNull() {
		var ip net.IP
		ip, ipNet, err = net.ParseCIDR(o.LoopbackIpv6.ValueString())
		if err != nil {
			diags.AddError("failed parsing generic system IPv6 loopback address", err.Error())
			return
		}

		ipNet.IP = ip // overwrite subnet address in the parsed object with host address
	}

	err = bp.SetGenericSystemLoopbackIpv6(ctx, apstra.ObjectId(o.Id.ValueString()), ipNet, 0)
	if err != nil {
		diags.AddError("failed setting generic system IPv6 loopback address", err.Error())
	}
}

// setPortChannelIdMinMax sets or clears the generic system Po ID min/max depending on the zero value of
// o.PortChannelIdMin and o.PortChannelIdMax (null/unknown/0 will "clear" the value from the web UI).
func (o *DatacenterGenericSystem) setPortChannelIdMinMax(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	err := bp.SetGenericSystemPortChannelMinMax(ctx, apstra.ObjectId(o.Id.ValueString()), int(o.PortChannelIdMin.ValueInt64()),
		int(o.PortChannelIdMax.ValueInt64()))
	if err != nil {
		diags.AddError("failed setting generic system Port Channel Id Min and Max", err.Error())
	}
}

func interfacesFromLinkIds(ctx context.Context, linkIds []apstra.ObjectId, bp *apstra.TwoStageL3ClosClient) ([]apstra.ObjectId, error) {
	linkIdStringVals := make(apstra.QEStringValIsIn, len(linkIds))
	for i, linkId := range linkIds {
		linkIdStringVals[i] = linkId.String()
	}

	query := new(apstra.PathQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeLink.QEEAttribute(),
			{Key: "id", Value: linkIdStringVals},
		}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_interface")},
		})

	var queryResult struct {
		Items []struct {
			Interface struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_interface"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResult)
	if err != nil {
		return nil, fmt.Errorf("failed while querying for interfaces from link IDs - %w", err)
	}

	result := make([]apstra.ObjectId, len(queryResult.Items))
	for i, item := range queryResult.Items {
		result[i] = item.Interface.Id
	}

	return result, nil
}

func (o *DatacenterGenericSystem) ClearConnectivityTemplatesFromLinks(ctx context.Context, linkIds []apstra.ObjectId, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// first learn the interface IDs from the link IDs. This will give us both ends of each link, but that's okay.
	interfaceIds, err := interfacesFromLinkIds(ctx, linkIds, bp)
	if err != nil {
		diags.AddError(
			"failed determining interface ids from link ids while attempting to handle the link deletion error",
			err.Error())
		return
	}

	// now collect all interface-to-CT assignments
	interfaceToCts, err := bp.GetAllInterfacesConnectivityTemplates(ctx)
	if err != nil {
		diags.AddError(
			"failed determining current connectivity template assignments while attempting to handle the link deletion error",
			err.Error())
		return
	}

	// create a new assignments map which will clear the problem CTs
	newAssignments := make(map[apstra.ObjectId]map[apstra.ObjectId]bool)
	for _, interfaceId := range interfaceIds {
		if ctIds, ok := interfaceToCts[interfaceId]; ok {
			assignment := make(map[apstra.ObjectId]bool, len(ctIds))
			for _, ctId := range ctIds {
				assignment[ctId] = false
			}
			newAssignments[interfaceId] = assignment
		}
	}

	// send the new assignments to apstra
	err = bp.SetApplicationPointsConnectivityTemplates(ctx, newAssignments)
	if err != nil {
		diags.AddError(
			"failed clearing connectivity templates from interfaces while attempting to handle the link deletion error",
			err.Error())
		return
	}
}

func (o *DatacenterGenericSystem) ReadSwitchInterfaceApplicationPoints(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	nodeName := "n_system"
	physicalLinkName := "n_physical_link"
	physicalLinkTag := "n_physical_link_tag"
	physicalIntfName := "n_physical_intf"
	localLagIntfName := "n_local_lag_intf"
	sharedLagIntfName := "n_shared_lag_intf"

	// Graph traversal: server >---> physical link
	physicalLinkQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
			{Key: "name", Value: apstra.QEStringVal(nodeName)},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_type", Value: apstra.QEStringVal("ethernet")},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeLink.QEEAttribute(),
			{Key: "link_type", Value: apstra.QEStringVal("ethernet")},
			{Key: "name", Value: apstra.QEStringVal(physicalLinkName)},
		})

	// Graph traversal: physical link >---> switch
	physicalIntfQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal(physicalLinkName)}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_type", Value: apstra.QEStringVal("ethernet")},
			{Key: "name", Value: apstra.QEStringVal(physicalIntfName)},
		}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "system_type", Value: apstra.QEStringVal("switch")},
		})

	// Graph traversal: physical link >---> tag
	linkTagQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal(physicalLinkName)}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeTag.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeTag.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal(physicalLinkTag)},
		})

	// Graph traversal: physical switch interface >---> local lag interface
	localLagQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal(physicalIntfName)}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOf.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_type", Value: apstra.QEStringVal("port_channel")},
			{Key: "po_control_protocol", Value: apstra.QENone(true)},
			{Key: "name", Value: apstra.QEStringVal(localLagIntfName)},
		})

	// Graph traversal: local lag interface >---> shared (esi/mlag) lag interface
	redundantLagQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal(localLagIntfName)}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOf.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_type", Value: apstra.QEStringVal("port_channel")},
			{Key: "po_control_protocol", Value: apstra.QENone(false)},
			{Key: "name", Value: apstra.QEStringVal(sharedLagIntfName)},
		})

	query := new(apstra.MatchQuery).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		Match(physicalLinkQuery).
		Match(new(apstra.MatchQuery).
			Match(physicalIntfQuery).
			Optional(linkTagQuery)).
		Optional(new(apstra.MatchQuery).
			Match(localLagQuery).
			Optional(redundantLagQuery))

	var queryResult struct {
		Items []struct {
			PhysicalLink struct {
				Id         apstra.ObjectId `json:"id"`
				GroupLabel string          `json:"group_label"`
			} `json:"n_physical_link"`
			PhysicalIntf struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_physical_intf"`
			LocalLagIntf *struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_local_lag_intf"`
			SharedLagIntf *struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_shared_lag_intf"`
			PhysicalLinkTag *struct {
				Label string `json:"label"`
			} `json:"n_physical_link_tag"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResult)
	if err != nil {
		diags.AddError(
			"Error while executing query for application point IDs",
			fmt.Sprintf("Graph Query: %q\n\nError: %s", query.String(), err.Error()),
		)
		return
	}

	groupLabelToIntf := make(map[string]apstra.ObjectId)
	tagToIntfMap := make(map[string]map[apstra.ObjectId]struct{})
	for _, item := range queryResult.Items {
		switch {
		case item.SharedLagIntf != nil: // link is part of an esi or mlag aggregate
			groupLabelToIntf[item.PhysicalLink.GroupLabel] = item.SharedLagIntf.Id
			if item.PhysicalLinkTag != nil {
				if _, ok := tagToIntfMap[item.PhysicalLinkTag.Label]; !ok {
					tagToIntfMap[item.PhysicalLinkTag.Label] = make(map[apstra.ObjectId]struct{})
				}
				tagToIntfMap[item.PhysicalLinkTag.Label][item.SharedLagIntf.Id] = struct{}{}
			}
		case item.LocalLagIntf != nil: // link is part of a lag without esi or mlag
			groupLabelToIntf[item.PhysicalLink.GroupLabel] = item.LocalLagIntf.Id
			if item.PhysicalLinkTag != nil {
				if _, ok := tagToIntfMap[item.PhysicalLinkTag.Label]; !ok {
					tagToIntfMap[item.PhysicalLinkTag.Label] = make(map[apstra.ObjectId]struct{})
				}
				tagToIntfMap[item.PhysicalLinkTag.Label][item.LocalLagIntf.Id] = struct{}{}
			}
		default: // link is a standalone physical link
			groupLabelToIntf[item.PhysicalLink.GroupLabel] = item.PhysicalIntf.Id
			if item.PhysicalLinkTag != nil {
				if _, ok := tagToIntfMap[item.PhysicalLinkTag.Label]; !ok {
					tagToIntfMap[item.PhysicalLinkTag.Label] = make(map[apstra.ObjectId]struct{})
				}
				tagToIntfMap[item.PhysicalLinkTag.Label][item.PhysicalIntf.Id] = struct{}{}
			}
		}
	}

	o.AppPointsByGroupLabel = utils.MapValueOrNull(ctx, types.StringType, groupLabelToIntf, diags)

	tagToIntfSlice := make(map[string][]apstra.ObjectId)
	for tag, intfMap := range tagToIntfMap {
		for intf := range intfMap {
			tagToIntfSlice[tag] = append(tagToIntfSlice[tag], intf)
		}
	}
	o.AppPointsByTag = utils.MapValueOrNull(ctx, types.SetType{ElemType: types.StringType}, tagToIntfSlice, diags)
}

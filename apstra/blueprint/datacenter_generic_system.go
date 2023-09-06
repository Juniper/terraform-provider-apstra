package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"sort"
)

type DatacenterGenericSystem struct {
	Id          types.String `tfsdk:"id"`
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Name        types.String `tfsdk:"name"`
	Hostname    types.String `tfsdk:"hostname"`
	Tags        types.Set    `tfsdk:"tags"`
	Links       types.Set    `tfsdk:"links"`
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
				stringvalidator.RegexMatches(regexp.MustCompile("^[A-Za-z0-9.-]+$"),
					"only alphanumeric characters, '.' and '-' allowed."),
				stringvalidator.LengthBetween(1, 32),
			},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Tag labels to be applied to this Generic System. If a Tag doesn't exist " +
				"in the Blueprint it will be created automatically.",
			ElementType: types.StringType,
			Optional:    true,
			Validators:  []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
		"links": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Generic System link details",
			Required:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: DatacenterGenericSystemLink{}.ResourceAttributes(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				genericSystemLinkSetValidator{},
			},
		},
	}
}

func (o *DatacenterGenericSystem) CreateRequest(ctx context.Context, diags *diag.Diagnostics) *apstra.CreateLinksWithNewServerRequest {
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
				PortGroups: []apstra.LogicalDevicePortGroup{{Count: 1, Speed: "100M", Roles: 0}},
			}},
		},
	}

	// extract []DatacenterGenericSystemLink from the plan
	planLinks := o.links(ctx, diags)
	if diags.HasError() {
		return nil
	}

	// start building the request object
	request := apstra.CreateLinksWithNewServerRequest{
		Links: make([]apstra.CreateLinkRequest, len(planLinks)),
		Server: apstra.CreateLinksWithNewServerRequestServer{
			Hostname:      o.Hostname.ValueString(),
			Label:         o.Name.ValueString(),
			LogicalDevice: &bogusLdTemplateUsedInEveryRequest,
		},
	}

	// populate the tags in the request object without checking diags for errors
	diags.Append(o.Tags.ElementsAs(ctx, &request.Server.Tags, false)...)

	// populate each link in the request object
	for i, link := range planLinks {
		request.Links[i] = *link.request(ctx, diags) // collect all errors
	}

	return &request
}

func (o *DatacenterGenericSystem) links(ctx context.Context, diags *diag.Diagnostics) []DatacenterGenericSystemLink {
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
	stateLinks := o.links(ctx, diags)
	if diags.HasError() {
		return
	}
	stateLinksMap := make(map[string]*DatacenterGenericSystemLink, len(stateLinks))
	for i, link := range stateLinks {
		stateLinksMap[link.digest()] = &stateLinks[i]
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
		if link, ok := stateLinksMap[dcgsl.digest()]; ok {
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

// GetLabelAndHostname returns an error rather than appending to a
// []diag.Diagnosics because some callers might need to invoke RemoveResource()
// based on the error type.
func (o *DatacenterGenericSystem) GetLabelAndHostname(ctx context.Context, bp *apstra.TwoStageL3ClosClient) error {
	var node struct {
		Hostname string `json:"hostname"`
		Label    string `json:"label"`
	}
	err := bp.Client().GetNode(ctx, bp.Id(), apstra.ObjectId(o.Id.ValueString()), &node)
	if err != nil {
		return err
	}
	o.Hostname = types.StringValue(node.Hostname)
	o.Name = types.StringValue(node.Label)
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
		planLinksMap[link.digest()] = &planLinks[i]
	}
	stateLinksMap := make(map[string]*DatacenterGenericSystemLink, len(stateLinks))
	for i, link := range stateLinks {
		stateLinksMap[link.digest()] = &stateLinks[i]
	}

	// compare plan and state, make lists of links to add / check+update / delete
	var addLinks, updateLinksPlan, updateLinksState, delLinks []*DatacenterGenericSystemLink
	for digest := range planLinksMap {
		if _, ok := stateLinksMap[digest]; !ok {
			addLinks = append(addLinks, planLinksMap[digest])
		} else {
			// "updateLinks" is two slices: plan and state, so that we can
			// compare and change only required attributes.
			updateLinksPlan = append(updateLinksPlan, planLinksMap[digest])
			updateLinksState = append(updateLinksState, stateLinksMap[digest])
		}
	}
	for digest := range stateLinksMap {
		if _, ok := planLinksMap[digest]; !ok {
			delLinks = append(delLinks, stateLinksMap[digest])
		}
	}

	o.addLinksToSystem(ctx, addLinks, bp, diags)
	if diags.HasError() {
		return
	}

	o.deleteLinksFromSystem(ctx, delLinks, bp, diags)
	if diags.HasError() {
		return
	}

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

	err := bp.DeleteLinksFromSystem(ctx, linkIdsToDelete)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("failed deleting links %v from generic system %s", linkIdsToDelete, o.Id),
			err.Error())
	}
}

// updateLinkParams is a method on DatacenterGenericSystem (which has links
// embedded), but it does not operate on those links (all of the links). Rather
// it operates only on the links passed as a function argument because only
// those links need to be updated/validated.
func (o *DatacenterGenericSystem) updateLinkParams(ctx context.Context, planLinks, stateLinks []*DatacenterGenericSystemLink, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// one at a time, check/update each link
	for i, link := range planLinks {
		// we don't keep the link ID, but we have each link's target switch and
		// interface name. That's enough to uniquely identify the link from the
		// graph data store
		linkId := o.linkId(ctx, link, bp, diags)
		if diags.HasError() {
			return
		}

		link.updateParams(ctx, linkId, stateLinks[i], bp, diags) // collect all errors
	}
}

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

	digests := make(map[string]bool, len(links))      // track switch interfaces in use
	groupModes := make(map[string]string, len(links)) // track lag modes per group
	for _, link := range links {
		digest := link.digest()
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

		lagMode := link.LagMode.ValueString()
		groupLabel := link.GroupLabel.ValueString()
		if m, ok := groupModes[groupLabel]; ok {
			if m != lagMode {
				resp.Diagnostics.Append(
					validatordiag.InvalidAttributeCombinationDiagnostic(
						req.Path,
						fmt.Sprintf("interfaces with group label %q have mismatched 'lag_mode': %q and %q",
							groupLabel, m, lagMode),
					),
				)
				return
			}
		} else {
			groupModes[groupLabel] = lagMode
		}
	}
}

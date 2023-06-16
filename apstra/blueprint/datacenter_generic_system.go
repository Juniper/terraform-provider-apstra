package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
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
)

type DatacenterGenericSystem struct {
	Id          types.String `tfsdk:"id"`
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Label       types.String `tfsdk:"label"`
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
		"label": resourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in thw Apstra web UI.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthBetween(0, 65)},
		},
		"hostname": resourceSchema.StringAttribute{
			MarkdownDescription: "System hostname.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile("^[A-Za-z0-9.]+$"),
					"only underscore, dash and alphanumeric characters allowed."),
				stringvalidator.LengthBetween(0, 33),
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
			Label:         o.Label.ValueString(),
			LogicalDevice: &bogusLdTemplateUsedInEveryRequest,
		},
	}

	// populate the tags in the request object
	diags.Append(o.Tags.ElementsAs(ctx, &request.Server.Tags, false)...)

	// populate each link in the request object
	for i, link := range planLinks {
		request.Links[i] = *link.request(ctx, diags)
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

	var d diag.Diagnostics
	o.Tags, d = types.SetValueFrom(ctx, types.StringType, tags)
	diags.Append(d...)
}

func (o *DatacenterGenericSystem) ReadLinks(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// Extract the prior state into a map of links keyed by endpoint digest.
	// We need quick access to the prior data so that we know whether to
	// populate `group_label` into our result. The `group_label` field is an
	// "Optional" but NOT "Computed" attribute because of the complexity of
	// dealing with "Unknown" values in terraform SetNested attributes.
	//
	// If the user didn't set "group_label", we want to return `null`, rather
	// than the Apstra-assigned value. This map is how we tell the difference.
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

	oLinks := make([]attr.Value, len(apiLinks))
	for i, apiLink := range apiLinks {
		var dcgsl DatacenterGenericSystemLink
		// loadApiData gets everything except for the transform ID
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
		diags.Append(d...)
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
	o.Label = types.StringValue(node.Label)
	return nil
}

// UpdateLabelAndHostname uses the node patch API to set the label and
// hostname fields.
func (o *DatacenterGenericSystem) UpdateLabelAndHostname(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	node := struct {
		Hostname string `json:"hostname,omitempty"`
		Label    string `json:"label,omitempty"`
	}{
		Hostname: o.Hostname.ValueString(),
		Label:    o.Label.ValueString(),
	}

	err := bp.Client().PatchNode(ctx, bp.Id(), apstra.ObjectId(o.Id.ValueString()), &node, nil)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error patching node %s with hostname: %s and label %s", o.Id, o.Hostname, o.Label),
			err.Error())
	}
}

// UpdateTags uses the tagging API to set the tag set
func (o *DatacenterGenericSystem) UpdateTags(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	var planTags []string
	diags.Append(o.Tags.ElementsAs(ctx, &planTags, false)...)

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
	var addLinks, checkLinks, delLinks []*DatacenterGenericSystemLink
	for digest := range planLinksMap {
		if _, ok := stateLinksMap[digest]; !ok {
			addLinks = append(addLinks, planLinksMap[digest])
		} else {
			checkLinks = append(checkLinks, planLinksMap[digest])
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

	o.updateLinkParams(ctx, checkLinks, bp, diags)
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

	ids, err := bp.AddLinksToSystem(ctx, linkRequests)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed adding links to generic system %s", o.Id), err.Error())
	}
	_ = ids
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
func (o *DatacenterGenericSystem) updateLinkParams(ctx context.Context, links []*DatacenterGenericSystemLink, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	if len(links) == 0 {
		return
	}

	// one at a time, check/update each link
	for _, link := range links {
		// we don't keep the link ID, but we have each link's target switch and
		// interface name. That's enough to uniquely identify the link from the
		// graph data store
		linkId := o.linkId(ctx, link, bp, diags)
		if diags.HasError() {
			return
		}

		link.updateParams(ctx, linkId, bp, diags)
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
	return "ensures that links each use a unique combination of system_id + interface_name"
}

func (o genericSystemLinkSetValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o genericSystemLinkSetValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	var links []DatacenterGenericSystemLink
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &links, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	digests := make(map[string]bool, len(links))
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
	}
}

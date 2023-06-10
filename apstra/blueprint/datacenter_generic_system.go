package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
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
	"terraform-provider-apstra/apstra/utils"
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
			MarkdownDescription: "Name displayed in thw Apstra web UI.",
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
				// todo: validate that no combination of switch+port is used more than once
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

// UpdateLabelAndHostname uses the node patch API to update the label and
// hostname fields.
func (o *DatacenterGenericSystem) UpdateLabelAndHostname(ctx context.Context, state *DatacenterGenericSystem, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	if o.Hostname.Equal(state.Hostname) && o.Label.Equal(state.Label) {
		return
	}

	node := struct {
		Hostname string `json:"hostname"`
		Label    string `json:"label"`
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

func (o *DatacenterGenericSystem) UpdateTags(ctx context.Context, state *DatacenterGenericSystem, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	var planTags, stateTags []string
	diags.Append(o.Tags.ElementsAs(ctx, &planTags, false)...)
	diags.Append(state.Tags.ElementsAs(ctx, &stateTags, false)...)

	sort.Strings(planTags)
	sort.Strings(stateTags)

	if utils.SlicesMatch(planTags, stateTags) {
		return
	}

	err := bp.SetNodeTags(ctx, apstra.ObjectId(o.Id.ValueString()), planTags)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to update tags on %s", o.Id), err.Error())
	}
}

func (o *DatacenterGenericSystem) UpdateLinks(ctx context.Context, state *DatacenterGenericSystem, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	var planLinks, stateLinks []DatacenterGenericSystemLink
	diags.Append(o.Links.ElementsAs(ctx, &planLinks, false)...)
	diags.Append(state.Links.ElementsAs(ctx, &stateLinks, false)...)

	planLinksMap := make(map[string]*DatacenterGenericSystemLink, len(planLinks))
	for i, link := range planLinks {
		planLinksMap[link.digest()] = &planLinks[i]
	}

	stateLinksMap := make(map[string]*DatacenterGenericSystemLink, len(stateLinks))
	for i, link := range stateLinks {
		stateLinksMap[link.digest()] = &stateLinks[i]
	}

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

	o.addLinks(ctx, addLinks, bp, diags)
	if diags.HasError() {
		return
	}

}

func (o *DatacenterGenericSystem) addLinks(ctx context.Context, links []*DatacenterGenericSystemLink, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
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

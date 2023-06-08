package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	Id              types.String `tfsdk:"id"`
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	Label           types.String `tfsdk:"label"`
	Hostname        types.String `tfsdk:"hostname"`
	LogicalDeviceId types.String `tfsdk:"logical_device_id"`
	Tags            types.Set    `tfsdk:"tags"`
	Links           types.List   `tfsdk:"links"`
}

//func (o DatacenterGenericSystem) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
//	return map[string]dataSourceSchema.Attribute{
//	}
//}

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
		"logical_device_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Global Catalog ID of the logical device used to model this Generic System.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Tag labels to be applied to this Generic System. If a Tag doesn't exist " +
				"in the Blueprint it will be created automatically.",
			ElementType: types.StringType,
			Optional:    true,
			Validators:  []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
		"links": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "Generic System link details",
			Required:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: DatacenterGenericSystemLink{}.ResourceAttributes(),
			},
			Validators: []validator.List{listvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *DatacenterGenericSystem) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.CreateLinksWithNewServerRequest {
	links := o.links(ctx, diags)
	if diags.HasError() {
		return nil
	}

	result := apstra.CreateLinksWithNewServerRequest{
		Links: make([]apstra.CreateLinksWithNewServerRequestLink, len(links)),
		Server: apstra.CreateLinksWithNewServerRequestServer{
			Hostname:        o.Hostname.ValueString(),
			Label:           o.Label.ValueString(),
			LogicalDeviceId: apstra.ObjectId(o.LogicalDeviceId.ValueString()),
		},
	}

	diags.Append(o.Tags.ElementsAs(ctx, &result.Server.Tags, false)...)

	for i, link := range links {
		result.Links[i] = *link.Request(ctx, diags)
	}

	return &result
}

func (o *DatacenterGenericSystem) links(ctx context.Context, diags *diag.Diagnostics) []DatacenterGenericSystemLink {
	var result []DatacenterGenericSystemLink
	diags.Append(o.Links.ElementsAs(ctx, &result, false)...)
	return result
}

//// PopulateLinkInfo should be called only in Create() to insert each link
//// ID and group label assigned by apstra into the appropriate structure.
//func (o *DatacenterGenericSystem) PopulateLinkInfo(ctx context.Context, newLinkIds []apstra.ObjectId, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	o.populateLinkIds(ctx, newLinkIds, client, diags)
//	if diags.HasError() {
//		return
//	}
//
//	o.populateLinkGroupLabels(ctx, client, diags)
//}

// PopulateLinkInfo should be called only in Create(). It looks up the details of newLinkIds (the
// link IDs returned at generic system creation time)
// ID into the appropriate structure.
// func (o *DatacenterGenericSystem) PopulateLinkInfo(ctx context.Context, newLinkIds []apstra.ObjectId, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
func (o *DatacenterGenericSystem) PopulateLinkInfo(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	//// ensure the supplied link IDs have no repeating values
	//newLinkIdSet := make(map[string]bool, len(newLinkIds))
	//for _, id := range newLinkIds {
	//	newLinkIdSet[id.String()] = true
	//}
	//if len(newLinkIdSet) != len(newLinkIds) {
	//	diags.AddError("supplied slice of link IDs has repeat elements?", fmt.Sprintf("list: %v", newLinkIds))
	//	return
	//}

	// get the links actually attached to the server from the cabling
	// api and compare with the links reported at create time
	linkInfos, err := client.GetCablingMapLinksBySystem(ctx, apstra.ObjectId(o.Id.ValueString()))
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to get cabling map info for system %s", o.Id), err.Error())
		return
	}
	//linkInfoIdSet := make(map[string]bool, len(linkInfos))
	//for _, linkInfo := range linkInfos {
	//	linkInfoIdSet[linkInfo.Id.String()] = true
	//}
	//if !utils.MapsMatch(newLinkIdSet, linkInfoIdSet) {
	//	var linkInfoIds []string
	//	for k := range linkInfoIdSet {
	//		linkInfoIds = append(linkInfoIds, k)
	//	}
	//	diags.AddError(
	//		"new generic system link confusion",
	//		fmt.Sprintf("on creation, system had links %v now %v", newLinkIds, linkInfoIds))
	//}

	// organize the link info from the cabling api into a map for quick retrieval
	endpointDigestToLinkInfo := make(map[string]*apstra.CablingMapLink, len(linkInfos))
	for i, linkInfo := range linkInfos {
		if linkInfo.Type != apstra.LinkTypeEthernet {
			continue // not interested in aggregates or logical links here
		}

		ep := linkInfo.OppositeEndpointBySystemId(apstra.ObjectId(o.Id.ValueString()))
		if ep == nil {
			diags.AddError(
				"link lands on wrong system?",
				fmt.Sprintf("cabling map link %q doesn't have an endpoint on system %s", linkInfo.Id, o.Id))
			return
		}

		digest := ep.Digest()
		if digest == nil {
			diags.AddError(
				"incomplete link endpoint?",
				fmt.Sprintf("cabling map link %q endpoint on system %s missing required fields", linkInfo, o.Id))
			return
		}

		endpointDigestToLinkInfo[*digest] = &linkInfos[i]
	}

	// planLinks is the user-defined link info from the plan
	planLinks := o.links(ctx, diags)
	if diags.HasError() {
		return
	}

	// iterate over the planned links. Use each link's digest to identify the
	// corresponding link info from the cabling api. Store that link info's ID
	// in each planned link object.
	for i, planLink := range planLinks {
		// pull the matching API response from the map
		digest := planLink.digest()
		linkInfo, ok := endpointDigestToLinkInfo[digest]
		if !ok {
			diags.AddError(
				"planned link not found",
				fmt.Sprintf("link to %q not found among links to system %s", digest, o.Id),
			)
			return
		}

		// Finally! this is why we're here.
		planLinks[i].Id = types.StringValue(linkInfo.Id.String())
		planLinks[i].GroupLabel = types.StringValue(linkInfo.GroupLabel)
	}

	// re-pack planLinks into o.Links now that we've updated their ID fields.
	var d diag.Diagnostics
	o.Links, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DatacenterGenericSystemLink{}.attrTypes()}, planLinks)
	diags.Append(d...)
}

//func (o *DatacenterGenericSystem) PopulateGenericSystemInfo(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	gsQuery := new(apstra.MatchQuery).
//		SetBlueprintId(client.Id()).
//		SetBlueprintType(apstra.BlueprintTypeStaging).
//		SetClient(client.Client())
//
//	for _, link := range o.links(ctx, diags) {
//		if diags.HasError() {
//			return
//		}
//		gsQuery.Match(link.GenericSystemQuery())
//	}
//	if diags.HasError() {
//		return
//	}
//
//	gsResponse := DatacenterGenericSystemLink{}.GenericSystemQueryResponse()
//	err := gsQuery.Do(ctx, &gsResponse)
//	if err != nil {
//		diags.AddError("failed executing graph query to find Generic System ID",
//			err.Error())
//		return
//	}
//
//	if len(gsResponse.Items) != 1 {
//		diags.AddError(
//			fmt.Sprintf("expected exactly 1 graph query response, got %d", len(gsResponse.Items)),
//			fmt.Sprintf("query: %s", gsQuery.String()),
//		)
//		return
//	}
//
//	o.Hostname = types.StringValue(gsResponse.Items[0].Generic.Hostname)
//	o.Id = types.StringValue(gsResponse.Items[0].Generic.Id)
//	o.Label = types.StringValue(gsResponse.Items[0].Generic.Label)
//}

//func (o *DatacenterGenericSystem) populateLinkGroupLabels(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	links := o.links(ctx, diags)
//	if diags.HasError() {
//		return
//	}
//
//	// extract our link IDs as []string
//	linkIds := make([]string, len(links))
//	for i, link := range links {
//		if !utils.Known(link.Id) {
//			diags.AddError("missing link id when querying for link group labels",
//				fmt.Sprintf("link %d ID is %s", i, link.Id))
//			return
//		}
//
//		linkIds[i] = link.Id.ValueString()
//	}
//
//	// query for all link nodes in one shot
//	linkQuery := new(apstra.PathQuery).
//		SetBlueprintId(client.Id()).
//		SetBlueprintType(apstra.BlueprintTypeStaging).
//		SetClient(client.Client()).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeLink.QEEAttribute(),
//			{Key: "id", Value: apstra.QEStringValIsIn(linkIds)},
//			{Key: "name", Value: apstra.QEStringVal("n_link")},
//		})
//
//	var linkQueryResponse struct {
//		Items []struct {
//			Link struct {
//				Id         string `json:"id"`
//				GroupLabel string `json:"group_label"`
//			} `json:"n_link"`
//		} `json:"items"`
//	}
//
//	err := linkQuery.Do(ctx, &linkQueryResponse)
//	if err != nil {
//		diags.AddError("error querying for link group labels", err.Error())
//		return
//	}
//
//	// make a map of link ID -> group label
//	idToGroupLabel := make(map[string]string, len(linkIds))
//	for _, item := range linkQueryResponse.Items {
//		idToGroupLabel[item.Link.Id] = item.Link.GroupLabel
//	}
//
//	// ensure that every link ID is represented in the map
//	for _, id := range linkIds {
//		if _, ok := idToGroupLabel[id]; !ok {
//			diags.AddError(
//				fmt.Sprintf("error querying for link group labels : link %s not found", id),
//				fmt.Sprintf("query: %s", linkQuery.String()),
//			)
//			return
//		}
//	}
//
//	// populate the label into the links slice
//	for i, link := range links {
//		link.GroupLabel = types.StringValue(idToGroupLabel[link.Id.ValueString()])
//		links[i] = link
//	}
//
//	// pack the links slice back into o
//	var d diag.Diagnostics
//	o.Links, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DatacenterGenericSystemLink{}.attrTypes()}, links)
//	diags.Append(d...)
//}

func (o *DatacenterGenericSystem) ReadLogicalDevice(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	ldQuery := new(apstra.PathQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLogicalDevice.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeLogicalDevice.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_logical_device")},
		})

	var ldQueryResponse struct {
		Items []struct {
			LogicalDevice struct {
				Id string `json:"id"`
			} `json:"n_logical_device"`
		} `json:"items"`
	}

	err := ldQuery.Do(ctx, &ldQueryResponse)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed querying for logical device on node %s", o.Id), err.Error())
	}
	if len(ldQueryResponse.Items) != 1 {
		diags.AddError(
			fmt.Sprintf("expected 1 logical device from query, got %d", len(ldQueryResponse.Items)),
			fmt.Sprintf("query string: %s", ldQuery.String()),
		)
	}
	if diags.HasError() {
		return
	}

	o.Id = types.StringValue(ldQueryResponse.Items[0].LogicalDevice.Id)
}

func (o *DatacenterGenericSystem) ReadTags(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	tagQuery := new(apstra.PathQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
		}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeTag.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeTag.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_tag")},
		})

	var tagQueryResponse struct {
		Items []struct {
			Tag struct {
				Label string `json:"label"`
			} `json:"n_tag"`
		} `json:"items"`
	}

	err := tagQuery.Do(ctx, &tagQueryResponse)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed querying for tags on node %s", o.Id), err.Error())
		return
	}

	tags := make([]attr.Value, len(tagQueryResponse.Items))
	for i, item := range tagQueryResponse.Items {
		tags[i] = types.StringValue(item.Tag.Label)
	}
	o.Tags = types.SetValueMust(types.StringType, tags)
}

//func (o *DatacenterGenericSystem) ReadLinks(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	// get the list of links from the API
//	apiLinks, err := bp.GetCablingMapLinksBySystem(ctx, apstra.ObjectId(o.Id.ValueString()))
//	if err != nil {
//		diags.AddError(fmt.Sprintf("failed to fetch links from %s", o.Id), err.Error())
//		return
//	}
//
//	// eliminate non-Ethernet links (LAGs)
//	for i := len(apiLinks) - 1; i >= 0; i-- { // loop backwards through the slice
//		if apiLinks[i].Type != apstra.LinkTypeEthernet { // keep only Ethernet links
//			apiLinks[i] = apiLinks[len(apiLinks)-1] // overwrite unwanted element with last element
//			apiLinks = apiLinks[:len(apiLinks)-1]   // shorten the slice to eliminate the newly dup'ed last item.
//		}
//	}
//
//	//// get the list of links from o
//	//tfLinks := o.links(ctx, diags)
//	//if diags.HasError() {
//	//	return
//	//}
//
//	for i, apiLink := range apiLinks {
//		tfLinkIdx := linkIndex(tfLinks, apiLink.Id)
//		if tfLinkIdx < 0 {
//			var tfLink DatacenterGenericSystemLink
//			tfLink.LoadApiData(ctx, apiLink)
//		}
//	}
//
//	//linkQuery := new(apstra.MatchQuery).
//	//	SetBlueprintId(bp.Id()).
//	//	SetClient(bp.Client()).
//	//	SetBlueprintType(apstra.BlueprintTypeStaging).
//	//	Match(
//	//		new(apstra.PathQuery).
//	//			Node([]apstra.QEEAttribute{
//	//				apstra.NodeTypeSystem.QEEAttribute(),
//	//				{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
//	//			}).
//	//			Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
//	//			Node([]apstra.QEEAttribute{
//	//				apstra.NodeTypeInterface.QEEAttribute(),
//	//				{Key: "if_type", Value: apstra.QEStringVal("ethernet")},
//	//				{Key: "name", Value: apstra.QEStringVal("n_interface")},
//	//			})).
//	//	Match(new(apstra.PathQuery).
//	//		Node([]apstra.QEEAttribute{
//	//			apstra.NodeTypeInterface.QEEAttribute(),
//	//			{Key: "if_type", Value: apstra.QEStringVal("ethernet")},
//	//			{Key: "name", Value: apstra.QEStringVal("n_interface")},
//	//		}).
//	//		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
//	//		Node([]apstra.QEEAttribute{
//	//			apstra.NodeTypeLink.QEEAttribute(),
//	//			{Key: "link_type", Value: apstra.QEStringVal("ethernet")},
//	//			{Key: "name", Value: apstra.QEStringVal("n_link")},
//	//		})).
//	//	Match(new(apstra.PathQuery).
//	//		Node([]apstra.QEEAttribute{
//	//			apstra.NodeTypeLink.QEEAttribute(),
//	//			{Key: "link_type", Value: apstra.QEStringVal("ethernet")},
//	//			{Key: "name", Value: apstra.QEStringVal("n_link")},
//	//		}).
//	//		In([]apstra.QEEAttribute{apstra.RelationshipTypeTag.QEEAttribute()}).
//	//		Node([]apstra.QEEAttribute{
//	//			apstra.NodeTypeTag.QEEAttribute(),
//	//			{Key: "name", Value: apstra.QEStringVal("n_tag")},
//	//		}))
//	//
//	//var linkQueryResponse struct {
//	//	Items []struct {
//	//		Link struct {
//	//			Id         string `json:"id"`
//	//			GroupLabel string `json:"group_label"`
//	//		} `json:"n_link"`
//	//		Tag struct {
//	//			Label string `json:"label"`
//	//		} `json:"n_tag"`
//	//	} `json:"items"`
//	//}
//	//
//	//err := linkQuery.Do(ctx, &linkQueryResponse)
//	//if err != nil {
//	//	diags.AddError(fmt.Sprintf("failed querying for links on node %s", o.Id), err.Error())
//	//	return
//	//}
//	//
//	//linkIdToTagLabels := make(map[string][]string)
//	//for _, item := range linkQueryResponse.Items {
//	//	linkIdToTagLabels[item.Link.Id] = append(linkIdToTagLabels[item.Link.Id], item.Tag.Label)
//	//}
//
//}

func linkIndex(links []DatacenterGenericSystemLink, linkId apstra.ObjectId) int {
	for i, link := range links {
		if link.Id.ValueString() == linkId.String() {
			return i
		}
	}
	return -1
}

type key struct {
	ThingOne types.String `tfsdk:"thing_one"`
	ThingTwo types.Int64  `tfsdk:"thing_two"`
}

func (o key) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"thing_one": types.StringType,
		"thing_two": types.Int64Type,
	}
}

func keySliceToList(ctx context.Context, keysSliceIn []key, diags *diag.Diagnostics) types.List {
	keys, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: key{}.attrTypes()}, keysSliceIn)
	diags.Append(d...)
	return keys
}

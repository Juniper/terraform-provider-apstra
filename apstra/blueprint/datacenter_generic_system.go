package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"terraform-provider-apstra/apstra/utils"
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
			MarkdownDescription: "Tag labels to be applied to this generic system. If a tag doesn't exist " +
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

	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		return nil
	}

	result := apstra.CreateLinksWithNewServerRequest{
		Links: make([]apstra.SwitchLink, len(links)),
		Server: apstra.System{
			Hostname:        o.Hostname.ValueString(),
			Label:           o.Label.ValueString(),
			LogicalDeviceId: apstra.ObjectId(o.LogicalDeviceId.ValueString()),
			Tags:            tags,
		},
	}

	for i, link := range links {
		result.Links[i] = *link.Request(ctx, diags)
	}

	return &result
}

func (o *DatacenterGenericSystem) LoadApiData(_ context.Context, sz *apstra.SecurityZoneData, _ *diag.Diagnostics) {
	//o.Name = types.StringValue(sz.VrfName)
	//o.VlanId = types.Int64Value(int64(*sz.VlanId))
	//
	//if sz.RoutingPolicyId != "" {
	//	o.RoutingPolicyId = types.StringValue(sz.RoutingPolicyId.String())
	//} else {
	//	o.RoutingPolicyId = types.StringNull()
	//}
	//
	//if sz.VniId != nil {
	//	o.Vni = types.Int64Value(int64(*sz.VniId))
	//} else {
	//	o.Vni = types.Int64Null()
	//}
}

func (o *DatacenterGenericSystem) links(ctx context.Context, diags *diag.Diagnostics) []DatacenterGenericSystemLink {
	var result []DatacenterGenericSystemLink
	diags.Append(o.Links.ElementsAs(ctx, &result, false)...)
	return result
}

func (o *DatacenterGenericSystem) PopulateLinkInfo(ctx context.Context, newLinkIds []apstra.ObjectId, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	o.populateLinkIds(ctx, newLinkIds, client, diags)
	if diags.HasError() {
		return
	}

	o.populateLinkGroupLabels(ctx, client, diags)
}

func (o *DatacenterGenericSystem) populateLinkIds(ctx context.Context, newLinkIds []apstra.ObjectId, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	newLinkIdSet := make(map[string]bool, len(newLinkIds))
	for _, id := range newLinkIds {
		newLinkIdSet[id.String()] = true
	}

	if len(newLinkIdSet) != len(newLinkIds) {
		diags.AddError("returned link IDs has repeat elements?", fmt.Sprintf("list: %v", newLinkIds))
		return
	}

	links := o.links(ctx, diags)
	if diags.HasError() {
		return
	}

	usedLinkIdMap := make(map[string]bool, len(newLinkIds))
	for i, link := range links {
		endpoint := link.endpoint(ctx, diags)
		if diags.HasError() {
			return
		}

		linkId := endpoint.linkId(ctx, client, diags)
		if diags.HasError() {
			return
		}

		if !newLinkIdSet[linkId] {
			diags.AddError("requested switch port not connected to any new generic system links",
				fmt.Sprintf("switch %s port %q links: %v", endpoint.SystemId, endpoint.IfName, newLinkIds))
			return
		}

		if usedLinkIdMap[linkId] {
			diags.AddError("multiple switch ports connected to the same link?",
				fmt.Sprintf("link: %s switch %s port %q links: %v", linkId, endpoint.SystemId, endpoint.IfName, newLinkIds))
			return
		}
		usedLinkIdMap[linkId] = true

		links[i].Id = types.StringValue(linkId)
	}

	var d diag.Diagnostics
	o.Links, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DatacenterGenericSystemLink{}.attrTypes()}, links)
	diags.Append(d...)
}

func (o *DatacenterGenericSystem) PopulateGenericSystemInfo(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	gsQuery := new(apstra.MatchQuery).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetClient(client.Client())

	for _, link := range o.links(ctx, diags) {
		if diags.HasError() {
			return
		}
		gsQuery.Match(link.GenericSystemQuery())
	}
	if diags.HasError() {
		return
	}

	gsResponse := DatacenterGenericSystemLink{}.GenericSystemQueryResponse()
	err := gsQuery.Do(ctx, &gsResponse)
	if err != nil {
		diags.AddError("failed executing graph query to find Generic System ID",
			err.Error())
		return
	}

	if len(gsResponse.Items) != 1 {
		diags.AddError(
			fmt.Sprintf("expected exactly 1 graph query response, got %d", len(gsResponse.Items)),
			fmt.Sprintf("query: %s", gsQuery.String()),
		)
		return
	}

	o.Hostname = types.StringValue(gsResponse.Items[0].Generic.Hostname)
	o.Id = types.StringValue(gsResponse.Items[0].Generic.Id)
	o.Label = types.StringValue(gsResponse.Items[0].Generic.Label)
}

func (o *DatacenterGenericSystem) populateLinkGroupLabels(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	links := o.links(ctx, diags)
	if diags.HasError() {
		return
	}

	// extract our link IDs as []string
	linkIds := make([]string, len(links))
	for i, link := range links {
		if !utils.Known(link.Id) {
			diags.AddError("missing link id when querying for link group labels",
				fmt.Sprintf("link %d ID is %s", i, link.Id))
			return
		}

		linkIds[i] = link.Id.ValueString()
	}

	// query for all link nodes in one shot
	linkQuery := new(apstra.PathQuery).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetClient(client.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeLink.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringValIsIn(linkIds)},
			{Key: "name", Value: apstra.QEStringVal("n_link")},
		})

	var linkQueryResponse struct {
		Items []struct {
			Link struct {
				Id         string `json:"id"`
				GroupLabel string `json:"group_label"`
			} `json:"n_link"`
		} `json:"items"`
	}

	err := linkQuery.Do(ctx, &linkQueryResponse)
	if err != nil {
		diags.AddError("error querying for link group labels", err.Error())
		return
	}

	// make a map of link ID -> group label
	idToGroupLabel := make(map[string]string, len(linkIds))
	for _, item := range linkQueryResponse.Items {
		idToGroupLabel[item.Link.Id] = item.Link.GroupLabel
	}

	// ensure that every link ID is represented in the map
	for _, id := range linkIds {
		if _, ok := idToGroupLabel[id]; !ok {
			diags.AddError(
				fmt.Sprintf("error querying for link group labels : link %s not found", id),
				fmt.Sprintf("query: %s", linkQuery.String()),
			)
			return
		}
	}

	// populate the label into the links slice
	for i, link := range links {
		link.GroupLabel = types.StringValue(idToGroupLabel[link.Id.ValueString()])
		links[i] = link
	}

	// pack the links slice back into o
	var d diag.Diagnostics
	o.Links, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DatacenterGenericSystemLink{}.attrTypes()}, links)
	diags.Append(d...)
}

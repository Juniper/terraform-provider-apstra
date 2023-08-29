package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type NodesTypeSystem struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Filter      types.Object `tfsdk:"filter"`
	Ids         types.Set    `tfsdk:"ids"`
	QueryString types.String `tfsdk:"query_string"`
}

func (o NodesTypeSystem) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint to search.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"filter": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Filter used to select only desired node IDs. All specified attributes must match.",
			Optional:            true,
			Attributes:          NodeTypeSystemAttributes{}.DataSourceAttributesAsFilter(),
		},
		"query_string": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Graph DB query string based on the supplied filter; possibly useful for troubleshooting.",
			Computed:            true,
		},
		"ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs of matching `system` Graph DB nodes.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o *NodesTypeSystem) ReadFromApi(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	var queryResponse struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	query := o.query(ctx, diags).
		SetClient(client).
		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetBlueprintType(apstra.BlueprintTypeStaging)
	if diags.HasError() { // catch errors fro
		return
	}

	err := query.Do(ctx, &queryResponse)
	if err != nil {
		diags.AddError("Error executing Blueprint query", err.Error())
		return
	}

	ids := make([]attr.Value, len(queryResponse.Items))
	for i := range queryResponse.Items {
		ids[i] = types.StringValue(queryResponse.Items[i].System.Id)
	}

	o.Ids = types.SetValueMust(types.StringType, ids)
	o.QueryString = types.StringValue(query.String())
}

func (o *NodesTypeSystem) query(ctx context.Context, diags *diag.Diagnostics) *apstra.MatchQuery {
	var filter NodeTypeSystemAttributes
	if utils.Known(o.Filter) {
		diags.Append(o.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil
		}
	}

	var tagIds []string
	if utils.Known(filter.TagIds) {
		diags.Append(filter.TagIds.ElementsAs(ctx, &tagIds, false)...)
		if diags.HasError() {
			return nil
		}
	}

	systemNodeBaseAttributes := []apstra.QEEAttribute{
		{Key: "type", Value: apstra.QEStringVal("system")},
		{Key: "name", Value: apstra.QEStringVal("n_system")},
	}

	// []QEEAttribute to match the system hostname, label, role, etc... as specified by `filter`
	systemNodeAttributes := append(systemNodeBaseAttributes, filter.QEEAttributes()...)

	// []QEEAttribute to match the relationship between system and tag nodes
	relationshipAttributes := []apstra.QEEAttribute{{Key: "type", Value: apstra.QEStringVal("tag")}}

	// []QEEAttribute to match the tag node (further qualified in the loop below)
	tagNodeBaseAttributes := []apstra.QEEAttribute{{Key: "type", Value: apstra.QEStringVal("tag")}}

	// This is the query we actually want to execute. It's a `match()`
	// query-of-queries which selects the system node using
	// `systemNodeAttributes` and also selects paths from the system node to
	// each specified tag.
	query := new(apstra.MatchQuery)

	// first query: the system node with filter.
	query.Match(new(apstra.PathQuery).Node(systemNodeAttributes))

	// now add each tag-path query.
	for i := range tagIds {
		tagLabelAttribute := apstra.QEEAttribute{
			Key:   "label",
			Value: apstra.QEStringVal(tagIds[i]),
		}
		tagQuery := new(apstra.PathQuery).
			Node(systemNodeBaseAttributes).
			In(relationshipAttributes).
			Node(append(tagNodeBaseAttributes, tagLabelAttribute))
		query.Match(tagQuery)
	}

	return query
}

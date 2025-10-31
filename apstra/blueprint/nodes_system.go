package blueprint

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type NodesTypeSystem struct {
	BlueprintId  types.String `tfsdk:"blueprint_id"`
	Filter       types.Object `tfsdk:"filter"`
	Filters      types.List   `tfsdk:"filters"`
	Ids          types.Set    `tfsdk:"ids"`
	QueryStrings types.List   `tfsdk:"query_strings"`
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
			DeprecationMessage: "The `filter` attribute is deprecated and will be removed in a future " +
				"release. Please migrate your configuration to use `filters` instead.",
			Validators: []validator.Object{
				apstravalidator.AtMostNOf(1,
					path.MatchRelative(),
					path.MatchRoot("filters"),
				),
				apstravalidator.AtLeastNAttributes(
					1,
					"hostname", "id", "label", "role", "system_id", "system_type", "tag_ids",
				),
			},
		},
		"filters": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "List of filters used to select only desired node IDs. For a node " +
				"to match a filter, all specified attributes must match (each attribute within a " +
				"filter is AND-ed together). The returned node IDs represent the nodes matched by " +
				"all of the filters together (filters are OR-ed together).",
			Optional:   true,
			Validators: []validator.List{listvalidator.SizeAtLeast(1)},
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: NodeTypeSystemAttributes{}.DataSourceAttributesAsFilter(),
				Validators: []validator.Object{
					apstravalidator.AtLeastNAttributes(
						1,
						"hostname", "id", "label", "role", "system_id", "system_type", "tag_ids",
					),
				},
			},
		},
		"query_strings": dataSourceSchema.ListAttribute{
			MarkdownDescription: "Graph DB query strings based on the supplied filters; possibly useful for troubleshooting.",
			Computed:            true,
			ElementType:         types.StringType,
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

	// we're always going to perform at least one query, but we keep 'em as a slice
	var queries []apstra.MatchQuery
	switch {
	case utils.HasValue(o.Filter): // one query, because the user specified 'filter' (deprecated)
		var filter NodeTypeSystemAttributes
		if utils.HasValue(o.Filter) {
			diags.Append(o.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return
			}
		}

		queries = []apstra.MatchQuery{*filter.query(ctx, diags)}
		if diags.HasError() {
			return
		}
	case utils.HasValue(o.Filters): // many queries, because the user specified 'filters'
		var filters []NodeTypeSystemAttributes
		if utils.HasValue(o.Filters) {
			diags.Append(o.Filters.ElementsAs(ctx, &filters, false)...)
			if diags.HasError() {
				return
			}
		}

		queries = make([]apstra.MatchQuery, len(filters))
		for i, filter := range filters {
			queries[i] = *filter.query(ctx, diags)
			if diags.HasError() {
				return
			}
		}
	default: // one query because the user specified no filters - and we create a catchall
		queries = []apstra.MatchQuery{*NodeTypeSystemAttributes{}.query(ctx, diags)}
		if diags.HasError() {
			return
		}
	}

	// create a map of node IDs (for unique-ness) and a slice of query strings
	idMap := make(map[string]bool)
	queryStrings := make([]string, len(queries))

	// collect IDs and a query string for each filter/query
	for i, query := range queries {
		query. // flesh out the query with info needed to run it
			SetClient(client).
			SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
			SetBlueprintType(apstra.BlueprintTypeStaging)
		if diags.HasError() {
			return
		}

		// run the query
		err := query.Do(ctx, &queryResponse)
		if err != nil {
			diags.AddError("Error executing Blueprint query", err.Error())
			return
		}

		// collect the matching system IDs
		for j := range queryResponse.Items {
			idMap[queryResponse.Items[j].System.Id] = true
		}

		// save the query string
		queryStrings[i] = query.String()
	}

	// pull the IDs out of the map
	ids := make([]attr.Value, len(idMap))
	var i int
	for id := range idMap {
		ids[i] = types.StringValue(id)
		i++
	}

	o.Ids = types.SetValueMust(types.StringType, ids)
	o.QueryStrings = value.ListOrNull(ctx, types.StringType, queryStrings, diags)
}

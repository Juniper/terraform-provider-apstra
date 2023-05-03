package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-apstra/apstra/utils"
)

type Systems struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Filters     types.Object `tfsdk:"filters"`
	Ids         types.Set    `tfsdk:"ids"`
	QueryString types.String `tfsdk:"query_string"`
}

func (o Systems) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint to search.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"filters": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Filters used to select only desired node IDs.",
			Optional:            true,
			Attributes:          SystemNode{}.DataSourceAttributesAsFilter(),
		},
		"query_string": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Graph DB query string based on the supplied filters; possibly useful for troubleshooting.",
			Computed:            true,
		},
		"ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "IDs of matching `system` Graph DB nodes.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o Systems) Query(ctx context.Context, diags *diag.Diagnostics) *apstra.MatchQuery {
	var filters SystemNode
	if utils.Known(o.Filters) {
		diags.Append(o.Filters.As(ctx, &filters, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil
		}
	}

	var tagIds []string
	if utils.Known(filters.TagIds) {
		diags.Append(filters.TagIds.ElementsAs(ctx, &tagIds, false)...)
		if diags.HasError() {
			return nil
		}
	}

	systemNodeBaseAttributes := []apstra.QEEAttribute{
		{Key: "type", Value: apstra.QEStringVal("system")},
		{Key: "name", Value: apstra.QEStringVal("n_system")},
	}

	// []QEEAttribute to match the system hostname, label, role, etc... as specified by `filters`
	systemNodeAttributes := append(systemNodeBaseAttributes, filters.QEEAttributes()...)

	// []QEEAttribute to match the relationship between system and tag nodes
	relationshipAttributes := []apstra.QEEAttribute{{Key: "type", Value: apstra.QEStringVal("tag")}}

	// []QEEAttribute to match the tag node (further qualified in the loop below)
	tagNodeBaseAttributes := []apstra.QEEAttribute{{Key: "type", Value: apstra.QEStringVal("tag")}}

	// This is the query we actually want to execute. It's a `match()`
	// query-of-queries which selects the system node using
	// `systemNodeAttributes` and also selects paths from the system node to
	// each specified tag.
	query := new(apstra.MatchQuery)

	// first query: the system node with filters.
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

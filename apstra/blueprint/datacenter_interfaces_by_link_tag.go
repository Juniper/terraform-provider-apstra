package blueprint

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const interfacesByTagDefaultSystemType = "switch"

type InterfacesByLinkTag struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Tags        types.Set    `tfsdk:"tags"`
	SystemType  types.String `tfsdk:"system_type"`
	SystemRole  types.String `tfsdk:"system_role"`
	Ids         types.Set    `tfsdk:"ids"`
	GraphQuery  types.String `tfsdk:"graph_query"`
}

func (o InterfacesByLinkTag) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of required Tags",
			Required:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"system_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Used to specify which interface/end of the link we're "+
				"looking for. Default value is `%s`.", interfacesByTagDefaultSystemType),
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"system_role": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Used to further specify which interface/end of the link we're looking for when" +
				"both ends lead to the same type. For example, on a switch-to-switch link from spine to leaf, " +
				"specify either `spine` or `leaf`.",
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "A set of Apstra object IDs representing selected interfaces.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"graph_query": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The graph datastore query used to perform the lookup.",
			Computed:            true,
		},
	}
}

func (o InterfacesByLinkTag) RunQuery(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) ([]string, apstra.QEQuery) {
	// set default system type
	if o.SystemType.IsNull() {
		o.SystemType = types.StringValue(interfacesByTagDefaultSystemType)
	}

	// extract the tag set
	var tags []string
	diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		return nil, nil
	}

	// create a query which uses the tag set to select candiate interfaces
	tq := new(apstra.MatchQuery)
	for _, tag := range tags {
		tq.Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeTag.QEEAttribute(),
				{Key: "label", Value: apstra.QEStringVal(tag)},
			}).
			Out([]apstra.QEEAttribute{
				apstra.RelationshipTypeTag.QEEAttribute(),
			}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeLink.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("n_link")},
			}))
	}

	// create a query which follows discovered interfaces to the intended systems
	q1 := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_link")}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_type", Value: apstra.QEStringVal("ethernet")},
			{Key: "name", Value: apstra.QEStringVal("n_phy_interface")},
		}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()})
	if o.SystemRole.IsNull() {
		q1.Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "system_type", Value: apstra.QEStringVal(o.SystemType.ValueString())},
		})
	} else {
		q1.Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "system_type", Value: apstra.QEStringVal(o.SystemType.ValueString())},
			{Key: "role", Value: apstra.QEStringVal(o.SystemRole.ValueString())},
		})
	}

	// an optional query which finds aggregations
	q2 := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_phy_interface")}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOf.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_lag_interface")},
		})

	// an optional query which finds multi-chassis aggregations
	q3 := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_lag_interface")}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOf.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_mlag_interface")},
		})

	// assemble the final query
	query := new(apstra.MatchQuery).
		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetClient(client.Client()).
		Match(tq).
		Match(q1).
		Optional(q2).
		Optional(q3)

	var queryResult struct {
		Items []struct {
			PhyInterface struct {
				Id *string `json:"id"`
			} `json:"n_phy_interface"`
			LagInterface struct {
				Id *string `json:"id"`
			} `json:"n_lag_interface"`
			MLagInterface struct {
				Id *string `json:"id"`
			} `json:"n_mlag_interface"`
		} `json:"items"`
	}

	// execute the query
	err := query.Do(ctx, &queryResult)
	if err != nil {
		diags.AddError("failed executing graph query", err.Error())
		return nil, nil
	}

	// pack ID strings into a map
	idMap := make(map[string]struct{})
	for _, item := range queryResult.Items {
		switch {
		case item.MLagInterface.Id != nil:
			idMap[*item.MLagInterface.Id] = struct{}{}
		case item.LagInterface.Id != nil:
			idMap[*item.LagInterface.Id] = struct{}{}
		case item.PhyInterface.Id != nil:
			idMap[*item.PhyInterface.Id] = struct{}{}
		default:
			diags.AddWarning("graph query result included no interfaces",
				fmt.Sprintf("query: %q", query.String()))
		}
	}

	// convert the idMap into a slice
	result := make([]string, len(idMap))
	var i int
	for id := range idMap {
		result[i] = id
		i++
	}

	return result, query
}

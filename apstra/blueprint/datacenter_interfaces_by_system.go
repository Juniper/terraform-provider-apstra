package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InterfacesBySystem struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	SystemId    types.String `tfsdk:"system_id"`
	IfMap       types.Map    `tfsdk:"if_map"`
	GraphQuery  types.String `tfsdk:"graph_query"`
}

func (o InterfacesBySystem) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"system_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra System ID within the Blueprint.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"if_map": dataSourceSchema.MapAttribute{
			MarkdownDescription: "A map of Apstra object IDs representing selected Interfaces, keyed by Interface " +
				"name (xe-0/0/0, etc...)",
			Computed:    true,
			ElementType: types.StringType,
		},
		"graph_query": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The graph datastore query used to perform the lookup.",
			Computed:            true,
		},
	}
}

func (o InterfacesBySystem) RunQuery(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) (map[string]string, apstra.QEQuery) {
	//node(type='system', id='VQM_UdeyD4aKAQ9nrPM')
	//  .out(type='hosted_interfaces')
	//  .node(type='interface', if_name=not_none(), name='n_interface')
	query := new(apstra.PathQuery).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(o.SystemId.ValueString())},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_interface")},
			{Key: "id", Value: apstra.QENone(false)},
			{Key: "if_name", Value: apstra.QENone(false)},
		})

	var queryResult struct {
		Items []struct {
			Interface struct {
				Id     string `json:"id"`
				IfName string `json:"if_name"`
			} `json:"n_interface"`
		} `json:"items"`
	}

	// execute the query
	err := query.Do(ctx, &queryResult)
	if err != nil {
		diags.AddError("failed executing graph query", err.Error())
		return nil, nil
	}

	// prep the result
	result := make(map[string]string)
	for _, item := range queryResult.Items {
		if _, ok := result[item.Interface.IfName]; ok {
			diags.AddError("interface name collision",
				fmt.Sprintf("multiple interfaces with \"if_name=%s\" found by graph query %q",
					item.Interface.IfName, query.String()))
			return nil, nil
		}
		result[item.Interface.IfName] = item.Interface.Id
	}

	return result, query
}

package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterSvis struct {
	BlueprintId    types.String `tfsdk:"blueprint_id"`
	InterfaceToSvi types.Map    `tfsdk:"by_id"`
	NetworkToSvi   types.Map    `tfsdk:"by_virtual_network"`
	SystemToSvi    types.Map    `tfsdk:"by_system"`
	GraphQuery     types.String `tfsdk:"graph_query"`
}

func (o DatacenterSvis) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"by_id": dataSourceSchema.MapAttribute{
			MarkdownDescription: "A map of sets of SVI info keyed by SVI ID.",
			Computed:            true,
			ElementType:         types.ObjectType{AttrTypes: SviMapEntry{}.AttrTypes()},
		},
		"by_virtual_network": dataSourceSchema.MapAttribute{
			MarkdownDescription: "A map of sets of SVI info keyed by Virtual Network ID.",
			Computed:            true,
			ElementType: types.SetType{
				ElemType: types.ObjectType{AttrTypes: SviMapEntry{}.AttrTypes()},
			},
		},
		"by_system": dataSourceSchema.MapAttribute{
			MarkdownDescription: "A map of sets of SVI info keyed by System ID.",
			Computed:            true,
			ElementType: types.SetType{
				ElemType: types.ObjectType{AttrTypes: SviMapEntry{}.AttrTypes()},
			},
		},
		"graph_query": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The graph datastore query used to perform the lookup.",
			Computed:            true,
		},
	}
}

func (o DatacenterSvis) GetSviInfo(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) ([]SviMapEntry, apstra.QEQuery) {
	// query to find paths from VNs to instances of those VNs
	vnQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeVirtualNetwork.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_virtual_network")},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeInstantiatedBy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeVirtualNetworkInstance.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_vn_instance")},
		})

	// query to find paths from switches to SVIs
	vnInstanceQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_system")},
			{Key: "system_type", Value: apstra.QEStringVal("switch")},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedVnInstances.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeVirtualNetworkInstance.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_vn_instance")},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeMemberInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_interface")},
		})

	// final query
	query := new(apstra.MatchQuery).
		SetClient(client.Client()).
		SetBlueprintId(client.Id()).
		Match(vnQuery).
		Match(vnInstanceQuery)

	var queryResult struct {
		Items []struct {
			Interface struct {
				Id       string  `json:"id"`
				IfName   string  `json:"if_name"`
				IPv4Addr *string `json:"ipv4_addr"`
				IPv6Addr *string `json:"ipv6_addr"`
			} `json:"n_interface"`
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
			VirtualNetwork struct {
				Id string `json:"id"`
			} `json:"n_virtual_network"`
			VirtualNetworkInstance struct {
				IPv4Mode string `json:"ipv4_mode"`
				IPv6Mode string `json:"ipv6_mode"`
			} `json:"n_vn_instance"`
		} `json:"items"`
	}

	// execute the query
	err := query.Do(ctx, &queryResult)
	if err != nil {
		diags.AddError("failed executing graph query", err.Error())
		return nil, nil
	}

	// prep the result
	result := make([]SviMapEntry, len(queryResult.Items))
	for i, item := range queryResult.Items {
		result[i] = SviMapEntry{
			SystemId:  types.StringValue(item.System.Id),
			Id:        types.StringValue(item.Interface.Id),
			Name:      types.StringValue(item.Interface.IfName),
			Ipv4Addr:  types.StringPointerValue(item.Interface.IPv4Addr),
			Ipv6Addr:  types.StringPointerValue(item.Interface.IPv6Addr),
			Ipv4Mode:  types.StringValue(item.VirtualNetworkInstance.IPv4Mode),
			Ipv6Mode:  types.StringValue(item.VirtualNetworkInstance.IPv6Mode),
			NetworkId: types.StringValue(item.VirtualNetwork.Id),
		}
	}

	return result, query
}

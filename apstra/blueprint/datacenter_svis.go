package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterSvis struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	SviMap      types.Map    `tfsdk:"svi_map"`
	GraphQuery  types.String `tfsdk:"graph_query"`
}

func (o DatacenterSvis) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"svi_map": dataSourceSchema.MapAttribute{
			MarkdownDescription: "A map of sets of SVI info keyed by Virtual Network ID.",
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

func (o DatacenterSvis) RunQuery(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) (map[string]types.Set, apstra.QEQuery) {
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
	sliceMap := make(map[string][]attr.Value)
	for _, item := range queryResult.Items {
		attrVal, d := types.ObjectValueFrom(ctx, SviMapEntry{}.AttrTypes(), &SviMapEntry{
			SystemId: types.StringValue(item.System.Id),
			SviId:    types.StringValue(item.Interface.Id),
			Name:     types.StringValue(item.Interface.IfName),
			Ipv4Addr: types.StringPointerValue(item.Interface.IPv4Addr),
			Ipv6Addr: types.StringPointerValue(item.Interface.IPv6Addr),
			Ipv4Mode: types.StringValue(item.VirtualNetworkInstance.IPv4Mode),
			Ipv6Mode: types.StringValue(item.VirtualNetworkInstance.IPv6Mode),
		})
		diags.Append(d...)
		if diags.HasError() {
			return nil, nil
		}

		sliceMap[item.VirtualNetwork.Id] = append(sliceMap[item.VirtualNetwork.Id], attrVal)
	}

	result := make(map[string]types.Set, len(sliceMap))
	for k, v := range sliceMap {
		result[k] = types.SetValueMust(types.ObjectType{AttrTypes: SviMapEntry{}.AttrTypes()}, v)
	}

	return result, query
}

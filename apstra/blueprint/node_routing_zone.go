package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
)

type NodeTypeSecurityZone struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Id          types.String `tfsdk:"id"`
	Attributes  types.Object `tfsdk:"attributes"`
}

func (o NodeTypeSecurityZone) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID",
			Required:            true,
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph datastore node `id`",
			Required:            true,
		},
		"attributes": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Attributes of a `security_zone` (graph datastore term for `routing_zone`) graph node.",
			Computed:            true,
			Attributes:          NodeTypeRoutingZoneAttributes{}.DataSourceAttributes(),
		},
	}
}

func (o *NodeTypeSecurityZone) ReadFromApi(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	var err error

	type securityZoneNode struct {
		Id      string `json:"id"`
		Label   string `json:"label"`
		SzType  string `json:"sz_type"`
		VlanId  int64  `json:"vlan_id"`
		VniId   int64  `json:"vni_id"`
		VrfId   int64  `json:"vrf_id"`
		VrfName string `json:"vrf_name"`
	}

	type dhcpPolicyNode struct {
		DhcpServers []string `json:"dhcp_servers"`
	}

	type routingPolicyNode struct {
		Id string `json:"id"`
	}

	rpQuery := new(apstra.PathQuery).
		SetClient(client).
		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeSecurityZone.String())},
			{Key: "name", Value: apstra.QEStringVal("n_security_zone")},
		}).
		Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("policy")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeRoutingPolicy.String())},
			{Key: "name", Value: apstra.QEStringVal("n_routing_policy")},
		})

	rpResponse := &struct {
		Items []struct {
			SecurityZone  securityZoneNode  `json:"n_security_zone"`
			RoutingPolicy routingPolicyNode `json:"n_routing_policy"`
		} `json:"items"`
	}{}

	dhcpQuery := new(apstra.PathQuery).
		SetClient(client).
		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeSecurityZone.String())},
		}).
		Out([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("policy")},
		}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypePolicy.String())},
			{Key: "name", Value: apstra.QEStringVal("n_dhcp_policy")},
		})

	dhcpResponse := &struct {
		Items []struct {
			DhcpPolicy dhcpPolicyNode `json:"n_dhcp_policy"`
		} `json:"items"`
	}{}

	err = rpQuery.Do(ctx, rpResponse)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error querying graph datastore for blueprint %s", o.BlueprintId),
			err.Error())
		return
	}
	if len(rpResponse.Items) != 1 {
		diags.AddError("failed querying graph datastore for Routing Zone with Routing Policy",
			fmt.Sprintf("expected 1 result, got %d using query: %q", len(rpResponse.Items), rpQuery.String()))
		return
	}

	// rpRespnose.Items has exactly 1 element. Pull out the interesting bits.
	sz := rpResponse.Items[0].SecurityZone
	rp := rpResponse.Items[0].RoutingPolicy

	err = dhcpQuery.Do(ctx, dhcpResponse)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error querying graph datastore for blueprint %s", o.BlueprintId),
			err.Error())
		return
	}
	var dhcpServers []string
	switch len(dhcpResponse.Items) {
	case 0: // dhcpServers slice stays nil
	case 1:
		dhcpServers = dhcpResponse.Items[0].DhcpPolicy.DhcpServers
	default:
		diags.AddError("failed querying graph datastore for Routing Zone with DHCP Policy",
			fmt.Sprintf("expected no more than 1 result, got %d using query: %q", len(dhcpResponse.Items), dhcpQuery.String()))
		return
	}

	o.Attributes = types.ObjectValueMust(NodeTypeRoutingZoneAttributes{}.AttrTypes(), map[string]attr.Value{
		"id":                types.StringValue(sz.Id),
		"name":              types.StringValue(sz.Label),
		"type":              types.StringValue(sz.SzType),
		"vlan_id":           types.Int64Value(sz.VlanId),
		"vni":               types.Int64Value(sz.VniId),
		"vrf_id":            types.Int64Value(sz.VrfId),
		"vrf_name":          types.StringValue(sz.VrfName),
		"dhcp_servers":      utils.SetValueOrNull(ctx, types.StringType, dhcpServers, diags),
		"routing_policy_id": types.StringValue(rp.Id),
	})
}

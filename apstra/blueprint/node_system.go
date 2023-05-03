package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NodeTypeSystem struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Id          types.String `tfsdk:"id"`
	Attributes  types.Object `tfsdk:"attributes"`
}

func (o NodeTypeSystem) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID",
			Required:            true,
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `ID`",
			Required:            true,
		},
		"attributes": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Attributes of a `system` Graph DB node.",
			Computed:            true,
			Attributes:          NodeTypeSystemAttributes{}.DataSourceAttributes(),
		},
	}
}

func (o *NodeTypeSystem) ReadFromApi(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	type node struct {
		Id         string `json:"id"`
		Hostname   string `json:"hostname"`
		Label      string `json:"label"`
		Role       string `json:"role"`
		SystemId   string `json:"system_id"`
		SystemType string `json:"system_type"`
	}
	nodeResponse := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}

	err := client.GetNodes(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), apstra.NodeTypeSystem, nodeResponse)
	if err != nil {
		diags.AddError("error fetching blueprint nodes", err.Error())
		return
	}

	// pick out the desired node from the node slice in the response object
	desiredNode := new(node)
	for _, n := range nodeResponse.Nodes {
		if n.Id == o.Id.ValueString() {
			desiredNode = &n
			break
		}
	}

	if desiredNode == nil {
		diags.AddError("node not found",
			fmt.Sprintf("node %q not found in blueprint %q",
				o.Id.ValueString(), o.BlueprintId.ValueString()))
		return
	}

	tagResponse := &struct {
		Items []struct {
			Tag struct {
				Label string `json:"label"`
			} `json:"n_tag"`
		} `json:"items"`
	}{}

	err = new(apstra.PathQuery).
		SetClient(client).
		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(desiredNode.Id)}}).
		In([]apstra.QEEAttribute{{Key: "type", Value: apstra.QEStringVal("tag")}}).
		Node([]apstra.QEEAttribute{
			{Key: "type", Value: apstra.QEStringVal("tag")},
			{Key: "name", Value: apstra.QEStringVal("n_tag")},
		}).
		Do(ctx, tagResponse)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error querying graph db tags for node %q", desiredNode.Id),
			err.Error())
		return
	}

	tags := make([]attr.Value, len(tagResponse.Items))
	for i := range tagResponse.Items {
		tags[i] = types.StringValue(tagResponse.Items[i].Tag.Label)
	}

	o.Attributes = types.ObjectValueMust(NodeTypeSystemAttributes{}.AttrTypes(), map[string]attr.Value{
		"id":          types.StringValue(desiredNode.Id),
		"hostname":    types.StringValue(desiredNode.Hostname),
		"label":       types.StringValue(desiredNode.Label),
		"role":        types.StringValue(desiredNode.Role),
		"system_id":   types.StringValue(desiredNode.SystemId),
		"system_type": types.StringValue(desiredNode.SystemType),
		"tag_ids":     types.SetValueMust(types.StringType, tags),
	})
}

package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NodeTypeSystem struct {
	BlueprintId      types.String `tfsdk:"blueprint_id"`
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	NullWhenNotFound types.Bool   `tfsdk:"null_when_not_found"`
	Attributes       types.Object `tfsdk:"attributes"`
}

func (o NodeTypeSystem) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID",
			Required:            true,
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Graph DB node `id` field. Required when `name` is omitted.",
			Optional:            true,
			Validators: []validator.String{stringvalidator.ExactlyOneOf(path.Expressions{
				path.MatchRelative(),
				path.MatchRoot("name"),
			}...)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Web UI name (Graph DB `label` field). Required when `id` is omitted.",
			Optional:            true,
		},
		"null_when_not_found": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "When `true` and the specified object is not found, rather than raising an error, " +
				"computed values are set to `null`.",
			Optional: true,
		},
		"attributes": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Attributes of a `system` Graph DB node.",
			Computed:            true,
			Attributes:          NodeTypeSystemAttributes{}.DataSourceAttributes(),
		},
	}
}

func (o *NodeTypeSystem) AttributesFromApi(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
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
	var desiredNode node
	var ok bool
	switch {
	case !o.Id.IsNull():
		desiredNode, ok = nodeResponse.Nodes[o.Id.ValueString()]
		if !ok {
			if o.NullWhenNotFound.ValueBool() {
				o.Attributes = types.ObjectNull(NodeTypeSystemAttributes{}.AttrTypes())
				return
			}
			diags.AddError("Node not found",
				fmt.Sprintf("Node with ID %q not found in blueprint %q",
					o.Id.ValueString(), o.BlueprintId.ValueString()))
		}
	case !o.Name.IsNull():
		for _, n := range nodeResponse.Nodes {
			if n.Label == o.Name.ValueString() {
				desiredNode = n
				ok = true
				break
			}
		}
		if !ok {
			if o.NullWhenNotFound.ValueBool() {
				o.Attributes = types.ObjectNull(NodeTypeSystemAttributes{}.AttrTypes())
				return
			}
			diags.AddError("Node not found",
				fmt.Sprintf("Node with Name %q not found in blueprint %q",
					o.Name.ValueString(), o.BlueprintId.ValueString()))
		}
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
		"id":          utils.StringValueOrNull(ctx, desiredNode.Id, diags),
		"hostname":    utils.StringValueOrNull(ctx, desiredNode.Hostname, diags),
		"label":       utils.StringValueOrNull(ctx, desiredNode.Label, diags),
		"role":        utils.StringValueOrNull(ctx, desiredNode.Role, diags),
		"system_id":   utils.StringValueOrNull(ctx, desiredNode.SystemId, diags),
		"system_type": utils.StringValueOrNull(ctx, desiredNode.SystemType, diags),
		"tag_ids":     utils.SetValueOrNull(ctx, types.StringType, tags, diags),
	})
}

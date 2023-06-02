package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterGenericSystemLinkEndpoint struct {
	SystemId         types.String `tfsdk:"system_id"`
	IfName           types.String `tfsdk:"if_name"`
	TransformationId types.Int64  `tfsdk:"transformation_id"`
}

func (o DatacenterGenericSystemLinkEndpoint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"system_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Graph Node ID of the Leaf Switch or Access Switch where the link connects.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"if_name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name of the physical interface where the link connects (\"ge-0/0/1\" or similar).",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"transformation_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "Transformation ID sets the operational mode of an interface.",
			Required:            true,
			Validators:          []validator.Int64{int64validator.AtLeast(1)},
		},
	}
}

func (o DatacenterGenericSystemLinkEndpoint) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"system_id":         types.StringType,
		"if_name":           types.StringType,
		"transformation_id": types.Int64Type,
	}
}

func (o DatacenterGenericSystemLinkEndpoint) Request(_ context.Context, _ *diag.Diagnostics) apstra.SwitchLinkEndpoint {
	return apstra.SwitchLinkEndpoint{
		TransformationId: int(o.TransformationId.ValueInt64()),
		SystemId:         apstra.ObjectId(o.SystemId.ValueString()),
		IfName:           o.IfName.ValueString(),
	}
}

func (o *DatacenterGenericSystemLinkEndpoint) linkId(ctx context.Context, client *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) string {
	query := new(apstra.PathQuery).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetClient(client.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(o.SystemId.ValueString())},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_name", Value: apstra.QEStringVal(o.IfName.ValueString())},
			{Key: "if_type", Value: apstra.QEStringVal("ethernet")},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeLink.QEEAttribute(),
			{Key: "link_type", Value: apstra.QEStringVal("ethernet")},
			{Key: "name", Value: apstra.QEStringVal("n_link")},
		})

	var result struct {
		Items []struct {
			Link struct {
				Id string `json:"id"`
			} `json:"n_link"`
		} `json:"items"`
	}

	err := query.Do(ctx, &result)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("failed querying graph datastore for new link to switch %s : %s",
				o.SystemId, o.IfName), err.Error())
		return ""
	}
	if len(result.Items) != 1 {
		diags.AddError(
			fmt.Sprintf("expected exactly one result from query, got %d", len(result.Items)),
			fmt.Sprintf("query: %s", query.String()))
		return ""
	}

	return result.Items[0].Link.Id
}

func (o *DatacenterGenericSystemLinkEndpoint) request(_ context.Context, _ *diag.Diagnostics) *apstra.SwitchLinkEndpoint {
	return &apstra.SwitchLinkEndpoint{
		TransformationId: int(o.TransformationId.ValueInt64()),
		SystemId:         apstra.ObjectId(o.SystemId.ValueString()),
		IfName:           o.IfName.ValueString(),
	}
}

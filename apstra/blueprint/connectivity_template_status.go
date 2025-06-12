package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type ConnectivityTemplateStatus struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Status                types.String `tfsdk:"status"`
	ApplicationPointCount types.Int64  `tfsdk:"application_point_count"`
	Tags                  types.Set    `tfsdk:"tags"`
}

func (o ConnectivityTemplateStatus) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                      types.StringType,
		"name":                    types.StringType,
		"description":             types.StringType,
		"status":                  types.StringType,
		"application_point_count": types.Int64Type,
		"tags":                    types.SetType{ElemType: types.StringType},
	}
}

func (o ConnectivityTemplateStatus) DataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Graph node ID of the Connectivity Template",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Connectivity Template, as displayed in the web UI",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the Connectivity Template, as displayed in the web UI",
			Computed:            true,
		},
		"status": schema.StringAttribute{
			MarkdownDescription: fmt.Sprintf(
				"Status of the Connectivity Template - One of: [`%s`]",
				strings.Join(utils.StringersToFriendlyStrings(enum.EndpointPolicyStatuses.Members()), "`, `"),
			),
			Computed: true,
		},
		"application_point_count": schema.Int64Attribute{
			MarkdownDescription: "Count of Application Points to which Connectivity Template has been assigned",
			Computed:            true,
		},
		"tags": schema.SetAttribute{
			MarkdownDescription: "Tags associated with this Connectivity Template",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

func (o *ConnectivityTemplateStatus) LoadApiData(ctx context.Context, in apstra.EndpointPolicyStatus, diags *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Label)
	o.Description = types.StringValue(in.Description)
	o.Status = types.StringValue(in.Status.String())
	o.ApplicationPointCount = types.Int64Value(int64(in.AppPointsCount))
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, in.Tags, diags)
}

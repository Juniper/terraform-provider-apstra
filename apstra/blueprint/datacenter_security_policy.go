package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterSecurityPolicy struct {
	BlueprintId   types.String `tfsdk:"blueprint_id"`
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	SrcAppPointId types.String `tfsdk:"source_application_point_id"`
	DstAppPointId types.String `tfsdk:"destination_application_point_id"`
	Rules         types.List   `tfsdk:"rules"`
	Tags          types.Set    `tfsdk:"tags"`
}

func (o DatacenterSecurityPolicy) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Security Policy by ID. Required when `name` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up a Security Policy by name. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Description of the Security Policy as seen in the Web UI.",
			Computed:            true,
		},
		"enabled": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the Security Policy is enabled.",
			Computed:            true,
		},
		"source_application_point_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of the source Application Point (Virtual Network ID, Routing Zone ID, etc...)",
			Computed:            true,
		},
		"destination_application_point_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of the source Application Point (Virtual Network ID, Routing Zone ID, etc...)",
			Computed:            true,
		},
		"rules": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "List of Rules associated with the Security Policy.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: DatacenterSecurityPolicyRule{}.DataSourceAttributes(),
			},
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags associated with the Security Policy.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o *DatacenterSecurityPolicy) Read(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) error {
	var api *apstra.Policy
	var err error

	if o.Id.IsNull() {
		api, err = bp.GetPolicyByLabel(ctx, o.Name.ValueString())
	} else {
		api, err = bp.GetPolicy(ctx, apstra.ObjectId(o.Id.ValueString()))
	}
	if err != nil {
		return err
	}

	o.Id = types.StringValue(api.Id.String())
	o.loadApiData(ctx, api.Data, diags)

	return nil
}

func (o *DatacenterSecurityPolicy) loadApiData(ctx context.Context, data *apstra.PolicyData, diags *diag.Diagnostics) {
	var srcAppPointId, dstAppPointId types.String
	if data.SrcApplicationPoint != nil {
		srcAppPointId = utils.StringValueOrNull(ctx, data.SrcApplicationPoint.Id.String(), diags)
	}
	if data.DstApplicationPoint != nil {
		dstAppPointId = utils.StringValueOrNull(ctx, data.DstApplicationPoint.Id.String(), diags)
	}

	o.Name = types.StringValue(data.Label)
	o.Description = utils.StringValueOrNull(ctx, data.Description, diags)
	o.Enabled = types.BoolValue(data.Enabled)
	o.SrcAppPointId = srcAppPointId
	o.DstAppPointId = dstAppPointId
	o.Rules = newPolicyRuleList(ctx, data.Rules, diags)
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, data.Tags, diags)
}

package blueprint

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Tag struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (o Tag) AttrTypes(_ context.Context, _ *diag.Diagnostic) map[string]attr.Type {
	return map[string]attr.Type{
		"blueprint_id": types.StringType,
		"name":         types.StringType,
		"description":  types.StringType,
	}
}

func (o Tag) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID",
			Required:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Tag name",
			Required:            true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Tag description",
			Computed:            true,
		},
	}
}

func (o Tag) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Does not apply in filter context - ignore",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Tag name",
			Optional:            true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Tag description",
			Optional:            true,
		},
	}
}

func (o Tag) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag name",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag description",
			Optional:            true,
		},
	}
}

func (o Tag) Request(_ context.Context, _ *diag.Diagnostics) apstra.TwoStageL3ClosTagData {
	return apstra.TwoStageL3ClosTagData{
		Label:       o.Name.ValueString(),
		Description: o.Description.ValueString(),
	}
}

func (o *Tag) LoadApiData(_ context.Context, data *apstra.TwoStageL3ClosTagData, _ *diag.Diagnostics) {
	o.Name = types.StringValue(data.Label)
	o.Description = types.StringNull()
	if data.Description != "" {
		o.Description = types.StringValue(data.Description)
	}
}

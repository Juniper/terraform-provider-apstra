package blueprint

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectivityTemplate struct {
	Id          types.String `tfsdk:"id"`
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.Set    `tfsdk:"tags"`
	Primitives  types.List   `tfsdk:"primitives"`
}

func (o ConnectivityTemplate) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID.",
			Computed:            true,
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in web UI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Description displayed in web UI.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tag labels",
			ElementType:         types.StringType,
			Optional:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
		},
		"primitives": resourceSchema.ListAttribute{
			MarkdownDescription: "List of Connectivity Template Primitives expressed as JSON strings.",
			ElementType:         types.StringType,
			Required:            true,
			Validators:          []validator.List{listvalidator.SizeAtLeast(1)},
		},
	}
}

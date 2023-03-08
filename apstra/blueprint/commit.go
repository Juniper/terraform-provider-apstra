package blueprint

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Commit struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Comment     types.String `tfsdk:"comment"`
}

func (o Commit) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "", // todo
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"comment": resourceSchema.StringAttribute{
			MarkdownDescription: "", // todo
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

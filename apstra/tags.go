package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tagData struct {
	Name        string `tfsdk:"name"`
	Description string `tfsdk:"description"`
}

func (o tagData) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag name (label) field.",
			Computed:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Tag description field.",
			Computed:            true,
		},
	}
}

func (o tagData) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
	}
}

func newTagDataSet(ctx context.Context, in []goapstra.DesignTagData, diags *diag.Diagnostics) types.Set {
	if len(in) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: tagData{}.attrTypes()})
	}

	tags := make([]tagData, len(in))
	for i, tag := range in {
		tags[i] = tagData{
			Name:        tag.Label,
			Description: tag.Description,
		}
	}

	result, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: tagData{}.attrTypes()}, &tags)
	diags.Append(d...)

	return result
}

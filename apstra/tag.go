package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tagData struct {
	Name        string `tfsdk:"name"`
	Description string `tfsdk:"description"`
}

func (o *tagData) parseApi(in *goapstra.DesignTagData) {
	o.Name = in.Label
	o.Description = in.Description
}

func (o tagData) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":        types.StringType,
			"description": types.StringType,
		},
	}
}

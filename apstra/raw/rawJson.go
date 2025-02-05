package raw

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RawJson struct {
	Id           types.String         `tfsdk:"id"`
	Url          types.String         `tfsdk:"url"`
	UpdateMethod types.String         `tfsdk:"update_method"`
	Payload      jsontypes.Normalized `tfsdk:"payload"`
}

func (o *RawJson) ResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "The ID of the raw JSON object. We attempt to determine the ID from the API response. " +
				"If the ID can be anticipated, it is possible to specify it here.",
			Computed:      true,
			Optional:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"url": schema.StringAttribute{
			MarkdownDescription: "The API URL associated with the raw JSON object.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"update_method": schema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("The method used to update the JSON object. Must be one of "+
				"`%s` or `%s`. Default: `%s`", http.MethodPut, http.MethodPatch, http.MethodPut),
			Computed:   true,
			Default:    stringdefault.StaticString(http.MethodPut),
			Validators: []validator.String{stringvalidator.OneOf(http.MethodPut, http.MethodPatch)},
		},
		"payload": schema.StringAttribute{
			CustomType:          jsontypes.NormalizedType{},
			MarkdownDescription: "JSON payload used to create and update the raw JSON object.",
			Required:            true,
		},
	}
}

package connectivitytemplate

import (
	"context"
	"encoding/json"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ Primitive = &VnSingle{}

type VnSingle struct {
	VnId      types.String `tfsdk:"vn_id"`
	Tagged    types.Bool   `tfsdk:"tagged"`
	Primitive types.String `tfsdk:"primitive"`
}

func (o VnSingle) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"vn_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network ID",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"tagged": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the VN should mark frames belonging to " +
				"the VN with 802.1Q tags. Default: `false`",
			Optional: true,
		},
		"primitive": dataSourceSchema.StringAttribute{
			MarkdownDescription: "JSON output for use in the `primitives` field of an " +
				"`apstra_datacenter_connectivity_template` resource or a different Connectivity " +
				"Template Primitive data source",
			Computed: true,
		},
	}
}

func (o VnSingle) Render(_ context.Context, diags *diag.Diagnostics) string {
	obj := vnSinglePrototype{
		VnId:   o.VnId.ValueString(),
		Tagged: o.Tagged.ValueBool(),
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling VnSingle primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&RenderedPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachSingleVlan.String(),
		Data:          data,
	})

	return string(data)
}

func (o VnSingle) connectivityTemplateAttributes() (apstra.ConnectivityTemplateAttributes, error) {
	vnId := apstra.ObjectId(o.VnId.ValueString())
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan{
		Tagged:   o.Tagged.ValueBool(),
		VnNodeId: &vnId,
	}, nil
}

type vnSinglePrototype struct {
	VnId   string `json:"vn_id"`
	Tagged bool   `json:"tagged"`
}

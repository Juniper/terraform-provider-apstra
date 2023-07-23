package connectivitytemplate

import (
	"context"
	"encoding/json"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
				"Template JsonPrimitive data source",
			Computed: true,
		},
	}
}

func (o VnSingle) Marshal(_ context.Context, diags *diag.Diagnostics) string {
	obj := vnSinglePrototype{
		VnId:   o.VnId.ValueString(),
		Tagged: o.Tagged.ValueBool(),
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling VnSingle primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&TfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachSingleVlan.String(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

var _ JsonPrimitive = &vnSinglePrototype{}

type vnSinglePrototype struct {
	VnId   string `json:"vn_id"`
	Tagged bool   `json:"tagged"`
}

func (o vnSinglePrototype) attributes(_ context.Context, _ path.Path, _ *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	vnId := apstra.ObjectId(o.VnId)
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan{
		Tagged:   o.Tagged,
		VnNodeId: &vnId,
	}
}

func (o vnSinglePrototype) SdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	attributes := o.attributes(ctx, path, diags)
	if diags.HasError() {
		return nil
	}

	return &apstra.ConnectivityTemplatePrimitive{
		Id:          nil, // calculated later
		Attributes:  attributes,
		Subpolicies: nil, // this primitive has no children
		BatchId:     nil, // this primitive has no children
		PipelineId:  nil, // calculated later
	}
}

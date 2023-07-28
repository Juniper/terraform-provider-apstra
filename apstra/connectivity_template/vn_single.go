package connectivitytemplate

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"sort"
	"terraform-provider-apstra/apstra/utils"
)

var _ Primitive = &VnSingle{}

type VnSingle struct {
	VnId      types.String `tfsdk:"vn_id"`
	Tagged    types.Bool   `tfsdk:"tagged"`
	Primitive types.String `tfsdk:"primitive"`
	Children  types.Set    `tfsdk:"children"`
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
		"children": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of JSON strings describing Connectivity Template Primitives " +
				"which are children of this Connectivity Template JsonPrimitive. Use the `primitive` " +
				"attribute of other Connectivity Template Primitives data sources here.",
			ElementType: types.StringType,
			Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
			Optional:    true,
		},
	}
}

func (o VnSingle) Marshal(ctx context.Context, diags *diag.Diagnostics) string {
	var children []string
	diags.Append(o.Children.ElementsAs(ctx, &children, false)...)
	if diags.HasError() {
		return ""
	}

	// sort the children by their SHA1 sums for easier comparison of nested strings
	sort.Slice(children, func(i, j int) bool {
		sum1 := sha1.Sum([]byte(children[i]))
		sum2 := sha1.Sum([]byte(children[j]))
		return bytes.Compare(sum1[:], sum2[:]) >= 0
	})

	obj := vnSinglePrototype{
		VnId:     o.VnId.ValueString(),
		Tagged:   o.Tagged.ValueBool(),
		Children: children,
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling VnSingle primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachSingleVlan.String(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o *VnSingle) loadSdkPrimitive(ctx context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	attributes, ok := in.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan)
	if !ok {
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %T", in))
		return
	}

	if attributes.VnNodeId != nil {
		o.VnId = types.StringValue(attributes.VnNodeId.String())
	} else {
		o.VnId = types.StringNull()
	}
	o.Tagged = types.BoolValue(attributes.Tagged)
	o.Children = utils.SetValueOrNull(ctx, types.StringType, SdkPrimitivesToJsonStrings(ctx, in.Subpolicies, diags), diags)
}

var _ JsonPrimitive = &vnSinglePrototype{}

type vnSinglePrototype struct {
	VnId     string   `json:"vn_id"`
	Tagged   bool     `json:"tagged"`
	Children []string `json:"children,omitempty"`
}

func (o vnSinglePrototype) attributes(_ context.Context, _ path.Path, _ *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	vnId := apstra.ObjectId(o.VnId)
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachSingleVlan{
		Tagged:   o.Tagged,
		VnNodeId: &vnId,
	}
}

func (o vnSinglePrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	attributes := o.attributes(ctx, path, diags)
	if diags.HasError() {
		return nil
	}

	children := ChildPrimitivesFromListOfJsonStrings(ctx, o.Children, path, diags)
	if diags.HasError() {
		return nil
	}

	return &apstra.ConnectivityTemplatePrimitive{
		Id:          nil, // calculated later
		Attributes:  attributes,
		Subpolicies: children,
		BatchId:     nil, // calculated later
		PipelineId:  nil, // calculated later
	}
}

package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
)

var _ Primitive = &VnMultiple{}

type VnMultiple struct {
	UntaggedVnId types.String `tfsdk:"untagged_vn_id"`
	TaggedVnIds  types.Set    `tfsdk:"tagged_vn_ids"`
	Primitive    types.String `tfsdk:"primitive"`
}

func (o VnMultiple) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"untagged_vn_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network ID which should be presented without VLAN tags",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.DifferentFromValues(path.MatchRoot("tagged_vn_ids")),
			},
		},
		"tagged_vn_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Virtual Network IDs which should be presented with VLAN tags",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				setvalidator.SizeAtLeast(1),
			},
		},
		"primitive": dataSourceSchema.StringAttribute{
			MarkdownDescription: "JSON output for use in the `primitives` field of an " +
				"`apstra_datacenter_connectivity_template` resource or a different Connectivity " +
				"Template JsonPrimitive data source",
			Computed: true,
		},
	}
}

func (o VnMultiple) Marshal(ctx context.Context, diags *diag.Diagnostics) string {
	var untaggedVnId *apstra.ObjectId
	if !o.UntaggedVnId.IsNull() {
		vnId := apstra.ObjectId(o.UntaggedVnId.ValueString())
		untaggedVnId = &vnId
	}

	var taggedVnIds []apstra.ObjectId
	diags.Append(o.TaggedVnIds.ElementsAs(ctx, &taggedVnIds, false)...)
	if diags.HasError() {
		return ""
	}

	if taggedVnIds == nil {
		// nil slice causes API errors
		taggedVnIds = []apstra.ObjectId{}
	}

	obj := vnMultiplePrototype{
		UntaggedVnId: untaggedVnId,
		TaggedVnIds:  taggedVnIds,
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling VnMultiple primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachMultipleVLAN.String(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o *VnMultiple) loadSdkPrimitive(_ context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	attributes, ok := in.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachMultipleVlan)
	if !ok {
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %T", in))
		return
	}

	if attributes.UntaggedVnNodeId != nil {
		o.UntaggedVnId = types.StringValue(attributes.UntaggedVnNodeId.String())
	} else {
		o.UntaggedVnId = types.StringNull()
	}

	taggedVnIds := make([]attr.Value, len(attributes.TaggedVnNodeIds))
	for i, id := range attributes.TaggedVnNodeIds {
		taggedVnIds[i] = types.StringValue(id.String())
	}
	o.TaggedVnIds = types.SetValueMust(types.StringType, taggedVnIds)
}

var _ JsonPrimitive = &vnMultiplePrototype{}

type vnMultiplePrototype struct {
	UntaggedVnId *apstra.ObjectId  `json:"untagged_vn_id"`
	TaggedVnIds  []apstra.ObjectId `json:"tagged_vn_ids"`
}

func (o vnMultiplePrototype) attributes(_ context.Context, _ path.Path, _ *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachMultipleVlan{
		UntaggedVnNodeId: o.UntaggedVnId,
		TaggedVnNodeIds:  o.TaggedVnIds,
	}
}

func (o vnMultiplePrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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

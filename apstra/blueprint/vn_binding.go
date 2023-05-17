package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/utils"
)

type VnBinding struct {
	VlanId    types.Int64  `tfsdk:"vlan_id"`
	LeafId    types.String `tfsdk:"leaf_id"`
	AccessIds types.Set    `tfsdk:"access_ids"`
}

func (o VnBinding) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"vlan_id":    types.Int64Type,
		"leaf_id":    types.StringType,
		"access_ids": types.SetType{ElemType: types.StringType},
	}

}

func (o VnBinding) DataSourceAttributesConstructorOutput() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"vlan_id": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The value supplied as `vlan_id` at the root of this datasource " +
				"configuration, if any. May be `null`, in which case Apstra will choose.",
			Computed: true,
		},
		"leaf_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "A graph db node ID representing a leaf switch `system` node or " +
				"`redundancy_group` node.",
			Computed: true,
		},
		"access_ids": dataSourceSchema.SetAttribute{
			MarkdownDescription: "A set of zero or more graph db node IDs representing access " +
				"switch `system` nodes or a `redundancy_group` nodes.",
			Computed:    true,
			ElementType: types.StringType,
		},
	}
}

func (o VnBinding) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"vlan_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "When not specified, Apstra will choose the VLAN to be used on each switch.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.Int64{int64validator.Between(design.VlanMin-1, design.VlanMax+1)},
		},
		"leaf_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The graph db node ID of the leaf switch `system` node (nonredundant " +
				"leaf switch) or `redundancy_group` node (MLAG or ESI LAG leaf switches) to which this " +
				"VN should be bound.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"access_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "The graph db node ID of the access switch `system` node (nonredundant " +
				"access switch) or `redundancy_group` node (ESI LAG access switches) beneath `leaf_id` " +
				"to which this VN should be bound.",
			Optional:    true,
			ElementType: types.StringType,
		},
	}
}

func (o VnBinding) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.VnBinding {
	var vlanId *apstra.Vlan
	if utils.Known(o.VlanId) {
		v := apstra.Vlan(o.VlanId.ValueInt64())
		vlanId = &v
	}

	var result apstra.VnBinding
	result.SystemId = apstra.ObjectId(o.LeafId.ValueString())
	result.VlanId = vlanId
	diags.Append(o.AccessIds.ElementsAs(ctx, &result.AccessSwitchNodeIds, false)...)
	return &result
}

func (o *VnBinding) LoadApiData(ctx context.Context, in apstra.VnBinding, diags *diag.Diagnostics) {
	o.LeafId = types.StringValue(in.SystemId.String())
	o.VlanId = utils.Int64ValueOrNull(ctx, in.VlanId, diags)
	o.AccessIds = utils.SetValueOrNull(ctx, types.StringType, in.AccessSwitchNodeIds, diags)
}

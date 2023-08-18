package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ Primitive = &RoutingPolicy{}

type RoutingPolicy struct {
	Label           types.String `tfsdk:"label"`
	RoutingPolicyId types.String `tfsdk:"routing_policy_id"`
	Primitive       types.String `tfsdk:"primitive"`
}

func (o RoutingPolicy) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"label": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Primitive label displayed in the web UI",
			Optional:            true,
		},
		"routing_policy_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of Routing Policy to be attached.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"primitive": dataSourceSchema.StringAttribute{
			MarkdownDescription: "JSON output for use in the `primitives` field of an " +
				"`apstra_datacenter_connectivity_template` resource or a different Connectivity " +
				"Template Primitive data source",
			Computed: true,
		},
	}
}

func (o RoutingPolicy) Marshal(_ context.Context, diags *diag.Diagnostics) string {
	obj := routingPolicyPrototype{}
	if !o.RoutingPolicyId.IsNull() {
		id := o.RoutingPolicyId.ValueString()
		obj.RoutingPolicyId = &id
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling RoutingPolicy primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachExistingRoutingPolicy.String(),
		Label:         o.Label.ValueString(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o *RoutingPolicy) loadSdkPrimitive(ctx context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	attributes, ok := in.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy)
	if !ok {
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %T", in))
		return
	}

	o.loadSdkPrimitiveAttributes(ctx, attributes, diags)
	if diags.HasError() {
		return
	}

	o.Label = types.StringValue(in.Label)
}

func (o *RoutingPolicy) loadSdkPrimitiveAttributes(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy, _ *diag.Diagnostics) {
	o.RoutingPolicyId = types.StringNull()
	if in.RpToAttach != nil {
		o.RoutingPolicyId = types.StringValue(in.RpToAttach.String())
	}
}

var _ JsonPrimitive = &routingPolicyPrototype{}

type routingPolicyPrototype struct {
	Label           string  `json:"label,omitempty"`
	RoutingPolicyId *string `json:"routing_policy_id"`
}

func (o routingPolicyPrototype) attributes(_ context.Context, _ path.Path, _ *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	rpId := apstra.ObjectId(*o.RoutingPolicyId)
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachExistingRoutingPolicy{
		RpToAttach: &rpId,
	}
}

func (o routingPolicyPrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
	attributes := o.attributes(ctx, path, diags)
	if diags.HasError() {
		return nil
	}

	return &apstra.ConnectivityTemplatePrimitive{
		Id:          nil, // calculated later
		Label:       o.Label,
		Attributes:  attributes,
		Subpolicies: nil, // this primitive has no children
		BatchId:     nil, // this primitive has no children
		PipelineId:  nil, // calculated later
	}
}

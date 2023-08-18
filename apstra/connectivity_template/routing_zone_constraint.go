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

var _ Primitive = &RoutingZoneConstraint{}

type RoutingZoneConstraint struct {
	Label                   types.String `tfsdk:"label"`
	RoutingZoneConstraintId types.String `tfsdk:"routing_zone_constraint_id"`
	Primitive               types.String `tfsdk:"primitive"`
}

func (o RoutingZoneConstraint) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"label": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Primitive label displayed in the web UI",
			Optional:            true,
		},
		"routing_zone_constraint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Object ID of Routing Zone Constraint to be attached.",
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

func (o RoutingZoneConstraint) Marshal(_ context.Context, diags *diag.Diagnostics) string {
	obj := routingZoneConstraintPrototype{}
	if !o.RoutingZoneConstraintId.IsNull() {
		id := o.RoutingZoneConstraintId.ValueString()
		obj.RoutingZoneConstraintId = &id
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		diags.AddError("failed marshaling RoutingZoneConstraint primitive data", err.Error())
		return ""
	}

	data, err = json.Marshal(&tfCfgPrimitive{
		PrimitiveType: apstra.CtPrimitivePolicyTypeNameAttachRoutingZoneConstraint.String(),
		Label:         o.Label.ValueString(),
		Data:          data,
	})
	if err != nil {
		diags.AddError("failed marshaling primitive", err.Error())
		return ""
	}

	return string(data)
}

func (o *RoutingZoneConstraint) loadSdkPrimitive(ctx context.Context, in apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) {
	attributes, ok := in.Attributes.(*apstra.ConnectivityTemplatePrimitiveAttributesAttachRoutingZoneConstraint)
	if !ok {
		diags.AddError("failed loading SDK primitive due to wrong attribute type", fmt.Sprintf("unexpected type %T", in))
		return
	}

	o.loadSdkPrimitiveAttributes(ctx, attributes, diags)
	if diags.HasError() {
		return
	}
}

func (o *RoutingZoneConstraint) loadSdkPrimitiveAttributes(_ context.Context, in *apstra.ConnectivityTemplatePrimitiveAttributesAttachRoutingZoneConstraint, _ *diag.Diagnostics) {
	o.RoutingZoneConstraintId = types.StringNull()
	if in.RoutingZoneConstraint != nil {
		o.RoutingZoneConstraintId = types.StringValue(in.RoutingZoneConstraint.String())
	}
}

var _ JsonPrimitive = &routingZoneConstraintPrototype{}

type routingZoneConstraintPrototype struct {
	RoutingZoneConstraintId *string `json:"routing_zone_constraint_id"`
}

func (o routingZoneConstraintPrototype) attributes(_ context.Context, _ path.Path, _ *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes {
	rzcId := apstra.ObjectId(*o.RoutingZoneConstraintId)
	return &apstra.ConnectivityTemplatePrimitiveAttributesAttachRoutingZoneConstraint{
		RoutingZoneConstraint: &rzcId,
	}
}

func (o routingZoneConstraintPrototype) ToSdkPrimitive(ctx context.Context, path path.Path, diags *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive {
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

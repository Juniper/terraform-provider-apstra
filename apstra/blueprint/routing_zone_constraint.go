package blueprint

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterRoutingZoneConstraint struct {
	Id                         types.String `tfsdk:"id"`
	BlueprintId                types.String `tfsdk:"blueprint_id"`
	Name                       types.String `tfsdk:"name"`
	MaxCountConstraint         types.Int64  `tfsdk:"max_count_constraint"`
	RoutingZonesListConstraint types.String `tfsdk:"routing_zones_list_constraint"`
	Constraints                types.Set    `tfsdk:"constraints"`
}

func (o DatacenterRoutingZoneConstraint) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID. Required when `name` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI. Required when `id` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"max_count_constraint": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The maximum number of Routing Zones that the Application Point can be part of.",
			Computed:            true,
		},
		"routing_zones_list_constraint": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf(
				"Routing Zone constraint mode. One of: %s.", strings.Join(
					[]string{
						"`" + rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeAllow) + "`",
						"`" + rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeDeny) + "`",
						"`" + rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeNone) + "`",
					}, ", "),
			),
			Computed: true,
		},
		"constraints": dataSourceSchema.SetAttribute{
			MarkdownDescription: fmt.Sprintf("When `%s` instance constraint mode is chosen, only VNs from selected "+
				"Routing Zones are allowed to have endpoints on the interface(s) the policy is applied to. The permitted "+
				"Routing Zones may be specified directly or indirectly (via Routing Zone Groups)",
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeAllow),
			),
			Computed:    true,
			ElementType: types.StringType,
		},
	}
}

func (o DatacenterRoutingZoneConstraint) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI.",
			Optional:            true,
		},
		"max_count_constraint": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The maximum number of Routing Zones that the Application Point can be part of.",
			Optional:            true,
		},
		"routing_zones_list_constraint": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf(
				"Routing Zone constraint mode. One of: %s.", strings.Join(
					[]string{
						"`" + rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeAllow) + "`",
						"`" + rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeDeny) + "`",
						"`" + rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeNone) + "`",
					}, ", "),
			),
			Optional: true,
			Validators: []validator.String{stringvalidator.OneOf( // validated b/c this runs through rosetta
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeAllow),
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeDeny),
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeNone),
			)},
		},
		"constraints": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Routing Zone IDs. All Routing Zones supplied here are used to match the " +
				"Routing Zone Constraint, but a matching Routing Zone Constraintmay have additional Security Zones " +
				"not enumerated in this set.",
			Optional:    true,
			ElementType: types.StringType,
			Validators:  []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
	}
}

func (o DatacenterRoutingZoneConstraint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name displayed in the Apstra web UI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"max_count_constraint": resourceSchema.Int64Attribute{
			MarkdownDescription: "The maximum number of Routing Zones that the Application Point can be part of.",
			Optional:            true,
			Validators:          []validator.Int64{int64validator.Between(0, 255)},
		},
		"routing_zones_list_constraint": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Instance constraint mode.\n"+
				"- `%s` - only allow the specified routing zones (add specific routing zones to allow)\n"+
				"- `%s` - denies allocation of specified routing zones (add specific routing zones to deny)\n"+
				"- `%s` - no additional constraints on routing zones (any routing zones)",
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeAllow),
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeDeny),
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeNone),
			),
			Required: true,
			Validators: []validator.String{stringvalidator.OneOf(
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeAllow),
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeDeny),
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeNone),
			)},
		},
		"constraints": resourceSchema.SetAttribute{
			MarkdownDescription: fmt.Sprintf("When `%s` instance constraint mode is chosen, only VNs from selected "+
				"Routing Zones are allowed to have endpoints on the interface(s) the policy is applied to. The permitted "+
				"Routing Zones may be specified directly or indirectly (via Routing Zone Groups)",
				rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeAllow),
			),
			Optional:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				apstravalidator.ForbiddenWhenValueIs(
					path.MatchRoot("routing_zones_list_constraint"),
					types.StringValue(rosetta.StringersToFriendlyString(enum.RoutingZoneConstraintModeNone)),
				),
			},
		},
	}
}

func (o DatacenterRoutingZoneConstraint) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.RoutingZoneConstraintData {
	result := apstra.RoutingZoneConstraintData{
		Label: o.Name.ValueString(),
	}

	// set result.Mode
	err := rosetta.ApiStringerFromFriendlyString(&result.Mode, o.RoutingZonesListConstraint.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("failed converting %s to API type", o.RoutingZonesListConstraint), err.Error())
		return nil
	}

	// set result.MaxRoutingZones
	if !o.MaxCountConstraint.IsNull() {
		result.MaxRoutingZones = pointer.To(int(o.MaxCountConstraint.ValueInt64()))
	}

	// set result.RoutingZoneIds
	diags.Append(o.Constraints.ElementsAs(ctx, &result.RoutingZoneIds, false)...)

	return &result
}

func (o *DatacenterRoutingZoneConstraint) LoadApiData(ctx context.Context, in apstra.RoutingZoneConstraintData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(in.Label)
	if in.MaxRoutingZones == nil {
		o.MaxCountConstraint = types.Int64Null()
	} else {
		o.MaxCountConstraint = types.Int64Value(int64(*in.MaxRoutingZones))
	}
	o.RoutingZonesListConstraint = types.StringValue(in.Mode.String())
	o.Constraints = utils.SetValueOrNull(ctx, types.StringType, in.RoutingZoneIds, diags)
}

func (o DatacenterRoutingZoneConstraint) Query(ctx context.Context, rzcResultName string, diags *diag.Diagnostics) *apstra.MatchQuery {
	rzcNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal(rzcResultName)}
	nodeAttributes := []apstra.QEEAttribute{rzcNameAttr, apstra.NodeTypeRoutingZoneConstraint.QEEAttribute()}

	// add the name to the match, if any
	if !o.Name.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{Key: "label", Value: apstra.QEStringVal(o.Name.ValueString())})
	}

	// add the max to the match, if any
	if !o.MaxCountConstraint.IsNull() {
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{Key: "max_count_constraint", Value: apstra.QEIntVal(o.MaxCountConstraint.ValueInt64())})
	}

	// add the mode to the match, if any
	if !o.RoutingZonesListConstraint.IsNull() {
		var rzcm enum.RoutingZoneConstraintMode
		err := rosetta.ApiStringerFromFriendlyString(&rzcm, o.RoutingZonesListConstraint.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("failed converting %s to API type", o.RoutingZonesListConstraint), err.Error())
			return nil
		}
		nodeAttributes = append(nodeAttributes, apstra.QEEAttribute{Key: "routing_zones_list_constraint", Value: apstra.QEStringVal(rzcm.String())})
	}

	query := new(apstra.MatchQuery).Match(new(apstra.PathQuery).Node(nodeAttributes))

	var rzIds []string
	diags.Append(o.Constraints.ElementsAs(ctx, &rzIds, false)...)
	if diags.HasError() {
		return nil
	}

	for _, rzId := range rzIds {
		query.Match(new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{rzcNameAttr}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeConstraint.QEEAttribute()}).
			Node([]apstra.QEEAttribute{apstra.NodeTypeSecurityZone.QEEAttribute(), {Key: "id", Value: apstra.QEStringVal(rzId)}}))
	}

	return query
}

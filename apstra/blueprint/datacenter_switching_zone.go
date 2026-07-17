package blueprint

import (
	"context"
	"fmt"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/datacenter"
	"github.com/Juniper/apstra-go-sdk/enum"
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
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

type DatacenterSwitchingZone struct {
	ID          types.String            `tfsdk:"id"`
	BlueprintID types.String            `tfsdk:"blueprint_id"`
	Name        types.String            `tfsdk:"name"`
	MACVRFName  types.String            `tfsdk:"mac_vrf_name"`
	Description types.String            `tfsdk:"description"`
	ServiceType types.String            `tfsdk:"service_type"`
	RouteTarget customtypes.RouteTarget `tfsdk:"route_target"`
	Tags        types.Set               `tfsdk:"tags"`
}

func (o DatacenterSwitchingZone) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID. Required when `name` and `mac_vrf_name` are omitted.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
					path.MatchRoot("mac_vrf_name"),
				}...),
			},
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Switching Zone Name. Required when `id` and `mac_vrf_name` are omitted.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"mac_vrf_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "MAC VRF name used on network devices and visible in the web UI. Required when `id` and `name` are omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Description displayed in the web UI.",
			Computed:            true,
		},
		"service_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "MAC VRF Service Type",
			Computed:            true,
		},
		"route_target": dataSourceSchema.StringAttribute{
			CustomType:          customtypes.RouteTargetType{},
			MarkdownDescription: "Route Target unique to the MAC-VRF Switching Zone.",
			Computed:            true,
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags associated with the Switching Zone.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (o DatacenterSwitchingZone) DataSourceFilterAttributes() map[string]dataSourceSchema.Attribute {
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
			MarkdownDescription: "Switching Zone Name.",
			Optional:            true,
		},
		"mac_vrf_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "MAC VRF name used on network devices and visible in the web UI.",
			Optional:            true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Description displayed in the web UI.",
			Optional:            true,
		},
		"service_type": dataSourceSchema.StringAttribute{
			MarkdownDescription: "MAC VRF Service Type",
			Optional:            true,
		},
		"route_target": dataSourceSchema.StringAttribute{
			CustomType:          customtypes.RouteTargetType{},
			MarkdownDescription: "Route Target unique to the MAC-VRF Switching Zone.",
			Optional:            true,
		},
		"tags": dataSourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags. All tags supplied here are used to match the Switching Zone, " +
				"but a matching Switching Zone may have additional tags not enumerated in this set.",
			Optional:    true,
			ElementType: types.StringType,
			Validators:  []validator.Set{setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))},
		},
	}
}

func (o DatacenterSwitchingZone) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Switching Zone Name.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"mac_vrf_name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name of the MAC-VRF instance as rendered in systems. Must be unique across all " +
				"MAC-VRFs. Reserved values are \"default\" for the default switch instance and \"evpn-1\" for the " +
				"MAC-VRF associated with EVPN routing zones by default.",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 15),
				stringvalidator.RegexMatches(apstraregexp.AlphaNumW2HLConstraint, apstraregexp.AlphaNumW2HLConstraintMsg),
				stringvalidator.NoneOf("default", "evpn-1"),
			},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Description displayed in the web UI.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 240),
				stringvalidator.RegexMatches(apstraregexp.RenderableDescription, apstraregexp.RenderableDescriptionMsg),
			},
		},
		"service_type": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Type of VLAN service that the MAC-VRF routing instance provides. "+
				"The selected option controls how the routing instance maps VLANs to the forwarding instances on the "+
				"device. VLAN bundle service allows multiple broadcast domains to map to a single bridge domain. "+
				"Multiple VLANs are mapped to a single EVPN instance (EVI) and share the same bridge table in the "+
				"MAC-VRF table, thus reducing the number of routes and labels stored in the table. The vlan-bundle "+
				"service-type does not support VLAN normalization at the egress leaf. VLAN-aware service allows "+
				"multiple VLANs to be mapped to a single EVI. Each VLAN has a different bridge table. This is default "+
				"service type that can be used for both enterprise- and SP-style configurations. Must be one of [`%s`]",
				strings.Join(enum.SwitchingZoneMACVRFServiceTypes.Values(), "`,`")),
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:    []validator.String{stringvalidator.OneOf(enum.SwitchingZoneMACVRFServiceTypes.Values()...)},
		},
		"route_target": resourceSchema.StringAttribute{
			CustomType: customtypes.RouteTargetType{},
			MarkdownDescription: "Route Target unique to the MAC-VRF switching zone. Value will be auto-generated if " +
				"omitted. Examples: `64496:100`, `192.0.2.1:200`",
			Optional: true,
			Computed: true,
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Tags applied to the Switching Zone.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *DatacenterSwitchingZone) Request(ctx context.Context, diags *diag.Diagnostics) datacenter.SwitchingZone {
	result := datacenter.SwitchingZone{
		Label:             o.Name.ValueStringPointer(),
		MACVRFDescription: o.Description.ValueStringPointer(),
		MACVRFName:        o.MACVRFName.ValueStringPointer(),
		// MACVRFServiceType: nil, // handled below
		// RouteTarget:       nil, // handled below
		// Tags:              nil, // handled below
	}

	_ = result.SetID(o.ID.ValueString()) // Ignoring error b/c we know id is empty. Also, setting empty ID in Create() is okay.

	if utils.HasValue(o.ServiceType) {
		result.MACVRFServiceType = new(enum.SwitchingZoneMACVRFServiceType)
		err := result.MACVRFServiceType.FromString(o.ServiceType.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("service_type"), "parsing service_type", err.Error())
		}
	}

	if utils.HasValue(o.RouteTarget) {
		result.RouteTarget = new(datacenter.RouteTarget)
		err := result.RouteTarget.UnmarshalText([]byte(o.RouteTarget.ValueString()))
		if err != nil {
			diags.AddAttributeError(path.Root("route_target"), "parsing route_target value: "+o.RouteTarget.String(), err.Error())
		}
	}

	if o.RouteTarget.IsUnknown() { // we need to clear the value from the API by sending `null` JSON payload.
		result.RouteTarget = new(datacenter.RouteTarget)
		err := result.RouteTarget.UnmarshalText([]byte("null"))
		if err != nil {
			diags.AddAttributeError(path.Root("route_target"), `preparing null route_target value`, err.Error())
		}
	}

	diags.Append(o.Tags.ElementsAs(ctx, &result.Tags, false)...)

	return result
}

func (o *DatacenterSwitchingZone) LoadApiData(ctx context.Context, sz datacenter.SwitchingZone, diags *diag.Diagnostics) {
	o.ID = types.StringPointerValue(sz.ID())
	o.Name = types.StringPointerValue(sz.Label)
	o.MACVRFName = types.StringPointerValue(sz.MACVRFName)
	o.Description = types.StringPointerValue(sz.MACVRFDescription)

	o.ServiceType = types.StringNull()
	if sz.MACVRFServiceType != nil {
		o.ServiceType = types.StringValue(sz.MACVRFServiceType.String())
	}

	o.RouteTarget = customtypes.NewRouteTargetNull()
	if sz.RouteTarget != nil {
		b, err := sz.RouteTarget.MarshalText()
		if err != nil {
			diags.AddAttributeError(path.Root("route_target"), "marshaling route target returned by API", err.Error())
		}
		o.RouteTarget = customtypes.NewRouteTargetValue(string(b))
	}

	o.Tags = value.SetOrNull(ctx, types.StringType, sz.Tags, diags)
}

func (o *DatacenterSwitchingZone) Query(szResultName string) *apstra.MatchQuery {
	matchQuery := new(apstra.MatchQuery)
	matchQuery.Match(new(apstra.PathQuery).Node(o.szNodeQueryAttributes(szResultName)))

	for _, tag := range o.Tags.Elements() {
		tagQuery := new(apstra.PathQuery).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSwitchingZone.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal(szResultName)},
			}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeTag.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeTag.QEEAttribute(),
				{Key: "label", Value: apstra.QEStringVal(tag.(types.String).ValueString())},
			})

		matchQuery.Match(tagQuery)
	}

	return matchQuery
}

func (o *DatacenterSwitchingZone) szNodeQueryAttributes(name string) []apstra.QEEAttribute {
	result := []apstra.QEEAttribute{
		{Key: "type", Value: apstra.QEStringVal(apstra.NodeTypeSwitchingZone.String())},
	}

	if name != "" {
		result = append(result, apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal(name)})
	}

	if utils.HasValue(o.Name) {
		result = append(result, apstra.QEEAttribute{Key: "label", Value: apstra.QEStringVal(o.Name.ValueString())})
	}

	if utils.HasValue(o.MACVRFName) {
		result = append(result, apstra.QEEAttribute{Key: "mac_vrf_name", Value: apstra.QEStringVal(o.MACVRFName.ValueString())})
	}

	if utils.HasValue(o.Description) {
		result = append(result, apstra.QEEAttribute{Key: "mac_vrf_description", Value: apstra.QEStringVal(o.Description.ValueString())})
	}

	if utils.HasValue(o.ServiceType) {
		result = append(result, apstra.QEEAttribute{Key: "mac_vrf_service_type", Value: apstra.QEStringVal(o.ServiceType.ValueString())})
	}

	if utils.HasValue(o.RouteTarget) {
		result = append(result, apstra.QEEAttribute{Key: "route_target", Value: apstra.QEStringVal(o.RouteTarget.ValueString())})
	}

	return result
}

// ReadComputedValues exists to populate unknown computed values (and only those values) during Create and Update.
// It does this by calling the API and then filling in unknown values.
func (o *DatacenterSwitchingZone) ReadComputedValues(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	if !o.RouteTarget.IsUnknown() {
		return // all possibly unknown values are known - we have nothing to do
	}

	// GET from the API
	sz, err := bp.GetSwitchingZone(ctx, o.ID.ValueString())
	if err != nil {
		diags.AddError("failed while reading Routing Zone from API", err.Error())
		return
	}

	// load API response into temporary object
	var temp DatacenterSwitchingZone
	temp.LoadApiData(ctx, sz, diags)

	// populate computed values
	if o.RouteTarget.IsUnknown() {
		o.RouteTarget = temp.RouteTarget
	}
}

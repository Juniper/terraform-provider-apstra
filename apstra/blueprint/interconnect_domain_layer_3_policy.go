package blueprint

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	customtypes "github.com/Juniper/terraform-provider-apstra/apstra/custom_types"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InterconnectDomainL3Policy struct {
	BlueprintID          types.String            `tfsdk:"blueprint_id"`
	InterconnectDomainID types.String            `tfsdk:"interconnect_domain_id"`
	RoutingZoneID        types.String            `tfsdk:"routing_zone_id"`
	EnabledForType5      types.Bool              `tfsdk:"enabled_for_type_5"`
	RoutingPolicyID      types.String            `tfsdk:"routing_policy_id"`
	RouteTarget          customtypes.RouteTarget `tfsdk:"route_target"`
}

func (l3p InterconnectDomainL3Policy) DatasourceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"blueprint_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"interconnect_domain_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Interconnect Domain.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"routing_zone_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Routing Zone.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"enabled_for_type_5": datasourceSchema.BoolAttribute{
			MarkdownDescription: "To prevent unintended route advertisement, advertising Layer-3 networks " +
				"is a two step process. This boolean allows the Routing Zone to advertise EVPN Type-5 routes. " +
				"You must still select the Virtual Networks you wish to individually advertise.",
			Computed: true,
		},
		"routing_policy_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Select the routing policy to be applied to DCI for this Routing Zone (VRF).",
			Computed:            true,
		},
		"route_target": datasourceSchema.StringAttribute{
			MarkdownDescription: "All Interconnect Gateways MUST use the same Interconnect Route Target (iRT). " +
				"The iRT is an additional unique RT for the Interconnect Domain.",
			CustomType: customtypes.RouteTargetType{},
			Computed:   true,
		},
	}
}

func (l3p InterconnectDomainL3Policy) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
		},
		"interconnect_domain_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Interconnect Domain.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Routing Zone.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
		},
		"enabled_for_type_5": resourceSchema.BoolAttribute{
			MarkdownDescription: "To prevent unintended route advertisement, advertising Layer-3 networks " +
				"is a two step process. This boolean allows the Routing Zone to advertise EVPN Type-5 routes. " +
				"You must still select the Virtual Networks you wish to individually advertise.",
			Required: true,
		},
		"routing_policy_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Select the routing policy to be applied to DCI for this Routing Zone (VRF).",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"route_target": resourceSchema.StringAttribute{
			MarkdownDescription: "All Interconnect Gateways MUST use the same Interconnect Route Target (iRT). " +
				"The iRT is an additional unique RT for the Interconnect Domain.",
			CustomType: customtypes.RouteTargetType{},
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (l3p InterconnectDomainL3Policy) Request(_ context.Context, _ *diag.Diagnostics) apstra.EVPNInterconnectGroup {
	result := apstra.EVPNInterconnectGroup{
		InterconnectSecurityZones: map[string]apstra.InterconnectSecurityZone{
			l3p.RoutingZoneID.ValueString(): {
				L3Enabled:       l3p.EnabledForType5.ValueBool(),
				RouteTarget:     l3p.RouteTarget.ValueStringPointer(),
				RoutingPolicyId: l3p.RoutingPolicyID.ValueStringPointer(),
			},
		},
	}
	_ = result.SetID(l3p.InterconnectDomainID.ValueString())
	return result
}

func (l3p *InterconnectDomainL3Policy) Read(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	g, err := bp.GetEVPNInterconnectGroup(ctx, l3p.InterconnectDomainID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			diags.AddError("Not Found", fmt.Sprintf("Interconnect Domain %s of Blueprint %s not found", l3p.InterconnectDomainID, l3p.BlueprintID))
			return
		}
		diags.AddError("Failed to fetch Interconnect Domain", err.Error())
		return
	}

	if l3Policy, ok := g.InterconnectSecurityZones[l3p.RoutingZoneID.ValueString()]; ok {
		l3p.EnabledForType5 = types.BoolValue(l3Policy.L3Enabled)
		l3p.RoutingPolicyID = types.StringPointerValue(l3Policy.RoutingPolicyId)
		l3p.RouteTarget = customtypes.NewRouteTargetPointerValue(l3Policy.RouteTarget)
	} else {
		diags.AddError("Layer 3 Policy not found", fmt.Sprintf("Interconnect Domain %s of Blueprint %s does not have Interconnect Domain policy settings for Routing Zone %s", l3p.InterconnectDomainID, l3p.BlueprintID, l3p.RoutingZoneID))
	}
}

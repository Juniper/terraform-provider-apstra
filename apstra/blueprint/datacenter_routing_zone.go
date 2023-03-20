package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"terraform-provider-apstra/apstra/design"
	"terraform-provider-apstra/apstra/resources"
)

type DatacenterRoutingZone struct {
	Id              types.String `tfsdk:"id"`
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	Name            types.String `tfsdk:"name"`
	VlanId          types.Int64  `tfsdk:"vlan_id"`
	VniId           types.Int64  `tfsdk:"vni_id"`
	RoutingPolicyId types.String `tfsdk:"routing_policy_id"`
}

func (o DatacenterRoutingZone) ResourceAttributes() map[string]resourceSchema.Attribute {
	nameRE, _ := regexp.Compile("^[A-Za-z0-9_-]+$")
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
			MarkdownDescription: "VRF name displayed in thw Apstra web UI.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(nameRE, "only underscore, dash and alphanumeric characters allowed."),
				stringvalidator.LengthBetween(1, 15),
			},
		},
		"vlan_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "Used for VLAN tagged Layer 3 links on external connections. " +
				"Leave this field blank to have it automatically assigned from a static pool in the " +
				"range of 2-4094), or enter a specific value.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.Int64{int64validator.Between(design.VlanMin-1, design.VlanMax+1)},
		},
		"vni_id": resourceSchema.Int64Attribute{
			MarkdownDescription: "VxLAN VNI associated with the routing zone. Leave this field blank to have it " +
				"automatically assigned from an allocated resource pool, or enter a specific value.",
			Optional:   true,
			Validators: []validator.Int64{int64validator.Between(resources.VniMin-1, resources.VniMax+1)},
		},
		"routing_policy_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Non-EVPN blueprints must use the default policy, so this field must be null. " +
				"Set this attribute in an EVPN blueprint to use a non-default policy.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (o *DatacenterRoutingZone) Request(_ context.Context, _ *diag.Diagnostics) *goapstra.SecurityZoneData {
	var vlan *goapstra.Vlan
	if !o.VlanId.IsNull() && !o.VlanId.IsUnknown() {
		v := goapstra.Vlan(o.VlanId.ValueInt64())
		vlan = &v
	}

	var vni *int
	if !o.VniId.IsNull() && !o.VniId.IsUnknown() {
		v := int(o.VniId.ValueInt64())
		vni = &v
	}

	// todo:
	//var routingPolicyId goapstra.ObjectId
	//if o.RoutingPolicyId.IsNull() {
	//	routingPolicyId = client.GetDefaultRoutingZone()
	//} else {
	//	routingPolicyId = o.RoutingPolicyId.ValueString()
	//}

	return &goapstra.SecurityZoneData{
		SzType:          goapstra.SecurityZoneTypeEVPN,
		VrfName:         o.Name.ValueString(),
		Label:           o.Name.ValueString(),
		RoutingPolicyId: goapstra.ObjectId(o.RoutingPolicyId.ValueString()),
		VlanId:          vlan,
		VniId:           vni,
	}
}

func (o *DatacenterRoutingZone) LoadApiData(ctx context.Context, sz *goapstra.SecurityZoneData, diags *diag.Diagnostics) {
	o.Name = types.StringValue(sz.VrfName)
	o.VlanId = types.Int64Value(int64(*sz.VlanId))

	if sz.RoutingPolicyId != "" {
		o.RoutingPolicyId = types.StringValue(sz.RoutingPolicyId.String())
	} else {
		o.RoutingPolicyId = types.StringNull()
	}

	if sz.VniId != nil {
		o.VniId = types.Int64Value(int64(*sz.VniId))
	} else {
		o.VniId = types.Int64Null()
	}
}

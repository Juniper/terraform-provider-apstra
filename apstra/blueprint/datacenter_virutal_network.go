package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/design"
)

type DatacenterVirtualNetwork struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	BlueprintId   types.String `tfsdk:"blueprint_id"`
	RoutingZoneId types.String `tfsdk:"routing_zone_id"`
	Type          types.String `tfsdk:"type"`
	LeafSwitchIds types.List   `tfsdk:"leaf_switch_ids"`
}

func (o DatacenterVirtualNetwork) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Name",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 30),
				stringvalidator.RegexMatches(regexp.MustCompile(design.AlphaNumericRegexp), "valid characters are: "+design.AlphaNumericChars),
			},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Type",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(apstra.VnTypeVxlan.String()),
			Validators:          []validator.String{apstravalidator.OneOfStringers(apstra.AllVirtualNetworkTypes())},
		},
		"routing_zone_id": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Routing Zone ID (required when `type == %s`", apstra.VnTypeVxlan),
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				apstravalidator.StringRequiredWhenValueIs(path.MatchRelative().AtParent().AtName("type"), fmt.Sprintf("%q", apstra.VnTypeVxlan)),
			},
		},
		"leaf_switch_ids": resourceSchema.ListAttribute{
			MarkdownDescription: "Graph DB node IDs of Leaf Switches to which this Virtual Network should be bound",
			Required:            true, // todo: can become optional when access_switch_ids added
			ElementType:         types.StringType,
			Validators: []validator.List{
				//listvalidator.AtLeastOneOf(),
				listvalidator.SizeAtLeast(1),
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}

func (o *DatacenterVirtualNetwork) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.VirtualNetworkData {
	var err error

	var vnType apstra.VnType
	err = vnType.FromString(o.Type.ValueString())
	if err != nil {
		diags.Append(
			validatordiag.BugInProviderDiagnostic(
				fmt.Sprintf("error parsing virtual network type %q - %s", o.Type.String(), err.Error())))
		return nil
	}

	var leafSwitchNodeIds []apstra.ObjectId
	o.LeafSwitchIds.ElementsAs(ctx, &leafSwitchNodeIds, false)

	vnBindings := make([]apstra.VnBinding, len(leafSwitchNodeIds))
	for i := range leafSwitchNodeIds {
		vnBindings[i].SystemId = leafSwitchNodeIds[i]
	}

	return &apstra.VirtualNetworkData{
		DhcpService:               false,
		Ipv4Enabled:               false,
		Ipv4Subnet:                nil,
		Ipv6Enabled:               false,
		Ipv6Subnet:                nil,
		Label:                     o.Name.ValueString(),
		ReservedVlanId:            nil,
		RouteTarget:               "",
		RtPolicy:                  nil,
		SecurityZoneId:            apstra.ObjectId(o.RoutingZoneId.ValueString()),
		SviIps:                    nil,
		VirtualGatewayIpv4:        nil,
		VirtualGatewayIpv6:        nil,
		VirtualGatewayIpv4Enabled: false,
		VirtualGatewayIpv6Enabled: false,
		VnBindings:                vnBindings,
		VnId:                      nil,
		VnType:                    vnType,
		VirtualMac:                nil,
	}
}

func (o *DatacenterVirtualNetwork) LoadApiData(_ context.Context, in *apstra.VirtualNetworkData, _ *diag.Diagnostics) {
	leafSwitchIds := make([]attr.Value, len(in.VnBindings))
	for i, vnBinding := range in.VnBindings {
		leafSwitchIds[i] = types.StringValue(vnBinding.SystemId.String())
	}

	o.Name = types.StringValue(in.Label)
	o.Type = types.StringValue(in.VnType.String())
	o.RoutingZoneId = types.StringValue(in.SecurityZoneId.String())
	o.LeafSwitchIds = types.ListValueMust(types.StringType, leafSwitchIds)
}

//func (o DatacenterVirtualNetwork) DataSourceAttributes() map[string]datasourceSchema.Attribute {
//	return map[string]datasourceSchema.Attribute{
//		"id": datasourceSchema.StringAttribute{
//			MarkdownDescription: "Apstra graph node ID.",
//			Computed:            true,
//		},
//		"name": datasourceSchema.StringAttribute{
//			MarkdownDescription: "Virtual Network Name",
//			Required:            true,
//			Validators:          []validator.String{stringvalidator.LengthBetween(1, 30)},
//		},
//		"type": datasourceSchema.StringAttribute{
//			MarkdownDescription: "Virtual Network Type",
//			Optional:            true,
//			Computed:            true,
//			Validators:          []validator.String{apstravalidator.OneOfStringers(apstra.AllNodeDeployModes())},
//		},
//		"blueprint_id": datasourceSchema.StringAttribute{},
//	}
//}

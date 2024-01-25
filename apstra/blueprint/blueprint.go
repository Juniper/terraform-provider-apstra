package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apstraplanmodifier "github.com/Juniper/terraform-provider-apstra/apstra/apstra_plan_modifier"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type Blueprint struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	TemplateId            types.String `tfsdk:"template_id"`
	FabricAddressing      types.String `tfsdk:"fabric_addressing"`
	Status                types.String `tfsdk:"status"`
	SuperspineCount       types.Int64  `tfsdk:"superspine_count"`
	SpineCount            types.Int64  `tfsdk:"spine_count"`
	LeafCount             types.Int64  `tfsdk:"leaf_switch_count"`
	AccessCount           types.Int64  `tfsdk:"access_switch_count"`
	GenericCount          types.Int64  `tfsdk:"generic_system_count"`
	ExternalCount         types.Int64  `tfsdk:"external_router_count"`
	HasUncommittedChanges types.Bool   `tfsdk:"has_uncommitted_changes"`
	Version               types.Int64  `tfsdk:"version"`
	BuildWarningsCount    types.Int64  `tfsdk:"build_warnings_count"`
	BuildErrorsCount      types.Int64  `tfsdk:"build_errors_count"`
	EsiMacMsb             types.Int64  `tfsdk:"esi_mac_msb"`
	Ipv6Applications      types.Bool   `tfsdk:"ipv6_applications"`
	FabricMtu             types.Int64  `tfsdk:"fabric_mtu"`
}

func (o Blueprint) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Blueprint. Required when `name` is omitted.",
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
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name of the Blueprint. Required when `id` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"template_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Template ID will always be null in 'data source' context.",
			Computed:            true,
		},
		"fabric_addressing": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Addressing scheme for both superspine/spine and spine/leaf links.",
			Computed:            true,
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status of the Blueprint",
			Computed:            true,
		},
		"superspine_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "For 5-stage topologies, the count of superspine devices",
			Computed:            true,
		},
		"spine_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of spine devices in the topology.",
			Computed:            true,
		},
		"leaf_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of leaf switches in the topology.",
			Computed:            true,
		},
		"access_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of access switches in the topology.",
			Computed:            true,
		},
		"generic_system_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of generic systems in the topology.",
			Computed:            true,
		},
		"external_router_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of external routers attached to the topology.",
			Computed:            true,
		},
		"has_uncommitted_changes": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the staging blueprint has uncommitted changes.",
			Computed:            true,
		},
		"version": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Currently active blueprint version",
			Computed:            true,
		},
		"build_warnings_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build warnings.",
			Computed:            true,
		},
		"build_errors_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build errors.",
			Computed:            true,
		},
		"esi_mac_msb": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "ESI MAC address most significant byte.",
			Computed:            true,
		},
		"ipv6_applications": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Enables support for IPv6 virtual networks and IPv6 external " +
				"connectivity points. This adds resource requirements and device configurations, " +
				"including IPv6 loopback addresses on leafs, spines and superspines, IPv6 addresses " +
				"for MLAG SVI subnets and IPv6 addresses for leaf L3 peer links.",
			Computed: true,
		},
		"fabric_mtu": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "MTU of fabric links. Requires Apstra 4.2 or later.",
			Computed:            true,
		},
	}
}

func (o Blueprint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint ID assigned by Apstra.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Blueprint name.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"template_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of Rack Based Template used to instantiate the Blueprint.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"fabric_addressing": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Addressing scheme for both superspine/spine and spine/leaf links. Only "+
				"applicable to Apstra versions 4.1.1 and later. Must be one of: %s",
				strings.Join([]string{
					apstra.AddressingSchemeIp4.String(),
					apstra.AddressingSchemeIp6.String(),
					apstra.AddressingSchemeIp46.String(),
				}, ", ")),
			Optional: true,
			Computed: true,
			Validators: []validator.String{stringvalidator.OneOf(
				apstra.AddressingSchemeIp4.String(),
				apstra.AddressingSchemeIp6.String(),
				apstra.AddressingSchemeIp46.String())},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplace(),
			},
		},
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status of the Blueprint",
			Computed:            true,
		},
		"superspine_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "For 5-stage topologies, the count of superspine devices",
			Computed:            true,
		},
		"spine_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of spine devices in the topology.",
			Computed:            true,
		},
		"leaf_switch_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of leaf switches in the topology.",
			Computed:            true,
		},
		"access_switch_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of access switches in the topology.",
			Computed:            true,
		},
		"generic_system_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of generic systems in the topology.",
			Computed:            true,
		},
		"external_router_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of external routers attached to the topology.",
			Computed:            true,
		},
		"has_uncommitted_changes": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether the staging blueprint has uncommitted changes.",
			Computed:            true,
		},
		"version": resourceSchema.Int64Attribute{
			MarkdownDescription: "Currently active blueprint version",
			Computed:            true,
		},
		"build_warnings_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build warnings.",
			Computed:            true,
		},
		"build_errors_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of build errors.",
			Computed:            true,
		},
		"esi_mac_msb": resourceSchema.Int64Attribute{
			MarkdownDescription: "ESI MAC address most significant byte. Must be an even number " +
				"between 0 and 254 inclusive.",
			Optional: true,
			Computed: true,
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
				int64validator.AtMost(254),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
		"ipv6_applications": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables support for IPv6 virtual networks and IPv6 external " +
				"connectivity points. This adds resource requirements and device configurations, " +
				"including IPv6 loopback addresses on leafs, spines and superspines, IPv6 addresses " +
				"for MLAG SVI subnets and IPv6 addresses for leaf L3 peer links. This option cannot " +
				"be disabled without re-creating the Blueprint.",
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
				boolplanmodifier.RequiresReplaceIf(
					apstraplanmodifier.BoolRequiresReplaceWhenSwitchingTo(false),
					"Switching from \"false\" to \"true\" requires the Blueprint to be replaced",
					"Switching from `false` to `true` requires the Blueprint to be replaced",
				),
			},
		},
		"fabric_mtu": resourceSchema.Int64Attribute{
			MarkdownDescription: "MTU of fabric links. Must be an even number between 1280 and 9216. Requires Apstra 4.2 or later.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				int64validator.Between(1280, 9216),
				apstravalidator.MustBeEvenOrOdd(true),
			},
		},
	}
}

func (o Blueprint) Request(_ context.Context, diags *diag.Diagnostics) *apstra.CreateBlueprintFromTemplateRequest {
	// start with a nil pointer for fabric addressing policy
	var fabricAddressingPolicy *apstra.BlueprintRequestFabricAddressingPolicy

	// if the user supplied either value, we must create the fabric addressing policy object
	if !o.FabricAddressing.IsUnknown() || !o.FabricMtu.IsUnknown() {
		fabricAddressingPolicy = &apstra.BlueprintRequestFabricAddressingPolicy{
			SpineSuperspineLinks: apstra.AddressingSchemeIp4, // sensible default
			SpineLeafLinks:       apstra.AddressingSchemeIp4, // sensible default
		}

		if !o.FabricAddressing.IsNull() {
			var fabricAddressing apstra.AddressingScheme
			err := fabricAddressing.FromString(o.FabricAddressing.ValueString())
			if err != nil {
				diags.AddError(
					constants.ErrProviderBug,
					fmt.Sprintf("error parsing fabric_addressing %q - %s",
						o.FabricAddressing.ValueString(), err.Error()))
				return nil
			}
			fabricAddressingPolicy.SpineSuperspineLinks = fabricAddressing
			fabricAddressingPolicy.SpineLeafLinks = fabricAddressing
		}

		if utils.Known(o.FabricMtu) {
			fabricMtu := uint16(o.FabricMtu.ValueInt64())
			fabricAddressingPolicy.FabricL3Mtu = &fabricMtu
		}
	}

	return &apstra.CreateBlueprintFromTemplateRequest{
		RefDesign:              apstra.RefDesignTwoStageL3Clos,
		Label:                  o.Name.ValueString(),
		TemplateId:             apstra.ObjectId(o.TemplateId.ValueString()),
		FabricAddressingPolicy: fabricAddressingPolicy,
	}
}

func (o Blueprint) FabricAddressingRequest(_ context.Context, _ *diag.Diagnostics) *apstra.TwoStageL3ClosFabricAddressingPolicy {
	var result apstra.TwoStageL3ClosFabricAddressingPolicy

	if utils.Known(o.Ipv6Applications) {
		ipv6Enabled := o.Ipv6Applications.ValueBool()
		result.Ipv6Enabled = &ipv6Enabled
	}

	if utils.Known(o.EsiMacMsb) {
		esiMacMsb := uint8(o.EsiMacMsb.ValueInt64())
		result.EsiMacMsb = &esiMacMsb
	}

	if utils.Known(o.FabricMtu) {
		fabricMtu := uint16(o.FabricMtu.ValueInt64())
		result.FabricL3Mtu = &fabricMtu
	}

	return &result
}

func (o *Blueprint) LoadApiData(_ context.Context, in *apstra.BlueprintStatus, _ *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Label)
	o.Status = types.StringValue(in.Status)
	o.SuperspineCount = types.Int64Value(int64(in.SuperspineCount))
	o.SpineCount = types.Int64Value(int64(in.SpineCount))
	o.LeafCount = types.Int64Value(int64(in.LeafCount))
	o.AccessCount = types.Int64Value(int64(in.AccessCount))
	o.GenericCount = types.Int64Value(int64(in.GenericCount))
	o.ExternalCount = types.Int64Value(int64(in.ExternalRouterCount))
	o.HasUncommittedChanges = types.BoolValue(in.HasUncommittedChanges)
	o.Version = types.Int64Value(int64(in.Version))
	o.BuildErrorsCount = types.Int64Value(int64(in.BuildErrorsCount))
	o.BuildWarningsCount = types.Int64Value(int64(in.BuildWarningsCount))
}

func (o *Blueprint) LoadFabricAddressingPolicy(_ context.Context, in *apstra.TwoStageL3ClosFabricAddressingPolicy, _ *diag.Diagnostics) {
	o.EsiMacMsb = types.Int64Null()
	if in.EsiMacMsb != nil {
		o.EsiMacMsb = types.Int64Value(int64(*in.EsiMacMsb))
	}

	o.FabricMtu = types.Int64Null()
	if in.FabricL3Mtu != nil {
		o.FabricMtu = types.Int64Value(int64(*in.FabricL3Mtu))
	}

	o.Ipv6Applications = types.BoolValue(false)
	if in.Ipv6Enabled != nil {
		o.Ipv6Applications = types.BoolValue(*in.Ipv6Enabled)
	}
}

func (o *Blueprint) SetName(ctx context.Context, bpClient *apstra.TwoStageL3ClosClient, state *Blueprint, diags *diag.Diagnostics) {
	if o.Name.Equal(state.Name) {
		// nothing to do
		return
	}

	type node struct {
		Label string          `json:"label,omitempty"`
		Id    apstra.ObjectId `json:"id,omitempty"`
	}
	response := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}

	err := bpClient.GetNodes(ctx, apstra.NodeTypeMetadata, response)
	if err != nil {
		diags.AddError(
			fmt.Sprintf(errApiGetWithTypeAndId, "Blueprint Node", bpClient.Id()),
			err.Error(),
		)
		return
	}
	if len(response.Nodes) != 1 {
		diags.AddError(fmt.Sprintf("wrong number of %s nodes", apstra.NodeTypeMetadata.String()),
			fmt.Sprintf("expecting 1 got %d nodes", len(response.Nodes)))
		return
	}
	var nodeId apstra.ObjectId
	for _, v := range response.Nodes {
		nodeId = v.Id
	}
	err = bpClient.PatchNode(ctx, nodeId, &node{Label: o.Name.ValueString()}, nil)
	if err != nil {
		diags.AddError(
			fmt.Sprintf(errApiPatchWithTypeAndId, bpClient.Id(), nodeId),
			err.Error(),
		)
		return
	}
}

func (o *Blueprint) SetFabricAddressingPolicy(ctx context.Context, bpClient *apstra.TwoStageL3ClosClient, state *Blueprint, diags *diag.Diagnostics) {
	switch {
	case utils.Known(o.EsiMacMsb): // we have a value; do not return in default action
	case utils.Known(o.FabricMtu): // we have a value; do not return in default action
	case utils.Known(o.Ipv6Applications): // we have a value; do not return in default action
	default:
		return // no relevant values set in the plan
	}

	if state != nil {
		switch {
		case utils.Known(o.EsiMacMsb) && !o.EsiMacMsb.Equal(state.EsiMacMsb): // plan and state not in agreement
		case utils.Known(o.FabricMtu) && !o.FabricMtu.Equal(state.FabricMtu): // plan and state not in agreement
		case utils.Known(o.Ipv6Applications) && !o.Ipv6Applications.Equal(state.Ipv6Applications): // plan and state not in agreement
		default:
			return // no plan values represent changes from the current state
		}
	}

	fapRequest := o.FabricAddressingRequest(ctx, diags)
	if diags.HasError() {
		return
	}

	if fapRequest == nil {
		// nothing to do
		return
	}

	err := bpClient.SetFabricAddressingPolicy(ctx, fapRequest)
	if err != nil {
		diags.AddError("failed setting blueprint fabric addressing policy", err.Error())
		return
	}
}

func (o *Blueprint) MinMaxApiVersions(_ context.Context, diags *diag.Diagnostics) (*version.Version, *version.Version) {
	var min, max *version.Version
	var err error
	if !o.FabricAddressing.IsNull() {
		min, err = version.NewVersion("4.1.1")
	}
	if err != nil {
		diags.AddError(constants.ErrProviderBug,
			fmt.Sprintf("error parsing min/max version - %s", err.Error()))
	}

	return min, max
}

func (o *Blueprint) CheckCompatibility(_ context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	if compatibility.FabricL3MtuForbiddenInRequest(client.ApiVersion()) && !o.FabricMtu.IsUnknown() {
		diags.AddAttributeError(
			path.Root("fabric_mtu"),
			constants.ErrApiCompatibility,
			"`fabric_mtu` requires Apstra 4.2.0 or later",
		)
	}
}

func (o *Blueprint) GetFabricLinkIpVersion(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	query := new(apstra.PathQuery).
		SetClient(bp.Client()).
		SetBlueprintId(bp.Id()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeFabricAddressingPolicy.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_fabric_addressing_policy")},
		})

	var result struct {
		Items []struct {
			FabricAddressingPolicy struct {
				SpineLeafLinks       string `json:"spine_leaf_links"`
				SpineSuperspineLinks string `json:"spine_superspine_links"`
			} `json:"n_fabric_addressing_policy"`
		} `json:"items"`
	}

	err := query.Do(ctx, &result)
	if err != nil {
		diags.AddError("failed querying for blueprint fabric addressing policy", err.Error())
		return
	}

	switch len(result.Items) {
	case 0:
		diags.AddError(
			"failed querying for blueprint fabric addressing policy",
			fmt.Sprintf("query produced no results: %q", query.String()))
		return
	case 1:
		// expected case handled below
	default:
		diags.AddError(
			"failed querying for blueprint fabric addressing policy",
			fmt.Sprintf("query produced %d results (expected 1): %q", len(result.Items), query.String()))
		return
	}

	if result.Items[0].FabricAddressingPolicy.SpineLeafLinks != result.Items[0].FabricAddressingPolicy.SpineSuperspineLinks {
		diags.AddError(
			"failed querying for blueprint fabric addressing policy",
			fmt.Sprintf("spine_leaf_links addressing does not match spine_superspine_links addressing:\n"+
				"%q vs. %q\nquery: %q",
				result.Items[0].FabricAddressingPolicy.SpineLeafLinks,
				result.Items[0].FabricAddressingPolicy.SpineSuperspineLinks,
				query.String()))
		return
	}

	var addressingScheme apstra.AddressingScheme
	err = addressingScheme.FromString(result.Items[0].FabricAddressingPolicy.SpineLeafLinks)
	if err != nil {
		diags.AddError("failed to parse fabric addressing", err.Error())
		return
	}

	o.FabricAddressing = types.StringValue(addressingScheme.String())
}

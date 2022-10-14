package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"reflect"
	"strings"
)

const defaultFabricAddressingPolicy = goapstra.AddressingSchemeIp4

var _ resource.ResourceWithConfigure = &resourceBlueprint{}
var _ resource.ResourceWithValidateConfig = &resourceBlueprint{}

type resourceBlueprint struct {
	client *goapstra.Client
}

func (o *resourceBlueprint) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (o *resourceBlueprint) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errResourceConfigureProviderDataDetail,
			fmt.Sprintf(errResourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *resourceBlueprint) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This resource instantiates a `Datacenter` Blueprint from a template.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Blueprint ID assigned by Apstra.",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"name": {
				MarkdownDescription: "Blueprint name.",
				Type:                types.StringType,
				Required:            true,
			},
			"template_id": {
				MarkdownDescription: "ID # of a rack-based, pod-based or L3-collapsed template. This parameter cannot be modified in place.",
				Type:                types.StringType,
				Required:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			// todo validate client version
			// todo validate null with rack-baed and pod-based template
			"superspine_spine_addressing": {
				MarkdownDescription: fmt.Sprintf("Addressing scheme for superspine/spine links in a fabric "+
					"built from a `%s` template. Only applicable to Apstra versions 4.1.1 and later.",
					goapstra.TemplateTypePodBased),
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.OneOf(
					goapstra.AddressingSchemeIp4.String(),
					goapstra.AddressingSchemeIp6.String(),
					goapstra.AddressingSchemeIp46.String())},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			// todo validate client version
			// todo validate null with rack-baed and pod-based template
			"spine_leaf_addressing": {
				MarkdownDescription: fmt.Sprintf("Addressing scheme for spine/leaf links in a fabric "+
					"built from a `%s` or `%s templates. Only applicable to Apstra versions 4.1.1 and later.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypeRackBased),
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{stringvalidator.OneOf(
					goapstra.AddressingSchemeIp4.String(),
					goapstra.AddressingSchemeIp6.String(),
					goapstra.AddressingSchemeIp46.String())},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"superspine_asn_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the ASN Pool(s) to be used on superspine switches "+
					"in blueprints built from `%s` templates.",
					goapstra.TemplateTypePodBased),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"spine_asn_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the ASN Pool(s) to be used on spine switches "+
					"in blueprints built from `%s` or `%s` templates.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypeRackBased),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"leaf_asn_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the ASN Pool(s) to be used on leaf switches "+
					"in blueprints built from `%s`, `%s` or `%s` templates.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypeRackBased, goapstra.TemplateTypeL3Collapsed),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"access_asn_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the ASN Pool(s) to be used on access switches "+
					"in blueprints featuring access switches concigured for `%s` redundancy mode.",
					goapstra.AccessRedundancyProtocolEsi),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},

			"superspine_loopback_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for superspine switch "+
					"loopback interfaces in blueprints built from %s templates.", goapstra.TemplateTypePodBased),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"spine_loopback_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for spine switch "+
					"loopback interfaces in blueprints built from %s or %s templates.",
					goapstra.TemplateTypePodBased.String(), goapstra.TemplateTypeRackBased),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"leaf_loopback_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for leaf switch "+
					"loopback interfaces in blueprints built from %s, %s or %s templates.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypeRackBased, goapstra.TemplateTypeL3Collapsed),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"access_loopback_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for access switch "+
					"peer-link interfaces in blueprints featuring access switches concigured for `%s` redundancy mode.",
					goapstra.AccessRedundancyProtocolEsi),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},

			"superspine_spine_ip4_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for superspine/spine "+
					"fabric links in blueprints built from `%s` templates.", goapstra.TemplateTypePodBased),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"spine_leaf_ip4_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for superspine/spine "+
					"fabric links in blueprints built from `%s` or `%s` templates.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypeRackBased),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"leaf_leaf_ip4_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for leaf/leaf "+
					"fabric links in blueprints built from `%s` templates.", goapstra.TemplateTypeL3Collapsed),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},

			"leaf_mlag_peer_link_ip4_pool_ids": {
				MarkdownDescription: "ID(s) of the IPv4 Pool(s) to be used on MLAG peer links between leaf switches.",
				Type:                types.SetType{ElemType: types.StringType},
				Optional:            true,
				Computed:            true,
				Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"access_esi_peer_link_ip4_pool_ids": {
				MarkdownDescription: "ID(s) of the IPv4 Pool(s) to be used on ESI LAG peer links between access switches.",
				Type:                types.SetType{ElemType: types.StringType},
				Optional:            true,
				Computed:            true,
				Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"vtep_ip4_pool_ids": {
				MarkdownDescription: "Unclear what this is for.", //todo
				Type:                types.SetType{ElemType: types.StringType},
				Optional:            true,
				Computed:            true,
				Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},

			"superspine_spine_ip6_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for superspine/spine "+
					"fabric links in blueprints built from `%s` templates and using addressing mode `%s` or `%s`.",
					goapstra.TemplateTypePodBased, goapstra.AddressingSchemeIp6, goapstra.AddressingSchemeIp46),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"spine_leaf_ip6_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for spine/leaf fabric "+
					"links in blueprints built from `%s` or `%s` templates and using addressing mode `%s` or `%s`.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypePodBased,
					goapstra.AddressingSchemeIp6, goapstra.AddressingSchemeIp46),
				Type:          types.SetType{ElemType: types.StringType},
				Optional:      true,
				Computed:      true,
				Validators:    []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},

			"template_type": {
				MarkdownDescription: fmt.Sprintf("Type (`%s`) of the template specified by `template_id`.",
					strings.Join([]string{
						goapstra.TemplateTypePodBased.String(),
						goapstra.TemplateTypeRackBased.String(),
						goapstra.TemplateTypeL3Collapsed.String()}, "`, `")),
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			},
			"switches": {
				MarkdownDescription: "Map of switch info keyed by blueprint role (e.g., 'spine1', or" +
					" '[a_esi_001_leaf1](https://www.juniper.net/documentation/us/en/software/apstra4.1/apstra-user-guide/images/blueprints/staged/physical/build1/resource_stage_401.png)'",
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					"interface_map_id": {
						MarkdownDescription: "links a Logical Device (design element) to a Device Profile which" +
							"describes a hardware model. Optional when only a single interface map references the " +
							"logical device underpinning the node in question.",
						Type:          types.StringType,
						Optional:      true,
						Computed:      true,
						PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
					},
					"device_key": {
						MarkdownDescription: "Unique ID for a device, generally the serial number.",
						Type:                types.StringType,
						Required:            true,
					},
					"device_profile_id": {
						MarkdownDescription: "Device Profile ", //todo
						Type:                types.StringType,
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
					},
					"system_node_id": {
						MarkdownDescription: "ID number of the blueprint graphdb node representing this system.",
						Type:                types.StringType,
						Computed:            true,
						PlanModifiers:       tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
					},
				}),
				Optional: true,
			},
		},
	}, nil
}

func (o *resourceBlueprint) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if o.client == nil { // cannot proceed without a client
		return
	}

	var config rBlueprint
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// ensure template ID exists and learn its type
	templateType, err := o.client.GetTemplateType(ctx, goapstra.ObjectId(config.TemplateId.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error determining template type of template '%s'", config.TemplateId.Value),
			err.Error())
		return
	}

	// save the just-learned template type to our config object b/c it's needed for validation
	config.TemplateType = types.String{Value: templateType.String()}

	// validate resource pools are appropriate for the template type
	config.validateConfigResourcePools(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// validate addressing schemes are appropriate for the template type
	config.validateConfigAddressingSchemes(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// todo: catch case where addressing scheme and fabric ip[46] pools are misaligned

	// ensure ASN pools from the plan exist on Apstra
	var asnPools []attr.Value
	asnPools = append(asnPools, config.SuperspineAsnPoolIds.Elems...)
	asnPools = append(asnPools, config.SpineAsnPoolIds.Elems...)
	asnPools = append(asnPools, config.LeafAsnPoolIds.Elems...)
	asnPools = append(asnPools, config.AccessAsnPoolIds.Elems...)
	missing := findMissingAsnPools(ctx, asnPools, o.client, &resp.Diagnostics)
	if len(missing) > 0 {
		resp.Diagnostics.AddError("cannot assign ASN pool",
			fmt.Sprintf("requested pool id(s) %s not found", missing))
	}

	// ensure Ip4 pools from the plan exist on Apstra
	var ip4Pools []attr.Value
	ip4Pools = append(ip4Pools, config.SuperspineIp4PoolIds.Elems...)   // superspine loopback
	ip4Pools = append(ip4Pools, config.SpineIp4PoolIds.Elems...)        // spine loopback
	ip4Pools = append(ip4Pools, config.LeafIp4PoolIds.Elems...)         // leaf loopback
	ip4Pools = append(ip4Pools, config.AccessIp4PoolIds.Elems...)       // access loopback
	ip4Pools = append(ip4Pools, config.SuperspineSpinePoolIp4.Elems...) // superspine fabric
	ip4Pools = append(ip4Pools, config.SpineLeafPoolIp4.Elems...)       // spine fabric
	ip4Pools = append(ip4Pools, config.LeafLeafPoolIp4.Elems...)        // leaf-only fabric
	ip4Pools = append(ip4Pools, config.LeafMlagPeerIp4.Elems...)        // leaf peer link
	ip4Pools = append(ip4Pools, config.AccessEsiPeerIp4.Elems...)       // access peer link
	ip4Pools = append(ip4Pools, config.VtepIps.Elems...)                // vtep
	missing = findMissingIp4Pools(ctx, ip4Pools, o.client, &resp.Diagnostics)
	if len(missing) > 0 {
		resp.Diagnostics.AddError("cannot assign IPv4 pool",
			fmt.Sprintf("requested pool id(s) %s not found", missing))
	}

	// ensure Ip6 pools from the plan exist on Apstra
	var ip6Pools []attr.Value
	ip6Pools = append(ip6Pools, config.SuperspineSpinePoolIp6.Elems...) // superspine fabric
	ip6Pools = append(ip6Pools, config.SpineLeafPoolIp6.Elems...)       // spine fabric
	missing = findMissingIp6Pools(ctx, ip6Pools, o.client, &resp.Diagnostics)
	if len(missing) > 0 {
		resp.Diagnostics.AddError("cannot assign IPv6 pool",
			fmt.Sprintf("requested pool(s) %s not found", missing))
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// populate device profile IDs, detect errors along the way
	config.populateDeviceProfileIds(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// populate interface map IDs, detect errors along the way
	config.populateInterfaceMapIds(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceBlueprint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rBlueprint
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// ensure template ID exists and learn its type
	templateType, err := o.client.GetTemplateType(ctx, goapstra.ObjectId(plan.TemplateId.Value))
	if err != nil {
		resp.Diagnostics.AddError("error determining template type", err.Error())
		return
	}

	// save the just-learned template type to our config object b/c it's needed later
	plan.TemplateType = types.String{Value: templateType.String()}

	// compute the device profile of each switch the user told us about (use device key)
	plan.populateDeviceProfileIds(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// create blueprint
	blueprintId, err := o.client.CreateBlueprintFromTemplate(ctx, &goapstra.CreateBlueprintFromTemplateRequest{
		RefDesign:  goapstra.RefDesignDatacenter,
		Label:      plan.Name.Value,
		TemplateId: goapstra.ObjectId(plan.TemplateId.Value),
		FabricAddressingPolicy: &goapstra.FabricAddressingPolicy{
			SpineSuperspineLinks: parseFabricAddressingPolicy(plan.SuperspineSpineAddressing, &resp.Diagnostics),
			SpineLeafLinks:       parseFabricAddressingPolicy(plan.SpineLeafAddressing, &resp.Diagnostics),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("error creating Blueprint", err.Error())
	}
	if resp.Diagnostics.HasError() {
		// this check catches errors from multiple levels above, do not merge into "err != nil" check
		return
	}

	// record the blueprint ID
	plan.Id = types.String{Value: string(blueprintId)}

	// create a client specific to the reference design
	blueprint, err := o.client.NewTwoStageL3ClosClient(ctx, blueprintId)
	if err != nil {
		resp.Diagnostics.AddError("error getting blueprint client", err.Error())
		return
	}

	// set user-configured resource group allocations
	for _, tag := range listOfResourceGroupAllocationTags() {
		plan.setApstraPoolAllocationByTfsdkTag(ctx, tag, blueprint, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// warn the user about any omitted resource group allocations
	warnMissingResourceGroupAllocations(ctx, blueprint, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// compute the blueprint "system" node IDs (switches)
	plan.populateSystemNodeIds(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// compute the interface map IDs
	plan.populateInterfaceMapIds(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set interface map assignments (selects hardware model, but not specific instance)
	plan.assignInterfaceMaps(ctx, blueprint, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// warn about switches discovered in the graph db, and which do not appear in the tf config
	plan.warnSwitchConfigVsBlueprint(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// assign switches (managed devices) to blueprint system nodes
	plan.assignManagedDevices(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//// structure we'll use when assigning interface maps to switches
	//ifmapAssignments := make(goapstra.SystemIdToInterfaceMapAssignment)
	//
	// assign details of each configured switch (don't add elements to the plan.Switches map)
	//	- DeviceKey : required user input
	//	- InterfaceMap : optional user input - if only one option, we'll auto-assign
	//	- DeviceProfile : a.k.a. aos_hcl_model - determined from InterfaceMap, represents physical device/model
	//	- SystemNodeId : id of the "type='system', system_type="switch" graph db node representing a spine/leaf/etc...
	//for switchLabel, switchPlan := range plan.Switches {
	//	// fetch the switch graph db node ID and candidate interface maps
	//	systemNodeId, ifmapCandidates, err := getSystemNodeIdAndIfmapCandidates(ctx, r.p.client, blueprintId, switchLabel)
	//	if err != nil {
	//		resp.Diagnostics.AddWarning("error fetching interface map candidates", err.Error())
	//		continue
	//	}
	//
	//	// save the SystemNodeId (1:1 relationship with switchLabel in graph db)
	//	switchPlan.SystemNodeId = types.String{Value: systemNodeId}
	//
	//	// validate/choose interface map, build ifmap assignment structure
	//	if !switchPlan.InterfaceMap.Null && !switchPlan.InterfaceMap.Unknown && !(switchPlan.InterfaceMap.Value == "") {
	//		// user gave us an interface map label they'd like to use
	//		ifmapNodeId := ifmapCandidateFromCandidates(switchPlan.InterfaceMap.Value, ifmapCandidates)
	//		if ifmapNodeId != nil {
	//			ifmapAssignments[systemNodeId] = ifmapNodeId.id
	//			switchPlan.DeviceProfile = types.String{Value: ifmapNodeId.deviceProfileId}
	//		} else {
	//			resp.Diagnostics.AddWarning(
	//				"invalid interface map",
	//				fmt.Sprintf("interface map '%s' not found among candidates for node '%s'",
	//					switchPlan.InterfaceMap.Value, switchLabel))
	//		}
	//	} else {
	//		// user didn't give us an interface map label; try to find a default
	//		switch len(ifmapCandidates) {
	//		case 0: // no candidates!
	//			resp.Diagnostics.AddWarning(
	//				"interface map not specified, and no candidates found",
	//				fmt.Sprintf("no candidate interface maps found for node '%s'", switchLabel))
	//		case 1: // exact match; we can work with this
	//			ifmapAssignments[systemNodeId] = ifmapCandidates[0].id
	//			switchPlan.InterfaceMap = types.String{Value: ifmapCandidates[0].label}
	//			switchPlan.DeviceProfile = types.String{Value: ifmapCandidates[0].deviceProfileId}
	//		default: // multiple match!
	//			sb := strings.Builder{}
	//			sb.WriteString(fmt.Sprintf("'%s'", ifmapCandidates[0].label))
	//			for _, candidate := range ifmapCandidates[1:] {
	//				sb.WriteString(fmt.Sprintf(", '%s'", candidate.label))
	//			}
	//			resp.Diagnostics.AddWarning(
	//				"cannot assign interface map",
	//				fmt.Sprintf("node '%s' has %d interface map candidates. Please choose one of ['%s']",
	//					switchLabel, len(ifmapCandidates), sb.String()))
	//		}
	//	}
	//
	//	plan.Switches[switchLabel] = switchPlan
	//}

	//// assign previously-selected interface maps
	//err = refDesignClient.SetInterfaceMapAssignments(ctx, ifmapAssignments)
	//if err != nil {
	//	if err != nil {
	//		resp.Diagnostics.AddError("error assigning interface maps", err.Error())
	//		return
	//	}
	//}

	//// having assigned interface maps, link physical assets to graph db 'switch' nodes
	//var patch struct {
	//	SystemId string `json:"system_id"`
	//}
	//for _, switchPlan := range plan.Switches {
	//	patch.SystemId = switchPlan.DeviceKey.Value
	//	err = r.p.client.PatchNode(ctx, blueprintId, goapstra.ObjectId(switchPlan.SystemNodeId.Value), &patch, nil)
	//	if err != nil {
	//		resp.Diagnostics.AddWarning("failed to assign switch device", err.Error())
	//	}
	//}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceBlueprint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state rBlueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// some interesting details are in blueprintStatus
	blueprintStatus, err := o.client.GetBlueprintStatus(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error fetching blueprint", err.Error())
		return
	}

	// create a client specific to the reference design
	blueprint, err := o.client.NewTwoStageL3ClosClient(ctx, blueprintStatus.Id)
	if err != nil {
		resp.Diagnostics.AddError("error getting blueprint client", err.Error())
		return
	}

	// create new state object with some obvious values
	newState := rBlueprint{
		Name:                      types.String{Value: blueprintStatus.Label},
		Id:                        state.Id,                        // blindly copy because immutable
		TemplateId:                state.TemplateId,                // blindly copy because resource.RequiresReplace()
		TemplateType:              state.TemplateType,              // blindly copy because resource.RequiresReplace()
		SuperspineSpineAddressing: state.SuperspineSpineAddressing, // blindly copy because resource.RequiresReplace()
		SpineLeafAddressing:       state.SpineLeafAddressing,       // blindly copy because resource.RequiresReplace()
	}

	// collect resource pool values into new state object
	for _, tag := range listOfResourceGroupAllocationTags() {
		newState.readPoolAllocationFromApstraIntoElementByTfsdkTag(ctx, tag, blueprint, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// read switch info from Apstra, then delete any switches unknown to the state file
	newState.readSwitchesFromGraphDb(ctx, o.client, &resp.Diagnostics)
	for sl := range newState.Switches {
		if _, ok := state.Switches[sl]; !ok {
			delete(newState.Switches, sl)
		}
	}

	// Set state
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (o *resourceBlueprint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Retrieve state
	var state rBlueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve plan
	var plan rBlueprint
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a client specific to the reference design
	blueprint, err := o.client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError("error getting blueprint client", err.Error())
		return
	}

	// name change?
	if state.Name.Value != plan.Name.Value {
		setBlueprintName(ctx, blueprint, plan.Name.Value, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// reset resource group allocations as needed
	for _, tag := range listOfResourceGroupAllocationTags() {
		plan.updateResourcePoolAllocationByTfsdkTag(ctx, tag, blueprint, &state, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// compute the device profile of each switch the user told us about (use device key)
	plan.populateDeviceProfileIds(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// compute the blueprint "system" node IDs for the planned switches
	plan.populateSystemNodeIds(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// compute the interface map IDs for the planned switches
	plan.populateInterfaceMapIds(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// prepare two lists. changed switches appear on
	// both lists, will be wiped out then added as new
	switchesToDeleteOrChange := make(map[string]struct{}) // these switches will be completely wiped out
	switchesToAddOrChange := make(map[string]struct{})    // these switches will be added as new

	// accumulate list elements from plan
	for switchLabel := range plan.Switches {
		if _, found := state.Switches[switchLabel]; !found {
			switchesToAddOrChange[switchLabel] = struct{}{}
		}
	}

	// accumulate list elements from state
	for switchLabel := range state.Switches {
		if _, found := plan.Switches[switchLabel]; !found {
			switchesToDeleteOrChange[switchLabel] = struct{}{}
			continue
		}
		if !state.Switches[switchLabel].Equal(plan.Switches[switchLabel]) {
			switchesToAddOrChange[switchLabel] = struct{}{}
			switchesToDeleteOrChange[switchLabel] = struct{}{}
		}
	}

	// wipe out device allocation for delete/change switches
	for label := range switchesToDeleteOrChange {
		nodeId := state.Switches[label].Attrs["system_node_id"].(types.String)
		state.releaseManagedDevice(ctx, nodeId.Value, o.client, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// wipe out node->interface map assignments for delete/change switches
	assignments := make(goapstra.SystemIdToInterfaceMapAssignment, len(switchesToDeleteOrChange))
	for label := range switchesToDeleteOrChange {
		nodeId := state.Switches[label].Attrs["system_node_id"].(types.String)
		assignments[nodeId.Value] = nil
	}
	err = blueprint.SetInterfaceMapAssignments(ctx, assignments)
	if err != nil {
		resp.Diagnostics.AddError("error clearing interface map assignment", err.Error())
	}

	// create interface_map assignments for add/change switches
	for label := range switchesToAddOrChange {
		nodeId := plan.Switches[label].Attrs["system_node_id"].(types.String)
		interfaceMapId := plan.Switches[label].Attrs["interface_map_id"].(types.String)
		assignments[nodeId.Value] = interfaceMapId.Value
	}
	err = blueprint.SetInterfaceMapAssignments(ctx, assignments)
	if err != nil {
		resp.Diagnostics.AddError("error setting interface map assignment", err.Error())
	}

	// assert device allocation for entire plan
	plan.assignManagedDevices(ctx, o.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (o *resourceBlueprint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rBlueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.DeleteBlueprint(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("error deleting blueprint", err.Error())
		return
	}
}

// getSwitchLabelId queries the graph db for 'switch' type systems, returns
// map[string]string (map[label]id)
func getSwitchLabelId(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId) (map[string]string, error) {
	var switchQr struct {
		Items []struct {
			System struct {
				Label string `json:"label"`
				Id    string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"name", goapstra.QEStringVal("n_system")},
			{"system_type", goapstra.QEStringVal("switch")},
		}).
		Do(&switchQr)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(switchQr.Items))
	for _, item := range switchQr.Items {
		result[item.System.Label] = item.System.Id
	}

	return result, nil
}

// populateDeviceProfileIds used the user supplied device_key for each switch to
// populate device_profile_id (hardware type)
func (o *rBlueprint) populateDeviceProfileIds(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	allSystemsInfo := getAllSystemsInfo(ctx, client, diags)
	if diags.HasError() {
		return
	}

	for switchLabel, plannedSwitch := range o.Switches {
		deviceKey := plannedSwitch.Attrs["device_key"].(types.String).Value
		var msi goapstra.ManagedSystemInfo
		var found bool
		if msi, found = allSystemsInfo[deviceKey]; !found {
			diags.AddAttributeError(
				path.Root("switches").AtMapKey(switchLabel),
				"managed device not found",
				fmt.Sprintf("Switch with device_key '%s' not found among managed devices", deviceKey),
			)
		}
		deviceProfileId := string(msi.Facts.AosHclModel)
		o.Switches[switchLabel].Attrs["device_profile_id"] = types.String{Value: deviceProfileId}
	}
}

// populateInterfaceMapIds attempts to populate (and validate when known) the
// interface_map_id for each switch using either the design API's global catalog
// (blueprint not created yet) or graphDB elements (blueprint exists).
func (o *rBlueprint) populateInterfaceMapIds(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	if o.Id.IsNull() {
		o.populateInterfaceMapIdsFromGC(ctx, client, diags)
	} else {
		o.populateInterfaceMapIdsFromBP(ctx, client, diags)
	}
}

func (o *rBlueprint) populateInterfaceMapIdsFromGC(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	for switchLabel, plannedSwitch := range o.Switches {
		devProfile := plannedSwitch.Attrs["device_profile_id"].(types.String)
		ifMapId := plannedSwitch.Attrs["interface_map_id"].(types.String)

		// sanity check
		if devProfile.IsUnknown() {
			diags.AddError(errProviderBug,
				fmt.Sprintf("attempt to populateInterfaceMapIdsFromGC for switch '%s' while device profile is unknown",
					switchLabel))
			return
		}

		if !ifMapId.IsNull() {
			assertInterfaceMapSupportsDeviceProfile(ctx, client, ifMapId.Value, devProfile.Value, diags)
			if diags.HasError() {
				return
			}
			o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Value: ifMapId.Value}
			continue
		}

		// todo: try to populate interface map ID
		//  until we have a way to parse the template to learn logical device types, this is impossible
		o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Null: true}
	}
}

func (o *rBlueprint) populateInterfaceMapIdsFromBP(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	// structure for receiving results of blueprint query
	var candidateInterfaceMapsQR struct {
		Items []struct {
			InterfaceMap struct {
				Id string `json:"id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}

	for switchLabel, plannedSwitch := range o.Switches {
		devProfile := plannedSwitch.Attrs["device_profile_id"].(types.String)
		ifMapId := plannedSwitch.Attrs["interface_map_id"].(types.String)

		// sanity check
		if devProfile.IsUnknown() {
			diags.AddError(errProviderBug,
				fmt.Sprintf("attempt to populateInterfaceMapIdsFromBP for switch '%s' while device profile is unknown",
					switchLabel))
			return
		}

		query := client.NewQuery(goapstra.ObjectId(o.Id.Value)).
			SetContext(ctx).
			Node([]goapstra.QEEAttribute{
				{"type", goapstra.QEStringVal("system")},
				{"label", goapstra.QEStringVal(switchLabel)},
				//{"name", goapstra.QEStringVal("n_system")},
			}).
			Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
			Node([]goapstra.QEEAttribute{
				{"type", goapstra.QEStringVal("logical_device")},
			}).
			In([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}})

		if ifMapId.IsUnknown() {
			query = query.Node([]goapstra.QEEAttribute{
				{"type", goapstra.QEStringVal("interface_map")},
				{"name", goapstra.QEStringVal("n_interface_map")},
			})
		} else {
			query = query.Node([]goapstra.QEEAttribute{
				{"type", goapstra.QEStringVal("interface_map")},
				{"name", goapstra.QEStringVal("n_interface_map")},
				{"id", goapstra.QEStringVal(ifMapId.Value)},
			})
		}

		err := query.Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("device_profile")}}).
			Node([]goapstra.QEEAttribute{
				{"type", goapstra.QEStringVal("device_profile")},
				{"device_profile_id", goapstra.QEStringVal(devProfile.Value)},
			}).
			Do(&candidateInterfaceMapsQR)
		if err != nil {
			diags.AddError("error running interface map query", err.Error())
			return
		}

		switch len(candidateInterfaceMapsQR.Items) {
		case 0:
			diags.AddAttributeError(path.Root("switches").AtMapKey(switchLabel),
				"unable to assign interface_map",
				fmt.Sprintf("no interface_map links system '%s' to device profile '%s'",
					switchLabel, devProfile.Value))
		case 1:
			o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Value: candidateInterfaceMapsQR.Items[0].InterfaceMap.Id}
		default:
		}
	}
}

func (o *rBlueprint) populateSystemNodeIds(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	var candidateSystemsQR struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}

	for switchLabel := range o.Switches {
		err := client.NewQuery(goapstra.ObjectId(o.Id.Value)).
			SetContext(ctx).
			Node([]goapstra.QEEAttribute{
				{"type", goapstra.QEStringVal("system")},
				{"label", goapstra.QEStringVal(switchLabel)},
				{"name", goapstra.QEStringVal("n_system")},
			}).
			Do(&candidateSystemsQR)
		if err != nil {
			diags.AddError("error querying for bp system node", err.Error())
		}

		switch len(candidateSystemsQR.Items) {
		case 0:
			diags.AddError("switch node not found in blueprint",
				fmt.Sprintf("switch/system node with label '%s' not found in blueprint", switchLabel))
			return
		case 1:
			// no error case
		default:
			diags.AddError("multiple switches found in blueprint",
				fmt.Sprintf("switch/system node with label '%s': %d matches found in blueprint",
					switchLabel, len(candidateSystemsQR.Items)))
			return
		}

		o.Switches[switchLabel].Attrs["system_node_id"] = types.String{Value: candidateSystemsQR.Items[0].System.Id}
	}
}

func (o *rBlueprint) assignInterfaceMaps(ctx context.Context, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	assignments := make(goapstra.SystemIdToInterfaceMapAssignment, len(o.Switches))
	for k, v := range o.Switches {
		switch {
		case v.Attrs["system_node_id"].IsUnknown():
			diags.AddError(errProviderBug, fmt.Sprintf("switch '%s' system_id is unknown", k))
		case v.Attrs["system_node_id"].IsNull():
			diags.AddError(errProviderBug, fmt.Sprintf("switch '%s' system_id is null", k))
		case v.Attrs["interface_map_id"].IsUnknown():
			diags.AddError(errProviderBug, fmt.Sprintf("switch '%s' interface_map_id is unknown", k))
		case v.Attrs["interface_map_id"].IsNull():
			assignments[v.Attrs["system_node_id"].(types.String).Value] = nil
		default:
			assignments[v.Attrs["system_node_id"].(types.String).Value] = v.Attrs["interface_map_id"].(types.String).Value
		}

		err := client.SetInterfaceMapAssignments(ctx, assignments)
		if err != nil {
			diags.AddError("error assigning interface maps", err.Error())
		}
	}

	err := client.SetInterfaceMapAssignments(ctx, assignments)
	if err != nil {
		diags.AddError("error assigning interface maps", err.Error())
	}
}

func (o *rBlueprint) assignManagedDevices(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	//// having assigned interface maps, link physical assets to graph db 'switch' nodes
	var patch struct {
		SystemId string `json:"system_id"`
	}
	bpId := goapstra.ObjectId(o.Id.Value)
	for _, plannedSwitch := range o.Switches {
		patch.SystemId = plannedSwitch.Attrs["device_key"].(types.String).Value
		nodeId := goapstra.ObjectId(plannedSwitch.Attrs["system_node_id"].(types.String).Value)
		err := client.PatchNode(ctx, bpId, nodeId, &patch, nil)
		if err != nil {
			diags.AddWarning(fmt.Sprintf("failed to assign switch device for node '%s'", nodeId), err.Error())
		}
	}
}

func (o *rBlueprint) releaseManagedDevice(ctx context.Context, nodeId string, client *goapstra.Client, diags *diag.Diagnostics) {
	var patch struct {
		_ interface{} `json:"system_id"`
	}
	err := client.PatchNode(ctx, goapstra.ObjectId(o.Id.Value), goapstra.ObjectId(nodeId), &patch, nil)
	if err != nil {
		diags.AddWarning(fmt.Sprintf("failed to assign switch device for node '%s'", nodeId), err.Error())
	}
}

func assertInterfaceMapSupportsDeviceProfile(ctx context.Context, client *goapstra.Client, ifMapId string, devProfileId string, diags *diag.Diagnostics) {
	ifMap, err := client.GetInterfaceMapDigest(ctx, goapstra.ObjectId(ifMapId))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			diags.AddError("interface map not found",
				fmt.Sprintf("interfacem map with id '%s' not found", ifMapId))
		}
		diags.AddError(fmt.Sprintf("error fetching interface map '%s'", ifMapId), err.Error())
	}
	if string(ifMap.DeviceProfile.Id) != devProfileId {
		diags.AddError(
			errInvalidConfig,
			fmt.Sprintf("interface map '%s' works with device profile '%s', not '%s'",
				ifMapId, ifMap.DeviceProfile.Id, devProfileId))
	}
}

//func assertInterfaceMapExists(ctx context.Context, client *goapstra.Client, id string, diags *diag.Diagnostics) {
//	_, err := client.GetInterfaceMapDigest(ctx, goapstra.ObjectId(id))
//	if err != nil {
//		var ace goapstra.ApstraClientErr
//		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
//			diags.AddError("interface map not found",
//				fmt.Sprintf("interfacem map with id '%s' not found", id))
//		}
//		diags.AddError(fmt.Sprintf("error fetching interface map '%s'", id), err.Error())
//	}
//}

// populateSwitchNodeAndInterfaceMapIds
func (o *rBlueprint) populateSwitchNodeAndInterfaceMapIds(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	if o.Id.IsUnknown() {
		diags.AddError(errProviderBug, "attempt to populateSwitchNodeAndInterfaceMapIds while blueprint ID is unknown")
		return
	}

	for switchLabel, plannedSwitch := range o.Switches {
		nodeId, ldId, err := getSystemNodeIdAndLogicalDeviceId(ctx, client, goapstra.ObjectId(o.Id.Value), switchLabel)
		if err != nil {
			diags.AddAttributeError(
				path.Root("switches").AtMapKey(switchLabel),
				fmt.Sprintf("error fetching node ID for switch '%s'", switchLabel),
				err.Error())
			return
		}
		// save the system node ID
		o.Switches[switchLabel].Attrs["system_node_id"] = types.String{Value: nodeId}

		// device profile should be known at this time
		if plannedSwitch.Attrs["device_profile_id"].IsUnknown() {
			diags.AddAttributeWarning(
				path.Root("switches").AtMapKey(switchLabel),
				"device profile unknown",
				fmt.Sprintf("device profile for '%s' unknown - this is probably a bug", switchLabel))
			continue
		}

		// fetch interface map candidates
		deviceProfile := plannedSwitch.Attrs["device_profile_id"].(types.String).Value
		imaps, err := client.GetInterfaceMapDigestsLogicalDeviceAndDeviceProfile(ctx, goapstra.ObjectId(ldId), goapstra.ObjectId(deviceProfile))
		if err != nil {
			diags.AddAttributeError(
				path.Root("switches").AtMapKey(switchLabel),
				fmt.Sprintf("error fetching interface map digests for switch '%s'", switchLabel),
				err.Error())
		}
		switch len(imaps) {
		case 0:
			diags.AddAttributeError(
				path.Root("switches").AtMapKey(switchLabel),
				fmt.Sprintf("switch '%s': could not find interface map", switchLabel),
				fmt.Sprintf("no interface maps link logical device '%s' to device profile '%s'",
					ldId, switchLabel))
		case 1:
			o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Value: string(imaps[0].Id)}
		default:
			o.Switches[switchLabel].Attrs["interface_map_id"] = types.String{Null: true}
			imapIds := make([]string, len(imaps))
			for i, imap := range imaps {
				imapIds[i] = string(imap.Id)
			}
			diags.AddAttributeWarning(
				path.Root("switches").AtMapKey(switchLabel),
				fmt.Sprintf("switch '%s': multiple interface map candidates", switchLabel),
				fmt.Sprintf("please configure 'interface_map_id' using one of the valid candidate IDs: '%s'",
					strings.Join(imapIds, "', '")))
		}
	}
}

func (o *rBlueprint) warnSwitchConfigVsBlueprint(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	switchLabelToGraphDbId, err := getSwitchLabelId(ctx, client, goapstra.ObjectId(o.Id.Value))
	if err != nil {
		diags.AddError("error getting blueprint switch inventory", err.Error())
		return
	}

	var missing []string
	for switchLabel := range switchLabelToGraphDbId {
		if _, found := o.Switches[switchLabel]; !found {
			missing = append(missing, switchLabel)
		}
	}

	var extra []string
	for switchLabel := range o.Switches {
		if _, found := switchLabelToGraphDbId[switchLabel]; !found {
			extra = append(extra, switchLabel)
		}
	}

	// warn about missing switches
	if len(missing) != 0 {
		diags.AddAttributeWarning(
			path.Root("switches"),
			"switch missing from plan",
			fmt.Sprintf("blueprint expects the following switches: ['%s']",
				strings.Join(missing, "', '")))
	}

	// warn about extraneous switches mentioned in config
	if len(extra) != 0 {
		diags.AddAttributeWarning(
			path.Root("switches"),
			"extraneous switches found in configuration",
			fmt.Sprintf("please remove switches not needed by blueprint: '%s'",
				strings.Join(extra, "', '")))
	}
}

type ifmapInfo struct {
	id              string
	label           string
	deviceProfileId string
}

// getSystemNodeIdAndLogicalDeviceId takes the 'label' field representing a
// graph db node with "type='system', returns its node id and the linked
// logical device Id
func getSystemNodeIdAndLogicalDeviceId(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, label string) (string, string, error) {
	var systemAndLogicalDeviceQR struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
			LogicalDevice struct {
				Id string `json:"id"`
			} `json:"n_logical_device"`
		} `json:"items"`
	}

	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(label)},
			{"name", goapstra.QEStringVal("n_system")},
		}).
		Out([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("logical_device")},
		}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("logical_device")},
			{"name", goapstra.QEStringVal("n_logical_device")},
		}).
		Do(&systemAndLogicalDeviceQR)
	if err != nil {
		return "", "", err
	}

	switch len(systemAndLogicalDeviceQR.Items) {
	case 0:
		return "", "", fmt.Errorf("query result for 'system' node, with label '%s' empty", label)
	case 1:
		// expected behavior - no error
	default:
		return "", "", fmt.Errorf("query result for 'system' node, with label '%s' returned %d results",
			label, len(systemAndLogicalDeviceQR.Items))
	}

	return systemAndLogicalDeviceQR.Items[0].System.Id, systemAndLogicalDeviceQR.Items[0].LogicalDevice.Id, nil
}

// getSystemNodeIdAndIfmapCandidates takes the 'label' field representing a
// graph db node with "type='system', returns the node id and a []ifmapInfo
// representing candidate interface maps for that system.
func getSystemNodeIdAndIfmapCandidates(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, label string) (string, []ifmapInfo, error) {
	var candidateInterfaceMapsQR struct {
		Items []struct {
			System struct {
				Id string `json:"id"`
			} `json:"n_system"`
			InterfaceMap struct {
				Id              string `json:"id"`
				Label           string `json:"label"`
				DeviceProfileId string `json:"device_profile_id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(label)},
			{"name", goapstra.QEStringVal("n_system")},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("logical_device")},
		}).
		In([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"name", goapstra.QEStringVal("n_interface_map")},
		}).
		Do(&candidateInterfaceMapsQR)
	if err != nil {
		return "", nil, err
	}

	var systemNodeId string
	var candidates []ifmapInfo
	for _, item := range candidateInterfaceMapsQR.Items {
		if item.System.Id == "" {
			return "", nil, fmt.Errorf("graph db search for \"type='system', label='%s'\" found match with empty 'id' field", label)
		}
		if systemNodeId != "" && systemNodeId != item.System.Id {
			return "", nil,
				fmt.Errorf("graph db search for \"type='system', label='%s'\" found nodes with different 'id' fields: '%s' and '%s'",
					label, systemNodeId, item.System.Id)
		}
		if systemNodeId == "" {
			systemNodeId = item.System.Id
		}
		candidates = append(candidates, ifmapInfo{
			label:           item.InterfaceMap.Label,
			id:              item.InterfaceMap.Id,
			deviceProfileId: item.InterfaceMap.DeviceProfileId,
		})
	}

	return systemNodeId, candidates, nil
}

// ifmapCandidateFromCandidates finds an interface map (by label) within a
// []ifmapInfo, returns pointer to it, nil if not found.
func ifmapCandidateFromCandidates(label string, candidates []ifmapInfo) *ifmapInfo {
	for _, candidate := range candidates {
		if label == candidate.label {
			return &candidate
		}
	}
	return nil
}

func getNodeInterfaceMap(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, label string) (*ifmapInfo, error) {
	var interfaceMapQR struct {
		Items []struct {
			InterfaceMap struct {
				Id              string `json:"id"`
				Label           string `json:"label"`
				DeviceProfileId string `json:"device_profile_id"`
			} `json:"n_interface_map"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(label)},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("interface_map")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"name", goapstra.QEStringVal("n_interface_map")},
		}).
		Do(&interfaceMapQR)
	if err != nil {
		return nil, err
	}
	if len(interfaceMapQR.Items) != 1 {
		return nil, fmt.Errorf("expected exactly one interface map, got %d", len(interfaceMapQR.Items))
	}
	return &ifmapInfo{
		id:              interfaceMapQR.Items[0].InterfaceMap.Id,
		label:           interfaceMapQR.Items[0].InterfaceMap.Label,
		deviceProfileId: interfaceMapQR.Items[0].InterfaceMap.DeviceProfileId,
	}, nil
}

func getSwitchFromGraphDb(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, label string, diags *diag.Diagnostics) *types.Object {
	// query for system node on its own
	var systemOnlyQR struct {
		Items []struct {
			System struct {
				NodeId   string `json:"id"`
				Label    string `json:"label"`
				SystemID string `json:"system_id"`
			} `json:"n_system"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"label", goapstra.QEStringVal(label)},
			{"name", goapstra.QEStringVal("n_system")},
		}).
		Do(&systemOnlyQR)
	if err != nil {
		diags.AddError("error querying blueprint node", err.Error())
		return nil
	}
	if len(systemOnlyQR.Items) != 1 {
		diags.AddError("error querying blueprint node",
			fmt.Sprintf("expected exactly one system node, got %d", len(systemOnlyQR.Items)))
		return nil
	}

	// query for system node with interface map
	type interfaceMapQrItem struct {
		InterfaceMap struct {
			Id              string `json:"id"`
			DeviceProfileId string `json:"device_profile_id"`
		} `json:"n_interface_map"`
	}
	var interfaceMapQR struct {
		Items []interfaceMapQrItem `json:"items"`
	}
	err = client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"id", goapstra.QEStringVal(systemOnlyQR.Items[0].System.NodeId)},
		}).
		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("interface_map")}}).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("interface_map")},
			{"name", goapstra.QEStringVal("n_interface_map")},
		}).
		Do(&interfaceMapQR)
	if err != nil {
		diags.AddError("error querying blueprint node", err.Error())
		return nil
	}
	switch len(interfaceMapQR.Items) {
	case 0: // slam an empty interfaceMapQrItem{} in there so we have something to read
		interfaceMapQR.Items = append(interfaceMapQR.Items, interfaceMapQrItem{})
	case 1: // this is the expected case - no special handling
	default: // this should never happen
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("node '%s' linked to more than one (%d) interface maps",
				systemOnlyQR.Items[0].System.NodeId, len(interfaceMapQR.Items)))
	}

	result := newSwitchObject()
	result.Attrs["device_key"] = types.String{Value: systemOnlyQR.Items[0].System.SystemID}
	result.Attrs["system_node_id"] = types.String{Value: systemOnlyQR.Items[0].System.NodeId}
	result.Attrs["interface_map_id"] = types.String{Value: interfaceMapQR.Items[0].InterfaceMap.Id}
	result.Attrs["device_profile_id"] = types.String{Value: interfaceMapQR.Items[0].InterfaceMap.DeviceProfileId}

	// flag any result elements with empty value as null
	for k, _ := range switchElementSchema() {
		if result.Attrs[k].(types.String).Value == "" {
			result.Attrs[k] = types.String{Null: true}
		}
	}

	return &result
}

type rBlueprint struct {
	Id                        types.String            `tfsdk:"id"`
	Name                      types.String            `tfsdk:"name"`
	TemplateId                types.String            `tfsdk:"template_id"`
	TemplateType              types.String            `tfsdk:"template_type"`
	SuperspineSpineAddressing types.String            `tfsdk:"superspine_spine_addressing"`
	SpineLeafAddressing       types.String            `tfsdk:"spine_leaf_addressing"`
	Switches                  map[string]types.Object `tfsdk:"switches"`

	// asn pools
	SuperspineAsnPoolIds types.Set `tfsdk:"superspine_asn_pool_ids"` // asn superspine
	SpineAsnPoolIds      types.Set `tfsdk:"spine_asn_pool_ids"`      // asn spine
	LeafAsnPoolIds       types.Set `tfsdk:"leaf_asn_pool_ids"`       // asn leaf
	AccessAsnPoolIds     types.Set `tfsdk:"access_asn_pool_ids"`     // asn access

	// loopback ipv4 pools
	SuperspineIp4PoolIds types.Set `tfsdk:"superspine_loopback_pool_ids"` // loopback superspine
	SpineIp4PoolIds      types.Set `tfsdk:"spine_loopback_pool_ids"`      // loopback spine
	LeafIp4PoolIds       types.Set `tfsdk:"leaf_loopback_pool_ids"`       // loopback leaf
	AccessIp4PoolIds     types.Set `tfsdk:"access_loopback_pool_ids"`     // loopback access

	// fabric ipv4 pools
	SuperspineSpinePoolIp4 types.Set `tfsdk:"superspine_spine_ip4_pool_ids"` // fabric superspine/spine
	SpineLeafPoolIp4       types.Set `tfsdk:"spine_leaf_ip4_pool_ids"`       // fabric spine/leaf
	LeafLeafPoolIp4        types.Set `tfsdk:"leaf_leaf_ip4_pool_ids"`        // fabric leaf/leaf

	// other ipv4 pools
	LeafMlagPeerIp4  types.Set `tfsdk:"leaf_mlag_peer_link_ip4_pool_ids"`  // peer link leaf
	AccessEsiPeerIp4 types.Set `tfsdk:"access_esi_peer_link_ip4_pool_ids"` // peer link access
	VtepIps          types.Set `tfsdk:"vtep_ip4_pool_ids"`                 // vtep

	// fabric ipv6 pools
	SuperspineSpinePoolIp6 types.Set `tfsdk:"superspine_spine_ip6_pool_ids"` //
	SpineLeafPoolIp6       types.Set `tfsdk:"spine_leaf_ip6_pool_ids"`
}

func (o *rBlueprint) validateConfigResourcePools(diags *diag.Diagnostics) {
	// throws errors if pools permitted only in pod-based
	// templates are found in the configuration
	errOnPodBasedOnlyPools := func() {
		switch {
		case !o.SuperspineAsnPoolIds.Null:
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SuperspineAsnPoolIds", diags)))
		case !o.SuperspineIp4PoolIds.Null:
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SuperspineIp4PoolIds", diags)))
		case !o.SuperspineSpinePoolIp4.Null:
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SuperspineSpinePoolIp4", diags)))
		case !o.SuperspineSpinePoolIp6.Null:
			diags.AddError(
				errInvalidConfig, fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SuperspineSpinePoolIp6", diags)))
		}
	}

	// throws errors if pools permitted only in l3Collapsed
	// templates are found in the configuration
	errOnL3CollapsedOnlyPools := func() {
		if !o.LeafLeafPoolIp4.Null {
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "LeafLeafPoolIp4", diags)))
		}
	}

	// throws errors if pools permitted forbidden by l3Collapsed
	// templates are found in the configuration
	errOnL3CollapsedForbiddenPools := func() {
		errOnPodBasedOnlyPools()
		switch {
		case !o.SpineAsnPoolIds.Null:
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SpineAsnPoolIds", diags)))
		case !o.SpineIp4PoolIds.Null:
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SpineIp4PoolIds", diags)))
		case !o.SpineLeafPoolIp4.Null:
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SpineLeafPoolIp4", diags)))
		case !o.SpineLeafPoolIp6.Null:
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SpineLeafPoolIp6", diags)))
		case !o.LeafMlagPeerIp4.Null:
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "LeafMlagPeerIp4", diags)))
		}
	}

	switch o.TemplateType.Value {
	case goapstra.TemplateTypePodBased.String():
		errOnL3CollapsedOnlyPools()
	case goapstra.TemplateTypeRackBased.String():
		errOnPodBasedOnlyPools()
		errOnL3CollapsedOnlyPools()
	case goapstra.TemplateTypeL3Collapsed.String():
		errOnL3CollapsedForbiddenPools()
	default:
		diags.AddError(errProviderBug, fmt.Sprintf("unexpected template type '%s'", o.TemplateType.Value))
		return
	}
}

func (o *rBlueprint) validateConfigAddressingSchemes(diags *diag.Diagnostics) {
	// throws errors if addressing scheme permitted only in pod-based
	// templates is found in the configuration
	errOnPodBasedOnlyAddressingScheme := func() {
		if !o.SuperspineSpineAddressing.Null {
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SuperspineSpineAddressing", diags)))
		}
	}

	// throws errors if any addressing scheme (these are permitted only in
	// pod-baed and rack-based templates) is found in the configuration
	errOnAnyAddressingScheme := func() {
		errOnPodBasedOnlyAddressingScheme()
		if !o.SpineLeafAddressing.Null {
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "SpineLeafAddressing", diags)))
		}
	}

	switch o.TemplateType.Value {
	case goapstra.TemplateTypePodBased.String():
	case goapstra.TemplateTypeRackBased.String():
		errOnPodBasedOnlyAddressingScheme()
	case goapstra.TemplateTypeL3Collapsed.String():
		errOnAnyAddressingScheme()
	default:
		diags.AddError(errProviderBug, fmt.Sprintf("unexpected template type '%s'", o.TemplateType.Value))
		return
	}
}

// resourceTypeNameFromResourceGroupName guesses a resource type name
// (asn/ip/ipv6/possibly others) based on the resource group name. Both type and
// name are required to uniquely identify a resource group allocation, but so
// far (fingers crossed) the group names (e.g. "leaf_asns") supply enough of a
// clue to determine the resource type ("asn"). Using this lookup function saves
// functions which interact with resource groups from the hassle of keeping
// track of resource type.
func resourceTypeNameFromResourceGroupName(in goapstra.ResourceGroupName, diags *diag.Diagnostics) goapstra.ResourceType {
	switch in {
	case goapstra.ResourceGroupNameSuperspineAsn:
		return goapstra.ResourceTypeAsnPool
	case goapstra.ResourceGroupNameSpineAsn:
		return goapstra.ResourceTypeAsnPool
	case goapstra.ResourceGroupNameLeafAsn:
		return goapstra.ResourceTypeAsnPool
	case goapstra.ResourceGroupNameAccessAsn:
		return goapstra.ResourceTypeAsnPool

	case goapstra.ResourceGroupNameSuperspineIp4:
		return goapstra.ResourceTypeIp4Pool
	case goapstra.ResourceGroupNameSpineIp4:
		return goapstra.ResourceTypeIp4Pool
	case goapstra.ResourceGroupNameLeafIp4:
		return goapstra.ResourceTypeIp4Pool
	case goapstra.ResourceGroupNameAccessIp4:
		return goapstra.ResourceTypeIp4Pool

	case goapstra.ResourceGroupNameSuperspineSpineIp4:
		return goapstra.ResourceTypeIp4Pool
	case goapstra.ResourceGroupNameSpineLeafIp4:
		return goapstra.ResourceTypeIp4Pool
	case goapstra.ResourceGroupNameLeafLeafIp4:
		return goapstra.ResourceTypeIp4Pool

	case goapstra.ResourceGroupNameMlagDomainSviSubnets:
		return goapstra.ResourceTypeIp4Pool
	case goapstra.ResourceGroupNameAccessAccessIps:
		return goapstra.ResourceTypeIp4Pool
	case goapstra.ResourceGroupNameVtepIps:
		return goapstra.ResourceTypeIp4Pool

	case goapstra.ResourceGroupNameSuperspineSpineIp6:
		return goapstra.ResourceTypeIp6Pool
	case goapstra.ResourceGroupNameSpineLeafIp6:
		return goapstra.ResourceTypeIp6Pool
	}
	diags.AddError(errProviderBug, fmt.Sprintf("unable to determine type of resource group '%s'", in))
	return goapstra.ResourceTypeUnknown
}

// extractResourcePoolElementByTfsdkTag returns the value (types.Set)
//identified by fieldName (a tfsdk tag) found within the rBlueprint object
func (o *rBlueprint) extractResourcePoolElementByTfsdkTag(fieldName string, diags *diag.Diagnostics) types.Set {
	v := reflect.ValueOf(o).Elem()
	// It's possible we can cache this, which is why precompute all these ahead of time.
	findTfsdkName := func(t reflect.StructTag) string {
		if tfsdkTag, ok := t.Lookup("tfsdk"); ok {
			return tfsdkTag
		}
		diags.AddError(errProviderBug, fmt.Sprintf("attempt to lookupg nonexistent tfsdk tag '%s'", fieldName))
		return ""
	}
	fieldNames := map[string]int{}
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		tag := typeField.Tag
		tname := findTfsdkName(tag)
		fieldNames[tname] = i
	}

	fieldNum, ok := fieldNames[fieldName]
	if !ok {
		diags.AddError(errProviderBug, fmt.Sprintf("field '%s' does not exist within the provided item", fieldName))
	}
	fieldVal := v.Field(fieldNum)
	return fieldVal.Interface().(types.Set)
}

// setResourcePoolElementByTfsdkTag sets value (types.Set) into the named field
// of the rBlueprint object by tfsdk tag
func (o *rBlueprint) setResourcePoolElementByTfsdkTag(fieldName string, value types.Set, diags *diag.Diagnostics) {
	v := reflect.ValueOf(o).Elem()
	findTfsdkName := func(t reflect.StructTag) string {
		if tfsdkTag, ok := t.Lookup("tfsdk"); ok {
			return tfsdkTag
		}
		diags.AddError(errProviderBug, fmt.Sprintf("attempt to lookupg nonexistent tfsdk tag '%s'", fieldName))
		return ""
	}
	fieldNames := map[string]int{}
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		tag := typeField.Tag
		tname := findTfsdkName(tag)
		fieldNames[tname] = i
	}

	fieldNum, ok := fieldNames[fieldName]
	if !ok {
		diags.AddError(errProviderBug, fmt.Sprintf("field '%s' does not exist within the provided item", fieldName))
	}
	fieldVal := v.Field(fieldNum)
	fieldVal.Set(reflect.ValueOf(value))
}

// readPoolAllocationFromApstraIntoElementByTfsdkTag retrieves a pool
// allocation from apstra and sets the appropriate element in the
// rBlueprint structure
func (o *rBlueprint) readPoolAllocationFromApstraIntoElementByTfsdkTag(ctx context.Context, tag string, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	rgn := tfsdkTagToRgn(tag, diags)
	rg := &goapstra.ResourceGroup{
		Type: resourceTypeNameFromResourceGroupName(rgn, diags),
		Name: rgn,
	}
	if diags.HasError() {
		return
	}
	rga, err := client.GetResourceAllocation(ctx, rg)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			// apstra doesn't know about this resource allocation
			o.setResourcePoolElementByTfsdkTag(tag, types.Set{Null: true, ElemType: types.StringType}, diags)
			return
		}
		diags.AddError("error getting resource group allocation", err.Error())
		return
	}

	poolIds := types.Set{
		ElemType: types.StringType,
		Elems:    make([]attr.Value, len(rga.PoolIds)),
	}

	for i, poolId := range rga.PoolIds {
		poolIds.Elems[i] = types.String{Value: string(poolId)}
	}

	o.setResourcePoolElementByTfsdkTag(tag, poolIds, diags)
}

// setApstraPoolAllocationByTfsdkTag reads the named pool allocation element
// from the rBlueprint object and sets that value in apstra.
func (o *rBlueprint) setApstraPoolAllocationByTfsdkTag(ctx context.Context, tag string, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// extract the poolSet matching 'tag' from o
	poolSet := o.extractResourcePoolElementByTfsdkTag(tag, diags)

	// create a goapstra.ResourceGroupAllocation object for (a) query and (b) assignment
	rga := newRga(tfsdkTagToRgn(tag, diags), &poolSet, diags)

	// get apstra's opinion on the resource group in question
	_, err := client.GetResourceAllocation(ctx, &rga.ResourceGroup)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			// the blueprint doesn't need/want this resource group.
			if poolSet.IsNull() {
				return // apstra does not want, and the set is null - nothing to do
			}
			if poolSet.IsUnknown() {
				// can't have unknown values in TF state
				o.setResourcePoolElementByTfsdkTag(tag, types.Set{ElemType: types.StringType, Null: true}, diags)
				return
			}
			// blueprint doesn't want it, but the pool is non-null (appears in the TF config) warn the user.
			diags.AddWarning(warnUnwantedResourceSummary, fmt.Sprintf(warnUnwantedResourceDetail, tag))
			// overwrite the planned value so that TF state doesn't reflect the unwanted group
			o.setResourcePoolElementByTfsdkTag(tag, types.Set{ElemType: types.StringType, Null: true}, diags)
			return
		} else {
			// not a 404
			diags.AddWarning("error reading resource allocation", err.Error())
			return
		}
	}

	if poolSet.IsUnknown() && len(poolSet.Elems) != 0 {
		diags.AddError("oh snap", "an unknown, but non-empty poolSet")
	}

	if poolSet.IsUnknown() {
		// can't have unknown values in TF state
		o.setResourcePoolElementByTfsdkTag(tag, types.Set{ElemType: types.StringType, Null: true}, diags)
		return
	}

	// Apstra was expecting something, so set it
	err = client.SetResourceAllocation(ctx, rga)
	if err != nil {
		diags.AddError(errSettingAllocation, err.Error())
	}
}

func (o *rBlueprint) updateResourcePoolAllocationByTfsdkTag(ctx context.Context, tag string, client *goapstra.TwoStageL3ClosClient, state *rBlueprint, diags *diag.Diagnostics) {
	planPool := o.extractResourcePoolElementByTfsdkTag(tag, diags)
	statePool := state.extractResourcePoolElementByTfsdkTag(tag, diags)
	if diags.HasError() {
		return
	}

	if setOfAttrValuesMatch(planPool, statePool) {
		// no change; set plan = state
		o.setResourcePoolElementByTfsdkTag(tag, statePool, diags)
	} else {
		// edit needed
		o.setApstraPoolAllocationByTfsdkTag(ctx, tag, client, diags) // use plan to update apstra
		planPool.Unknown = false                                     // mark planed object as !Unknown
		o.setResourcePoolElementByTfsdkTag(tag, planPool, diags)     // update plan with planned object
	}
}

func (o *rBlueprint) readSwitchesFromGraphDb(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	// get the list of switch roles (spine1, leaf2...) from the blueprint
	switchLabels := listSwitches(ctx, client, goapstra.ObjectId(o.Id.Value), diags)
	if diags.HasError() {
		return
	}

	// collect switch info into newState
	o.Switches = make(map[string]types.Object, len(switchLabels))
	for _, switchLabel := range switchLabels {
		sfgdb := getSwitchFromGraphDb(ctx, client, goapstra.ObjectId(o.Id.Value), switchLabel, diags)
		if diags.HasError() {
			return
		}
		o.Switches[switchLabel] = *sfgdb
	}
}

func setBlueprintName(ctx context.Context, client *goapstra.TwoStageL3ClosClient, name string, diags *diag.Diagnostics) {
	type node struct {
		Label string            `json:"label,omitempty"`
		Id    goapstra.ObjectId `json:"id,omitempty"`
	}
	response := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}
	err := client.GetNodes(ctx, goapstra.NodeTypeMetadata, response)
	if err != nil {
		diags.AddError("error querying blueprint nodes", err.Error())
		return
	}
	if len(response.Nodes) != 1 {
		diags.AddError(fmt.Sprintf("wrong number of %s nodes", goapstra.NodeTypeMetadata.String()),
			fmt.Sprintf("expecting 1 got %d nodes", len(response.Nodes)))
		return
	}
	var nodeId goapstra.ObjectId
	for _, v := range response.Nodes {
		nodeId = v.Id
	}
	err = client.PatchNode(ctx, nodeId, &node{Label: name}, nil)
	if err != nil {
		diags.AddError("error setting blueprint name", err.Error())
		return
	}

}

func parseFabricAddressingPolicy(in types.String, diags *diag.Diagnostics) goapstra.AddressingScheme {
	if in.IsNull() {
		return defaultFabricAddressingPolicy
	}
	switch in.Value {
	case goapstra.AddressingSchemeIp4.String():
		return goapstra.AddressingSchemeIp4
	case goapstra.AddressingSchemeIp46.String():
		return goapstra.AddressingSchemeIp46
	case goapstra.AddressingSchemeIp6.String():
		return goapstra.AddressingSchemeIp6
	}
	diags.AddWarning(errProviderBug, fmt.Sprintf("cannot handle '%s' when parsing fabric addressing policy", in))
	return -1
}

// tfsdkTagToRgn is a simple lookup of tfsdk tag to goapstra.ResourceGroupName.
// Any lookup misses are a provider bug.
func tfsdkTagToRgn(tag string, diags *diag.Diagnostics) goapstra.ResourceGroupName {
	switch tag {
	case "superspine_asn_pool_ids":
		return goapstra.ResourceGroupNameSuperspineAsn
	case "spine_asn_pool_ids":
		return goapstra.ResourceGroupNameSpineAsn
	case "leaf_asn_pool_ids":
		return goapstra.ResourceGroupNameLeafAsn
	case "access_asn_pool_ids":
		return goapstra.ResourceGroupNameAccessAsn
	case "superspine_loopback_pool_ids":
		return goapstra.ResourceGroupNameSuperspineIp4
	case "spine_loopback_pool_ids":
		return goapstra.ResourceGroupNameSpineIp4
	case "leaf_loopback_pool_ids":
		return goapstra.ResourceGroupNameLeafIp4
	case "access_loopback_pool_ids":
		return goapstra.ResourceGroupNameAccessIp4
	case "superspine_spine_ip4_pool_ids":
		return goapstra.ResourceGroupNameSuperspineSpineIp4
	case "spine_leaf_ip4_pool_ids":
		return goapstra.ResourceGroupNameSpineLeafIp4
	case "leaf_leaf_ip4_pool_ids":
		return goapstra.ResourceGroupNameLeafLeafIp4
	case "leaf_mlag_peer_link_ip4_pool_ids":
		return goapstra.ResourceGroupNameMlagDomainSviSubnets
	case "access_esi_peer_link_ip4_pool_ids":
		return goapstra.ResourceGroupNameAccessAccessIps
	case "vtep_ip4_pool_ids":
		return goapstra.ResourceGroupNameVtepIps
	case "superspine_spine_ip6_pool_ids":
		return goapstra.ResourceGroupNameSuperspineSpineIp6
	case "spine_leaf_ip6_pool_ids":
		return goapstra.ResourceGroupNameSpineLeafIp6
	}
	diags.AddError(errProviderBug, fmt.Sprintf("tfsdk tag '%s' unknown", tag))
	return goapstra.ResourceGroupNameUnknown
}

// rgnToTfsdkTag is a simple lookup from goapstraResourceGroupName to the tfsdk
// tag used to represent it.
// Any lookup misses are a provider bug.
func rgnToTfsdkTag(rgn goapstra.ResourceGroupName, diags *diag.Diagnostics) string {
	switch rgn {
	case goapstra.ResourceGroupNameSuperspineAsn:
		return "superspine_asn_pool_ids"
	case goapstra.ResourceGroupNameSpineAsn:
		return "spine_asn_pool_ids"
	case goapstra.ResourceGroupNameLeafAsn:
		return "leaf_asn_pool_ids"
	case goapstra.ResourceGroupNameAccessAsn:
		return "access_asn_pool_ids"
	case goapstra.ResourceGroupNameSuperspineIp4:
		return "superspine_loopback_pool_ids"
	case goapstra.ResourceGroupNameSpineIp4:
		return "spine_loopback_pool_ids"
	case goapstra.ResourceGroupNameLeafIp4:
		return "leaf_loopback_pool_ids"
	case goapstra.ResourceGroupNameAccessIp4:
		return "access_loopback_pool_ids"
	case goapstra.ResourceGroupNameSuperspineSpineIp4:
		return "superspine_spine_ip4_pool_ids"
	case goapstra.ResourceGroupNameSpineLeafIp4:
		return "spine_leaf_ip4_pool_ids"
	case goapstra.ResourceGroupNameLeafLeafIp4:
		return "leaf_leaf_ip4_pool_ids"
	case goapstra.ResourceGroupNameMlagDomainSviSubnets:
		return "leaf_mlag_peer_link_ip4_pool_ids"
	case goapstra.ResourceGroupNameAccessAccessIps:
		return "access_esi_peer_link_ip4_pool_ids"
	case goapstra.ResourceGroupNameVtepIps:
		return "vtep_ip4_pool_ids"
	case goapstra.ResourceGroupNameSuperspineSpineIp6:
		return "superspine_spine_ip6_pool_ids"
	case goapstra.ResourceGroupNameSpineLeafIp6:
		return "spine_leaf_ip6_pool_ids"
	}
	diags.AddError(errProviderBug, fmt.Sprintf("resource group name '%s' unknown", rgn.String()))
	return ""
}

func warnMissingResourceGroupAllocations(ctx context.Context, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	allocations, err := client.GetResourceAllocations(ctx)
	if err != nil {
		diags.AddError("error fetching resource group allocations", err.Error())
		return
	}

	var missing []string
	for _, allocation := range allocations {
		if allocation.IsEmpty() {
			missing = append(missing, fmt.Sprintf("%q", rgnToTfsdkTag(allocation.ResourceGroup.Name, diags)))
		}
	}
	if len(missing) != 0 {
		diags.AddWarning(warnMissingResourceSummary, fmt.Sprintf(warnMissingResourceDetail, strings.Join(missing, ", ")))
	}
}

// listOfResourceGroupAllocationTags returns the full list of tfsdk tags
// representing potential resource group allocations for a "datacenter"
// blueprint. This could probably be rewritten as a "reflect" operation against
// an rBlueprint which extracts tags ending in "_pool_ids".
func listOfResourceGroupAllocationTags() []string {
	return []string{
		"superspine_asn_pool_ids",
		"spine_asn_pool_ids",
		"leaf_asn_pool_ids",
		"access_asn_pool_ids",
		"superspine_loopback_pool_ids",
		"spine_loopback_pool_ids",
		"leaf_loopback_pool_ids",
		"access_loopback_pool_ids",
		"superspine_spine_ip4_pool_ids",
		"spine_leaf_ip4_pool_ids",
		"leaf_leaf_ip4_pool_ids",
		"leaf_mlag_peer_link_ip4_pool_ids",
		"access_esi_peer_link_ip4_pool_ids",
		"vtep_ip4_pool_ids",
		"superspine_spine_ip6_pool_ids",
		"spine_leaf_ip6_pool_ids",
	}
}

// getAllSystemsInfo returns map[string]goapstra.ManagedSystemInfo keyed by
// device_key (switch serial number)
func getAllSystemsInfo(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) map[string]goapstra.ManagedSystemInfo {
	// pull SystemInfo for all switches managed by apstra
	asi, err := client.GetAllSystemsInfo(ctx) // pull all managed systems info from Apstra
	if err != nil {
		diags.AddError("get managed system info", err.Error())
		return nil
	}

	// organize the []ManagedSystemInfo into a map by device key (serial number)
	deviceKeyToSystemInfo := make(map[string]goapstra.ManagedSystemInfo, len(asi)) // map-ify the Apstra output
	for _, si := range asi {
		deviceKeyToSystemInfo[si.DeviceKey] = si
	}
	return deviceKeyToSystemInfo
}

func newSwitchObject() types.Object {
	return types.Object{
		AttrTypes: switchElementSchema(),
		Attrs:     make(map[string]attr.Value, len(switchElementSchema())),
	}
}

func switchElementSchema() map[string]attr.Type {
	return map[string]attr.Type{
		"interface_map_id":  types.StringType,
		"device_key":        types.StringType,
		"device_profile_id": types.StringType,
		"system_node_id":    types.StringType,
	}
}

// listSwitches returns a []string enumerating switch roles (spine1,
// leaf2_1, etc...) in the indicated blueprint
func listSwitches(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, diags *diag.Diagnostics) []string {
	var switchQr struct {
		Items []struct {
			System struct {
				Label string `json:"label"`
				Id    string `json:"id"`
			} `json:"n_system"`
		} `json:"items"`
	}
	err := client.NewQuery(bpId).
		SetContext(ctx).
		Node([]goapstra.QEEAttribute{
			{"type", goapstra.QEStringVal("system")},
			{"name", goapstra.QEStringVal("n_system")},
			{"system_type", goapstra.QEStringVal("switch")},
		}).
		Do(&switchQr)
	if err != nil {
		diags.AddError("error querying graphdb for switch nodes", err.Error())
		return nil
	}

	result := make([]string, len(switchQr.Items))
	for i := range switchQr.Items {
		result[i] = switchQr.Items[i].System.Label
	}
	return result
}

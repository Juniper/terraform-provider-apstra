package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

var _ resource.ResourceWithConfigure = &resourceBlueprint{}
var _ resource.ResourceWithValidateConfig = &resourceBlueprint{}

type resourceBlueprint struct {
	client *goapstra.Client
}

func (o *resourceBlueprint) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (o *resourceBlueprint) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"spine_asn_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the ASN Pool(s) to be used on spine switches "+
					"in blueprints built from `%s` or `%s` templates.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypeRackBased),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"leaf_asn_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the ASN Pool(s) to be used on leaf switches "+
					"in blueprints built from `%s`, `%s` or `%s` templates.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypeRackBased, goapstra.TemplateTypeL3Collapsed),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"access_asn_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the ASN Pool(s) to be used on access switches "+
					"in blueprints featuring access switches concigured for `%s` redundancy mode.",
					goapstra.AccessRedundancyProtocolEsi),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},

			"superspine_loopback_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for superspine switch "+
					"loopback interfaces in blueprints built from %s templates.", goapstra.TemplateTypePodBased),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"spine_loopback_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for spine switch "+
					"loopback interfaces in blueprints built from %s or %s templates.",
					goapstra.TemplateTypePodBased.String(), goapstra.TemplateTypeRackBased),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"leaf_loopback_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for leaf switch "+
					"loopback interfaces in blueprints built from %s, %s or %s templates.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypeRackBased, goapstra.TemplateTypeL3Collapsed),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"access_loopback_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for access switch "+
					"peer-link interfaces in blueprints featuring access switches concigured for `%s` redundancy mode.",
					goapstra.AccessRedundancyProtocolEsi),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},

			"superspine_spine_ip4_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for superspine/spine "+
					"fabric links in blueprints built from `%s` templates.", goapstra.TemplateTypePodBased),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"spine_leaf_ip4_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for superspine/spine "+
					"fabric links in blueprints built from `%s` or `%s` templates.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypeRackBased),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"leaf_leaf_ip4_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for leaf/leaf "+
					"fabric links in blueprints built from `%s` templates.", goapstra.TemplateTypeL3Collapsed),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},

			"leaf_mlag_peer_link_ip4_pool_ids": {
				// todo validate
				MarkdownDescription: "ID(s) of the IPv4 Pool(s) to be used on MLAG peer links between leaf switches.",
				Type:                types.SetType{ElemType: types.StringType},
				Optional:            true,
				Computed:            true,
				Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"access_esi_peer_link_ip4_pool_ids": {
				// todo validate
				MarkdownDescription: "ID(s) of the IPv4 Pool(s) to be used on ESI LAG peer links between access switches.",
				Type:                types.SetType{ElemType: types.StringType},
				Optional:            true,
				Computed:            true,
				Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"vtep_ip4_pool_ids": {
				// todo validate
				MarkdownDescription: "Unclear what this is for.", //todo
				Type:                types.SetType{ElemType: types.StringType},
				Optional:            true,
				Computed:            true,
				Validators:          []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},

			"superspine_spine_ip6_pool_ids": {
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for superspine/spine "+
					"fabric links in blueprints built from `%s` templates and using addressing mode `%s` or `%s`.",
					goapstra.TemplateTypePodBased, goapstra.AddressingSchemeIp6, goapstra.AddressingSchemeIp46),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
			},
			"spine_leaf_ip6_pool_ids": {
				// todo validate
				MarkdownDescription: fmt.Sprintf("ID(s) of the IPv4 Pool(s) to be used for spine/leaf fabric "+
					"links in blueprints built from `%s` or `%s` templates and using addressing mode `%s` or `%s`.",
					goapstra.TemplateTypePodBased, goapstra.TemplateTypePodBased,
					goapstra.AddressingSchemeIp6, goapstra.AddressingSchemeIp46),
				Type:       types.SetType{ElemType: types.StringType},
				Optional:   true,
				Computed:   true,
				Validators: []tfsdk.AttributeValidator{setvalidator.SizeAtLeast(1)},
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
			//"switches": {
			//	Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
			//		"interface_map_id": {
			//			Type:          types.StringType,
			//			Optional:      true,
			//			Computed:      true,
			//			PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			//		},
			//		"device_key": {
			//			Type:     types.StringType,
			//			Required: true,
			//		},
			//		"device_profile": {
			//			Type:     types.StringType,
			//			Computed: true,
			//		},
			//		"system_node_id": {
			//			Type:          types.StringType,
			//			Computed:      true,
			//			PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
			//		},
			//	}),
			//	Optional: true,
			//},
		},
	}, nil
}

func (o *resourceBlueprint) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config rBlueprint
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if o.client == nil { // cannot proceed without a client
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

	// validate addressing schemes are appropriate for the template type
	config.validateConfigAddressingSchemes(&resp.Diagnostics)

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
	ip4Pools = append(ip4Pools, config.LeafLeafPoolIpv4.Elems...)       // leaf-only fabric
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
}

func (o *resourceBlueprint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
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

	//// ensure switches (by device key) exist on Apstra
	//asi, err := r.p.client.GetAllSystemsInfo(ctx) // pull all managed systems info from Apstra
	//if err != nil {
	//	resp.Diagnostics.AddError("get managed system info", err.Error())
	//	return
	//}
	//deviceKeyToSystemInfo := make(map[string]*goapstra.ManagedSystemInfo) // map-ify the Apstra output
	//for _, si := range asi {
	//	deviceKeyToSystemInfo[si.DeviceKey] = &si
	//}
	//// check each planned switch exists in Apstra, and save the aos_hcl_model (device profile)
	//for switchLabel, switchFromPlan := range plan.Switches {
	//	if si, found := deviceKeyToSystemInfo[switchFromPlan.DeviceKey.Value]; !found {
	//		resp.Diagnostics.AddError("switch not found",
	//			fmt.Sprintf("no switch with device_key '%s' exists on Apstra", switchFromPlan.DeviceKey.Value))
	//		return
	//	} else {
	//		switchFromPlan.DeviceProfile = types.String{Value: si.Facts.AosHclModel}
	//		plan.Switches[switchLabel] = switchFromPlan
	//	}
	//}

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
		return
	}

	// create a client specific to the reference design
	blueprint, err := o.client.NewTwoStageL3ClosClient(ctx, blueprintId)
	if err != nil {
		resp.Diagnostics.AddError("error getting blueprint client", err.Error())
		return
	}

	// set user-configured resource group allocations
	plan.setResourceAllocations(ctx, blueprint, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//// warn about switches discovered in the graph db, and which do not appear in the tf config
	//err = warnAboutSwitchesMissingFromPlan(ctx, r.p.client, blueprintId, plan.Switches, &resp.Diagnostics)
	//if err != nil {
	//	resp.Diagnostics.AddError("error while inventorying switches after blueprint creation", err.Error())
	//	return
	//}

	//// structure we'll use when assigning interface maps to switches
	//ifmapAssignments := make(goapstra.SystemIdToInterfaceMapAssignment)
	//
	//// assign details of each configured switch (don't add elements to the plan.Switches map)
	////	- DeviceKey : required user input
	////	- InterfaceMap : optional user input - if only one option, we'll auto-assign
	////	- DeviceProfile : a.k.a. aos_hcl_model - determined from InterfaceMap, represents physical device/model
	////	- SystemNodeId : id of the "type='system', system_type="switch" graph db node representing a spine/leaf/etc...
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

	diags = resp.State.Set(ctx, &rBlueprint{
		Id:           types.String{Value: string(blueprintId)},
		Name:         types.String{Value: plan.Name.Value},
		TemplateId:   types.String{Value: plan.TemplateId.Value},
		TemplateType: types.String{Value: templateType.String()},

		SuperspineAsnPoolIds:   plan.SuperspineAsnPoolIds,   // ASN
		SpineAsnPoolIds:        plan.SpineAsnPoolIds,        // ASN
		LeafAsnPoolIds:         plan.LeafAsnPoolIds,         // ASN
		AccessAsnPoolIds:       plan.AccessAsnPoolIds,       // ASN
		SuperspineIp4PoolIds:   plan.SuperspineIp4PoolIds,   // loopback
		SpineIp4PoolIds:        plan.SpineIp4PoolIds,        // loopback
		LeafIp4PoolIds:         plan.LeafIp4PoolIds,         // loopback
		AccessIp4PoolIds:       plan.AccessIp4PoolIds,       // loopback
		SuperspineSpinePoolIp4: plan.SuperspineSpinePoolIp4, // fabric v4
		SpineLeafPoolIp4:       plan.SpineLeafPoolIp4,       // fabric v4
		LeafLeafPoolIpv4:       plan.LeafLeafPoolIpv4,       // fabric v4
		LeafMlagPeerIp4:        plan.LeafMlagPeerIp4,        // other v4
		AccessEsiPeerIp4:       plan.AccessIp4PoolIds,       // other v4
		VtepIps:                plan.VtepIps,                // other v4
		SuperspineSpinePoolIp6: plan.SuperspineSpinePoolIp6, // fabric v6
		SpineLeafPoolIp6:       plan.SpineLeafPoolIp6,       // fabric v6
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (o *resourceBlueprint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
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

	state.Name = types.String{Value: blueprintStatus.Label}
	state.getResourceAllocations(ctx, blueprint, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	//// get switch info
	//for switchLabel, stateSwitch := range state.Switches {
	//	// assign details of each known switch (don't add elements to the state.Switches map)
	//	//	- DeviceKey : required user input
	//	//	- InterfaceMap : optional user input - if only one option, we'll auto-assign
	//	//	- DeviceProfile : a.k.a. aos_hcl_model - determined from InterfaceMap, represents physical device/model
	//	//	- SystemNodeId : id of the "type='system', system_type="switch" graph db node representing a spine/leaf/etc...
	//	systemInfo, err := getSystemNodeInfo(ctx, r.p.client, blueprintStatus.Id, switchLabel)
	//	if err != nil {
	//		resp.Diagnostics.AddError(
	//			fmt.Sprintf("error while reading info for system node '%s'", switchLabel),
	//			err.Error())
	//	}
	//	stateSwitch.SystemNodeId = types.String{Value: systemInfo.id}
	//	stateSwitch.DeviceKey = types.String{Value: systemInfo.systemId}
	//	interfaceMap, err := getNodeInterfaceMap(ctx, r.p.client, blueprintStatus.Id, switchLabel)
	//	if err != nil {
	//		resp.Diagnostics.AddError(
	//			fmt.Sprintf("error while reading interface map for node '%s'", switchLabel),
	//			err.Error())
	//	}
	//	stateSwitch.InterfaceMap = types.String{Value: interfaceMap.label}
	//	stateSwitch.DeviceProfile = types.String{Value: interfaceMap.deviceProfileId}
	//	state.Switches[switchLabel] = stateSwitch
	//}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (o *resourceBlueprint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
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

	//// ensure planned switches (by device key) exist on Apstra
	//asi, err := o.client.GetAllSystemsInfo(ctx) // pull all managed systems info from Apstra
	//if err != nil {
	//	resp.Diagnostics.AddError("get managed system info", err.Error())
	//	return
	//}
	//deviceKeyToSystemInfo := make(map[string]*goapstra.ManagedSystemInfo) // map-ify the Apstra output
	//for _, si := range asi {
	//	deviceKeyToSystemInfo[si.DeviceKey] = &si
	//}
	//// check each planned switch exists in Apstra, and save the aos_hcl_model (device profile)
	//for switchLabel, switchFromPlan := range plan.Switches {
	//	if si, found := deviceKeyToSystemInfo[switchFromPlan.DeviceKey.Value]; !found {
	//		resp.Diagnostics.AddError("switch not found",
	//			fmt.Sprintf("no switch with device_key '%s' exists on Apstra", switchFromPlan.DeviceKey.Value))
	//		return
	//	} else {
	//		switchFromPlan.DeviceProfile = types.String{Value: si.Facts.AosHclModel}
	//		plan.Switches[switchLabel] = switchFromPlan
	//	}
	//}

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

	if !setOfAttrValuesMatch(plan.SuperspineAsnPoolIds, state.SuperspineAsnPoolIds) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameSuperspineAsn, plan.LeafAsnPoolIds, &resp.Diagnostics)
	}
	if !setOfAttrValuesMatch(plan.SpineAsnPoolIds, state.SpineAsnPoolIds) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameSpineAsn, plan.LeafAsnPoolIds, &resp.Diagnostics)
	}
	if !setOfAttrValuesMatch(plan.LeafAsnPoolIds, state.LeafAsnPoolIds) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameLeafAsn, plan.LeafAsnPoolIds, &resp.Diagnostics)
	}
	if !setOfAttrValuesMatch(plan.AccessAsnPoolIds, state.AccessAsnPoolIds) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameAccessAsn, plan.LeafAsnPoolIds, &resp.Diagnostics)
	}

	if !setOfAttrValuesMatch(plan.SuperspineIp4PoolIds, state.SuperspineIp4PoolIds) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameSuperspineIp4, plan.LeafIp4PoolIds, &resp.Diagnostics)
	}
	if !setOfAttrValuesMatch(plan.SpineIp4PoolIds, state.SpineIp4PoolIds) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameSpineIp4, plan.LeafIp4PoolIds, &resp.Diagnostics)
	}
	if !setOfAttrValuesMatch(plan.LeafIp4PoolIds, state.LeafIp4PoolIds) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameLeafIp4, plan.LeafIp4PoolIds, &resp.Diagnostics)
	}
	if !setOfAttrValuesMatch(plan.AccessIp4PoolIds, state.AccessIp4PoolIds) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameAccessIp4, plan.LeafIp4PoolIds, &resp.Diagnostics)
	}

	if !setOfAttrValuesMatch(plan.SuperspineSpinePoolIp4, state.SuperspineSpinePoolIp4) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameSuperspineSpineIp4, plan.SuperspineSpinePoolIp4, &resp.Diagnostics)
	}
	if !setOfAttrValuesMatch(plan.SpineLeafPoolIp4, state.SpineLeafPoolIp4) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameSpineLeafIp4, plan.SpineLeafPoolIp4, &resp.Diagnostics)
	}
	if !setOfAttrValuesMatch(plan.AccessEsiPeerIp4, state.AccessEsiPeerIp4) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameAccessAccessIps, plan.AccessEsiPeerIp4, &resp.Diagnostics)
	}

	if !setOfAttrValuesMatch(plan.SuperspineSpinePoolIp6, state.SuperspineSpinePoolIp6) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameSuperspineSpineIp6, plan.SuperspineSpinePoolIp6, &resp.Diagnostics)
	}
	if !setOfAttrValuesMatch(plan.SpineLeafPoolIp6, state.SpineLeafPoolIp6) {
		setResourceAllocation(ctx, blueprint, goapstra.ResourceGroupNameSpineLeafIp6, plan.SpineLeafPoolIp6, &resp.Diagnostics)
	}

	//// combine switch labels from plan and state into a single set (map of empty struct)
	//combinedSwitchLabels := make(map[string]struct{})
	//for stateSwitchLabel := range state.Switches {
	//	combinedSwitchLabels[stateSwitchLabel] = struct{}{}
	//}
	//for planSwitchLabel := range plan.Switches {
	//	combinedSwitchLabels[planSwitchLabel] = struct{}{}
	//}

	//// structure we'll use when assigning interface maps to switches
	//ifmapReassignments := make(goapstra.SystemIdToInterfaceMapAssignment)

	//// loop over all switches: plan and/or state
	//for switchLabel := range combinedSwitchLabels {
	//	// compare details of each switch
	//	//	- DeviceKey : required user input - changeable
	//	//	- InterfaceMap : optional user input - changeable
	//	//	- DeviceProfile : a.k.a. aos_hcl_model - changeable
	//	//	- SystemNodeId : id of the "type='system', system_type="switch" graph db node representing a spine/leaf/etc...
	//
	//	// fetch the switch graph db node ID and candidate interface maps
	//	systemNodeId, ifmapCandidates, err := getSystemNodeIdAndIfmapCandidates(ctx, r.p.client, goapstra.ObjectId(state.Id.Value), switchLabel)
	//	if err != nil {
	//		resp.Diagnostics.AddWarning("error fetching interface map candidates", err.Error())
	//		continue
	//	}
	//	var foundInPlan, foundInState bool
	//	var planSwitch, stateSwitch Switch
	//	planSwitch, foundInPlan = plan.Switches[switchLabel]
	//	stateSwitch, foundInState = state.Switches[switchLabel]
	//	switch {
	//	case foundInPlan && foundInState: // the normal case: switch exists in plan and state
	//		if planSwitch.SystemNodeId.Value != stateSwitch.SystemNodeId.Value {
	//			resp.Diagnostics.AddError(
	//				fmt.Sprintf("node graph entry for %s changed", switchLabel),
	//				fmt.Sprintf("change: '%s'->'%s' this isn't supposed to happen",
	//					planSwitch.SystemNodeId.Value, stateSwitch.SystemNodeId.Value))
	//			return
	//		}
	//		if (planSwitch.DeviceKey.Value != stateSwitch.DeviceKey.Value) || // device change?
	//			(planSwitch.InterfaceMap.Value != stateSwitch.InterfaceMap.Value) {
	//			// clear existing system id from switch node
	//			var patch struct {
	//				SystemId interface{} `json:"system_id"`
	//			}
	//			patch.SystemId = nil
	//			err = r.p.client.PatchNode(ctx, goapstra.ObjectId(plan.Id.Value), goapstra.ObjectId(planSwitch.SystemNodeId.Value), &patch, nil)
	//			if err != nil {
	//				resp.Diagnostics.AddWarning("failed to revoke switch device", err.Error())
	//			}
	//
	//			// proceed as in Create()
	//			// validate/choose interface map, build ifmap assignment structure
	//			if !planSwitch.InterfaceMap.Null && !planSwitch.InterfaceMap.Unknown && !(planSwitch.InterfaceMap.Value == "") {
	//				// user gave us an interface map label they'd like to use
	//				ifmapNodeId := ifmapCandidateFromCandidates(planSwitch.InterfaceMap.Value, ifmapCandidates)
	//				if ifmapNodeId != nil {
	//					ifmapReassignments[systemNodeId] = ifmapNodeId.id
	//					planSwitch.DeviceProfile = types.String{Value: ifmapNodeId.deviceProfileId}
	//				} else {
	//					resp.Diagnostics.AddWarning(
	//						"invalid interface map",
	//						fmt.Sprintf("interface map '%s' not found among candidates for node '%s'",
	//							planSwitch.InterfaceMap.Value, switchLabel))
	//				}
	//			} else {
	//				// user didn't give us an interface map label; try to find a default
	//				switch len(ifmapCandidates) {
	//				case 0: // no candidates!
	//					resp.Diagnostics.AddWarning(
	//						"interface map not specified, and no candidates found",
	//						fmt.Sprintf("no candidate interface maps found for node '%s'", switchLabel))
	//				case 1: // exact match; we can work with this
	//					ifmapReassignments[systemNodeId] = ifmapCandidates[0].id
	//					planSwitch.InterfaceMap = types.String{Value: ifmapCandidates[0].label}
	//					planSwitch.DeviceProfile = types.String{Value: ifmapCandidates[0].deviceProfileId}
	//				default: // multiple match!
	//					sb := strings.Builder{}
	//					sb.WriteString(fmt.Sprintf("'%s'", ifmapCandidates[0].label))
	//					for _, candidate := range ifmapCandidates[1:] {
	//						sb.WriteString(fmt.Sprintf(", '%s'", candidate.label))
	//					}
	//					resp.Diagnostics.AddWarning(
	//						"cannot assign interface map",
	//						fmt.Sprintf("node '%s' has %d interface map candidates. Please choose one of ['%s']",
	//							switchLabel, len(ifmapCandidates), sb.String()))
	//				}
	//			}
	//		}
	//		state.Switches[switchLabel] = planSwitch
	//
	//	case foundInPlan && !foundInState: // new switch
	//		// save the SystemNodeId (1:1 relationship with switchLabel in graph db)
	//		planSwitch.SystemNodeId = types.String{Value: systemNodeId}
	//
	//		// validate/choose interface map, build ifmap assignment structure
	//		if !planSwitch.InterfaceMap.Null && !planSwitch.InterfaceMap.Unknown && !(planSwitch.InterfaceMap.Value == "") {
	//			// user gave us an interface map label they'd like to use
	//			ifmapNodeId := ifmapCandidateFromCandidates(planSwitch.InterfaceMap.Value, ifmapCandidates)
	//			if ifmapNodeId != nil {
	//				ifmapReassignments[systemNodeId] = ifmapNodeId.id
	//				planSwitch.DeviceProfile = types.String{Value: ifmapNodeId.deviceProfileId}
	//			} else {
	//				resp.Diagnostics.AddWarning(
	//					"invalid interface map",
	//					fmt.Sprintf("interface map '%s' not found among candidates for node '%s'",
	//						planSwitch.InterfaceMap.Value, switchLabel))
	//			}
	//		} else {
	//			// user didn't give us an interface map label; try to find a default
	//			switch len(ifmapCandidates) {
	//			case 0: // no candidates!
	//				resp.Diagnostics.AddWarning(
	//					"interface map not specified, and no candidates found",
	//					fmt.Sprintf("no candidate interface maps found for node '%s'", switchLabel))
	//			case 1: // exact match; we can work with this
	//				ifmapReassignments[systemNodeId] = ifmapCandidates[0].id
	//				planSwitch.InterfaceMap = types.String{Value: ifmapCandidates[0].label}
	//				planSwitch.DeviceProfile = types.String{Value: ifmapCandidates[0].deviceProfileId}
	//			default: // multiple match!
	//				sb := strings.Builder{}
	//				sb.WriteString(fmt.Sprintf("'%s'", ifmapCandidates[0].label))
	//				for _, candidate := range ifmapCandidates[1:] {
	//					sb.WriteString(fmt.Sprintf(", '%s'", candidate.label))
	//				}
	//				resp.Diagnostics.AddWarning(
	//					"cannot assign interface map",
	//					fmt.Sprintf("node '%s' has %d interface map candidates. Please choose one of ['%s']",
	//						switchLabel, len(ifmapCandidates), sb.String()))
	//			}
	//		}
	//
	//		state.Switches[switchLabel] = stateSwitch
	//
	//	case !foundInPlan && foundInState: // deleted switch
	//		resp.Diagnostics.AddWarning("request to delete switch not yet supported",
	//			fmt.Sprintf("cannot delete removed switch '%s'", switchLabel))
	//	}
	//}

	//// assign previously-selected interface maps
	//err = refDesignClient.SetInterfaceMapAssignments(ctx, ifmapReassignments)
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
	//	err = r.p.client.PatchNode(ctx, goapstra.ObjectId(state.Id.Value), goapstra.ObjectId(switchPlan.SystemNodeId.Value), &patch, nil)
	//	if err != nil {
	//		resp.Diagnostics.AddWarning("failed to assign switch device", err.Error())
	//	}
	//}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (o *resourceBlueprint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
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

func ip4PoolsExist(ctx context.Context, in []types.String, client *goapstra.Client) (string, error) {
	poolsPerApi, err := client.ListIp4PoolIds(ctx)
	if err != nil {
		return "", fmt.Errorf("error listing available resource pool IDs")
	}

testPool:
	for _, testPool := range in {
		for _, apiPool := range poolsPerApi {
			if goapstra.ObjectId(testPool.Value) == apiPool {
				continue testPool // this one's good, check the next testPool
			}
		}
		return testPool.Value, nil // we looked at every apiPool, none matched testPool
	}
	return "", nil
}

//type switchLabelToInterfaceMap map[string]struct {
//	label string
//	id    string
//}

// getSwitchLabelId queries the graph db for 'switch' type systems, returns
// map[string]string (map[label]id)
func getSwitchLabelId(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId) (map[string]string, error) {
	var switchQr struct {
		Count int `json:"count"`
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

	result := make(map[string]string, switchQr.Count)
	for _, item := range switchQr.Items {
		result[item.System.Label] = item.System.Id
	}

	return result, nil
}

//type switchLabelToCandidateInterfaceMaps map[string][]struct {
//	Id    string
//	Label string
//}
//
//func (o *switchLabelToCandidateInterfaceMaps) string() (string, error) {
//	data, err := json.Marshal(o)
//	if err != nil {
//		return "", err
//	}
//	return string(data), nil
//}

//// getSwitchCandidateInterfaceMaps queries the graph db for
//// 'switch' type systems and their candidate interface maps.
//// It returns switchLabelToCandidateInterfaceMaps.
//func getSwitchCandidateInterfaceMaps(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId) (switchLabelToCandidateInterfaceMaps, error) {
//	var candidateInterfaceMapsQR struct {
//		Items []struct {
//			System struct {
//				Label string `json:"label"`
//			} `json:"n_system"`
//			InterfaceMap struct {
//				Id    string `json:"id"`
//				Label string `json:"label"`
//			} `json:"n_interface_map"`
//		} `json:"items"`
//	}
//	err := client.NewQuery(bpId).
//		SetContext(ctx).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("system")},
//			{"name", goapstra.QEStringVal("n_system")},
//			{"system_type", goapstra.QEStringVal("switch")},
//		}).
//		Out([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("logical_device")},
//		}).
//		In([]goapstra.QEEAttribute{{"type", goapstra.QEStringVal("logical_device")}}).
//		Node([]goapstra.QEEAttribute{
//			{"type", goapstra.QEStringVal("interface_map")},
//			{"name", goapstra.QEStringVal("n_interface_map")},
//		}).
//		Do(&candidateInterfaceMapsQR)
//	if err != nil {
//		return nil, err
//	}
//
//	result := make(switchLabelToCandidateInterfaceMaps)
//
//	for _, item := range candidateInterfaceMapsQR.Items {
//		mapEntry := result[item.System.Label]
//		mapEntry = append(mapEntry, struct {
//			Id    string
//			Label string
//		}{Id: item.InterfaceMap.Id, Label: item.InterfaceMap.Label})
//		result[item.System.Label] = mapEntry
//	}
//
//	return result, nil
//}

func warnAboutSwitchesMissingFromPlan(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, switches map[string]Switch, diag *diag.Diagnostics) error {
	switchLabelToGraphDbId, err := getSwitchLabelId(ctx, client, bpId)
	if err != nil {
		return err
	}
	var missing []string
	for switchLabel := range switchLabelToGraphDbId {
		if _, found := switches[switchLabel]; !found {
			missing = append(missing, switchLabel)
		}
	}
	if len(missing) != 0 {
		diag.AddWarning("switch missing from plan",
			fmt.Sprintf("blueprint expects the following switches: ['%s']", strings.Join(missing, "', '")))
	}
	return nil
}

type ifmapInfo struct {
	id              string
	label           string
	deviceProfileId string
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

type systemNodeInfo struct {
	id       string
	label    string
	systemId string
}

func getSystemNodeInfo(ctx context.Context, client *goapstra.Client, bpId goapstra.ObjectId, label string) (*systemNodeInfo, error) {
	var systemQR struct {
		Items []struct {
			System struct {
				Id       string `json:"id"`
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
		}).Do(&systemQR)
	if err != nil {
		return nil, err
	}
	if len(systemQR.Items) != 1 {
		return nil, fmt.Errorf("expected exactly one system node, got %d", len(systemQR.Items))
	}
	return &systemNodeInfo{
		id:       systemQR.Items[0].System.Id,
		label:    systemQR.Items[0].System.Label,
		systemId: systemQR.Items[0].System.SystemID,
	}, nil
}

type rBlueprint struct {
	Id                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	TemplateId                types.String `tfsdk:"template_id"`
	TemplateType              types.String `tfsdk:"template_type"`
	SuperspineSpineAddressing types.String `tfsdk:"superspine_spine_addressing"`
	SpineLeafAddressing       types.String `tfsdk:"spine_leaf_addressing"`

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
	LeafLeafPoolIpv4       types.Set `tfsdk:"leaf_leaf_ip4_pool_ids"`        // fabric leaf/leaf

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
		if !o.LeafLeafPoolIpv4.Null {
			diags.AddError(
				errInvalidConfig,
				fmt.Sprintf(errTemplateTypeInvalidElement,
					o.TemplateId.Value,
					o.TemplateType.Value,
					getTfsdkTag(o, "LeafLeafPoolIpv4", diags)))
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

func (o *rBlueprint) setResourceAllocations(ctx context.Context, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	hashResourceGroup := func(in *goapstra.ResourceGroup) [16]byte {
		return md5.Sum([]byte(in.Type.String() + ":" + in.Name.String()))
	}

	newRga := func(name goapstra.ResourceGroupName, poolIds types.Set) *goapstra.ResourceGroupAllocation {
		return &goapstra.ResourceGroupAllocation{
			ResourceGroup: goapstra.ResourceGroup{
				Type: resourceTypeNameFromResourceGroupName(name, diags),
				Name: name,
			},
			PoolIds: sliceAttrValueToSliceObjectId(poolIds.Elems),
		}
	}

	// rgaMap is a map of all resource group allocations we might possibly request,
	// keyed by tfsdk tag. At least some of these will be Empty or Null.
	rgaMap := make(map[string]goapstra.ResourceGroupAllocation, 16) // 16 groups below

	// add switch ASN resource pools to the resource group allocation slice
	rgaMap[getTfsdkTag(o, "SuperspineAsnPoolIds", diags)] = *newRga(goapstra.ResourceGroupNameSuperspineAsn, o.SuperspineAsnPoolIds)
	rgaMap[getTfsdkTag(o, "SpineAsnPoolIds", diags)] = *newRga(goapstra.ResourceGroupNameSpineAsn, o.SpineAsnPoolIds)
	rgaMap[getTfsdkTag(o, "LeafAsnPoolIds", diags)] = *newRga(goapstra.ResourceGroupNameLeafAsn, o.LeafAsnPoolIds)
	rgaMap[getTfsdkTag(o, "AccessAsnPoolIds", diags)] = *newRga(goapstra.ResourceGroupNameAccessAsn, o.AccessAsnPoolIds)

	// add switch loopback resource pools to the resource group allocation slice
	rgaMap[getTfsdkTag(o, "SuperspineIp4PoolIds", diags)] = *newRga(goapstra.ResourceGroupNameSuperspineIp4, o.SuperspineIp4PoolIds)
	rgaMap[getTfsdkTag(o, "SpineIp4PoolIds", diags)] = *newRga(goapstra.ResourceGroupNameSpineIp4, o.SpineIp4PoolIds)
	rgaMap[getTfsdkTag(o, "LeafIp4PoolIds", diags)] = *newRga(goapstra.ResourceGroupNameLeafIp4, o.LeafIp4PoolIds)
	rgaMap[getTfsdkTag(o, "AccessIp4PoolIds", diags)] = *newRga(goapstra.ResourceGroupNameAccessIp4, o.AccessIp4PoolIds)

	// add fabric ipv4 resource pools to the resource group allocation slice
	rgaMap[getTfsdkTag(o, "SuperspineSpinePoolIp4", diags)] = *newRga(goapstra.ResourceGroupNameSuperspineSpineIp4, o.SuperspineSpinePoolIp4)
	rgaMap[getTfsdkTag(o, "SpineLeafPoolIp4", diags)] = *newRga(goapstra.ResourceGroupNameSpineLeafIp4, o.SpineLeafPoolIp4)
	rgaMap[getTfsdkTag(o, "LeafLeafPoolIpv4", diags)] = *newRga(goapstra.ResourceGroupNameLeafLeafIp4, o.LeafLeafPoolIpv4)

	// add other ipv4 resource pools to the resource group allocation slice
	rgaMap[getTfsdkTag(o, "LeafMlagPeerIp4", diags)] = *newRga(goapstra.ResourceGroupNameMlagDomainSviSubnets, o.LeafMlagPeerIp4)
	rgaMap[getTfsdkTag(o, "AccessEsiPeerIp4", diags)] = *newRga(goapstra.ResourceGroupNameAccessAccessIps, o.AccessEsiPeerIp4)
	rgaMap[getTfsdkTag(o, "VtepIps", diags)] = *newRga(goapstra.ResourceGroupNameVtepIps, o.VtepIps)

	// add fabric ipv6 resource pools to the resource group allocation slice
	rgaMap[getTfsdkTag(o, "SuperspineSpinePoolIp6", diags)] = *newRga(goapstra.ResourceGroupNameSuperspineSpineIp6, o.SuperspineSpinePoolIp6)
	rgaMap[getTfsdkTag(o, "SpineLeafPoolIp6", diags)] = *newRga(goapstra.ResourceGroupNameSpineLeafIp6, o.SpineLeafPoolIp6)

	// fetch all resource group allocations the blueprint is expecting
	blueprintResources, err := client.GetAllResourceAllocations(ctx)
	if err != nil {
		diags.AddError("error fetching blueprint resource allocations", err.Error())
		return
	}

	// convert to map for easy lookup
	rgHashMap := make(map[[16]byte]struct{}, len(blueprintResources))
	for _, rg := range blueprintResources {
		rgHashMap[hashResourceGroup(&rg.ResourceGroup)] = struct{}{}
	}

	for k, rga := range rgaMap {
		if rga.PoolIds == nil || len(rga.PoolIds) == 0 {
			continue // skip empty or nil groups
		}
		if _, ok := rgHashMap[hashResourceGroup(&rga.ResourceGroup)]; !ok {
			diags.AddWarning(
				"unused resource group",
				fmt.Sprintf("blueprint '%s' doesn't need resource group '%s'", client.Id(), k))
			continue
		}
		err = client.SetResourceAllocation(ctx, &rga)
		if err != nil {
			diags.AddError("error setting resource group allocation", err.Error())
		}
	}
}

func (o *rBlueprint) getResourceAllocations(ctx context.Context, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	// newSet turns a ResourceGroupAllocation into a types.Set with ElemType StringType
	newSet := func(in *goapstra.ResourceGroupAllocation) types.Set {
		stringers := make([]fmt.Stringer, len(in.PoolIds))
		for i, o := range in.PoolIds {
			stringers[i] = o
		}
		return types.Set{
			ElemType: types.StringType,
			Elems:    sliceStringerToSliceAttrValue(stringers),
		}
	}

	o.SuperspineAsnPoolIds = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameSuperspineAsn, diags))        // ASN
	o.SpineAsnPoolIds = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameSpineAsn, diags))                  // ASN
	o.LeafAsnPoolIds = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameLeafAsn, diags))                    // ASN
	o.AccessAsnPoolIds = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameAccessAsn, diags))                // ASN
	o.SuperspineIp4PoolIds = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameSuperspineIp4, diags))        // loopback
	o.SpineIp4PoolIds = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameSpineIp4, diags))                  // loopback
	o.LeafIp4PoolIds = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameLeafIp4, diags))                    // loopback
	o.AccessIp4PoolIds = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameAccessIp4, diags))                // loopback
	o.SuperspineSpinePoolIp4 = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameSuperspineSpineIp4, diags)) // fabric 4
	o.SpineLeafPoolIp4 = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameSpineLeafIp4, diags))             // fabric 4
	o.LeafLeafPoolIpv4 = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameLeafLeafIp4, diags))              // fabric 4
	o.LeafMlagPeerIp4 = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameMlagDomainSviSubnets, diags))      // other 4
	o.AccessEsiPeerIp4 = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameAccessAccessIps, diags))          // other 4
	o.VtepIps = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameVtepIps, diags))                           // other 4
	o.SuperspineSpinePoolIp6 = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameSuperspineSpineIp6, diags)) // fabric 6
	o.SpineLeafPoolIp6 = newSet(getResourceAllocation(ctx, client, goapstra.ResourceGroupNameSpineLeafIp6, diags))             // fabric 6
}

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
	case goapstra.ResourceGroupNameAccessAccessIps:
		return goapstra.ResourceTypeIp4Pool

	case goapstra.ResourceGroupNameSuperspineSpineIp6:
		return goapstra.ResourceTypeIp6Pool
	case goapstra.ResourceGroupNameSpineLeafIp6:
		return goapstra.ResourceTypeIp6Pool
	}
	diags.AddError(errProviderBug, fmt.Sprintf("unable to determine type of resource group '%s'", in))
	return goapstra.ResourceTypeUnknown
}

func getResourceAllocation(ctx context.Context, client *goapstra.TwoStageL3ClosClient, resourceName goapstra.ResourceGroupName, diags *diag.Diagnostics) *goapstra.ResourceGroupAllocation {
	resourceType := resourceTypeNameFromResourceGroupName(resourceName, diags)
	if resourceType == goapstra.ResourceTypeUnknown && diags.HasError() {
		return nil
	}

	allocation, err := client.GetResourceAllocation(ctx, &goapstra.ResourceGroup{
		Type: resourceType,
		Name: resourceName,
	})
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			return nil
		}
		diags.AddError(fmt.Sprintf(errReadingAllocation, goapstra.ResourceTypeAsnPool.String(), goapstra.ResourceGroupNameSpineAsn.String(), client.Id()), err.Error())
		return nil
	}
	return allocation
}

func setResourceAllocations(ctx context.Context, client *goapstra.TwoStageL3ClosClient, allocations []goapstra.ResourceGroupAllocation, diags *diag.Diagnostics) {
	hashResourceGroup := func(in *goapstra.ResourceGroup) [16]byte {
		return md5.Sum([]byte(in.Type.String() + ":" + in.Name.String()))
	}

	blueprintResources, err := client.GetAllResourceAllocations(ctx)
	if err != nil {
		diags.AddError("error fetching blueprint resource allocations", err.Error())
		return
	}

	rgHashMap := make(map[[16]byte]struct{}, len(blueprintResources))
	for _, rg := range allocations {
		rgHashMap[hashResourceGroup(&rg.ResourceGroup)] = struct{}{}
	}

	for _, allocation := range allocations {
		if _, ok := rgHashMap[hashResourceGroup(&allocation.ResourceGroup)]; !ok {
			diags.AddWarning("")
		}
	}

}

func setResourceAllocation(ctx context.Context, client *goapstra.TwoStageL3ClosClient, resourceName goapstra.ResourceGroupName, vals types.Set, diags *diag.Diagnostics) {
	if vals.IsNull() || vals.IsUnknown() {
		return
	}

	resourceType := resourceTypeNameFromResourceGroupName(resourceName, diags)
	if resourceType == goapstra.ResourceTypeUnknown && diags.HasError() {
		return
	}

	var allAllocations []goapstra.ResourceGroupAllocation
	allocationIsNeeded := func(aType goapstra.ResourceGroup, aName goapstra.ResourceGroupName) {
		if allAllocations == nil {
			allAllocations, err := client.GetAllResourceAllcations(ctx)
			if err != nil {
				//oops
			}
		}
	}

	client.GetResourceAllocation()

	err := client.SetResourceAllocation(ctx, &goapstra.ResourceGroupAllocation{
		ResourceGroup: goapstra.ResourceGroup{
			Type: resourceType,
			Name: resourceName,
		},
		PoolIds: sliceAttrValueToSliceObjectId(vals.Elems),
	})
	if err != nil {
		diags.AddError(fmt.Sprintf(errReadingAllocation, goapstra.ResourceTypeAsnPool.String(), goapstra.ResourceGroupNameSpineAsn.String(), client.Id()), err.Error())
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

package blueprint

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/netip"
	"regexp"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	apstraregexp "github.com/Juniper/terraform-provider-apstra/apstra/regexp"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeviceAllocationSystemAttributes struct {
	Asn          types.Int64          `tfsdk:"asn"`
	Name         types.String         `tfsdk:"name"`
	Hostname     types.String         `tfsdk:"hostname"`
	LoopbackIpv4 cidrtypes.IPv4Prefix `tfsdk:"loopback_ipv4"`
	LoopbackIpv6 cidrtypes.IPv6Prefix `tfsdk:"loopback_ipv6"`
	Tags         types.Set            `tfsdk:"tags"`
	DeployMode   types.String         `tfsdk:"deploy_mode"`
}

func (o DeviceAllocationSystemAttributes) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"asn":           types.Int64Type,
		"name":          types.StringType,
		"hostname":      types.StringType,
		"loopback_ipv4": cidrtypes.IPv4PrefixType{},
		"loopback_ipv6": cidrtypes.IPv6PrefixType{},
		"tags":          types.SetType{ElemType: types.StringType},
		"deploy_mode":   types.StringType,
	}
}

func (o DeviceAllocationSystemAttributes) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Web UI label for the system node.",
			Validators:          []validator.String{stringvalidator.LengthBetween(1, 64)},
		},
		"hostname": resourceSchema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Hostname of the System node.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(apstraregexp.HostNameConstraint, apstraregexp.HostNameConstraintMsg),
				stringvalidator.LengthBetween(1, 32),
			},
		},
		"asn": resourceSchema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "ASN of the system node. Setting ASN is supported only for Spine and Leaf switches.",
			Computed:            true,
			Validators:          []validator.Int64{int64validator.Between(1, math.MaxUint32)},
		},
		"loopback_ipv4": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv4 address of loopback interface in CIDR notation, must use 32-bit mask. Setting " +
				"loopback addresses is supported only for Spine and Leaf switches.",
			CustomType: cidrtypes.IPv4PrefixType{},
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile("/32$"), "must use a 32-bit mask")},
		},
		"loopback_ipv6": resourceSchema.StringAttribute{
			MarkdownDescription: "IPv6 address of loopback interface in CIDR notation, must use 128-bit mask. Setting " +
				"loopback addresses is supported only for Spine and Leaf switches. IPv6 must be enabled in the " +
				"Blueprint to use this attribute.",
			CustomType: cidrtypes.IPv6PrefixType{},
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile("/128$"), "must use a 128-bit mask")},
		},
		"tags": resourceSchema.SetAttribute{
			MarkdownDescription: "Tag labels to be applied to the System node. If a Tag doesn't exist " +
				"in the Blueprint it will be created automatically.",
			ElementType: types.StringType,
			Optional:    true,
			Computed:    false, // the user controls this field directly
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"deploy_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "Set the [deploy mode](https://www.juniper.net/documentation/us/en/software/apstra4.1/apstra-user-guide/topics/topic-map/datacenter-deploy-mode-set.html) " +
				"of the associated fabric node.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.AllNodeDeployModes()...)},
		},
	}
}

func (o *DeviceAllocationSystemAttributes) IsEmpty() bool {
	if o.Asn.IsNull() &&
		o.Name.IsNull() &&
		o.Hostname.IsNull() &&
		o.LoopbackIpv4.IsNull() &&
		o.Tags.IsNull() &&
		o.DeployMode.IsNull() &&
		o.LoopbackIpv6.IsNull() {
		return true
	}

	return false
}

func (o *DeviceAllocationSystemAttributes) ValidateConfig(_ context.Context, experimental types.Bool, diags *diag.Diagnostics) {
	if o.IsEmpty() {
		diags.AddAttributeError(path.Root("system_attributes"), constants.ErrInvalidConfig,
			"Object may be omitted, but must not be empty")
		return
	}
}

func (o *DeviceAllocationSystemAttributes) Get(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId types.String, diags *diag.Diagnostics) {
	nId := apstra.ObjectId(nodeId.ValueString())

	o.getAsn(ctx, bp, nId, diags)
	if diags.HasError() {
		return
	}

	o.getLoopback0Addresses(ctx, bp, nId, diags)
	if diags.HasError() {
		return
	}

	o.getProperties(ctx, bp, nId, diags)
	if diags.HasError() {
		return
	}

	if !utils.HasValue(o.Tags) {
		tags, err := bp.GetNodeTags(ctx, nId)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to readtags from node %s", nodeId), err.Error())
			return
		}
		o.Tags = utils.SetValueOrNull(ctx, types.StringType, tags, diags)
	}
}

func (o *DeviceAllocationSystemAttributes) getAsn(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if utils.HasValue(o.Asn) {
		return
	}

	rawJson := getDomainNode(ctx, bp, nodeId, diags)
	if diags.HasError() {
		return
	}

	o.Asn = types.Int64Null() // set to null value in case we bail later

	if len(rawJson) == 0 {
		return // no domain node found
	}

	var domainNode struct {
		DomainId *string `json:"domain_id"`
	}

	err := json.Unmarshal(rawJson, &domainNode)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed while unpacking system %q domain node", nodeId), err.Error())
		return
	}

	if domainNode.DomainId == nil {
		return // no domain id found in domain node
	}

	asn, err := strconv.ParseUint(*domainNode.DomainId, 10, 32)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse ASN response from API: %q", *domainNode.DomainId), err.Error())
		return
	}

	o.Asn = types.Int64Value(int64(asn))
}

// getLoopback0Addresses loads the IPv4 and IPv6 values of Loopback0 from the API. It does not overwrite known values.
func (o *DeviceAllocationSystemAttributes) getLoopback0Addresses(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if utils.HasValue(o.LoopbackIpv4) && utils.HasValue(o.LoopbackIpv6) {
		return
	}

	idx := 0 // loopback 0

	query := new(apstra.PathQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(nodeId)},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_type", Value: apstra.QEStringVal("loopback")},
			{Key: "loopback_id", Value: apstra.QEIntVal(idx)},
			{Key: "name", Value: apstra.QEStringVal("n_interface")},
		})

	var queryResult struct {
		Items []struct {
			Node struct {
				IPv4Addr *string `json:"ipv4_addr"`
				IPv6Addr *string `json:"ipv6_addr"`
			} `json:"n_interface"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResult)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed while querying for system %q loopback %d node", nodeId, idx), err.Error())
		return
	}

	switch len(queryResult.Items) {
	case 0:
		// no loopback node found
		o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixNull()
		o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixNull()
		return
	case 1:
		// this is the normal case - handled below
	default:
		diags.AddError(
			fmt.Sprintf("unexpected rewult while querying for system %q loopback node %d", nodeId, idx),
			fmt.Sprintf("node has graph relationships with %d loopback nodes with id %d", len(queryResult.Items), idx),
		)
		return
	}

	if !utils.HasValue(o.LoopbackIpv4) { // don't overwrite known values (apparently?)
		o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixNull()
		if queryResult.Items[0].Node.IPv4Addr != nil && *queryResult.Items[0].Node.IPv4Addr != "" {
			_, _, err := net.ParseCIDR(*queryResult.Items[0].Node.IPv4Addr)
			if err != nil {
				diags.AddError(
					fmt.Sprintf("failed to parse API response value for `ipv4_addr`: %q", *queryResult.Items[0].Node.IPv4Addr),
					err.Error())
				return
			}
			o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixValue(*queryResult.Items[0].Node.IPv4Addr)
		}
	}

	if !utils.HasValue(o.LoopbackIpv6) { // don't overwrite known values (apparently?)
		o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixNull()
		if queryResult.Items[0].Node.IPv6Addr != nil && *queryResult.Items[0].Node.IPv6Addr != "" {
			_, _, err := net.ParseCIDR(*queryResult.Items[0].Node.IPv6Addr)
			if err != nil {
				diags.AddError(
					fmt.Sprintf("failed to parse API response value for `ipv6_addr`: %q", *queryResult.Items[0].Node.IPv6Addr),
					err.Error())
				return
			}
			o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixValue(*queryResult.Items[0].Node.IPv6Addr)
		}
	}
}

func (o *DeviceAllocationSystemAttributes) getProperties(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if utils.HasValue(o.DeployMode) && utils.HasValue(o.Hostname) && utils.HasValue(o.Name) {
		return
	}

	var node struct {
		DeployMode string `json:"deploy_mode,omitempty"`
		Hostname   string `json:"hostname,omitempty"`
		Label      string `json:"label,omitempty"`
	}

	err := bp.Client().GetNode(ctx, bp.Id(), nodeId, &node)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to read node %s properties", nodeId), err.Error())
		return
	}

	var deployMode enum.DeployMode
	err = deployMode.FromString(node.DeployMode)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse node %q deploy mode %q", nodeId, node.DeployMode), err.Error())
		return
	}

	if !utils.HasValue(o.DeployMode) {
		o.DeployMode = types.StringValue(rosetta.StringersToFriendlyString(deployMode))
	}
	if !utils.HasValue(o.Hostname) {
		o.Hostname = types.StringValue(node.Hostname)
	}
	if !utils.HasValue(o.Name) {
		o.Name = types.StringValue(node.Label)
	}
}

func (o *DeviceAllocationSystemAttributes) Set(ctx context.Context, state *DeviceAllocationSystemAttributes, bp *apstra.TwoStageL3ClosClient, nodeId types.String, diags *diag.Diagnostics) {
	if state == nil || !o.Asn.Equal(state.Asn) {
		o.setAsn(ctx, bp, apstra.ObjectId(nodeId.ValueString()), diags)
	}

	if state == nil || !o.LoopbackIpv4.Equal(state.LoopbackIpv4) || !o.LoopbackIpv6.Equal(state.LoopbackIpv6) {
		o.setLoopbacks(ctx, bp, apstra.ObjectId(nodeId.ValueString()), diags)
	}

	if state == nil || !o.Name.Equal(state.Name) || !o.Hostname.Equal(state.Hostname) || !o.DeployMode.Equal(state.DeployMode) {
		o.setProperties(ctx, bp, apstra.ObjectId(nodeId.ValueString()), diags)
	}

	if state == nil || !o.Tags.Equal(state.Tags) {
		o.setTags(ctx, state, bp, apstra.ObjectId(nodeId.ValueString()), diags)
	}
}

func (o *DeviceAllocationSystemAttributes) setAsn(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if !utils.HasValue(o.Asn) {
		return
	}

	rawJson := getDomainNode(ctx, bp, nodeId, diags)
	if diags.HasError() {
		return
	}

	if len(rawJson) == 0 {
		diags.AddAttributeError(
			path.Root("system_attributes").AtName("asn"), "Cannot set ASN",
			fmt.Sprintf("system %q has no associated domain (ASN) node -- is it a spine or leaf switch?", nodeId))
		return
	}

	var domainNode struct {
		Id *apstra.ObjectId `json:"id"`
	}

	err := json.Unmarshal(rawJson, &domainNode)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed while unpacking system %q domain node", nodeId), err.Error())
		return
	}

	if domainNode.Id == nil {
		diags.AddError(
			fmt.Sprintf("failed parsing domain node linked with node %q", nodeId),
			fmt.Sprintf("domain node has no field `id`: %q", string(rawJson)))
		return
	}

	patch := struct {
		DomainId string `json:"domain_id"`
	}{
		DomainId: strconv.FormatInt(o.Asn.ValueInt64(), 10),
	}

	err = bp.PatchNode(ctx, *domainNode.Id, &patch, nil)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed setting ASN to domain node %q", domainNode.Id), err.Error())
		return
	}
}

func getLoopbackNodeAndSecurityZoneIDs(ctx context.Context, bp *apstra.TwoStageL3ClosClient, systemNodeId apstra.ObjectId, loopIdx int, diags *diag.Diagnostics) (apstra.ObjectId, apstra.ObjectId) {
	query := new(apstra.PathQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(systemNodeId)},
		}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeInterface.QEEAttribute(),
			{Key: "if_type", Value: apstra.QEStringVal("loopback")},
			{Key: "loopback_id", Value: apstra.QEIntVal(loopIdx)},
			{Key: "name", Value: apstra.QEStringVal("n_interface")},
		}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeMemberInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeSecurityZoneInstance.QEEAttribute()}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeInstantiatedBy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSecurityZone.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_security_zone")},
		})

	var queryResponse struct {
		Items []struct {
			Interface struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_interface"`
			SecurityZone struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_security_zone"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResponse)
	if err != nil {
		diags.AddError("failed while querying for loopback interface and security zone", err.Error())
		return "", ""
	}
	if len(queryResponse.Items) != 1 {
		diags.AddError(
			fmt.Sprintf("expected exactly one loopback and security zone node pair, got %d", len(queryResponse.Items)),
			fmt.Sprintf("graph query: %q", query),
		)
		return "", ""
	}

	ifId, szId := queryResponse.Items[0].Interface.Id, queryResponse.Items[0].SecurityZone.Id

	if ifId == "" {
		diags.AddError(
			"got empty interface ID",
			fmt.Sprintf("graph query: %q", query),
		)
	}
	if szId == "" {
		diags.AddError(
			"got empty security zone ID",
			fmt.Sprintf("graph query: %q", query),
		)
	}

	return ifId, szId
}

func (o *DeviceAllocationSystemAttributes) legacySetLoopbacks(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	patch := &struct {
		IPv4Addr string `json:"ipv4_addr,omitempty"`
		IPv6Addr string `json:"ipv6_addr,omitempty"`
	}{
		IPv4Addr: o.LoopbackIpv4.ValueString(),
		IPv6Addr: o.LoopbackIpv6.ValueString(),
	}

	err := bp.PatchNode(ctx, nodeId, &patch, nil)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed setting loopback addresses to interface node %q", nodeId), err.Error())
		return
	}
}

func (o *DeviceAllocationSystemAttributes) setLoopbacks(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if !utils.HasValue(o.LoopbackIpv4) && !utils.HasValue(o.LoopbackIpv6) {
		return
	}

	idx := 0 // we always are interested in loopback 0

	loopbackNodeId, securityZoneId := getLoopbackNodeAndSecurityZoneIDs(ctx, bp, nodeId, idx, diags)
	if diags.HasError() {
		return
	}

	if compatibility.ApiNotSupportsSetLoopbackIps.Check(version.Must(version.NewVersion(bp.Client().ApiVersion()))) {
		// we must be talking to Apstra 4.x
		o.legacySetLoopbacks(ctx, bp, loopbackNodeId, diags)
		return
	}

	// Use new() here to ensure we have invalid non-nil prefix pointers. These will remove values from the API.
	ipv4Addr, ipv6Addr := new(netip.Prefix), new(netip.Prefix)
	if utils.HasValue(o.LoopbackIpv4) {
		ipv4Addr = utils.ToPtr(netip.MustParsePrefix(o.LoopbackIpv4.ValueString()))
	}
	if utils.HasValue(o.LoopbackIpv6) {
		ipv6Addr = utils.ToPtr(netip.MustParsePrefix(o.LoopbackIpv6.ValueString()))
	}

	err := bp.SetSecurityZoneLoopbacks(ctx, securityZoneId, map[apstra.ObjectId]apstra.SecurityZoneLoopback{
		loopbackNodeId: {
			IPv4Addr: ipv4Addr,
			IPv6Addr: ipv6Addr,
		},
	})
	if err != nil {
		diags.AddError("failed while setting loopback addresses", err.Error())
		return
	}
}

func (o *DeviceAllocationSystemAttributes) setProperties(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if !utils.HasValue(o.Name) && !utils.HasValue(o.Hostname) && !utils.HasValue(o.DeployMode) {
		return
	}

	if utils.HasValue(o.DeployMode) {
		var deployMode enum.DeployMode
		err := rosetta.ApiStringerFromFriendlyString(&deployMode, o.DeployMode.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error in rosetta function with deploy_mode = %s", o.DeployMode), err.Error())
			return
		}

		var deployModePayload *string
		if deployMode != enum.DeployModeNone {
			deployModePayload = utils.ToPtr(deployMode.String())
		}

		patch := struct {
			DeployMode *string `json:"deploy_mode"`
			Hostname   string  `json:"hostname,omitempty"`
			Label      string  `json:"label,omitempty"`
		}{
			DeployMode: deployModePayload,
			Hostname:   o.Hostname.ValueString(),
			Label:      o.Name.ValueString(),
		}

		err = bp.PatchNode(ctx, nodeId, &patch, nil)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed while patching system node %q", nodeId), err.Error())
			return
		}
	} else {
		patch := struct {
			Hostname string `json:"hostname,omitempty"`
			Label    string `json:"label,omitempty"`
		}{
			Hostname: o.Hostname.ValueString(),
			Label:    o.Name.ValueString(),
		}

		err := bp.PatchNode(ctx, nodeId, &patch, nil)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed while patching system node %q", nodeId), err.Error())
			return
		}
	}
}

func (o *DeviceAllocationSystemAttributes) setTags(ctx context.Context, state *DeviceAllocationSystemAttributes, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if len(o.Tags.Elements()) == 0 && (state == nil || len(state.Tags.Elements()) == 0) {
		// user supplied no tags (requiring us to clear them), but state indicates no tags exist
		return
	}

	var tags []string
	// extract tags from user config, if any. not a Computed value. If null, we clear the tags.
	if !o.Tags.IsNull() {
		diags.Append(o.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return
		}
	}

	err := bp.SetNodeTags(ctx, nodeId, tags)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed setting tags on node %q", nodeId), err.Error())
	}
}

func getDomainNode(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) json.RawMessage {
	query := new(apstra.PathQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(nodeId)},
		}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOfSystems.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeDomain.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_domain")},
		})

	var queryResult struct {
		Items []struct {
			Node json.RawMessage `json:"n_domain"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResult)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed while querying for system %q domain node", nodeId), err.Error())
		return nil
	}

	switch len(queryResult.Items) {
	case 0:
		return nil
	case 1:
		return queryResult.Items[0].Node
	default:
		diags.AddError(
			fmt.Sprintf("unexpected rewult while querying for system %q domain node", nodeId),
			fmt.Sprintf("node has graph relationships with %d domain nodes", len(queryResult.Items)),
		)
		return nil
	}
}

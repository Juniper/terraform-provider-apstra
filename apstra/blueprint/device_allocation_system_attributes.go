package blueprint

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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
				stringvalidator.RegexMatches(regexp.MustCompile("^[A-Za-z0-9.-]+$"),
					"only alphanumeric characters, '.' and '-' allowed."),
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

	if experimental.IsNull() {
		return // resource not yet configured
	}

	if !experimental.ValueBool() {
		diags.AddAttributeError(path.Root("system_attributes"), constants.ErrInvalidConfig,
			"Configuration requires `experimental = true` in provider configuration block")
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

	if !utils.Known(o.Tags) {
		tags, err := bp.GetNodeTags(ctx, nId)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to readtags from node %s", nodeId), err.Error())
			return
		}
		o.Tags = utils.SetValueOrNull(ctx, types.StringType, tags, diags)
	}
}

func (o *DeviceAllocationSystemAttributes) getAsn(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if utils.Known(o.Asn) {
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

func (o *DeviceAllocationSystemAttributes) getLoopback0Addresses(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if utils.Known(o.LoopbackIpv4) && utils.Known(o.LoopbackIpv6) {
		return
	}

	idx := 0
	rawJson := getLoopbackInterfaceNode(ctx, bp, nodeId, idx, diags)
	if diags.HasError() {
		return
	}

	if len(rawJson) == 0 {
		o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixNull()
		o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixNull()
		return // no loopback idx interface node found
	}

	var loopbackNode struct {
		IPv4Addr *string `json:"ipv4_addr"`
		IPv6Addr *string `json:"ipv6_addr"`
	}

	err := json.Unmarshal(rawJson, &loopbackNode)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("failed while unpacking system %q loopback %d interface node", nodeId, idx),
			err.Error(),
		)
		return
	}

	if loopbackNode.IPv4Addr == nil && loopbackNode.IPv6Addr == nil {
		o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixNull()
		o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixNull()
		return // no loopback IP addresses found in domain node
	}

	if !utils.Known(o.LoopbackIpv4) {
		o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixNull()
		if loopbackNode.IPv4Addr != nil && len(*loopbackNode.IPv4Addr) != 0 {
			_, _, err := net.ParseCIDR(*loopbackNode.IPv4Addr)
			if err != nil {
				diags.AddError(
					fmt.Sprintf("failed to parse `ipv4_addr` from API response %q", string(rawJson)),
					err.Error())
				return
			}
			o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixValue(*loopbackNode.IPv4Addr)
		}
	}

	if !utils.Known(o.LoopbackIpv6) {
		o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixNull()
		if loopbackNode.IPv6Addr != nil && len(*loopbackNode.IPv6Addr) != 0 {
			_, _, err := net.ParseCIDR(*loopbackNode.IPv6Addr)
			if err != nil {
				diags.AddError(
					fmt.Sprintf("failed to parse `ipv6_addr` from API response %q", string(rawJson)),
					err.Error())
				return
			}
			o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixValue(*loopbackNode.IPv6Addr)
		}
	}
}

func (o *DeviceAllocationSystemAttributes) getProperties(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if utils.Known(o.DeployMode) && utils.Known(o.Hostname) && utils.Known(o.Name) {
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

	var deployMode apstra.DeployMode
	err = deployMode.FromString(node.DeployMode)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse node %q deploy mode %q", nodeId, node.DeployMode), err.Error())
		return
	}

	if !utils.Known(o.DeployMode) {
		o.DeployMode = types.StringValue(utils.StringersToFriendlyString(deployMode))
	}
	if !utils.Known(o.Hostname) {
		o.Hostname = types.StringValue(node.Hostname)
	}
	if !utils.Known(o.Name) {
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
	if !utils.Known(o.Asn) {
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

func (o *DeviceAllocationSystemAttributes) setLoopbacks(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if !utils.Known(o.LoopbackIpv4) && !utils.Known(o.LoopbackIpv6) {
		return
	}

	idx := 0
	rawJson := getLoopbackInterfaceNode(ctx, bp, nodeId, idx, diags)
	if diags.HasError() {
		return
	}

	if len(rawJson) == 0 && utils.Known(o.LoopbackIpv4) {
		diags.AddAttributeError(
			path.Root("system_attributes").AtName("loopback_ipv4"),
			"Cannot set loopback address",
			fmt.Sprintf("system %q has no associated loopback %d node -- is it a spine or leaf switch?", nodeId, idx))
	}
	if len(rawJson) == 0 && utils.Known(o.LoopbackIpv6) {
		diags.AddAttributeError(
			path.Root("system_attributes").AtName("loopback_ipv6"),
			"Cannot set loopback address",
			fmt.Sprintf("system %q has no associated loopback %d node -- is it a spine or leaf switch?", nodeId, idx))
	}
	if diags.HasError() {
		return
	}

	var loopbackNode struct {
		Id *apstra.ObjectId `json:"id"`
	}

	err := json.Unmarshal(rawJson, &loopbackNode)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed while unpacking system %q loopback %d node", nodeId, idx), err.Error())
		return
	}

	if loopbackNode.Id == nil {
		diags.AddError(
			fmt.Sprintf("failed parsing loopback %d node linked with node %q", idx, nodeId),
			fmt.Sprintf("loopback %d node has no field `id`: %q", idx, string(rawJson)))
		return
	}

	patch := &struct {
		IPv4Addr string `json:"ipv4_addr,omitempty"`
		IPv6Addr string `json:"ipv6_addr,omitempty"`
	}{
		IPv4Addr: o.LoopbackIpv4.ValueString(),
		IPv6Addr: o.LoopbackIpv6.ValueString(),
	}

	err = bp.PatchNode(ctx, *loopbackNode.Id, &patch, nil)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed setting loopback addresses to interface node %q", loopbackNode.Id), err.Error())
		return
	}
}

func (o *DeviceAllocationSystemAttributes) setProperties(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if !utils.Known(o.Name) && !utils.Known(o.Hostname) && !utils.Known(o.DeployMode) {
		return
	}

	var deployMode apstra.DeployMode
	err := utils.ApiStringerFromFriendlyString(&deployMode, o.DeployMode.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("error in rosetta function with deploy_mode = %s", o.DeployMode), err.Error())
		return
	}

	node := struct {
		DeployMode string `json:"deploy_mode,omitempty"`
		Hostname   string `json:"hostname,omitempty"`
		Label      string `json:"label,omitempty"`
	}{
		DeployMode: deployMode.String(),
		Hostname:   o.Hostname.ValueString(),
		Label:      o.Name.ValueString(),
	}

	err = bp.PatchNode(ctx, nodeId, &node, nil)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed while patching system node %q", nodeId), err.Error())
		return
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

func getLoopbackInterfaceNode(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, idx int, diags *diag.Diagnostics) json.RawMessage {
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
			Node json.RawMessage `json:"n_interface"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResult)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed while querying for system %q loopback %d node", nodeId, idx), err.Error())
		return nil
	}

	switch len(queryResult.Items) {
	case 0:
		return nil
	case 1:
		return queryResult.Items[0].Node
	default:
		diags.AddError(
			fmt.Sprintf("unexpected rewult while querying for system %q loopback node %d", nodeId, idx),
			fmt.Sprintf("node has graph relationships with %d loopback nodes with id %d", len(queryResult.Items), idx),
		)
		return nil
	}
}

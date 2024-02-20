package blueprint

import (
	"context"
	"encoding/json"
	"fmt"
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
	"math"
	"regexp"
	"strconv"
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

func (o DeviceAllocationSystemAttributes) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":          types.StringType,
		"hostname":      types.StringType,
		"asn":           types.Int64Type,
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
				"loopback addresses is supported only for Spine and Leaf switches.",
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

	o.getLoopbacks(ctx, bp, nId, diags)
	if diags.HasError() {
		return
	}

	o.getProperties(ctx, bp, nId, diags)
	if diags.HasError() {
		return
	}

	tags, err := bp.GetNodeTags(ctx, nId)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to readtags from node %s", nodeId), err.Error())
		return
	}
	o.Tags = utils.SetValueOrNull(ctx, types.StringType, tags, diags)
}

func (o *DeviceAllocationSystemAttributes) getAsn(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	var domainNode *struct {
		DomainId *string `json:"domain_id"`
	}

	getDomainNode(ctx, bp, nodeId, domainNode, diags)
	if diags.HasError() {
		return
	}

	o.Asn = types.Int64Null()

	if domainNode != nil && domainNode.DomainId != nil {
		asn, err := strconv.ParseUint(*domainNode.DomainId, 10, 32)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to parse ASN response from API: %q", *domainNode.DomainId), err.Error())
			return
		}

		o.Asn = types.Int64Value(int64(asn))
	}
}

func (o *DeviceAllocationSystemAttributes) getLoopbacks(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	var loopbackNode *struct {
		IPv4Addr *string `json:"ipv4_addr"`
		IPv6Addr *string `json:"ipv6_addr"`
	}

	getLoopbackInterfaceNode(ctx, bp, nodeId, 0, loopbackNode, diags)
	if diags.HasError() {
		return
	}

	o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixNull()
	o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixNull()

	if loopbackNode != nil && loopbackNode.IPv4Addr != nil {
		o.LoopbackIpv4 = cidrtypes.NewIPv4PrefixValue(*loopbackNode.IPv4Addr)
	}

	if loopbackNode != nil && loopbackNode.IPv6Addr != nil {
		o.LoopbackIpv6 = cidrtypes.NewIPv6PrefixValue(*loopbackNode.IPv6Addr)
	}
}

func (o *DeviceAllocationSystemAttributes) getProperties(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
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

	var deployMode apstra.NodeDeployMode
	err = deployMode.FromString(node.DeployMode)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to parse node %q deploy mode %q", nodeId, node.DeployMode), err.Error())
		return
	}

	o.DeployMode = types.StringValue(utils.StringersToFriendlyString(deployMode))
	o.Hostname = types.StringValue(node.Hostname)
	o.Name = types.StringValue(node.Label)
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

	var domainNode *struct {
		Id apstra.ObjectId `json:"id"`
	}

	getDomainNode(ctx, bp, nodeId, domainNode, diags)
	if diags.HasError() {
		return
	}

	if domainNode == nil {
		diags.AddAttributeError(
			path.Root("system_attributes").AtName("asn"), "Cannot set ASN",
			fmt.Sprintf("system %q has no associated domain (ASN) node", nodeId))
		return
	}

	patch := struct {
		DomainId string `json:"domain_id"`
	}{
		DomainId: strconv.FormatInt(o.Asn.ValueInt64(), 10),
	}

	err := bp.PatchNode(ctx, domainNode.Id, &patch, nil)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed setting ASN to domain node %q", domainNode.Id), err.Error())
		return
	}
}

func (o *DeviceAllocationSystemAttributes) setLoopbacks(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if !utils.Known(o.LoopbackIpv4) && !utils.Known(o.LoopbackIpv6) {
		return
	}

	var loopbackNode *struct {
		Id apstra.ObjectId `json:"id"`
	}

	loIdx := 0

	getLoopbackInterfaceNode(ctx, bp, nodeId, loIdx, loopbackNode, diags)
	if diags.HasError() {
		return
	}

	if loopbackNode == nil && !o.LoopbackIpv4.IsNull() {
		diags.AddAttributeError(
			path.Root("system_attributes").AtName("loopback_ipv4"), "Cannot set IPv4 Loopback Address",
			fmt.Sprintf("system %q has no associated Loopback %d node", nodeId, loIdx))
	}
	if loopbackNode == nil && !o.LoopbackIpv6.IsNull() {
		diags.AddAttributeError(
			path.Root("system_attributes").AtName("loopback_ipv6"), "Cannot set IPv6 Loopback Address",
			fmt.Sprintf("system %q has no associated Loopback %d node", nodeId, loIdx))
	}
	if diags.HasError() {
		return
	}

	patch := &struct {
		IPv4Addr string `json:"ipv4_addr,omitempty"`
		IPv6Addr string `json:"ipv6_addr,omitempty"`
	}{
		IPv4Addr: o.LoopbackIpv4.ValueString(),
		IPv6Addr: o.LoopbackIpv6.ValueString(),
	}

	err := bp.PatchNode(ctx, loopbackNode.Id, &patch, nil)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed setting loopback addresses to interface node %q", loopbackNode.Id), err.Error())
		return
	}
}

func (o *DeviceAllocationSystemAttributes) setProperties(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	if !utils.Known(o.Name) && !utils.Known(o.Hostname) && !utils.Known(o.DeployMode) {
		return
	}

	node := struct {
		DeployMode string `json:"deploy_mode,omitempty"`
		Hostname   string `json:"hostname,omitempty"`
		Label      string `json:"label,omitempty"`
	}{
		DeployMode: o.DeployMode.ValueString(),
		Hostname:   o.Hostname.ValueString(),
		Label:      o.Name.ValueString(),
	}

	err := bp.PatchNode(ctx, nodeId, &node, nil)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed while patching system node %q", nodeId), err.Error())
		return
	}
}

func (o *DeviceAllocationSystemAttributes) setTags(ctx context.Context, state *DeviceAllocationSystemAttributes, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, diags *diag.Diagnostics) {
	// not a Computed value, so IsNull() will suffice
	if o.Tags.IsNull() && len(state.Tags.Elements()) == 0 {
		// tags not supplied by user indicates we intend to clear them
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

func getDomainNode(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, target any, diags *diag.Diagnostics) {
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
		return
	}

	switch len(queryResult.Items) {
	case 0:
	case 1:
		err = json.Unmarshal(queryResult.Items[0].Node, target)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed while unpacking system %q domain node", nodeId), err.Error())
		}
	default:
		diags.AddError(
			fmt.Sprintf("unexpected rewult while querying for system %q domain node", nodeId),
			fmt.Sprintf("node has graph relationships with %d domain nodes", len(queryResult.Items)),
		)
	}
}

func getLoopbackInterfaceNode(ctx context.Context, bp *apstra.TwoStageL3ClosClient, nodeId apstra.ObjectId, idx int, target any, diags *diag.Diagnostics) {
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
		return
	}

	switch len(queryResult.Items) {
	case 0:
	case 1:
		err = json.Unmarshal(queryResult.Items[0].Node, target)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed while unpacking system %q loopack node %d", nodeId, idx), err.Error())
		}
	default:
		diags.AddError(
			fmt.Sprintf("unexpected rewult while querying for system %q loopback node %d", nodeId, idx),
			fmt.Sprintf("node has graph relationships with %d loopback nodes with id %d", len(queryResult.Items), idx),
		)
	}
}

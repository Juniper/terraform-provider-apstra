package tfapstra

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"math"
	"net"
	"net/netip"

	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ resource.ResourceWithValidateConfig = &resourceDatacenterIpLinkAddressing{}
	_ resourceWithSetDcBpClientFunc       = &resourceDatacenterIpLinkAddressing{}
	_ resourceWithSetBpLockFunc           = &resourceDatacenterIpLinkAddressing{}
)

type resourceDatacenterIpLinkAddressing struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterIpLinkAddressing) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_ip_link_addressing"
}

func (o *resourceDatacenterIpLinkAddressing) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterIpLinkAddressing) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates IPv4 and IPv6 addressing on L3 " +
			"links within a Datacenter Blueprint fabric. It is intended for use with links created " +
			"as a side-effect of assigning Connectivity Templates containing IP Link primitives.",
		Attributes: blueprint.IpLinkAddressing{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterIpLinkAddressing) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config blueprint.IpLinkAddressing
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var switchIpv4, switchIpv6, genericIpv4, genericIpv6 netip.Prefix

	if !config.SwitchIpv4Addr.IsNull() {
		switchIpv4, _ = netip.ParsePrefix(config.SwitchIpv4Addr.ValueString())
	}
	if !config.SwitchIpv6Addr.IsNull() {
		switchIpv6, _ = netip.ParsePrefix(config.SwitchIpv6Addr.ValueString())
	}
	if !config.GenericIpv4Addr.IsNull() {
		genericIpv4, _ = netip.ParsePrefix(config.GenericIpv4Addr.ValueString())
	}
	if !config.GenericIpv6Addr.IsNull() {
		genericIpv6, _ = netip.ParsePrefix(config.GenericIpv6Addr.ValueString())
	}

	checkPrefix := func(prefix netip.Prefix, path path.Path) {
		if !prefix.IsValid() {
			return
		}

		if prefix.IsSingleIP() {
			resp.Diagnostics.AddAttributeError(
				path,
				constants.ErrInvalidConfig,
				fmt.Sprintf("Prefix length must support at least two hosts, got /%d", prefix.Bits()))
		}

		if prefix.Addr().Is4() && prefix.Bits() < 31 && prefix.Masked().Addr().String() == prefix.Addr().String() {
			resp.Diagnostics.AddAttributeError(
				path,
				constants.ErrInvalidConfig,
				fmt.Sprintf("Prefix must use a valid host address with the %q network, got %q (base network address)", prefix.Masked(), prefix),
			)
		}

		isBroadcast := func(p netip.Prefix) bool {
			if !p.Addr().Is4() {
				return false // only IPv4 has broadcast addresses
			}

			if binary.BigEndian.Uint32(p.Addr().AsSlice()) == math.MaxUint32 {
				return true // special case for 255.255.255.255
			}

			if p.Bits() >= 31 {
				return false // no broadcast on /31 and /32 interfaces
			}

			inverseMaskSlice := binary.BigEndian.AppendUint32(nil, uint32((1<<(32-p.Bits()))-1))
			broadcast := make(net.IP, 4)
			binary.BigEndian.PutUint32(broadcast, binary.BigEndian.Uint32(p.Addr().AsSlice())|binary.BigEndian.Uint32(inverseMaskSlice))

			return broadcast.Equal(p.Addr().AsSlice())
		}

		if prefix.Addr().Is4() && prefix.Bits() < 31 && isBroadcast(prefix) {
			resp.Diagnostics.AddAttributeError(
				path,
				errInvalidConfig,
				fmt.Sprintf("Prefix must use a valid host address within the %q network, got %q (broadcast address)", prefix.Masked(), prefix),
			)
		}
	}
	checkPrefix(switchIpv4, path.Root("switch_ipv4_address"))
	checkPrefix(switchIpv6, path.Root("switch_ipv6_address"))
	checkPrefix(genericIpv4, path.Root("generic_ipv4_address"))
	checkPrefix(genericIpv6, path.Root("generic_ipv6_address"))

	checkPrefixPair := func(switchPrefix, genericPrefix netip.Prefix) {
		if !switchPrefix.IsValid() || !genericPrefix.IsValid() {
			return
		}

		if switchPrefix.Bits() != genericPrefix.Bits() {
			var msg string
			if switchPrefix.Addr().Is4() {
				msg = "values of attributes `switch_ipv4_address` and `generic_ipv4_address` must specify the same prefix length. Got /%d and /%d."
			} else {
				msg = "values of attributes `switch_ipv6_address` and `generic_ipv6_address` must specify the same prefix length. Got /%d and /%d."
			}
			resp.Diagnostics.AddError(
				constants.ErrInvalidConfig,
				fmt.Sprintf(msg, switchPrefix.Bits(), genericPrefix.Bits()),
			)
		}

		if !switchPrefix.Contains(genericPrefix.Addr()) {
			var msg string
			if switchPrefix.Addr().Is4() {
				msg = "values of attributes `switch_ipv4_address` and `generic_ipv4_address` must fall withing the same network. Got %q and %q."
			} else {
				msg = "values of attributes `switch_ipv6_address` and `generic_ipv6_address` must fall withing the same network. Got %q and %q."
			}
			resp.Diagnostics.AddError(
				constants.ErrInvalidConfig,
				fmt.Sprintf(msg, switchPrefix, genericPrefix),
			)
		}
	}
	checkPrefixPair(switchIpv4, genericIpv4)
	checkPrefixPair(switchIpv6, genericIpv6)
}

func (o *resourceDatacenterIpLinkAddressing) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan blueprint.IpLinkAddressing
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to lock blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// fetch the link so that we can "compute" (learn) the switch/generic subinterface IDs
	link, err := bp.GetSubinterfaceLink(ctx, apstra.ObjectId(plan.LinkId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddAttributeError(
				path.Root("link_id"),
				"Link not found",
				fmt.Sprintf("Link %s not found in blueprint %s", plan.LinkId, plan.BlueprintId),
			)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("failed to fetch link %s info", plan.LinkId), err.Error())
		return
	}

	// LoadImmutableData only needs to be done once, and only in the Create() method.
	plan.LoadImmutableData(ctx, link, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a subinterface addressing request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// update the subinterface
	err = bp.UpdateSubinterfaces(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to add subinterface addressing", err.Error())
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterIpLinkAddressing) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state blueprint.IpLinkAddressing
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to create Blueprint client", err.Error())
		return
	}

	// fetch the details from the API
	apiData, err := bp.GetSubinterfaceLink(ctx, apstra.ObjectId(state.LinkId.ValueString()))
	if err != nil {
		if utils.IsApstra404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to fetch details for logical link node %s", state.LinkId), err.Error())
		return
	}

	// load the state
	state.LoadApiData(ctx, apiData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (o *resourceDatacenterIpLinkAddressing) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan blueprint.IpLinkAddressing
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("Blueprint %s not found", plan.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, plan.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to lock blueprint %q mutex", plan.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// create a subinterface addressing request
	request := plan.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// update the subinterface
	err = bp.UpdateSubinterfaces(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to add subinterface addressing", err.Error())
		return
	}

	// set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceDatacenterIpLinkAddressing) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blueprint.IpLinkAddressing
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return // 404 is okay
		}
		resp.Diagnostics.AddError("failed to create blueprint client", err.Error())
		return
	}

	// Lock the blueprint mutex.
	err = o.lockFunc(ctx, state.BlueprintId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to lock blueprint %q mutex", state.BlueprintId.ValueString()),
			err.Error())
		return
	}

	// collect the switch/server v4/v6 addressing types saved to private state during create
	pBytes, d := req.Private.GetKey(ctx, "ep_addr_types")
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	var private struct {
		SwitchIpv4AddressType  string `json:"switch_ipv4_address_type"`
		SwitchIpv6AddressType  string `json:"switch_ipv6_address_type"`
		GenericIpv4AddressType string `json:"generic_ipv4_address_type"`
		GenericIpv6AddressType string `json:"generic_ipv6_address_type"`
	}
	err = json.Unmarshal(pBytes, &private)
	if err != nil {
		resp.Diagnostics.AddError("failed unmarshaling private data", err.Error())
		return
	}

	// unpack the private state into apstra objects
	var switchIpv4AddressType, genericIpv4AddressType enum.InterfaceNumberingIpv4Type
	var switchIpv6AddressType, genericIpv6AddressType enum.InterfaceNumberingIpv6Type
	err = utils.ApiStringerFromFriendlyString(&switchIpv4AddressType, private.SwitchIpv4AddressType)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse private data switch_ipv4_address_type", err.Error())
		return
	}
	err = utils.ApiStringerFromFriendlyString(&switchIpv6AddressType, private.SwitchIpv6AddressType)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse private data switch_ipv6_address_type", err.Error())
		return
	}
	err = utils.ApiStringerFromFriendlyString(&genericIpv4AddressType, private.GenericIpv4AddressType)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse private data generic_ipv4_address_type", err.Error())
		return
	}
	err = utils.ApiStringerFromFriendlyString(&genericIpv6AddressType, private.GenericIpv6AddressType)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse private data generic_ipv6_address_type", err.Error())
		return
	}

	// create a subinterface addressing request which kills off IPv4 and IPv6
	// addressing for each subinterface associated with the link.
	request := map[apstra.ObjectId]apstra.TwoStageL3ClosSubinterface{
		apstra.ObjectId(state.SwitchIntfId.ValueString()): {
			Ipv4AddrType: switchIpv4AddressType,
			Ipv6AddrType: switchIpv6AddressType,
		},
		apstra.ObjectId(state.GenericIntfId.ValueString()): {
			Ipv4AddrType: genericIpv4AddressType,
			Ipv6AddrType: genericIpv6AddressType,
		},
	}

	// update the subinterfaces
	err = bp.UpdateSubinterfaces(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to clear logical link subinterface addressing", err.Error())
		return
	}
}

func (o *resourceDatacenterIpLinkAddressing) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterIpLinkAddressing) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}

package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceRackType{}
var _ datasource.DataSourceWithValidateConfig = &dataSourceRackType{}

type dataSourceRackType struct {
	client *goapstra.Client
}

func (o *dataSourceRackType) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack_type"
}

func (o *dataSourceRackType) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errDataSourceConfigureProviderDataDetail,
			fmt.Sprintf(errDataSourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *dataSourceRackType) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific Rack Type.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent on the user to ensure the criteria matches exactly one Rack Type. " +
			"Matching zero Rack Types or more than one Rack Type will produce an error.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Rack Type id.  Required when the Rack Type name is omitted.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Rack Type name displayed in the Apstra web UI.  Required when Rack Type id is omitted.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Rack Type description displayed in the Apstra web UI.",
				Computed:            true,
			},
			"fabric_connectivity_design": schema.StringAttribute{
				MarkdownDescription: "Indicates designs for which this Rack Type is intended.",
				Computed:            true,
			},
			"leaf_switches": schema.SetNestedAttribute{
				MarkdownDescription: "Details of Leaf Switches in this Rack Type.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: leafSwitchAttributes(),
				},
			},
			//"access_switches": schema.SetNestedAttribute{
			//	MarkdownDescription: "Details of Access Switches in this Rack Type.",
			//	Computed:            true,
			//	NestedObject: schema.NestedAttributeObject{
			//		Attributes: accessSwitchAttributes(),
			//	},
			//},
			//"generic_systems": schema.SetNestedAttribute{
			//	MarkdownDescription: "Details of Generic Systems in the Rack Type.",
			//	Computed:            true,
			//	NestedObject: schema.NestedAttributeObject{
			//		Attributes: genericSystemAttributes(),
			//	},
			//},
		},
	}
}

func (o *dataSourceRackType) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config dRackType
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (config.Name.IsNull() && config.Id.IsNull()) || (!config.Name.IsNull() && !config.Id.IsNull()) { // XOR
		resp.Diagnostics.AddError("configuration error", "exactly one of 'id' and 'name' must be specified")
		return
	}
}

func (o *dataSourceRackType) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dRackType
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var rt *goapstra.RackType
	var ace goapstra.ApstraClientErr

	// maybe the config gave us the rack type name?
	if !config.Name.IsNull() { // fetch rack type by name
		rt, err = o.client.GetRackTypeByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with name '%s' does not exist", config.Name.ValueString()))
			return
		}
	}

	// maybe the config gave us the rack type id?
	if !config.Id.IsNull() { // fetch rack type by ID
		rt, err = o.client.GetRackType(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with id '%s' does not exist", config.Id.ValueString()))
			return
		}
	}

	if err != nil { // catch errors other than 404 from above
		resp.Diagnostics.AddError("Error retrieving Rack Type", err.Error())
	}

	// catch problems which would crash the provider
	validateRackType(rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newState := &dRackType{}
	newState.parseApi(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func validateRackType(rt *goapstra.RackType, diags *diag.Diagnostics) {
	if rt.Data == nil {
		diags.AddError("rack type has no data", fmt.Sprintf("rack type '%s' data object is nil", rt.Id))
		return
	}

	for i := range rt.Data.LeafSwitches {
		validateLeafSwitch(rt, i, diags)
	}

	for i := range rt.Data.AccessSwitches {
		validateAccessSwitch(rt, i, diags)
	}

	for i := range rt.Data.GenericSystems {
		validateGenericSystem(rt, i, diags)
	}
}

func validateLeafSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	ls := rt.Data.LeafSwitches[i]
	if ls.RedundancyProtocol == goapstra.LeafRedundancyProtocolMlag && ls.MlagInfo == nil {
		diags.AddError("leaf switch MLAG Info missing",
			fmt.Sprintf("rack type '%s', leaf switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, ls.Label, ls.RedundancyProtocol.String()))
	}
	if ls.LogicalDevice == nil {
		diags.AddError("leaf switch logical device info missing",
			fmt.Sprintf("rack type '%s', leaf switch '%s' logical device is nil",
				rt.Id, ls.Label))
	}
}

func validateAccessSwitch(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	as := rt.Data.AccessSwitches[i]
	if as.RedundancyProtocol == goapstra.AccessRedundancyProtocolEsi && as.EsiLagInfo == nil {
		diags.AddError("access switch ESI LAG Info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' has '%s', but EsiLagInfo is nil",
				rt.Id, as.Label, as.RedundancyProtocol.String()))
	}
	if as.LogicalDevice == nil {
		diags.AddError("access switch logical device info missing",
			fmt.Sprintf("rack type '%s', access switch '%s' logical device is nil",
				rt.Id, as.Label))
	}
}

func validateGenericSystem(rt *goapstra.RackType, i int, diags *diag.Diagnostics) {
	gs := rt.Data.GenericSystems[i]
	if gs.LogicalDevice == nil {
		diags.AddError("generic system logical device info missing",
			fmt.Sprintf("rack type '%s', generic system '%s' logical device is nil",
				rt.Id, gs.Label))
	}
}

type dRackType struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Set    `tfsdk:"leaf_switches"`
	//AccessSwitches           types.Set    `tfsdk:"access_switches"` // todo re-enable this
	//GenericSystems           types.Set    `tfsdk:"generic_systems"` // todo re-enable this
}

func (o *dRackType) parseApi(ctx context.Context, in *goapstra.RackType, diags *diag.Diagnostics) {
	switch in.Data.FabricConnectivityDesign {
	case goapstra.FabricConnectivityDesignL3Collapsed: // supported FCD
	case goapstra.FabricConnectivityDesignL3Clos: // supported FCD
	default: // unsupported FCD
		diags.AddError(
			"unsupported fabric connectivity design",
			fmt.Sprintf("Rack Type '%s' has unsupported Fabric Connectivity Design '%s'",
				in.Id, in.Data.FabricConnectivityDesign.String()))
	}
	var d diag.Diagnostics

	leafSwitchSet := types.SetNull(dRackTypeLeafSwitch{}.attrType())
	if len(in.Data.LeafSwitches) > 0 {
		leafSwitches := make([]dRackTypeLeafSwitch, len(in.Data.LeafSwitches))
		for i := range in.Data.LeafSwitches {
			leafSwitches[i].loadApiResponse(ctx, &in.Data.LeafSwitches[i], in.Data.FabricConnectivityDesign, diags)
			if diags.HasError() {
				return
			}
		}
		leafSwitchSet, d = types.SetValueFrom(ctx, dRackTypeLeafSwitch{}.attrType(), leafSwitches)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}

	// todo re-enable this
	//accessSwitchSet := types.SetNull(dRackTypeAccessSwitch{}.attrType())
	//if len(in.Data.AccessSwitches) > 0 {
	//	accessSwitches := make([]dRackTypeAccessSwitch, len(in.Data.AccessSwitches))
	//	for i := range in.Data.AccessSwitches {
	//		accessSwitches[i].loadApiResponse(&in.Data.AccessSwitches[i])
	//if diags.HasError() {
	//	return
	//}
	//	}
	//	accessSwitchSet, d = types.SetValueFrom(ctx, dRackTypeAccessSwitch{}.attrType(), accessSwitches)
	//	diags.Append(d...)
	//if diags.HasError() {
	//	return
	//}
	//}

	// todo re-enable this
	//genericSystemSet := types.SetNull(dRackTypeGenericSystem{}.attrType())
	//if len(in.Data.GenericSystems) > 0 {
	//	genericSystems := make([]dRackTypeGenericSystem, len(in.Data.GenericSystems))
	//	for i := range in.Data.GenericSystems {
	//		genericSystems[i].loadApiResponse(&in.Data.GenericSystems[i])
	//if diags.HasError() {
	//	return
	//}
	//	}
	//	genericSystemSet, d = types.SetValueFrom(ctx, dRackTypeGenericSystem{}.attrType(), genericSystems)
	//	diags.Append(d...)
	//if diags.HasError() {
	//	return
	//}
	//}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)
	o.Description = types.StringValue(in.Data.Description)
	o.FabricConnectivityDesign = types.StringValue(in.Data.FabricConnectivityDesign.String())
	o.LeafSwitches = leafSwitchSet // todo re-enable this
	//o.AccessSwitches = accessSwitchSet // todo re-enable this
	//o.GenericSystems = genericSystemSet // todo re-enable this
}

type dRackTypeLeafSwitch struct {
	Name               types.String `tfsdk:"name"`
	SpineLinkCount     types.Int64  `tfsdk:"spine_link_count"`
	SpineLinkSpeed     types.String `tfsdk:"spine_link_speed"`
	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
	MlagInfo           types.Object `tfsdk:"mlag_info"`
	//LogicalDevice      types.Object `tfsdk:"logical_device"`
	TagData types.Set `tfsdk:"tag_data"`
}

func (o dRackTypeLeafSwitch) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"spine_link_count":    types.Int64Type,
		"spine_link_speed":    types.StringType,
		"redundancy_protocol": types.StringType,
		"mlag_info":           mlagInfo{}.attrType(),
		//"logical_device":      logicalDeviceData{}.attrType(),
		"tag_data": types.SetType{ElemType: tagData{}.attrType()},
	}
}

func (o dRackTypeLeafSwitch) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *dRackTypeLeafSwitch) loadApiResponse(ctx context.Context, in *goapstra.RackElementLeafSwitch, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) {
	var d diag.Diagnostics

	o.Name = types.StringValue(in.Label)
	switch fcd {
	case goapstra.FabricConnectivityDesignL3Collapsed:
		o.SpineLinkCount = types.Int64Null()
		o.SpineLinkSpeed = types.StringNull()
	case goapstra.FabricConnectivityDesignL3Clos:
		o.SpineLinkCount = types.Int64Value(int64(in.LinkPerSpineCount))
		o.SpineLinkSpeed = types.StringValue(string(in.LinkPerSpineSpeed))
	}

	if in.RedundancyProtocol == goapstra.LeafRedundancyProtocolNone {
		o.RedundancyProtocol = types.StringNull()
	} else {
		o.RedundancyProtocol = types.StringValue(in.RedundancyProtocol.String())
	}

	if in.MlagInfo != nil && in.MlagInfo.LeafLeafLinkCount > 0 {
		var mlagInfo mlagInfo
		mlagInfo.loadApiResponse(ctx, in.MlagInfo, diags)
		if diags.HasError() {
			return
		}
		o.MlagInfo, d = types.ObjectValueFrom(ctx, mlagInfo.attrTypes(), &mlagInfo)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	} else {
		o.MlagInfo = types.ObjectNull(mlagInfo{}.attrTypes())
	}

	o.TagData = newTagSet(ctx, in.Tags, diags)
	if diags.HasError() {
		return
	}

	// o.LogicalDevice.loadApiResponse(in.LogicalDevice)
	//
	//	if len(in.Tags) > 0 {
	//		o.TagData = make([]tagData, len(in.Tags)) // populated below
	//		for i := range in.Tags {
	//			o.TagData[i].loadApiResponse(&in.Tags[i])
	//		}
	//	}
}

type mlagInfo struct {
	MlagKeepaliveVLan       types.Int64  `tfsdk:"mlag_keepalive_vlan"`
	PeerLinkCount           types.Int64  `tfsdk:"peer_link_count"`
	PeerLinkSpeed           types.String `tfsdk:"peer_link_speed"`
	PeerLinkPortChannelId   types.Int64  `tfsdk:"peer_link_port_channel_id"`
	L3PeerLinkCount         types.Int64  `tfsdk:"l3_peer_link_count"`
	L3PeerLinkSpeed         types.String `tfsdk:"l3_peer_link_speed"`
	L3PeerLinkPortChannelId types.Int64  `tfsdk:"l3_peer_link_port_channel_id"`
}

func (o mlagInfo) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mlag_keepalive_vlan":          types.Int64Type,
		"peer_link_count":              types.Int64Type,
		"peer_link_speed":              types.StringType,
		"peer_link_port_channel_id":    types.Int64Type,
		"l3_peer_link_count":           types.Int64Type,
		"l3_peer_link_speed":           types.StringType,
		"l3_peer_link_port_channel_id": types.Int64Type}
}

func (o mlagInfo) attrType() attr.Type {
	return types.ObjectType{
		AttrTypes: o.attrTypes(),
	}
}

func (o *mlagInfo) loadApiResponse(ctx context.Context, in *goapstra.LeafMlagInfo, diags *diag.Diagnostics) {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to load mlagInfo from nil pointer")
	}

	var l3PeerLinkPortChannelId, l3PeerLinkCount types.Int64
	var l3PeerLinkSpeed types.String
	if in.LeafLeafL3LinkCount > 0 {
		l3PeerLinkPortChannelId = types.Int64Value(int64(in.LeafLeafL3LinkPortChannelId))
		l3PeerLinkCount = types.Int64Value(int64(in.LeafLeafL3LinkCount))
		l3PeerLinkSpeed = types.StringValue(string(in.LeafLeafL3LinkSpeed))
	} else {
		l3PeerLinkPortChannelId = types.Int64Null()
		l3PeerLinkCount = types.Int64Null()
		l3PeerLinkSpeed = types.StringNull()
	}

	o.MlagKeepaliveVLan = types.Int64Value(int64(in.MlagVlanId))
	o.PeerLinkCount = types.Int64Value(int64(in.LeafLeafLinkCount))
	o.PeerLinkSpeed = types.StringValue(string(in.LeafLeafLinkSpeed))
	o.PeerLinkPortChannelId = types.Int64Value(int64(in.LeafLeafLinkPortChannelId))
	o.L3PeerLinkCount = l3PeerLinkCount
	o.L3PeerLinkSpeed = l3PeerLinkSpeed
	o.L3PeerLinkPortChannelId = l3PeerLinkPortChannelId
}

//func (o *mlagInfo) request() *goapstra.LeafMlagInfo {
//	if o == nil {
//		return nil
//	}
//
//	var leafLeafL3LinkCount int
//	if o.L3PeerLinkCount != nil {
//		leafLeafL3LinkCount = int(*o.L3PeerLinkCount)
//	}
//
//	var leafLeafL3LinkPortChannelId int
//	if o.L3PeerLinkPortChannelId != nil {
//		leafLeafL3LinkPortChannelId = int(*o.L3PeerLinkPortChannelId)
//	}
//
//	var leafLeafLinkPortChannelId int
//	if o.PeerLinkPortChannelId != nil {
//		leafLeafLinkPortChannelId = int(*o.PeerLinkPortChannelId)
//	}
//
//	var leafLeafL3LinkSpeed goapstra.LogicalDevicePortSpeed
//	if o.L3PeerLinkSpeed != nil {
//		leafLeafL3LinkSpeed = goapstra.LogicalDevicePortSpeed(*o.L3PeerLinkSpeed)
//	}
//
//	return &goapstra.LeafMlagInfo{
//		LeafLeafL3LinkCount:         leafLeafL3LinkCount,
//		LeafLeafL3LinkPortChannelId: leafLeafL3LinkPortChannelId,
//		LeafLeafL3LinkSpeed:         leafLeafL3LinkSpeed,
//		LeafLeafLinkCount:           int(o.PeerLinkCount),
//		LeafLeafLinkPortChannelId:   leafLeafLinkPortChannelId,
//		LeafLeafLinkSpeed:           goapstra.LogicalDevicePortSpeed(o.PeerLinkSpeed),
//		MlagVlanId:                  int(o.MlagKeepaliveVLan),
//	}
//}

//type dRackTypeAccessSwitch struct {
//	Name               types.String `tfsdk:"name"`
//	Count              types.Int64  `tfsdk:"count"`
//	RedundancyProtocol types.String `tfsdk:"redundancy_protocol"`
//	EsiLagInfo         types.Object `tfsdk:"esi_lag_info"`
//	LogicalDevice      types.Object `tfsdk:"logical_device"`
//	TagData            types.Set    `tfsdk:"tag_data"`
//	Links              types.Set    `tfsdk:"links"`
//}
//
//func (o dRackTypeAccessSwitch) attrType() attr.Type {
//	return types.ObjectType{
//		AttrTypes: map[string]attr.Type{
//			"name":                types.StringType,
//			"count":               types.Int64Type,
//			"redundancy_protocol": types.StringType,
//			"esi_lag_info":        esiLagInfo{}.attrType(),
//			"logical_device":      logicalDeviceData{}.attrType(),
//			"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
//			"links":               types.SetType{ElemType: dRackLink{}.attrType()}}}
//}
//
//func (o *dRackTypeAccessSwitch) loadApiResponse(in *goapstra.RackElementAccessSwitch) {
//	o.Name = in.Label
//	o.Count = int64(in.InstanceCount)
//	if in.RedundancyProtocol != goapstra.AccessRedundancyProtocolNone {
//		redundancyProtocol := in.RedundancyProtocol.String()
//		o.RedundancyProtocol = &redundancyProtocol
//	}
//	if in.EsiLagInfo != nil {
//		o.EsiLagInfo = &esiLagInfo{}
//		o.EsiLagInfo.loadApiResponse(in.EsiLagInfo)
//	}
//	o.LogicalDevice.loadApiResponse(in.LogicalDevice)
//
//	if len(in.Tags) > 0 {
//		o.TagData = make([]tagData, len(in.Tags)) // populated below
//		for i := range in.Tags {
//			o.TagData[i].loadApiResponse(&in.Tags[i])
//		}
//	}
//
//	o.Links = make([]dRackLink, len(in.Links))
//	for i := range in.Links {
//		o.Links[i].loadApiResponse(&in.Links[i])
//	}
//}
//
//type esiLagInfo struct {
//	L3PeerLinkCount int64  `tfsdk:"l3_peer_link_count"`
//	L3PeerLinkSpeed string `tfsdk:"l3_peer_link_speed"`
//}
//
//func (o esiLagInfo) attrType() attr.Type {
//	return types.ObjectType{
//		AttrTypes: map[string]attr.Type{
//			"l3_peer_link_count": types.Int64Type,
//			"l3_peer_link_speed": types.StringType}}
//}
//
//func (o *esiLagInfo) loadApiResponse(in *goapstra.EsiLagInfo) {
//	o.L3PeerLinkCount = int64(in.AccessAccessLinkCount)
//	o.L3PeerLinkSpeed = string(in.AccessAccessLinkSpeed)
//}
//
//func dLinksAttributeSchema() schema.SetNestedAttribute {
//	return schema.SetNestedAttribute{
//		MarkdownDescription: "Details links from this Element to switches upstream switches within this Rack Type.",
//		Computed:            true,
//		Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
//		NestedObject: schema.NestedAttributeObject{
//			Attributes: map[string]schema.Attribute{
//				"name": schema.StringAttribute{
//					MarkdownDescription: "Name of this link.",
//					Computed:            true,
//				},
//				"target_switch_name": schema.StringAttribute{
//					MarkdownDescription: "The `name` of the switch in this Rack Type to which this Link connects.",
//					Computed:            true,
//				},
//				"lag_mode": schema.StringAttribute{
//					MarkdownDescription: "LAG negotiation mode of the Link.",
//					Computed:            true,
//				},
//				"links_per_switch": schema.Int64Attribute{
//					MarkdownDescription: "Number of Links to each switch.",
//					Computed:            true,
//				},
//				"speed": schema.StringAttribute{
//					MarkdownDescription: "Speed of this Link.",
//					Computed:            true,
//				},
//				"switch_peer": schema.StringAttribute{
//					MarkdownDescription: "For non-lAG connections to redundant switch pairs, this field selects the target switch.",
//					Computed:            true,
//				},
//				"tag_data": tagsDataAttributeSchema(),
//			},
//		},
//	}
//}

//type dRackLink struct {
//	Name             string    `tfsdk:"name"`
//	TargetSwitchName string    `tfsdk:"target_switch_name"`
//	LagMode          *string   `tfsdk:"lag_mode"`
//	LinksPerSwitch   int64     `tfsdk:"links_per_switch"`
//	Speed            string    `tfsdk:"speed"`
//	SwitchPeer       *string   `tfsdk:"switch_peer"`
//	TagData          []tagData `tfsdk:"tag_data"`
//}
//
//func (o dRackLink) attrType() attr.Type {
//	return types.ObjectType{
//		AttrTypes: map[string]attr.Type{
//			"name":               types.StringType,
//			"target_switch_name": types.StringType,
//			"lag_mode":           types.StringType,
//			"links_per_switch":   types.Int64Type,
//			"speed":              types.StringType,
//			"switch_peer":        types.StringType,
//			"tag_data":           types.SetType{ElemType: tagData{}.attrType()}}}
//}
//
//func (o *dRackLink) loadApiResponse(in *goapstra.RackLink) {
//	o.Name = in.Label
//	o.TargetSwitchName = in.TargetSwitchLabel
//	if in.LagMode != goapstra.RackLinkLagModeNone {
//		lagMode := in.LagMode.String()
//		o.LagMode = &lagMode
//	}
//	o.LinksPerSwitch = int64(in.LinkPerSwitchCount)
//	o.Speed = string(in.LinkSpeed)
//	if in.SwitchPeer != goapstra.RackLinkSwitchPeerNone {
//		switchPeer := in.SwitchPeer.String()
//		o.SwitchPeer = &switchPeer
//	}
//
//	if len(in.Tags) > 0 {
//		o.TagData = make([]tagData, len(in.Tags)) // populated below
//		for i := range in.Tags {
//			o.TagData[i].loadApiResponse(&in.Tags[i])
//		}
//	}
//}
//
//type dRackTypeGenericSystem struct {
//	Name             string            `tfsdk:"name"`
//	Count            int64             `tfsdk:"count"`
//	PortChannelIdMin int64             `tfsdk:"port_channel_id_min"`
//	PortChannelIdMax int64             `tfsdk:"port_channel_id_max"`
//	LogicalDevice    logicalDeviceData `tfsdk:"logical_device"`
//	TagData          []tagData         `tfsdk:"tag_data"`
//	Links            []dRackLink       `tfsdk:"links"`
//}
//
//func (o dRackTypeGenericSystem) attrType() attr.Type {
//	return types.ObjectType{
//		AttrTypes: map[string]attr.Type{
//			"name":                types.StringType,
//			"count":               types.Int64Type,
//			"port_channel_id_min": types.Int64Type,
//			"port_channel_id_max": types.Int64Type,
//			"logical_device":      logicalDeviceData{}.attrType(),
//			"tag_data":            types.SetType{ElemType: tagData{}.attrType()},
//			"links":               types.SetType{ElemType: dRackLink{}.attrType()}}}
//}
//
//func (o *dRackTypeGenericSystem) loadApiResponse(in *goapstra.RackElementGenericSystem) {
//	o.Name = in.Label
//	o.Count = int64(in.Count)
//	o.PortChannelIdMin = int64(in.PortChannelIdMin)
//	o.PortChannelIdMax = int64(in.PortChannelIdMax)
//	o.LogicalDevice.loadApiResponse(in.LogicalDevice)
//	o.Links = make([]dRackLink, len(in.Links))
//
//	if len(in.Tags) > 0 {
//		o.TagData = make([]tagData, len(in.Tags)) // populated below
//		for i := range in.Tags {
//			o.TagData[i].loadApiResponse(&in.Tags[i])
//		}
//	}
//
//	for i := range in.Links {
//		o.Links[i].loadApiResponse(&in.Links[i])
//	}
//}

func leafSwitchAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
			Computed:            true,
		},
		"spine_link_count": schema.Int64Attribute{
			MarkdownDescription: "Number of links to each spine switch.",
			Computed:            true,
		},
		"spine_link_speed": schema.StringAttribute{
			MarkdownDescription: "Speed of links to spine switches.",
			Computed:            true,
		},
		"redundancy_protocol": schema.StringAttribute{
			MarkdownDescription: "Indicates whether 'the switch' is actually a LAG-capable redundant pair and if so, what type.",
			Computed:            true,
		},
		"mlag_info": schema.SingleNestedAttribute{
			MarkdownDescription: "Details settings when the Leaf Switch is an MLAG-capable pair.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"mlag_keepalive_vlan": schema.Int64Attribute{
					MarkdownDescription: "MLAG keepalive VLAN ID.",
					Computed:            true,
				},
				"peer_link_count": schema.Int64Attribute{
					MarkdownDescription: "Number of links between MLAG devices.",
					Computed:            true,
				},
				"peer_link_speed": schema.StringAttribute{
					MarkdownDescription: "Speed of links between MLAG devices.",
					Computed:            true,
				},
				"peer_link_port_channel_id": schema.Int64Attribute{
					MarkdownDescription: "Peer link port-channel ID.",
					Computed:            true,
				},
				"l3_peer_link_count": schema.Int64Attribute{
					MarkdownDescription: "Number of L3 links between MLAG devices.",
					Computed:            true,
				},
				"l3_peer_link_speed": schema.StringAttribute{
					MarkdownDescription: "Speed of l3 links between MLAG devices.",
					Computed:            true,
				},
				"l3_peer_link_port_channel_id": schema.Int64Attribute{
					MarkdownDescription: "L3 peer link port-channel ID.",
					Computed:            true,
				},
			},
		},
		//"logical_device": logicalDeviceDataAttributeSchema(),
		"tag_data": tagsDataAttributeSchema(),
	}
}

//func accessSwitchAttributes() map[string]schema.Attribute {
//	return map[string]schema.Attribute{
//		"name": schema.StringAttribute{
//			MarkdownDescription: "Switch name, used when creating intra-rack links targeting this switch.",
//			Computed:            true,
//		},
//		"count": schema.Int64Attribute{
//			MarkdownDescription: "Count of Access Switches of this type.",
//			Computed:            true,
//		},
//		"redundancy_protocol": schema.StringAttribute{
//			MarkdownDescription: "Indicates whether 'the switch' is actually a LAG-capable redundant pair and if so, what type.",
//			Computed:            true,
//		},
//		"esi_lag_info": schema.SingleNestedAttribute{
//			MarkdownDescription: "Interconnect information for Access Switches in ESI-LAG redundancy mode.",
//			Computed:            true,
//			Attributes: map[string]schema.Attribute{
//				"l3_peer_link_count": schema.Int64Attribute{
//					MarkdownDescription: "Count of L3 links to ESI peer.",
//					Computed:            true,
//				},
//				"l3_peer_link_speed": schema.StringAttribute{
//					MarkdownDescription: "Speed of L3 links to ESI peer.",
//					Computed:            true,
//				},
//			},
//		},
//		"logical_device": logicalDeviceDataAttributeSchema(),
//		"tag_data":       tagsDataAttributeSchema(),
//		"links":          dLinksAttributeSchema(),
//	}
//}

//func genericSystemAttributes() map[string]schema.Attribute {
//	return map[string]schema.Attribute{
//		"name": schema.StringAttribute{
//			MarkdownDescription: "Generic name, must be unique within the rack-type.",
//			Computed:            true,
//		},
//		"count": schema.Int64Attribute{
//			MarkdownDescription: "Number of Generic Systems of this type.",
//			Computed:            true,
//		},
//		"port_channel_id_min": schema.Int64Attribute{
//			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
//			Computed:            true,
//		},
//		"port_channel_id_max": schema.Int64Attribute{
//			MarkdownDescription: "Port channel IDs are used when rendering leaf device port-channel configuration towards generic systems.",
//			Computed:            true,
//		},
//		"logical_device": logicalDeviceDataAttributeSchema(),
//		"tag_data":       tagsDataAttributeSchema(),
//		"links":          dLinksAttributeSchema(),
//	}
//}

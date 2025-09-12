package design

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RackType struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Map    `tfsdk:"leaf_switches"`
	AccessSwitches           types.Map    `tfsdk:"access_switches"`
	GenericSystems           types.Map    `tfsdk:"generic_systems"`
}

func (o RackType) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Rack Type. Required when `name` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Web UI name of the Type. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type description displayed in the Apstra web UI.",
			Computed:            true,
		},
		"fabric_connectivity_design": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Indicates designs for which this Rack Type is intended.",
			Computed:            true,
		},
		"leaf_switches": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of Leaf Switches in this Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: LeafSwitch{}.DataSourceAttributes(),
			},
		},
		"access_switches": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of Access Switches in this Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: AccessSwitch{}.DataSourceAttributes(),
			},
		},
		"generic_systems": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of Generic Systems in the Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: GenericSystem{}.DataSourceAttributes(),
			},
		},
	}
}

func (o RackType) DataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "IDs will always be `<null>` in nested contexts.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type name displayed in the Apstra web UI.",
			Computed:            true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type description displayed in the Apstra web UI.",
			Computed:            true,
		},
		"fabric_connectivity_design": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Indicates designs for which this Rack Type is intended.",
			Computed:            true,
		},
		"leaf_switches": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of Leaf Switches in this Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: LeafSwitch{}.DataSourceAttributes(),
			},
		},
		"access_switches": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of Access Switches in this Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: AccessSwitch{}.DataSourceAttributes(),
			},
		},
		"generic_systems": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of Generic Systems in the Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: GenericSystem{}.DataSourceAttributes(),
			},
		},
	}
}

func (o RackType) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Object ID for the Rack Type, assigned by Apstra.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type name, displayed in the Apstra web UI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthBetween(1, 17)},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type description, displayed in the Apstra web UI.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"fabric_connectivity_design": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Must be one of '%s'.", strings.Join(utils.FcdModes(), "', '")),
			Required:            true,
			Validators:          []validator.String{stringvalidator.OneOf(utils.FcdModes()...)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"leaf_switches": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Each Rack Type is required to have at least one Leaf Switch.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: LeafSwitch{}.ResourceAttributes(),
			},
		},
		"access_switches": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Access Switches are optional, link to Leaf Switches in the same rack",
			Optional:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: AccessSwitch{}.ResourceAttributes(),
			},
		},
		"generic_systems": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Generic Systems are optional rack elements not" +
				"managed by Apstra: Servers, routers, firewalls, etc...",
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: GenericSystem{}.ResourceAttributes(),
			},
		},
	}
}

func (o RackType) ResourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID will always be `<null>` in nested contexts.",
			Computed:            true,
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type name, displayed in the Apstra web UI.",
			Computed:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type description, displayed in the Apstra web UI.",
			Computed:            true,
		},
		"fabric_connectivity_design": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Must be one of '%s'.", strings.Join(utils.FcdModes(), "', '")),
			Computed:            true,
		},
		"leaf_switches": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Each Rack Type is required to have at least one Leaf Switch.",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: LeafSwitch{}.ResourceAttributesNested(),
			},
		},
		"access_switches": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Access Switches are optional, link to Leaf Switches in the same rack",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: AccessSwitch{}.ResourceAttributesNested(),
			},
		},
		"generic_systems": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Generic Systems are optional rack elements not" +
				"managed by Apstra: Servers, routers, firewalls, etc...",
			Computed: true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: GenericSystem{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o RackType) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                         types.StringType,
		"name":                       types.StringType,
		"description":                types.StringType,
		"fabric_connectivity_design": types.StringType,
		"leaf_switches":              types.MapType{ElemType: types.ObjectType{AttrTypes: LeafSwitch{}.AttrTypes()}},
		"access_switches":            types.MapType{ElemType: types.ObjectType{AttrTypes: AccessSwitch{}.AttrTypes()}},
		"generic_systems":            types.MapType{ElemType: types.ObjectType{AttrTypes: GenericSystem{}.AttrTypes()}},
	}
}

func (o *RackType) LoadApiData(ctx context.Context, in *apstra.RackTypeData, diags *diag.Diagnostics) {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to load RackType data from nil source")
		return
	}
	switch in.FabricConnectivityDesign {
	case enum.FabricConnectivityDesignL3Collapsed: // this FCD is supported
	case enum.FabricConnectivityDesignL3Clos: // this FCD is supported
	default: // this FCD is unsupported
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("Rack Type has unsupported Fabric Connectivity Design %q",
				in.FabricConnectivityDesign.String()))
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Description = utils.StringValueOrNull(ctx, in.Description, diags)
	o.FabricConnectivityDesign = types.StringValue(in.FabricConnectivityDesign.String())
	o.LeafSwitches = NewLeafSwitchMap(ctx, in.LeafSwitches, in.FabricConnectivityDesign, diags)
	o.AccessSwitches = NewAccessSwitchMap(ctx, in.AccessSwitches, diags)
	o.GenericSystems = NewGenericSystemMap(ctx, in.GenericSystems, diags)
}

func (o *RackType) GetFabricConnectivityDesign(_ context.Context, diags *diag.Diagnostics) enum.FabricConnectivityDesign {
	var fcd enum.FabricConnectivityDesign
	err := fcd.FromString(o.FabricConnectivityDesign.ValueString())
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing FCD '%s' - %s",
				o.FabricConnectivityDesign.ValueString(), err.Error()))
	}
	return fcd
}

func (o *RackType) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.RackTypeRequest {
	fcd := o.GetFabricConnectivityDesign(ctx, diags)
	if diags.HasError() {
		return nil
	}

	leafSwitches := o.LeafSwitchMap(ctx, diags)
	if diags.HasError() {
		return nil
	}

	accessSwitches := o.AccessSwitchMap(ctx, diags)
	if diags.HasError() {
		return nil
	}

	genericSystems := o.GenericSystemMap(ctx, diags)
	if diags.HasError() {
		return nil
	}

	var i int

	leafSwitchRequests := make([]apstra.RackElementLeafSwitchRequest, len(leafSwitches))
	i = 0
	for name, ls := range leafSwitches {
		req := ls.Request(ctx, path.Root("leaf_switches").AtMapKey(name), fcd, diags)
		if diags.HasError() {
			return nil
		}
		req.Label = name
		leafSwitchRequests[i] = *req
		i++
	}

	accessSwitchRequests := make([]apstra.RackElementAccessSwitchRequest, len(accessSwitches))
	i = 0
	for name, as := range accessSwitches {
		req := as.Request(ctx, path.Root("access_switches").AtMapKey(name), o, diags)
		if diags.HasError() {
			return nil
		}
		req.Label = name
		accessSwitchRequests[i] = *req
		i++
	}

	genericSystemsRequests := make([]apstra.RackElementGenericSystemRequest, len(genericSystems))
	i = 0
	for name, gs := range genericSystems {
		req := gs.Request(ctx, path.Root("generic_systems").AtMapKey(name), o, diags)
		if diags.HasError() {
			return nil
		}
		req.Label = name
		genericSystemsRequests[i] = *req
		i++
	}

	// sort the request slices so that leaf/access/generic wind up in predictable order
	sort.Slice(leafSwitchRequests, func(i, j int) bool {
		return leafSwitchRequests[i].Label < leafSwitchRequests[j].Label
	})
	sort.Slice(accessSwitchRequests, func(i, j int) bool {
		return accessSwitchRequests[i].Label < accessSwitchRequests[j].Label
	})
	sort.Slice(genericSystemsRequests, func(i, j int) bool {
		return genericSystemsRequests[i].Label < genericSystemsRequests[j].Label
	})

	return &apstra.RackTypeRequest{
		DisplayName:              o.Name.ValueString(),
		Description:              o.Description.ValueString(),
		FabricConnectivityDesign: fcd,
		LeafSwitches:             leafSwitchRequests,
		AccessSwitches:           accessSwitchRequests,
		GenericSystems:           genericSystemsRequests,
	}
}

func NewRackTypeObject(ctx context.Context, in *apstra.RackTypeData, diags *diag.Diagnostics) types.Object {
	var rt RackType
	rt.Id = types.StringNull()
	rt.LoadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(RackType{}.AttrTypes())
	}

	rtdObj, d := types.ObjectValueFrom(ctx, RackType{}.AttrTypes(), &rt)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(RackType{}.AttrTypes())
	}

	return rtdObj
}

func ValidateFcdSupport(_ context.Context, fcd enum.FabricConnectivityDesign, diags *diag.Diagnostics) {
	switch fcd {
	case enum.FabricConnectivityDesignL3Collapsed: // this FCD is supported
	case enum.FabricConnectivityDesignL3Clos: // this FCD is supported
	default: // this FCD is unsupported
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("Unsupported Fabric Connectivity Design '%s'",
				fcd.String()))
	}
}

func ValidateRackType(ctx context.Context, in *apstra.RackType, diags *diag.Diagnostics) {
	if in.Data == nil {
		diags.AddError("rack type has no data", fmt.Sprintf("rack type '%s' data object is nil", in.Id))
		return
	}

	ValidateFcdSupport(ctx, in.Data.FabricConnectivityDesign, diags)
	if diags.HasError() {
		return
	}

	for i := range in.Data.LeafSwitches {
		ValidateLeafSwitch(in, i, diags)
	}

	for i := range in.Data.AccessSwitches {
		ValidateAccessSwitch(in, i, diags)
	}

	for i := range in.Data.GenericSystems {
		ValidateGenericSystem(in, i, diags)
	}
}

func (o *RackType) LeafSwitchMap(ctx context.Context, diags *diag.Diagnostics) map[string]LeafSwitch {
	leafSwitches := make(map[string]LeafSwitch, len(o.LeafSwitches.Elements()))
	d := o.LeafSwitches.ElementsAs(ctx, &leafSwitches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	return leafSwitches
}

func (o *RackType) leafSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *LeafSwitch {
	leafSwitches := o.LeafSwitchMap(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if ls, ok := leafSwitches[requested]; ok {
		return &ls
	}

	return nil
}

func (o *RackType) AccessSwitchMap(ctx context.Context, diags *diag.Diagnostics) map[string]AccessSwitch {
	accessSwitches := make(map[string]AccessSwitch, len(o.AccessSwitches.Elements()))
	d := o.AccessSwitches.ElementsAs(ctx, &accessSwitches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	return accessSwitches
}

func (o *RackType) accessSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *AccessSwitch {
	accessSwitches := o.AccessSwitchMap(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if as, ok := accessSwitches[requested]; ok {
		return &as
	}

	return nil
}

func (o *RackType) GenericSystemMap(ctx context.Context, diags *diag.Diagnostics) map[string]GenericSystem {
	genericSystems := make(map[string]GenericSystem, len(o.GenericSystems.Elements()))
	d := o.GenericSystems.ElementsAs(ctx, &genericSystems, true)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	return genericSystems
}

//func (o *RackType) genericSystemByName(ctx context.Context, requested string, diags *diag.Diagnostics) *GenericSystem {
//	GenericSystemMap := o.GenericSystemMap(ctx, diags)
//	if diags.HasError() {
//		return nil
//	}
//
//	if gs, ok := GenericSystemMap[requested]; ok {
//		return &gs
//	}
//
//	return nil
//}

// CopyWriteOnlyElements copies elements (IDs of nested design API objects)
// from 'src' (plan or state - something which knows these facts) into 'o' a
// RackType to be used as state.
func (o *RackType) CopyWriteOnlyElements(ctx context.Context, src *RackType, diags *diag.Diagnostics) {
	// first extract native go structs from the TF set of objects
	dstLeafSwitches := o.LeafSwitchMap(ctx, diags)
	dstAccessSwitches := o.AccessSwitchMap(ctx, diags)
	dstGenericSystems := o.GenericSystemMap(ctx, diags)

	// invoke the CopyWriteOnlyElements on every leaf switch object
	for name, dstLeafSwitch := range dstLeafSwitches {
		srcLeafSwitch, ok := src.LeafSwitchMap(ctx, diags)[name]
		if !ok {
			continue
		}
		if diags.HasError() {
			return
		}

		dstLeafSwitch.CopyWriteOnlyElements(ctx, &srcLeafSwitch, diags)
		if diags.HasError() {
			return
		}
		dstLeafSwitches[name] = dstLeafSwitch
	}

	// invoke the CopyWriteOnlyElements on every access switch object
	for name, dstAccessSwitch := range dstAccessSwitches {
		srcAccessSwitch, ok := src.AccessSwitchMap(ctx, diags)[name]
		if !ok {
			continue
		}
		if diags.HasError() {
			return
		}

		dstAccessSwitch.CopyWriteOnlyElements(ctx, &srcAccessSwitch, diags)
		if diags.HasError() {
			return
		}
		dstAccessSwitches[name] = dstAccessSwitch
	}

	// invoke the CopyWriteOnlyElements on every generic system object
	for name, dstGenericSystem := range dstGenericSystems {
		srcGenericSystem, ok := src.GenericSystemMap(ctx, diags)[name]
		if !ok {
			continue
		}
		if diags.HasError() {
			return
		}

		dstGenericSystem.CopyWriteOnlyElements(ctx, &srcGenericSystem, diags)
		if diags.HasError() {
			return
		}
		dstGenericSystems[name] = dstGenericSystem
	}

	// transform the native go objects (with copied object IDs) back to TF set
	leafSwitchMap := utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: LeafSwitch{}.AttrTypes()}, dstLeafSwitches, diags)
	accessSwitchMap := utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: AccessSwitch{}.AttrTypes()}, dstAccessSwitches, diags)
	genericSystemMap := utils.MapValueOrNull(ctx, types.ObjectType{AttrTypes: GenericSystem{}.AttrTypes()}, dstGenericSystems, diags)
	if diags.HasError() {
		return
	}

	// save the TF sets into RackType
	o.LeafSwitches = leafSwitchMap
	o.AccessSwitches = accessSwitchMap
	o.GenericSystems = genericSystemMap
}

func (o *RackType) GetSwitchRedundancyProtocolByName(ctx context.Context, name string, path path.Path, diags *diag.Diagnostics) fmt.Stringer {
	leaf := o.leafSwitchByName(ctx, name, diags)
	access := o.accessSwitchByName(ctx, name, diags)
	if leaf == nil && access == nil {
		diags.AddAttributeError(path, errInvalidConfig,
			fmt.Sprintf("target switch %q not found in rack type %q", name, o.Id))
		return nil
	}
	if leaf != nil && access != nil {
		diags.AddError(errProviderBug, "link seems to be attached to both leaf and access switches")
		return nil
	}

	var leafRedundancyProtocol apstra.LeafRedundancyProtocol
	if leaf != nil {
		if leaf.RedundancyProtocol.IsNull() {
			return apstra.LeafRedundancyProtocolNone
		}
		err := leafRedundancyProtocol.FromString(leaf.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing leaf switch redundancy protocol", err.Error())
			return nil
		}
		return leafRedundancyProtocol
	}

	var accessRedundancyProtocol apstra.AccessRedundancyProtocol
	if access != nil {
		if !access.EsiLagInfo.IsNull() {
			return apstra.AccessRedundancyProtocolEsi
		}
		if access.RedundancyProtocol.IsNull() {
			return apstra.AccessRedundancyProtocolNone
		}
		err := accessRedundancyProtocol.FromString(access.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing access switch redundancy protocol", err.Error())
			return nil
		}
		return accessRedundancyProtocol
	}
	diags.AddError(errProviderBug, "somehow we've reached the end of GetSwitchRedundancyProtocolByName without finding a solution")
	return nil
}

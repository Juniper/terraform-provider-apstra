package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
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
	"strings"
)

type rackType struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Map    `tfsdk:"leaf_switches"`
	AccessSwitches           types.Map    `tfsdk:"access_switches"`
	GenericSystems           types.Map    `tfsdk:"generic_systems"`
}

func (o rackType) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type ID.  Required when the Rack Type name is omitted.",
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
			MarkdownDescription: "Rack Type name displayed in the Apstra web UI.  Required when Rack Type ID is omitted.",
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
				Attributes: leafSwitch{}.dataSourceAttributes(),
			},
		},
		"access_switches": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of Access Switches in this Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: accessSwitch{}.dataSourceAttributes(),
			},
		},
		"generic_systems": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of Generic Systems in the Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: genericSystem{}.dataSourceAttributes(),
			},
		},
	}
}

func (o rackType) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Object ID for the Rack Type, assigned by Apstra.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type name, displayed in the Apstra web UI.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type description, displayed in the Apstra web UI.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"fabric_connectivity_design": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Must be one of '%s'.", strings.Join(fcdModes(), "', '")),
			Required:            true,
			Validators:          []validator.String{stringvalidator.OneOf(fcdModes()...)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"leaf_switches": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Each Rack Type is required to have at least one Leaf Switch.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: leafSwitch{}.resourceAttributes(),
			},
		},
		"access_switches": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Access Switches are optional, link to Leaf Switches in the same rack",
			Optional:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: accessSwitch{}.resourceAttributes(),
			},
		},
		"generic_systems": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Generic Systems are rack elements not" +
				"managed by Apstra: Servers, routers, firewalls, etc...",
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: genericSystem{}.resourceAttributes(),
			},
		},
	}
}

func (o rackType) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                         types.StringType,
		"name":                       types.StringType,
		"description":                types.StringType,
		"fabric_connectivity_design": types.StringType,
		"leaf_switches":              types.MapType{ElemType: types.ObjectType{AttrTypes: leafSwitch{}.attrTypes()}},
		"access_switches":            types.MapType{ElemType: types.ObjectType{AttrTypes: accessSwitch{}.attrTypes()}},
		"generic_systems":            types.MapType{ElemType: types.ObjectType{AttrTypes: genericSystem{}.attrTypes()}},
	}
}

func (o *rackType) loadApiData(ctx context.Context, in *goapstra.RackTypeData, diags *diag.Diagnostics) {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to load rackType data from nil source")
		return
	}
	switch in.FabricConnectivityDesign {
	case goapstra.FabricConnectivityDesignL3Collapsed: // this FCD is supported
	case goapstra.FabricConnectivityDesignL3Clos: // this FCD is supported
	default: // this FCD is unsupported
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("Rack Type has unsupported Fabric Connectivity Design %q",
				in.FabricConnectivityDesign.String()))
	}

	//leafSwitches := make(map[string]leafSwitch, len(in.LeafSwitches))
	//for _, leafIn := range in.LeafSwitches {
	//	var ls leafSwitch
	//	ls.loadApiData(ctx, &leafIn, in.FabricConnectivityDesign, diags)
	//	leafSwitches[leafIn.Label] = ls
	//	if diags.HasError() {
	//		return
	//	}
	//}
	//
	//accessSwitches := make(map[string]accessSwitch, len(in.AccessSwitches))
	//for _, accessIn := range in.AccessSwitches {
	//	var as accessSwitch
	//	as.loadApiData(ctx, &accessIn, diags)
	//	accessSwitches[accessIn.Label] = as
	//	if diags.HasError() {
	//		return
	//	}
	//}
	//
	//genericSystems := make(map[string]genericSystem, len(in.GenericSystems))
	//for _, genericIn := range in.GenericSystems {
	//	var gs genericSystem
	//	gs.loadApiData(ctx, &genericIn, diags)
	//	genericSystems[genericIn.Label] = gs
	//	if diags.HasError() {
	//		return
	//	}
	//}

	o.Name = types.StringValue(in.DisplayName)
	o.Description = stringValueOrNull(ctx, in.Description, diags)
	o.FabricConnectivityDesign = types.StringValue(in.FabricConnectivityDesign.String())
	//o.LeafSwitches = mapValueOrNull(ctx, types.ObjectType{AttrTypes: leafSwitch{}.attrTypes()}, leafSwitches, diags)
	//o.AccessSwitches = mapValueOrNull(ctx, types.ObjectType{AttrTypes: accessSwitch{}.attrTypes()}, accessSwitches, diags)
	//o.GenericSystems = mapValueOrNull(ctx, types.ObjectType{AttrTypes: genericSystem{}.attrTypes()}, genericSystems, diags)
	//todo use these new functions
	o.LeafSwitches = newLeafSwitchMap(ctx, in.LeafSwitches, in.FabricConnectivityDesign, diags)
	o.AccessSwitches = newAccessSwitchMap(ctx, in.AccessSwitches, diags)
	o.GenericSystems = newGenericSystemMap(ctx, in.GenericSystems, diags)
}

func (o *rackType) fabricConnectivityDesign(_ context.Context, diags *diag.Diagnostics) goapstra.FabricConnectivityDesign {
	var fcd goapstra.FabricConnectivityDesign
	err := fcd.FromString(o.FabricConnectivityDesign.ValueString())
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing FCD '%s' - %s",
				o.FabricConnectivityDesign.ValueString(), err.Error()))
	}
	return fcd
}

func (o *rackType) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.RackTypeRequest {
	fcd := o.fabricConnectivityDesign(ctx, diags)
	if diags.HasError() {
		return nil
	}

	leafSwitches := o.leafSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	accessSwitches := o.accessSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	genericSystems := o.genericSystems(ctx, diags)
	if diags.HasError() {
		return nil
	}

	var i int

	leafSwitchRequests := make([]goapstra.RackElementLeafSwitchRequest, len(leafSwitches))
	i = 0
	for name, ls := range leafSwitches {
		req := ls.request(ctx, path.Root("leaf_switches").AtMapKey(name), fcd, diags)
		if diags.HasError() {
			return nil
		}
		req.Label = name
		leafSwitchRequests[i] = *req
		i++
	}

	accessSwitchRequests := make([]goapstra.RackElementAccessSwitchRequest, len(accessSwitches))
	i = 0
	for name, as := range accessSwitches {
		req := as.request(ctx, path.Root("access_switches").AtMapKey(name), o, diags)
		if diags.HasError() {
			return nil
		}
		req.Label = name
		accessSwitchRequests[i] = *req
		i++
	}

	genericSystemsRequests := make([]goapstra.RackElementGenericSystemRequest, len(genericSystems))
	i = 0
	for name, gs := range genericSystems {
		req := gs.request(ctx, path.Root("generic_systems").AtMapKey(name), o, diags)
		if diags.HasError() {
			return nil
		}
		req.Label = name
		genericSystemsRequests[i] = *req
		i++
	}

	return &goapstra.RackTypeRequest{
		DisplayName:              o.Name.ValueString(),
		Description:              o.Description.ValueString(),
		FabricConnectivityDesign: fcd,
		LeafSwitches:             leafSwitchRequests,
		AccessSwitches:           accessSwitchRequests,
		GenericSystems:           genericSystemsRequests,
	}
}

func newRackTypeObject(ctx context.Context, in *goapstra.RackTypeData, diags *diag.Diagnostics) types.Object {
	var rt rackType
	rt.Id = types.StringNull()
	rt.loadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(rackType{}.attrTypes())
	}

	rtdObj, d := types.ObjectValueFrom(ctx, rackType{}.attrTypes(), &rt)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(rackType{}.attrTypes())
	}

	return rtdObj
}

func validateFcdSupport(_ context.Context, fcd goapstra.FabricConnectivityDesign, diags *diag.Diagnostics) {
	switch fcd {
	case goapstra.FabricConnectivityDesignL3Collapsed: // this FCD is supported
	case goapstra.FabricConnectivityDesignL3Clos: // this FCD is supported
	default: // this FCD is unsupported
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("Unsupported Fabric Connectivity Design '%s'",
				fcd.String()))
	}
}

func validateRackType(ctx context.Context, in *goapstra.RackType, diags *diag.Diagnostics) {
	if in.Data == nil {
		diags.AddError("rack type has no data", fmt.Sprintf("rack type '%s' data object is nil", in.Id))
		return
	}

	validateFcdSupport(ctx, in.Data.FabricConnectivityDesign, diags)
	if diags.HasError() {
		return
	}

	for i := range in.Data.LeafSwitches {
		validateLeafSwitch(in, i, diags)
	}

	for i := range in.Data.AccessSwitches {
		validateAccessSwitch(in, i, diags)
	}

	for i := range in.Data.GenericSystems {
		validateGenericSystem(in, i, diags)
	}
}

func (o *rackType) leafSwitches(ctx context.Context, diags *diag.Diagnostics) map[string]leafSwitch {
	leafSwitches := make(map[string]leafSwitch, len(o.LeafSwitches.Elements()))
	d := o.LeafSwitches.ElementsAs(ctx, &leafSwitches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	return leafSwitches
}

func (o *rackType) leafSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *leafSwitch {
	leafSwitches := o.leafSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if ls, ok := leafSwitches[requested]; ok {
		return &ls
	}

	return nil
}

func (o *rackType) accessSwitches(ctx context.Context, diags *diag.Diagnostics) map[string]accessSwitch {
	accessSwitches := make(map[string]accessSwitch, len(o.AccessSwitches.Elements()))
	d := o.AccessSwitches.ElementsAs(ctx, &accessSwitches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	return accessSwitches
}

func (o *rackType) accessSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *accessSwitch {
	accessSwitches := o.accessSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if as, ok := accessSwitches[requested]; ok {
		return &as
	}

	return nil
}

func (o *rackType) genericSystems(ctx context.Context, diags *diag.Diagnostics) map[string]genericSystem {
	genericSystems := make(map[string]genericSystem, len(o.GenericSystems.Elements()))
	d := o.GenericSystems.ElementsAs(ctx, &genericSystems, true)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	return genericSystems
}

func (o *rackType) genericSystemByName(ctx context.Context, requested string, diags *diag.Diagnostics) *genericSystem {
	genericSystems := o.genericSystems(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if gs, ok := genericSystems[requested]; ok {
		return &gs
	}

	return nil
}

// copyWriteOnlyElements copies elements (IDs of nested design API objects)
// from 'src' (plan or state - something which knows these facts) into 'o' a
// rackType to be used as state.
func (o *rackType) copyWriteOnlyElements(ctx context.Context, src *rackType, diags *diag.Diagnostics) {
	// first extract native go structs from the TF set of objects
	dstLeafSwitches := o.leafSwitches(ctx, diags)
	dstAccessSwitches := o.accessSwitches(ctx, diags)
	dstGenericSystems := o.genericSystems(ctx, diags)

	// invoke the copyWriteOnlyElements on every leaf switch object
	for name, dstLeafSwitch := range dstLeafSwitches {
		srcLeafSwitch, ok := src.leafSwitches(ctx, diags)[name]
		if !ok {
			continue
		}
		if diags.HasError() {
			return
		}

		dstLeafSwitch.copyWriteOnlyElements(ctx, &srcLeafSwitch, diags)
		if diags.HasError() {
			return
		}
		dstLeafSwitches[name] = dstLeafSwitch
	}

	// invoke the copyWriteOnlyElements on every access switch object
	for name, dstAccessSwitch := range dstAccessSwitches {
		srcAccessSwitch, ok := src.accessSwitches(ctx, diags)[name]
		if !ok {
			continue
		}
		if diags.HasError() {
			return
		}

		dstAccessSwitch.copyWriteOnlyElements(ctx, &srcAccessSwitch, diags)
		if diags.HasError() {
			return
		}
		dstAccessSwitches[name] = dstAccessSwitch
	}

	// invoke the copyWriteOnlyElements on every generic system object
	for name, dstGenericSystem := range dstGenericSystems {
		srcGenericSystem, ok := src.genericSystems(ctx, diags)[name]
		if !ok {
			continue
		}
		if diags.HasError() {
			return
		}

		dstGenericSystem.copyWriteOnlyElements(ctx, &srcGenericSystem, diags)
		if diags.HasError() {
			return
		}
		dstGenericSystems[name] = dstGenericSystem
	}

	// transform the native go objects (with copied object IDs) back to TF set
	leafSwitchMap := mapValueOrNull(ctx, types.ObjectType{AttrTypes: leafSwitch{}.attrTypes()}, dstLeafSwitches, diags)
	accessSwitchMap := mapValueOrNull(ctx, types.ObjectType{AttrTypes: accessSwitch{}.attrTypes()}, dstAccessSwitches, diags)
	genericSystemMap := mapValueOrNull(ctx, types.ObjectType{AttrTypes: genericSystem{}.attrTypes()}, dstGenericSystems, diags)
	if diags.HasError() {
		return
	}

	// save the TF sets into rackType
	o.LeafSwitches = leafSwitchMap
	o.AccessSwitches = accessSwitchMap
	o.GenericSystems = genericSystemMap
}

func (o *rackType) getSwitchRedundancyProtocolByName(ctx context.Context, name string, path path.Path, diags *diag.Diagnostics) fmt.Stringer {
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

	var leafRedundancyProtocol goapstra.LeafRedundancyProtocol
	if leaf != nil {
		if leaf.RedundancyProtocol.IsNull() {
			return goapstra.LeafRedundancyProtocolNone
		}
		err := leafRedundancyProtocol.FromString(leaf.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing leaf switch redundancy protocol", err.Error())
			return nil
		}
		return leafRedundancyProtocol
	}

	var accessRedundancyProtocol goapstra.AccessRedundancyProtocol
	if access != nil {
		if !access.EsiLagInfo.IsNull() {
			return goapstra.AccessRedundancyProtocolEsi
		}
		if access.RedundancyProtocol.IsNull() {
			return goapstra.AccessRedundancyProtocolNone
		}
		err := accessRedundancyProtocol.FromString(access.RedundancyProtocol.ValueString())
		if err != nil {
			diags.AddAttributeError(path, "error parsing access switch redundancy protocol", err.Error())
			return nil
		}
		return accessRedundancyProtocol
	}
	diags.AddError(errProviderBug, "somehow we've reached the end of getSwitchRedundancyProtocolByName without finding a solution")
	return nil
}

// fcdModes returns permitted fabric_connectivity_design mode strings
func fcdModes() []string {
	return []string{
		goapstra.FabricConnectivityDesignL3Clos.String(),
		goapstra.FabricConnectivityDesignL3Collapsed.String()}
}

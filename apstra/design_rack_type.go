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

func (o rackType) dataSourceAttritbutes() map[string]dataSourceSchema.Attribute {
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

func (o rackType) resourceAttritbutes() map[string]resourceSchema.Attribute {
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
				Attributes: rRackTypeLeafSwitch{}.attributes(),
			},
		},
		"access_switches": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Access Switches are optional, link to Leaf Switches in the same rack",
			Optional:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: rRackTypeAccessSwitch{}.attributes(),
			},
		},
		"generic_systems": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Generic Systems are rack elements not" +
				"managed by Apstra: Servers, routers, firewalls, etc...",
			Optional:   true,
			Validators: []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: rRackTypeGenericSystem{}.attributes(),
			},
		},
	}
}

func (o *rackType) loadApiData(ctx context.Context, in *goapstra.RackTypeData, diags *diag.Diagnostics) {
	switch in.FabricConnectivityDesign {
	case goapstra.FabricConnectivityDesignL3Collapsed: // this FCD is supported
	case goapstra.FabricConnectivityDesignL3Clos: // this FCD is supported
	default: // this FCD is unsupported
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("Rack Type has unsupported Fabric Connectivity Design %q",
				in.FabricConnectivityDesign.String()))
	}

	leafSwitches := make(map[string]leafSwitch, len(in.LeafSwitches))
	for _, leafIn := range in.LeafSwitches {
		var ls leafSwitch
		ls.loadApiData(ctx, &leafIn, in.FabricConnectivityDesign, diags)
		leafSwitches[leafIn.Label] = ls
		if diags.HasError() {
			return
		}
	}

	accessSwitches := make(map[string]accessSwitch, len(in.AccessSwitches))
	for _, accessIn := range in.AccessSwitches {
		var as accessSwitch
		as.loadApiData(ctx, &accessIn, diags)
		accessSwitches[accessIn.Label] = as
		if diags.HasError() {
			return
		}
	}

	genericSystems := make(map[string]genericSystem, len(in.GenericSystems))
	for _, genericIn := range in.GenericSystems {
		var gs genericSystem
		gs.loadApiData(ctx, &genericIn, diags)
		genericSystems[genericIn.Label] = gs
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Description = stringValueOrNull(ctx, in.Description, diags)
	o.FabricConnectivityDesign = types.StringValue(in.FabricConnectivityDesign.String())
	o.LeafSwitches = mapValueOrNull(ctx, types.ObjectType{AttrTypes: leafSwitch{}.attrTypes()}, leafSwitches, diags)
	o.AccessSwitches = mapValueOrNull(ctx, types.ObjectType{AttrTypes: accessSwitch{}.attrTypes()}, accessSwitches, diags)
	o.GenericSystems = mapValueOrNull(ctx, types.ObjectType{AttrTypes: genericSystem{}.attrTypes()}, genericSystems, diags)
	// todo use these new functions
	//o.LeafSwitches = newLeafSwitchMap(ctx, in.LeafSwitches, in.FabricConnectivityDesign, diags)
	//o.AccessSwitches = newAccessSwitchMap(ctx, in.AccessSwitches, diags)
	//o.GenericSystems = newGenericSystemMap(ctx, in.GenericSystems, diags)
}

// todo delete everything below here
// todo delete everything below here
// todo delete everything below here
// todo delete everything below here
// todo delete everything below here
// todo delete everything below here

type rackTypeData struct {
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Map    `tfsdk:"leaf_switches"`
	AccessSwitches           types.Map    `tfsdk:"access_switches"`
	GenericSystems           types.Map    `tfsdk:"generic_systems"`
}

func (o rackTypeData) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Rack Type name displayed in the Apstra web UI.  Required when Rack Type ID is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
			},
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

//func (o rackTypeData) attributes() map[string]schema.Attribute {
//	return map[string]schema.Attribute{
//
//	}
//}

func (o *rackTypeData) loadApiResponse(ctx context.Context, in *goapstra.RackTypeData, diags *diag.Diagnostics) {
	switch in.FabricConnectivityDesign {
	case goapstra.FabricConnectivityDesignL3Collapsed: // this FCD is supported
	case goapstra.FabricConnectivityDesignL3Clos: // this FCD is supported
	default: // this FCD is unsupported
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("Rack Type Data has unsupported Fabric Connectivity Design %q",
				in.FabricConnectivityDesign.String()))
	}

	leafSwitches := make(map[string]leafSwitch, len(in.LeafSwitches))
	for _, leafIn := range in.LeafSwitches {
		var leafSwitch leafSwitch
		leafSwitch.loadApiData(ctx, &leafIn, in.FabricConnectivityDesign, diags)
		leafSwitches[leafIn.Label] = leafSwitch
		if diags.HasError() {
			return
		}
	}

	accessSwitches := make(map[string]accessSwitch, len(in.AccessSwitches))
	for _, accessIn := range in.AccessSwitches {
		var accessSwitch accessSwitch
		accessSwitch.loadApiData(ctx, &accessIn, diags)
		accessSwitches[accessIn.Label] = accessSwitch
		if diags.HasError() {
			return
		}
	}

	genericSystems := make(map[string]genericSystem, len(in.GenericSystems))
	for _, genericIn := range in.GenericSystems {
		var genericSystem genericSystem
		genericSystem.loadApiData(ctx, &genericIn, diags)
		genericSystems[genericIn.Label] = genericSystem
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Description = stringValueOrNull(ctx, in.Description, diags)
	o.FabricConnectivityDesign = types.StringValue(in.FabricConnectivityDesign.String())
	o.LeafSwitches = mapValueOrNull(ctx, types.ObjectType{AttrTypes: leafSwitch{}.attrTypes()}, leafSwitches, diags)
	o.AccessSwitches = mapValueOrNull(ctx, types.ObjectType{AttrTypes: accessSwitch{}.attrTypes()}, accessSwitches, diags)
	o.GenericSystems = mapValueOrNull(ctx, types.ObjectType{AttrTypes: genericSystem{}.attrTypes()}, genericSystems, diags)
}

func (o rackTypeData) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                       types.StringType,
		"description":                types.StringType,
		"fabric_connectivity_design": types.StringType,
		"leaf_switches":              types.MapType{ElemType: types.ObjectType{leafSwitch{}.attrTypes()}},
		"access_switches":            types.MapType{ElemType: types.ObjectType{accessSwitch{}.attrTypes()}},
		"generic_systems":            types.MapType{ElemType: types.ObjectType{genericSystem{}.attrTypes()}},
	}
}

func newRackTypeDataObject(ctx context.Context, in *goapstra.RackTypeData, diags *diag.Diagnostics) types.Object {
	var rtd rackTypeData
	rtd.loadApiResponse(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(rackTypeData{}.attrTypes())
	}

	rtdObj, d := types.ObjectValueFrom(ctx, rackTypeData{}.attrTypes(), &rtd)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(rackTypeData{}.attrTypes())
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

func (o *rackType) leafSwitches(ctx context.Context, diags *diag.Diagnostics) map[string]rRackTypeLeafSwitch {
	leafSwitches := make(map[string]rRackTypeLeafSwitch, len(o.LeafSwitches.Elements()))
	d := o.LeafSwitches.ElementsAs(ctx, &leafSwitches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	// copy the leaf switch name from the map key into the object's Name field
	for name, ls := range leafSwitches {
		ls.Name = types.StringValue(name)
		leafSwitches[name] = ls
	}
	return leafSwitches
}

func (o *rackType) leafSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackTypeLeafSwitch {
	leafSwitches := o.leafSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if ls, ok := leafSwitches[requested]; ok {
		return &ls
	}

	return nil
}

func (o *rackType) accessSwitches(ctx context.Context, diags *diag.Diagnostics) map[string]rRackTypeAccessSwitch {
	accessSwitches := make(map[string]rRackTypeAccessSwitch, len(o.AccessSwitches.Elements()))
	d := o.AccessSwitches.ElementsAs(ctx, &accessSwitches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	// copy the access switch name from the map key into the object's Name field
	for name, as := range accessSwitches {
		as.Name = types.StringValue(name)
		accessSwitches[name] = as
	}
	return accessSwitches
}

func (o *rackType) accessSwitchByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackTypeAccessSwitch {
	accessSwitches := o.accessSwitches(ctx, diags)
	if diags.HasError() {
		return nil
	}

	if as, ok := accessSwitches[requested]; ok {
		return &as
	}

	return nil
}

func (o *rackType) genericSystems(ctx context.Context, diags *diag.Diagnostics) map[string]rRackTypeGenericSystem {
	genericSystems := make(map[string]rRackTypeGenericSystem, len(o.GenericSystems.Elements()))
	d := o.GenericSystems.ElementsAs(ctx, &genericSystems, true)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	// copy the generic system name from the map key into the object's Name field
	for name, gs := range genericSystems {
		gs.Name = types.StringValue(name)
		genericSystems[name] = gs
	}
	return genericSystems
}

func (o *rackType) genericSystemByName(ctx context.Context, requested string, diags *diag.Diagnostics) *rRackTypeGenericSystem {
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
// rRackType to be used as state.
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
	leafSwitchMap := mapValueOrNull(ctx, types.ObjectType{AttrTypes: rRackTypeLeafSwitch{}.attrTypes()}, dstLeafSwitches, diags)
	accessSwitchMap := mapValueOrNull(ctx, types.ObjectType{AttrTypes: rRackTypeAccessSwitch{}.attrTypes()}, dstAccessSwitches, diags)
	genericSystemMap := mapValueOrNull(ctx, types.ObjectType{AttrTypes: rRackTypeGenericSystem{}.attrTypes()}, dstGenericSystems, diags)
	if diags.HasError() {
		return
	}

	// save the TF sets into rRackType
	o.LeafSwitches = leafSwitchMap
	o.AccessSwitches = accessSwitchMap
	o.GenericSystems = genericSystemMap
}

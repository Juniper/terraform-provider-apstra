package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rackTypeData struct {
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Map    `tfsdk:"leaf_switches"`
	AccessSwitches           types.Map    `tfsdk:"access_switches"`
	GenericSystems           types.Map    `tfsdk:"generic_systems"`
}

func (o rackTypeData) attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Rack Type name displayed in the Apstra web UI.  Required when Rack Type ID is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
			},
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Rack Type description displayed in the Apstra web UI.",
			Computed:            true,
		},
		"fabric_connectivity_design": schema.StringAttribute{
			MarkdownDescription: "Indicates designs for which this Rack Type is intended.",
			Computed:            true,
		},
		"leaf_switches": schema.MapNestedAttribute{
			MarkdownDescription: "A map of Leaf Switches in this Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: leafSwitchData{}.attributes(),
			},
		},
		"access_switches": schema.MapNestedAttribute{
			MarkdownDescription: "A map of Access Switches in this Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: accessSwitchData{}.attributes(),
			},
		},
		"generic_systems": schema.MapNestedAttribute{
			MarkdownDescription: "A map of Generic Systems in the Rack Type, keyed by name.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: dRackTypeGenericSystem{}.attributes(),
			},
		},
	}
}

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

	leafSwitches := make(map[string]leafSwitchData, len(in.LeafSwitches))
	for _, leafIn := range in.LeafSwitches {
		var leafSwitch leafSwitchData
		leafSwitch.loadApiResponse(ctx, &leafIn, in.FabricConnectivityDesign, diags)
		leafSwitches[leafIn.Label] = leafSwitch
		if diags.HasError() {
			return
		}
	}

	accessSwitches := make(map[string]accessSwitchData, len(in.AccessSwitches))
	for _, accessIn := range in.AccessSwitches {
		var accessSwitch accessSwitchData
		accessSwitch.loadApiResponse(ctx, &accessIn, diags)
		accessSwitches[accessIn.Label] = accessSwitch
		if diags.HasError() {
			return
		}
	}

	genericSystems := make(map[string]dRackTypeGenericSystem, len(in.GenericSystems))
	for _, genericIn := range in.GenericSystems {
		var genericSystem dRackTypeGenericSystem
		genericSystem.loadApiResponse(ctx, &genericIn, diags)
		genericSystems[genericIn.Label] = genericSystem
		if diags.HasError() {
			return
		}
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Description = stringValueOrNull(ctx, in.Description, diags)
	o.FabricConnectivityDesign = types.StringValue(in.FabricConnectivityDesign.String())
	o.LeafSwitches = mapValueOrNull(ctx, types.ObjectType{AttrTypes: leafSwitchData{}.attrTypes()}, leafSwitches, diags)
	o.AccessSwitches = mapValueOrNull(ctx, types.ObjectType{AttrTypes: accessSwitchData{}.attrTypes()}, accessSwitches, diags)
	o.GenericSystems = mapValueOrNull(ctx, types.ObjectType{AttrTypes: dRackTypeGenericSystem{}.attrTypes()}, genericSystems, diags)
}

func (o rackTypeData) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                       types.StringType,
		"description":                types.StringType,
		"fabric_connectivity_design": types.StringType,
		"leaf_switches":              types.MapType{ElemType: types.ObjectType{leafSwitchData{}.attrTypes()}},
		"access_switches":            types.MapType{ElemType: types.ObjectType{accessSwitchData{}.attrTypes()}},
		"generic_systems":            types.MapType{ElemType: types.ObjectType{dRackTypeGenericSystem{}.attrTypes()}},
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
		genericSystemData(in, i, diags)
	}
}

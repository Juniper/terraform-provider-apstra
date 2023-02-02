package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
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
					Attributes: dRackTypeLeafSwitch{}.schema(),
				},
			},
			"access_switches": schema.SetNestedAttribute{
				MarkdownDescription: "Details of Access Switches in this Rack Type.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dRackTypeAccessSwitch{}.schema(),
				},
			},
			"generic_systems": schema.SetNestedAttribute{
				MarkdownDescription: "Details of Generic Systems in the Rack Type.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dRackTypeGenericSystem{}.schema(),
				},
			},
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

type dRackType struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Set    `tfsdk:"leaf_switches"`
	AccessSwitches           types.Set    `tfsdk:"access_switches"`
	GenericSystems           types.Set    `tfsdk:"generic_systems"`
}

func (o *dRackType) parseApi(ctx context.Context, in *goapstra.RackType, diags *diag.Diagnostics) {
	switch in.Data.FabricConnectivityDesign {
	case goapstra.FabricConnectivityDesignL3Collapsed: // this FCD is supported
	case goapstra.FabricConnectivityDesignL3Clos: // this FCD is supported
	default: // this FCD is unsupported
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

	accessSwitchSet := types.SetNull(dRackTypeAccessSwitch{}.attrType())
	if len(in.Data.AccessSwitches) > 0 {
		accessSwitches := make([]dRackTypeAccessSwitch, len(in.Data.AccessSwitches))
		for i := range in.Data.AccessSwitches {
			accessSwitches[i].loadApiResponse(ctx, &in.Data.AccessSwitches[i], diags)
			if diags.HasError() {
				return
			}
		}
		accessSwitchSet, d = types.SetValueFrom(ctx, dRackTypeAccessSwitch{}.attrType(), accessSwitches)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}

	genericSystemSet := types.SetNull(dRackTypeGenericSystem{}.attrType())
	if len(in.Data.GenericSystems) > 0 {
		genericSystems := make([]dRackTypeGenericSystem, len(in.Data.GenericSystems))
		for i := range in.Data.GenericSystems {
			genericSystems[i].loadApiResponse(ctx, &in.Data.GenericSystems[i], diags)
			if diags.HasError() {
				return
			}
		}
		genericSystemSet, d = types.SetValueFrom(ctx, dRackTypeGenericSystem{}.attrType(), genericSystems)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)
	o.Description = types.StringValue(in.Data.Description)
	o.FabricConnectivityDesign = types.StringValue(in.Data.FabricConnectivityDesign.String())
	o.LeafSwitches = leafSwitchSet
	o.AccessSwitches = accessSwitchSet
	o.GenericSystems = genericSystemSet
}

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
			"leaf_switches": schema.MapNestedAttribute{
				MarkdownDescription: "A map of Leaf Switches in this Rack Type, keyed by name.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dRackTypeLeafSwitch{}.attributes(),
				},
			},
			"access_switches": schema.MapNestedAttribute{
				MarkdownDescription: "A map of Access Switches in this Rack Type, keyed by name.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dRackTypeAccessSwitch{}.attributes(),
				},
			},
			"generic_systems": schema.MapNestedAttribute{
				MarkdownDescription: "A map of Generic Systems in the Rack Type, keyed by name.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dRackTypeGenericSystem{}.attributes(),
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
	validateRackType(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	newState := &dRackType{}
	newState.loadApiResponse(ctx, rt, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

type dRackType struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	FabricConnectivityDesign types.String `tfsdk:"fabric_connectivity_design"`
	LeafSwitches             types.Map    `tfsdk:"leaf_switches"`
	AccessSwitches           types.Map    `tfsdk:"access_switches"`
	GenericSystems           types.Map    `tfsdk:"generic_systems"`
}

func (o *dRackType) loadApiResponse(ctx context.Context, in *goapstra.RackType, diags *diag.Diagnostics) {
	switch in.Data.FabricConnectivityDesign {
	case goapstra.FabricConnectivityDesignL3Collapsed: // this FCD is supported
	case goapstra.FabricConnectivityDesignL3Clos: // this FCD is supported
	default: // this FCD is unsupported
		diags.AddError(
			errProviderBug,
			fmt.Sprintf("Rack Type '%s' has unsupported Fabric Connectivity Design '%s'",
				in.Id, in.Data.FabricConnectivityDesign.String()))
	}

	leafSwitches := make(map[string]dRackTypeLeafSwitch, len(in.Data.LeafSwitches))
	for _, leafIn := range in.Data.LeafSwitches {
		var leafSwitch dRackTypeLeafSwitch
		leafSwitch.loadApiResponse(ctx, &leafIn, in.Data.FabricConnectivityDesign, diags)
		leafSwitches[leafIn.Label] = leafSwitch
		if diags.HasError() {
			return
		}
	}

	accessSwitches := make(map[string]dRackTypeAccessSwitch, len(in.Data.AccessSwitches))
	for _, accessIn := range in.Data.AccessSwitches {
		var accessSwitch dRackTypeAccessSwitch
		accessSwitch.loadApiResponse(ctx, &accessIn, diags)
		accessSwitches[accessIn.Label] = accessSwitch
		if diags.HasError() {
			return
		}
	}

	genericSystems := make(map[string]dRackTypeGenericSystem, len(in.Data.GenericSystems))
	for _, genericIn := range in.Data.GenericSystems {
		var genericSystem dRackTypeGenericSystem
		genericSystem.loadApiResponse(ctx, &genericIn, diags)
		genericSystems[genericIn.Label] = genericSystem
		if diags.HasError() {
			return
		}
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Data.DisplayName)
	o.Description = stringValueOrNull(ctx, in.Data.Description, diags)
	o.FabricConnectivityDesign = types.StringValue(in.Data.FabricConnectivityDesign.String())
	o.LeafSwitches = mapValueOrNull(ctx, dRackTypeLeafSwitch{}.attrType(), leafSwitches, diags)
	o.AccessSwitches = mapValueOrNull(ctx, dRackTypeAccessSwitch{}.attrType(), accessSwitches, diags)
	o.GenericSystems = mapValueOrNull(ctx, dRackTypeGenericSystem{}.attrType(), genericSystems, diags)
}

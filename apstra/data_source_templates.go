package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	_ "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceTemplates{}
var _ datasourceWithSetClient = &dataSourceTemplates{}

type dataSourceTemplates struct {
	client *apstra.Client
}

func (o *dataSourceTemplates) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_templates"
}

func (o *dataSourceTemplates) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceTemplates) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDesign + "This data source returns the ID numbers of Templates.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "A set of Apstra object ID numbers.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Optional filter to select only Templates of the specified type.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.OneOf(utils.AllTemplateTypes()...)},
			},
			"overlay_control_protocol": schema.StringAttribute{
				MarkdownDescription: "Optional filter to select only Templates with the specified Overlay Control Protocol.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.OneOf(utils.AllOverlayControlProtocols()...)},
			},
		},
	}
}

func (o *dataSourceTemplates) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config templates
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ids []apstra.ObjectId
	var err error
	//if config.Type.IsNull() && config.OverlayControlProtocol.IsNull() { // see todo in Schema(), then restore this
	if config.Type.IsNull() {
		ids, err = o.client.ListAllTemplateIds(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error listing Template IDs", err.Error())
			return
		}
	} else {
		allTemplates, err := o.client.GetAllTemplates(ctx)
		if err != nil {
			resp.Diagnostics.AddError("error retrieving Templates", err.Error())
			return
		}
		for _, template := range allTemplates {
			if !config.Type.IsNull() && config.Type.ValueString() != template.Type().String() {
				continue // filter out this template from the results
			}
			if !config.OverlayControlProtocol.IsNull() && config.OverlayControlProtocol.ValueString() != template.OverlayControlProtocol().String() {
				continue // filter out this template from the results
			}
			ids = append(ids, template.ID())
		}
	}

	idSet, diags := types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// save the list of IDs in the config object
	config.Ids = idSet

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

type templates struct {
	Ids                    types.Set    `tfsdk:"ids"`
	Type                   types.String `tfsdk:"type"`
	OverlayControlProtocol types.String `tfsdk:"overlay_control_protocol"`
}

func (o *dataSourceTemplates) setClient(client *apstra.Client) {
	o.client = client
}

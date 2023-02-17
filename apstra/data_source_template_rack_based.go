package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &dataSourceTemplateRackBased{}

type dataSourceTemplateRackBased struct {
	client *goapstra.Client
}

func (o *dataSourceTemplateRackBased) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack_based_template"
}

func (o *dataSourceTemplateRackBased) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (o *dataSourceTemplateRackBased) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source provides details of a specific Rack Based (3 stage) Template.\n\n" +
			"At least one optional attribute is required. " +
			"It is incumbent on the user to ensure the criteria matches exactly one Rack Based Template. " +
			"Matching zero Rack Based Templates or more than one Rack Based Template will produce an error.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Template ID.  Required when the Template name is omitted.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Template name displayed in the Apstra web UI.  Required when Template ID is omitted.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
				},
			},
			"spine": schema.SingleNestedAttribute{
				MarkdownDescription: "Spine layer details",
				Computed:            true,
				Attributes:          designTemplateSpine{}.dataSourceAttributes(),
			},
			"asn_allocation_scheme": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("%q is for 3-stage designs; %q is for 5-stage designs.",
					asnAllocationUnique, asnAllocationSingle),
				Computed: true,
			},
		},
	}
}

func (o *dataSourceTemplateRackBased) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var config dTemplateRackBased
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var trb *goapstra.TemplateRackBased
	var ace goapstra.ApstraClientErr

	// maybe the config gave us the rack type name?
	if !config.Name.IsNull() { // fetch rack type by name
		trb, err = o.client.GetRackBasedTemplateByName(ctx, config.Name.ValueString())
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Rack Based Template not found",
				fmt.Sprintf("Rack Based Template with name %q does not exist", config.Name.ValueString()))
			return
		}
	}

	// maybe the config gave us the rack type id?
	if !config.Id.IsNull() { // fetch rack type by ID
		trb, err = o.client.GetRackBasedTemplate(ctx, goapstra.ObjectId(config.Id.ValueString()))
		if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Rack Type not found",
				fmt.Sprintf("Rack Type with ID %q does not exist", config.Id.ValueString()))
			return
		}
	}

	// create state object
	var state dTemplateRackBased
	state.loadApiResponse(ctx, trb, &resp.Diagnostics)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type dTemplateRackBased struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Spine         types.Object `tfsdk:"spine"`
	AsnAllocation types.String `tfsdk:"asn_allocation_scheme"`
}

func (o *dTemplateRackBased) loadApiResponse(ctx context.Context, in *goapstra.TemplateRackBased, diags *diag.Diagnostics) {
	if in == nil || in.Data == nil {
		diags.AddError(errProviderBug, "attempt to load dTemplateRackBased from nil source")
		return
	}

	spine := newDesignTemplateSpineObject(ctx, &in.Data.Spine, diags)
	if diags.HasError() {
		return
	}

	o.Name = types.StringValue(in.Data.DisplayName)
	o.Id = types.StringValue(string(in.Id))
	o.AsnAllocation = types.StringValue(asnAllocationSchemeToString(in.Data.AsnAllocationPolicy.SpineAsnScheme, diags))
	if diags.HasError() {
		return
	}
	o.Spine = spine
}

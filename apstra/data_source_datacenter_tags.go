package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

var (
	_ datasource.DataSourceWithConfigure = &dataSourceDatacenterTags{}
	_ datasourceWithSetDcBpClientFunc    = &dataSourceDatacenterTags{}
)

type dataSourceDatacenterTags struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
}

func (o *dataSourceDatacenterTags) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_tags"
}

func (o *dataSourceDatacenterTags) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceDatacenterTags) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This data source returns the IDs of Tags within the specified Blueprint.",
		Attributes: map[string]schema.Attribute{
			"blueprint_id": schema.StringAttribute{
				MarkdownDescription: "Apstra Blueprint to search.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "List of filters used to select only desired Tags. " +
					"To match a filter, all specified attributes must match (each attribute within a " +
					"filter is AND-ed together). The returned IDs represent the Tags matched by " +
					"all of the filters together (filters are OR-ed together).",
				Optional:   true,
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: blueprint.Tag{}.DataSourceFilterAttributes(),
					Validators: []validator.Object{
						apstravalidator.AtLeastNAttributes(1, "name", "description"),
					},
				},
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "IDs of discovered `tag` Graph DB nodes.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"names": schema.SetAttribute{
				MarkdownDescription: "Names (labels) of discovered `tag` Graph DB nodes.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (o *dataSourceDatacenterTags) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config struct {
		BlueprintId types.String `tfsdk:"blueprint_id"`
		Ids         types.Set    `tfsdk:"ids"`
		Names       types.Set    `tfsdk:"names"`
		Filters     types.List   `tfsdk:"filters"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get a client for the datacenter reference design
	bp, err := o.getBpClientFunc(ctx, config.BlueprintId.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			resp.Diagnostics.AddError(fmt.Sprintf(errBpNotFoundSummary, config.BlueprintId), err.Error())
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf(errBpClientCreateSummary, config.BlueprintId), err.Error())
		return
	}

	// extract the filters
	var filters []blueprint.Tag
	resp.Diagnostics.Append(config.Filters.ElementsAs(ctx, &filters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// retrieve all tags from the API
	tags, err := bp.GetAllTags(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to get blueprint tags", err.Error())
		return
	}

	// collect matching IDs and names here
	var ids []attr.Value
	var names []attr.Value

	if len(filters) == 0 { // quick exit if user wants all tags
		config.Ids = utils.SetValueOrNull(ctx, types.StringType, maps.Keys(tags), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// collect the names and IDs
		ids = make([]attr.Value, 0, len(tags))
		names = make([]attr.Value, 0, len(tags))
		for k, v := range tags {
			ids = append(ids, types.StringValue(k.String()))
			names = append(names, types.StringValue(v.Data.Label))
		}

		// set the state and exit
		config.Ids = types.SetValueMust(types.StringType, ids)
		config.Names = types.SetValueMust(types.StringType, names)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	for _, filter := range filters {
		for k, v := range tags {
			if !filter.Name.IsNull() && filter.Name.ValueString() != v.Data.Label {
				continue // no match
			}
			if !filter.Description.IsNull() && filter.Description.ValueString() != v.Data.Description {
				continue // no match
			}

			// we got a match! add it to the list and delete it from the tags map so we never check it again
			ids = append(ids, types.StringValue(v.Id.String()))
			names = append(names, types.StringValue(v.Data.Label))
			delete(tags, k)
		}
	}

	// set the state
	config.Ids = types.SetValueMust(types.StringType, ids)
	config.Names = types.SetValueMust(types.StringType, names)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceDatacenterTags) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

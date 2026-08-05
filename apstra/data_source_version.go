package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure      = (*dataSourceDatacenterSwitchingZone)(nil)
	_ datasource.DataSourceWithValidateConfig = (*dataSourceDatacenterSwitchingZone)(nil)
	_ datasourceWithSetDcBpClientFunc         = (*dataSourceDatacenterSwitchingZone)(nil)
	_ datasourceWithSetClient                 = (*dataSourceDatacenterSwitchingZone)(nil)
)

type dataSourceVersion struct {
	client *apstra.Client
}

func (o *dataSourceVersion) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_version"
}

func (o *dataSourceVersion) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(ctx, o, req, resp)
}

func (o *dataSourceVersion) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryPlatform + "This data source returns the version of the Apstra service and " +
			"facilitates checking whether that version satisfies user-specified constraints.",
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				MarkdownDescription: "Apstra release version.",
				Computed:            true,
			},
			"checks": schema.MapAttribute{
				MarkdownDescription: "Map of version constraints to check against the Apstra release version. " +
					"The key is a user-defined name for the constraint, and the value is a version constraint " +
					"string. The version constraint string must be in a format that is compatible with " +
					"[the `go-version` package](https://pkg.go.dev/github.com/hashicorp/go-version#Constraints).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"results": schema.MapAttribute{
				MarkdownDescription: "Map of results from evaluating the user-specified version constraints in `checks`.",
				Computed:            true,
				ElementType:         types.BoolType,
			},
		},
	}
}

func (o *dataSourceVersion) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config struct {
		Version types.String `tfsdk:"version"`
		Checks  types.Map    `tfsdk:"checks"`
		Results types.Map    `tfsdk:"results"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var checks map[string]string
	resp.Diagnostics.Append(config.Checks.ElementsAs(ctx, &checks, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for k, v := range checks {
		_, err := version.NewConstraint(v)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("checks").AtMapKey(k),
				fmt.Sprintf("Cannot parse constraint %q: %q", k, v),
				err.Error(),
			)
		}
	}
}

func (o *dataSourceVersion) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var config struct {
		Version types.String `tfsdk:"version"`
		Checks  types.Map    `tfsdk:"checks"`
		Results types.Map    `tfsdk:"results"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the version string from the API.
	config.Version = types.StringValue(o.client.ApiVersion())

	if config.Checks.IsNull() { // No user-specified checks?
		// Set the result map null, set state and return.
		config.Results = types.MapNull(types.BoolType)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// Parse the version string from the API.
	apiVersion, err := version.NewVersion(config.Version.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Cannot parse version %q reported by Apstra", config.Version.ValueString()), err.Error())
		return
	}

	// Unpack the checks map (strings).
	var constraintStrings map[string]string
	resp.Diagnostics.Append(config.Checks.ElementsAs(ctx, &constraintStrings, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the result map based on the supplied constraints.
	results := make(map[string]attr.Value, len(constraintStrings))
	for k, v := range constraintStrings {
		constraint, err := version.NewConstraint(v)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("checks").AtMapKey(k),
				fmt.Sprintf("Cannot parse check/constraint[%q]: %q", k, v),
				err.Error(),
			)
			continue
		}

		results[k] = types.BoolValue(constraint.Check(apiVersion))
	}
	config.Results = types.MapValueMust(types.BoolType, results)

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (o *dataSourceVersion) setClient(client *apstra.Client) {
	o.client = client
}

package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type agentProfile struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Platform    types.String `tfsdk:"platform"`
	HasUsername types.Bool   `tfsdk:"has_username"`
	HasPassword types.Bool   `tfsdk:"has_password"`
	Packages    types.Map    `tfsdk:"packages"`
	OpenOptions types.Map    `tfsdk:"open_options"`
}

func (o *agentProfile) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.AgentProfileConfig {
	var platform string
	if o.Platform.IsNull() || o.Platform.IsUnknown() {
		platform = ""
	} else {
		platform = o.Platform.ValueString()
	}

	packages := make(goapstra.AgentPackages)
	diags.Append(o.Packages.ElementsAs(ctx, &packages, false)...)
	if diags.HasError() {
		return nil
	}

	options := make(map[string]string)
	diags.Append(o.OpenOptions.ElementsAs(ctx, &options, false)...)
	if diags.HasError() {
		return nil
	}

	return &goapstra.AgentProfileConfig{
		Label:       o.Name.ValueString(),
		Platform:    platform,
		Packages:    packages,
		OpenOptions: options,
	}
}

func (o *agentProfile) loadApiResponse(ctx context.Context, in *goapstra.AgentProfile, diags *diag.Diagnostics) {
	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Label)
	o.Platform = stringValueOrNull(ctx, in.Platform, diags)
	o.HasUsername = types.BoolValue(in.HasUsername)
	o.HasPassword = types.BoolValue(in.HasPassword)
	o.Packages = mapValueOrNull(ctx, types.StringType, in.Packages, diags)
	o.OpenOptions = mapValueOrNull(ctx, types.StringType, in.OpenOptions, diags)
}

package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func (o agentProfile) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up an Agent Profile by ID. Required when `name`is omitted.",
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
			MarkdownDescription: "Populate this field to look up an Agent Profile by name. Required when `id`is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"platform": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Indicates the platform supported by the Agent Profile.",
			Computed:            true,
		},
		"has_username": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether a username has been configured.",
			Computed:            true,
		},
		"has_password": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether a password has been configured.",
			Computed:            true,
		},
		"packages": dataSourceSchema.MapAttribute{
			MarkdownDescription: "Admin-provided software packages stored on the Apstra server applied to devices using the profile.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"open_options": dataSourceSchema.MapAttribute{
			MarkdownDescription: "Configured parameters for offbox agents",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
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

func (o *agentProfile) loadApiData(ctx context.Context, in *goapstra.AgentProfile, diags *diag.Diagnostics) {
	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Label)
	o.Platform = stringValueOrNull(ctx, in.Platform, diags)
	o.HasUsername = types.BoolValue(in.HasUsername)
	o.HasPassword = types.BoolValue(in.HasPassword)
	o.Packages = mapValueOrNull(ctx, types.StringType, in.Packages, diags)
	o.OpenOptions = mapValueOrNull(ctx, types.StringType, in.OpenOptions, diags)
}

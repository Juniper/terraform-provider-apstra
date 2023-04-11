package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-apstra/apstra/utils"
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

func (o agentProfile) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Agent Profile.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra name of the Agent Profile.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"has_username": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether a username has been set.",
			Computed:            true,
		},
		"has_password": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates whether a password has been set.",
			Computed:            true,
		},
		"platform": resourceSchema.StringAttribute{
			MarkdownDescription: "Indicates the platform supported by the Agent Profile.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.OneOf(utils.AgentProfilePlatforms()...)},
			Default:             stringdefault.StaticString(apstra.AgentPlatformJunos.String()),
		},
		"packages": resourceSchema.MapAttribute{
			MarkdownDescription: "List of [packages](https://www.juniper.net/documentation/us/en/software/apstra4.1/apstra-user-guide/topics/topic-map/packages.html) " +
				"to be included with agents deployed using this profile.",
			Optional:    true,
			ElementType: types.StringType,
			Validators:  []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
		"open_options": resourceSchema.MapAttribute{
			MarkdownDescription: "Passes configured parameters to offbox agents. For example, to use HTTPS as the " +
				"API connection from offbox agents to devices, use the key-value pair: proto-https - port-443.",
			Optional:    true,
			ElementType: types.StringType,
			Validators:  []validator.Map{mapvalidator.SizeAtLeast(1)},
		},
	}
}

func (o *agentProfile) request(ctx context.Context, diags *diag.Diagnostics) *apstra.AgentProfileConfig {
	var platform string
	if o.Platform.IsNull() || o.Platform.IsUnknown() {
		platform = ""
	} else {
		platform = o.Platform.ValueString()
	}

	packages := make(apstra.AgentPackages)
	diags.Append(o.Packages.ElementsAs(ctx, &packages, false)...)
	if diags.HasError() {
		return nil
	}

	options := make(map[string]string)
	diags.Append(o.OpenOptions.ElementsAs(ctx, &options, false)...)
	if diags.HasError() {
		return nil
	}

	return &apstra.AgentProfileConfig{
		Label:       o.Name.ValueString(),
		Platform:    platform,
		Packages:    packages,
		OpenOptions: options,
	}
}

func (o *agentProfile) loadApiData(ctx context.Context, in *apstra.AgentProfile, diags *diag.Diagnostics) {
	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.Label)
	o.Platform = utils.StringValueOrNull(ctx, in.Platform, diags)
	o.HasUsername = types.BoolValue(in.HasUsername)
	o.HasPassword = types.BoolValue(in.HasPassword)
	o.Packages = utils.MapValueOrNull(ctx, types.StringType, in.Packages, diags)
	o.OpenOptions = utils.MapValueOrNull(ctx, types.StringType, in.OpenOptions, diags)
}

package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

func (o agentProfile) reourceSchema() map[string]resourceSchema.Attribute {
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
			MarkdownDescription: "Device platform.",
			Optional:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				goapstra.AgentPlatformNXOS.String(),
				goapstra.AgentPlatformJunos.String(),
				goapstra.AgentPlatformEOS.String(),
			)},
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

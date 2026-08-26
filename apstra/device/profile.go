package device

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/device"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Profile struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (p Profile) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Device Profile.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{path.MatchRoot("name")}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra web UI name of the Device Profile. Note that Device Profile payloads are " +
				"very large, and by-name lookups require retrieving many Device Profiles from the Apstra API. It " +
				"may be necessary to increase the default `api_timeout` in the `provider` configuration block when" +
				"using this option.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

func (p *Profile) LoadAPIData(_ context.Context, in device.Profile, _ *diag.Diagnostics) {
	p.ID = types.StringPointerValue(in.ID())
	p.Name = types.StringValue(in.Label)
}

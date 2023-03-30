package design

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ validator.Object = ConfigletGeneratorValidator{}

type ConfigletGeneratorValidator struct {
}

func (o ConfigletGeneratorValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Ensures that the section name matches the config style.")
}

func (o ConfigletGeneratorValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ConfigletGeneratorValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var c ConfigletGenerator
	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &c, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	cg := c.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	valid := false
	for _, i := range cg.ConfigStyle.ValidSections() {
		if i == cg.Section {
			valid = true
			break
		}
	}
	if !valid {
		resp.Diagnostics.AddError("Invalid Section", fmt.Sprintf("Invalid Section %q used for Config Style %q", cg.Section.String(), cg.ConfigStyle.String()))
	}
}

func ValidateConfigletGenerator() validator.Object {
	return ConfigletGeneratorValidator{}
}

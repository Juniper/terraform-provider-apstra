package design

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ validator.Object = ConfigletGeneratorValidator{}

type ConfigletGeneratorValidator struct {
}

func (o ConfigletGeneratorValidator) Description(_ context.Context) string {
	return "Ensures that the section names matches the config style."
}

func (o ConfigletGeneratorValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o ConfigletGeneratorValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var generator ConfigletGenerator
	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &generator, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := generator.Request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !utils.ItemInSlice(request.Section, apstra.ValidConfigletSections(request.ConfigStyle)) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path.AtName("section"),
			fmt.Sprintf("Section %q not valid with config_style %q",
				request.Section.String(), request.ConfigStyle.String()),
		))
	}

	if !generator.FileName.IsNull() && request.Section != enum.ConfigletSectionFile {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path.AtName("filename"),
			fmt.Sprintf("'filename' attribute permitted only when section == %q",
				enum.ConfigletSectionFile.String()),
		))
	}
}

func ValidateConfigletGenerator() validator.Object {
	return ConfigletGeneratorValidator{}
}

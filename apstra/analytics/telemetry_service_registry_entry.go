package analytics

import (
	"context"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TelemetryServiceRegistryEntry struct {
	Name              types.String         `tfsdk:"name"`
	ApplicationSchema jsontypes.Normalized `tfsdk:"application_schema"`
	StorageSchemaPath types.String         `tfsdk:"storage_schema_path"`
	Builtin           types.Bool           `tfsdk:"built_in"`
	Description       types.String         `tfsdk:"description"`
	Version           types.String         `tfsdk:"version"`
}

func (o TelemetryServiceRegistryEntry) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Service Name. Used to identify the Service.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"application_schema": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Application Schema expressed in JSON",
			CustomType:          jsontypes.NormalizedType{},
			Computed:            true,
		},
		"storage_schema_path": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Storage Schema Path",
			Computed:            true,
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Description",
			Computed:            true,
		},
		"version": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Version",
			Computed:            true,
		},
		"built_in": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "True If built in.",
			Computed:            true,
		},
	}
}

func (o TelemetryServiceRegistryEntry) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Service Name. Used to identify the Service.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"application_schema": resourceSchema.StringAttribute{
			MarkdownDescription: "Application Schema expressed in JSON",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
		"storage_schema_path": resourceSchema.StringAttribute{
			MarkdownDescription: "Storage Schema Path. Must be one of:\n  - " + strings.Join([]string{rosetta.StringersToFriendlyString(enum.StorageSchemaPathIbaStringData), rosetta.StringersToFriendlyString(enum.StorageSchemaPathIbaIntegerData)}, "\n  - ") + "\n",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.OneOf(rosetta.StringersToFriendlyString(enum.StorageSchemaPathIbaStringData), rosetta.StringersToFriendlyString(enum.StorageSchemaPathIbaIntegerData)),
			},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Description",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"version": resourceSchema.StringAttribute{
			MarkdownDescription: "Version",
			Computed:            true,
		},
		"built_in": resourceSchema.BoolAttribute{
			MarkdownDescription: "Indicates if provided by Apstra",
			Computed:            true,
		},
	}
}

func (o *TelemetryServiceRegistryEntry) LoadApiData(ctx context.Context, in *apstra.TelemetryServiceRegistryEntry, diag *diag.Diagnostics) {
	o.Name = types.StringValue(in.ServiceName)
	o.Version = types.StringValue(in.Version)
	o.Description = value.StringOrNull(ctx, in.Description, diag)
	o.Builtin = types.BoolValue(in.Builtin)
	o.ApplicationSchema = jsontypes.NewNormalizedValue(string(in.ApplicationSchema))
	o.StorageSchemaPath = types.StringValue(rosetta.StringersToFriendlyString(in.StorageSchemaPath))
}

func (o *TelemetryServiceRegistryEntry) Request(_ context.Context, diags *diag.Diagnostics) *apstra.TelemetryServiceRegistryEntry {
	var storageSchemaPath enum.StorageSchemaPath
	err := rosetta.ApiStringerFromFriendlyString(&storageSchemaPath, o.StorageSchemaPath.ValueString())
	if err != nil {
		diags.AddError("Failed to Parse Storage Schema Path", err.Error())
		return nil
	}

	return &apstra.TelemetryServiceRegistryEntry{
		ServiceName:       o.Name.ValueString(),
		ApplicationSchema: []byte(o.ApplicationSchema.ValueString()),
		StorageSchemaPath: storageSchemaPath,
		Builtin:           o.Builtin.ValueBool(),
		Description:       o.Description.ValueString(),
		Version:           o.Version.ValueString(),
	}
}

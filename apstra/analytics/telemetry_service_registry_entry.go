package analytics

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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
	ServiceName       types.String `tfsdk:"service_name"`
	ApplicationSchema types.String `tfsdk:"application_schema"`
	StorageSchemaPath types.String `tfsdk:"storage_schema_path"`
	Builtin           types.Bool   `tfsdk:"built_in"`
	Description       types.String `tfsdk:"description"`
	Version           types.String `tfsdk:"version"`
}

func (o TelemetryServiceRegistryEntry) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"service_name": resourceSchema.StringAttribute{
			MarkdownDescription: "Service Name. Used to identify the Service.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"application_schema": resourceSchema.StringAttribute{
			MarkdownDescription: "Application Schema expressed in Json schema",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"storage_schema_path": resourceSchema.StringAttribute{
			MarkdownDescription: "Storage Schema Path",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Description",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"version": resourceSchema.StringAttribute{
			MarkdownDescription: "Version",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"built_in": resourceSchema.BoolAttribute{
			MarkdownDescription: "True If built in.",
			Computed:            true,
		},
	}
}

func (o TelemetryServiceRegistryEntry) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"service_name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Service Name. Used to identify the Service.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"application_schema": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Application Schema expressed in Json schema",
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

func (o *TelemetryServiceRegistryEntry) LoadApiData(ctx context.Context, in *apstra.TelemetryServiceRegistryEntry, diag *diag.Diagnostics) {
	o.ServiceName = types.StringValue(in.ServiceName)
	o.Version = types.StringValue(in.Version)
	o.Description = utils.StringValueOrNull(ctx, in.Description, diag)
	o.Builtin = types.BoolValue(in.Builtin)
	o.ApplicationSchema = types.StringValue(string(in.ApplicationSchema))
	o.StorageSchemaPath = types.StringValue(in.StorageSchemaPath.String())
}

func (o *TelemetryServiceRegistryEntry) Request(ctx context.Context, d *diag.Diagnostics) *apstra.TelemetryServiceRegistryEntry {
	var s apstra.StorageSchemaPath
	e := s.FromString(o.StorageSchemaPath.ValueString())
	if e != nil {
		d.AddError("Failed to Parse Storage Schema Path", e.Error())
		return nil
	}
	return &apstra.TelemetryServiceRegistryEntry{
		ServiceName:       o.ServiceName.ValueString(),
		ApplicationSchema: []byte(o.ApplicationSchema.ValueString()),
		StorageSchemaPath: s,
		Builtin:           o.Builtin.ValueBool(),
		Description:       o.Description.ValueString(),
		Version:           o.Version.ValueString(),
	}
}

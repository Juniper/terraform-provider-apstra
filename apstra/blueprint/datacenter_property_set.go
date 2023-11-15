package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatacenterPropertySet struct {
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Data            types.String `tfsdk:"data"`
	Stale           types.Bool   `tfsdk:"stale"`
	SyncWithCatalog types.Bool   `tfsdk:"sync_with_catalog"`
	Keys            types.Set    `tfsdk:"keys"`
}

func (o DatacenterPropertySet) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"blueprint_id":      types.StringType,
		"id":                types.StringType,
		"name":              types.StringType,
		"data":              types.StringType,
		"stale":             types.BoolType,
		"sync_with_catalog": types.BoolType,
		"keys":              types.SetType{ElemType: types.StringType},
	}
}

func (o DatacenterPropertySet) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify " +
				"the Blueprint that the Property Set has been imported into.",
			Required:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Populate this field to look up an imported Property Set by ID. Required when `name` is omitted.",
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
			MarkdownDescription: "Populate this field to look up an imported Property Set by `name`. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"data": dataSourceSchema.StringAttribute{
			MarkdownDescription: "A map of values in the Property Set in JSON format",
			Computed:            true,
		},
		"keys": dataSourceSchema.SetAttribute{
			MarkdownDescription: "List of Keys that have been imported.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"stale": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Stale as reported in the Web UI.",
			Computed:            true,
		},
		"sync_with_catalog": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Keep the data synchronized with the catalog. Has no meaning in the data source.",
			Computed:            true,
		},
	}
}

func (o DatacenterPropertySet) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID. Used to identify the Blueprint that the Property Set is imported into.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Property Set ID to be imported.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Property Set name as shown in the Web UI.",
			Computed:            true,
		},
		"keys": resourceSchema.SetAttribute{
			MarkdownDescription: "Subset of Keys to import, at least one Key is required.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
		},
		"data": resourceSchema.StringAttribute{
			MarkdownDescription: "A map of values in the Property Set in JSON format.",
			Computed:            true,
		},
		"stale": resourceSchema.BoolAttribute{
			MarkdownDescription: "Stale as reported in the Web UI.",
			Computed:            true,
		},
		"sync_with_catalog": resourceSchema.BoolAttribute{
			MarkdownDescription: "Keep the data synchronized with the catalog.On every apply, " +
				"check staleness and update data if required. Cannot be set if 'keys' is filled out",
			Optional: true,
			Validators: []validator.Bool{
				boolvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("keys"),
				}...),
			},
		},
	}
}

func (o *DatacenterPropertySet) LoadApiData(_ context.Context, in *apstra.TwoStageL3ClosPropertySet, diags *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Label)
	o.Data = types.StringValue(string(in.Values))
	o.Stale = types.BoolValue(in.Stale)
	keys, err := utils.GetKeysFromJSON(o.Data)
	if err != nil {
		diags.AddError("Error parsing Keys from API response", err.Error())
	}
	o.Keys = types.SetValueMust(types.StringType, keys)
}

package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"regexp"
	"strings"
	"terraform-provider-apstra/apstra/utils"
)

type DatacenterRoutingPolicy struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	BlueprintId     types.String `tfsdk:"blueprint_id"`
	ImportPolicy    types.String `tfsdk:"import_policy"`
	ExportPolicy    types.Object `tfsdk:"export_policy"`
	ExpectV4Default types.Bool   `tfsdk:"expect_default_ipv4"`
	ExpectV6Default types.Bool   `tfsdk:"expect_default_ipv6"`
	//AggregatePrefixes types.List   `tfsdk:"aggregate_prefixes"` // todo
	//ExtraImports      types.List   `tfsdk:"extra_imports"`      // todo
	//ExtraImports      types.List   `tfsdk:"extra_exports"`      // todo
}

func (o DatacenterRoutingPolicy) ResourceAttributes() map[string]resourceSchema.Attribute {
	nameRE, _ := regexp.Compile("^[A-Za-z0-9_-]+$")
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Web UI 'name' field.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 18),
				stringvalidator.RegexMatches(nameRE, "only underscore, dash and alphanumeric characters allowed."),
			},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Web UI 'description' field.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"import_policy": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("One of '%s'",
				strings.Join(utils.AllDcRoutingPolicyImportPolicy(), "', '")),
			Required:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.AllDcRoutingPolicyImportPolicy()...)},
		},
		"export_policy": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The export policy controls export of various types of fabric prefixes.",
			Optional:            true,
			Computed:            true,
			Attributes:          datacenterRoutingPolicyExport{}.resourceAttributes(),
		},
		"expect_default_ipv4": resourceSchema.BoolAttribute{
			MarkdownDescription: "Default IPv4 route is expected to be imported via protocol session using this " +
				"policy. Used for rendering route expectations.'",
			Optional: true,
			Computed: true,
		},
		"expect_default_ipv6": resourceSchema.BoolAttribute{
			MarkdownDescription: "Default IPv6 route is expected to be imported via protocol session using this " +
				"policy. Used for rendering route expectations.'",
			Optional: true,
			Computed: true,
		},
		//"aggregate_prefixes": resourceSchema.ListAttribute{}, // todo
		//"extra_imports": resourceSchema.ListAttribute{},      // todo
		//"extra_exports": resourceSchema.ListAttribute{},      // todo
	}
}

func (o *DatacenterRoutingPolicy) Request(ctx context.Context, diags *diag.Diagnostics) *goapstra.DcRoutingPolicyData {
	if o.ImportPolicy.IsUnknown() {
		o.ImportPolicy = types.StringValue(goapstra.DcRoutingPolicyImportPolicyDefaultOnly.String())
	}

	var importPolicy goapstra.DcRoutingPolicyImportPolicy
	err := importPolicy.FromString(o.ImportPolicy.ValueString())
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing routing import policy %q - %s", o.ImportPolicy.ValueString(), err.Error()))
		return nil
	}

	var exportPolicy datacenterRoutingPolicyExport
	d := o.ExportPolicy.As(ctx, &exportPolicy, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	if o.ExpectV4Default.IsUnknown() || o.ExpectV4Default.IsNull() {
		o.ExpectV4Default = types.BoolValue(false)
	}

	if o.ExpectV6Default.IsUnknown() || o.ExpectV6Default.IsNull() {
		o.ExpectV6Default = types.BoolValue(false)
	}

	return &goapstra.DcRoutingPolicyData{
		Label:                  o.Name.ValueString(),
		Description:            o.Description.ValueString(),
		PolicyType:             goapstra.DcRoutingPolicyTypeUser,
		ImportPolicy:           importPolicy,
		ExportPolicy:           *exportPolicy.request(),
		ExpectDefaultIpv4Route: o.ExpectV4Default.ValueBool(),
		ExpectDefaultIpv6Route: o.ExpectV6Default.ValueBool(),
		AggregatePrefixes:      nil, // todo
		ExtraImportRoutes:      nil, // todo
		ExtraExportRoutes:      nil, // todo
	}
}

func (o *DatacenterRoutingPolicy) LoadApiData(ctx context.Context, policyData *goapstra.DcRoutingPolicyData, diags *diag.Diagnostics) {
	var exportPolicy datacenterRoutingPolicyExport
	exportPolicy.loadApiData(ctx, &policyData.ExportPolicy, diags)
	if diags.HasError() {
		return
	}
	exportPolicyObj, d := types.ObjectValueFrom(ctx, exportPolicy.attrTypes(), &exportPolicy)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	o.Name = types.StringValue(policyData.Label)
	o.Description = types.StringValue(policyData.Description)
	o.ImportPolicy = types.StringValue(policyData.ImportPolicy.String())
	o.ExportPolicy = exportPolicyObj
	o.ExpectV4Default = types.BoolValue(policyData.ExpectDefaultIpv4Route)
	o.ExpectV6Default = types.BoolValue(policyData.ExpectDefaultIpv6Route)
}

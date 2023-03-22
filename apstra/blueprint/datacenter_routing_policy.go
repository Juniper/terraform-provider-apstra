package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"net"
	"regexp"
	"strings"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
	"terraform-provider-apstra/apstra/utils"
)

type DatacenterRoutingPolicy struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	BlueprintId       types.String `tfsdk:"blueprint_id"`
	ImportPolicy      types.String `tfsdk:"import_policy"`
	ExportPolicy      types.Object `tfsdk:"export_policy"`
	ExpectV4Default   types.Bool   `tfsdk:"expect_default_ipv4"`
	ExpectV6Default   types.Bool   `tfsdk:"expect_default_ipv6"`
	AggregatePrefixes types.Set    `tfsdk:"aggregate_prefixes"`
	//ExtraImports      types.List   `tfsdk:"extra_imports"`      // todo
	//ExtraExports      types.List   `tfsdk:"extra_exports"`      // todo
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
			Computed:   true,
			Optional:   true,
			Validators: []validator.String{stringvalidator.OneOf(utils.AllDcRoutingPolicyImportPolicy()...)},
			Default:    stringdefault.StaticString(goapstra.DcRoutingPolicyImportPolicyDefaultOnly.String()),
		},
		"export_policy": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The export policy controls export of various types of fabric prefixes.",
			Attributes:          datacenterRoutingPolicyExport{}.resourceAttributes(),
			Computed:            true,
			Optional:            true,
			Default: objectdefault.StaticValue(types.ObjectValueMust(
				datacenterRoutingPolicyExport{}.attrTypes(), datacenterRoutingPolicyExport{}.defaultObject())),
		},
		"expect_default_ipv4": resourceSchema.BoolAttribute{
			MarkdownDescription: "Default IPv4 route is expected to be imported via protocol session using this " +
				"policy. Used for rendering route expectations.'",
			Computed: true,
			Optional: true,
			Default:  booldefault.StaticBool(true),
		},
		"expect_default_ipv6": resourceSchema.BoolAttribute{
			MarkdownDescription: "Default IPv6 route is expected to be imported via protocol session using this " +
				"policy. Used for rendering route expectations.'",
			Computed: true,
			Optional: true,
			Default:  booldefault.StaticBool(true),
		},
		"aggregate_prefixes": resourceSchema.SetAttribute{
			MarkdownDescription: "BGP Aggregate routes to be imported into a routing zone (VRF) on all border " +
				"switches. This option can only be set on routing policies associated with routing zones, and cannot " +
				"be set on per-connectivity point policies. The aggregated routes are sent to all external router " +
				"peers in a SZ (VRF).",
			Optional:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(apstravalidator.ParseCidr(false, false)),
			},
		},
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

	exportPolicy := datacenterRoutingPolicyExport{}
	diags.Append(o.ExportPolicy.As(ctx, &exportPolicy, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	if o.ExpectV4Default.IsUnknown() || o.ExpectV4Default.IsNull() {
		o.ExpectV4Default = types.BoolValue(false)
	}

	if o.ExpectV6Default.IsUnknown() || o.ExpectV6Default.IsNull() {
		o.ExpectV6Default = types.BoolValue(false)
	}

	aggregatePrefixStrings := make([]string, len(o.AggregatePrefixes.Elements()))
	diags.Append(o.AggregatePrefixes.ElementsAs(ctx, &aggregatePrefixStrings, false)...)
	if diags.HasError() {
		return nil
	}

	aggregatePrefixes := make([]net.IPNet, len(aggregatePrefixStrings))
	for i := range aggregatePrefixStrings {
		_, netIp, err := net.ParseCIDR(aggregatePrefixStrings[i])
		if err != nil {
			diags.AddError(
				fmt.Sprintf("error parsing aggregate prefix string %q", aggregatePrefixStrings[i]),
				err.Error())
		}
		aggregatePrefixes[i] = *netIp
	}
	if diags.HasError() {
		return nil
	}

	return &goapstra.DcRoutingPolicyData{
		Label:                  o.Name.ValueString(),
		Description:            o.Description.ValueString(),
		PolicyType:             goapstra.DcRoutingPolicyTypeUser,
		ImportPolicy:           importPolicy,
		ExportPolicy:           *exportPolicy.request(),
		ExpectDefaultIpv4Route: o.ExpectV4Default.ValueBool(),
		ExpectDefaultIpv6Route: o.ExpectV6Default.ValueBool(),
		AggregatePrefixes:      aggregatePrefixes,
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

	aggregatePrefixStrings := make([]string, len(policyData.AggregatePrefixes))
	for i := range policyData.AggregatePrefixes {
		aggregatePrefixStrings[i] = policyData.AggregatePrefixes[i].String()
	}

	o.Name = types.StringValue(policyData.Label)
	o.Description = types.StringValue(policyData.Description)
	o.ImportPolicy = types.StringValue(policyData.ImportPolicy.String())
	o.ExportPolicy = exportPolicyObj
	o.ExpectV4Default = types.BoolValue(policyData.ExpectDefaultIpv4Route)
	o.ExpectV6Default = types.BoolValue(policyData.ExpectDefaultIpv6Route)
	o.AggregatePrefixes = utils.SetValueOrNull(ctx, types.StringType, aggregatePrefixStrings, diags)
}

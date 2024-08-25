package blueprint

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	AggregatePrefixes types.List   `tfsdk:"aggregate_prefixes"`
	ExtraImports      types.List   `tfsdk:"extra_imports"`
	ExtraExports      types.List   `tfsdk:"extra_exports"`
}

func (o DatacenterRoutingPolicy) ResourceAttributes() map[string]resourceSchema.Attribute {
	nameRE := regexp.MustCompile("^[A-Za-z0-9_-]+$")
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
				stringvalidator.LengthBetween(1, 17),
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
			Default:    stringdefault.StaticString(apstra.DcRoutingPolicyImportPolicyDefaultOnly.String()),
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
				"policy. Used for rendering route expectations.",
			Computed: true,
			Optional: true,
			Default:  booldefault.StaticBool(true),
		},
		"expect_default_ipv6": resourceSchema.BoolAttribute{
			MarkdownDescription: "Default IPv6 route is expected to be imported via protocol session using this " +
				"policy. Used for rendering route expectations.",
			Computed: true,
			Optional: true,
			Default:  booldefault.StaticBool(true),
		},
		"aggregate_prefixes": resourceSchema.ListAttribute{
			MarkdownDescription: "BGP Aggregate routes to be imported into a routing zone (VRF) on all border " +
				"switches. This option can only be set on routing policies associated with routing zones, and cannot " +
				"be set on per-connectivity point policies. The aggregated routes are sent to all external router " +
				"peers in a SZ (VRF).",
			Optional:    true,
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
				listvalidator.ValueStringsAre(apstravalidator.ParseCidr(false, false)),
			},
		},
		"extra_imports": resourceSchema.ListNestedAttribute{
			MarkdownDescription: fmt.Sprintf("User defined import routes will be used in addition to any "+
				"routes generated by the import policies. Prefixes specified here are additive to the import policy, "+
				"unless 'import_policy' is set to %q, in which only these routes will be imported.",
				apstra.DcRoutingPolicyImportPolicyExtraOnly),
			Optional: true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: prefixFilter{}.resourceAttributes(),
				Validators: []validator.Object{prefixFilterValidator()},
			},
			Validators: []validator.List{
				listvalidator.UniqueValues(),
				listvalidator.SizeAtLeast(1),
			},
		},
		"extra_exports": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "User defined export routes will be used in addition to any other routes specified " +
				"in export policies. These policies are additive. To advertise only extra routes, disable all export " +
				"types within 'export_policy', and only the extra prefixes specified here will be advertised.",
			Optional: true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: prefixFilter{}.resourceAttributes(),
				Validators: []validator.Object{prefixFilterValidator()},
			},
			Validators: []validator.List{
				listvalidator.UniqueValues(),
				listvalidator.SizeAtLeast(1),
			},
		},
	}
}

func (o DatacenterRoutingPolicy) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID. Required when `name` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Web UI `name` field. Required when `id` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthBetween(1, 17)},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Web UI 'description' field.",
			Computed:            true,
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
		},
		"import_policy": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("One of '%s'",
				strings.Join(utils.AllDcRoutingPolicyImportPolicy(), "', '")),
			Computed: true,
		},
		"export_policy": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The export policy controls export of various types of fabric prefixes.",
			Attributes:          datacenterRoutingPolicyExport{}.dataSourceAttributes(),
			Computed:            true,
		},
		"expect_default_ipv4": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Default IPv4 route is expected to be imported via protocol session using this " +
				"policy. Used for rendering route expectations.",
			Computed: true,
		},
		"expect_default_ipv6": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Default IPv6 route is expected to be imported via protocol session using this " +
				"policy. Used for rendering route expectations.",
			Computed: true,
		},
		"aggregate_prefixes": dataSourceSchema.ListAttribute{
			MarkdownDescription: "BGP Aggregate routes to be imported into a routing zone (VRF) on all border " +
				"switches. This option can only be set on routing policies associated with routing zones, and cannot " +
				"be set on per-connectivity point policies. The aggregated routes are sent to all external router " +
				"peers in a SZ (VRF).",
			Computed:    true,
			ElementType: types.StringType,
		},
		"extra_imports": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: fmt.Sprintf("User defined import routes will be used in addition to any "+
				"routes generated by the import policies. Prefixes specified here are additive to the import policy, "+
				"unless 'import_policy' is set to %q, in which only these routes will be imported.",
				apstra.DcRoutingPolicyImportPolicyExtraOnly),
			Computed: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: prefixFilter{}.dataSourceAttributes(),
				Validators: []validator.Object{prefixFilterValidator()},
			},
		},
		"extra_exports": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "User defined export routes will be used in addition to any other routes specified " +
				"in export policies. These policies are additive. To advertise only extra routes, disable all export " +
				"types within 'export_policy', and only the extra prefixes specified here will be advertised.",
			Computed: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: prefixFilter{}.dataSourceAttributes(),
				Validators: []validator.Object{prefixFilterValidator()},
			},
		},
	}
}

func (o DatacenterRoutingPolicy) DataSourceAttributesAsFilter() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Web UI `name` field.",
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthBetween(1, 17)},
		},
		"description": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Web UI 'description' field.",
			Optional:            true,
		},
		"blueprint_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Not applicable in filter context. Ignore.",
			Computed:            true,
		},
		"import_policy": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("One of '%s'",
				strings.Join(utils.AllDcRoutingPolicyImportPolicy(), "', '")),
			Optional: true,
		},
		"export_policy": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The export policy controls export of various types of fabric prefixes.",
			Attributes:          datacenterRoutingPolicyExport{}.dataSourceAttributesAsFilter(),
			Optional:            true,
		},
		"expect_default_ipv4": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Default IPv4 route is expected to be imported via protocol session using this " +
				"policy. Used for rendering route expectations.",
			Optional: true,
		},
		"expect_default_ipv6": dataSourceSchema.BoolAttribute{
			MarkdownDescription: "Default IPv6 route is expected to be imported via protocol session using this " +
				"policy. Used for rendering route expectations.",
			Optional: true,
		},
		"aggregate_prefixes": dataSourceSchema.ListAttribute{
			MarkdownDescription: "All `aggregate_prefixes` specified here are required for the filter to match, " +
				"but the list need not be an *exact match*. That is, a policy containting `10.1.0.0/16` and " +
				"`10.2.0.0/16` will match a filter which specifies only `10.1.0.0/16`",
			Optional:    true,
			ElementType: types.StringType,
		},
		"extra_imports": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "All `extra_imports` specified here are required for the filter to match, " +
				"using the same logic as `aggregate_prefixes`.",
			Optional: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: prefixFilter{}.dataSourceAttributesAsFilter(),
				Validators: []validator.Object{prefixFilterValidator()},
			},
		},
		"extra_exports": dataSourceSchema.ListNestedAttribute{
			MarkdownDescription: "All `extra_exports` specified here are required for the filter to match, " +
				"using the same logic as `aggregate_prefixes`.",
			Optional: true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: prefixFilter{}.dataSourceAttributesAsFilter(),
				Validators: []validator.Object{prefixFilterValidator()},
			},
		},
	}
}

func (o *DatacenterRoutingPolicy) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.DcRoutingPolicyData {
	if o.ImportPolicy.IsUnknown() {
		o.ImportPolicy = types.StringValue(apstra.DcRoutingPolicyImportPolicyDefaultOnly.String())
	}

	var importPolicy apstra.DcRoutingPolicyImportPolicy
	err := importPolicy.FromString(o.ImportPolicy.ValueString())
	if err != nil {
		diags.AddError(constants.ErrProviderBug,
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

	extraImports := make([]apstra.PrefixFilter, len(o.ExtraImports.Elements()))
	for i, value := range o.ExtraImports.Elements() {
		obj := value.(types.Object)
		var pf prefixFilter
		diags.Append(obj.As(ctx, &pf, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil
		}
		extraImports[i] = *pf.request(ctx, diags)
		if diags.HasError() {
			return nil
		}
	}

	extraExports := make([]apstra.PrefixFilter, len(o.ExtraExports.Elements()))
	for i, value := range o.ExtraExports.Elements() {
		obj := value.(types.Object)
		var pf prefixFilter
		diags.Append(obj.As(ctx, &pf, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil
		}
		extraExports[i] = *pf.request(ctx, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &apstra.DcRoutingPolicyData{
		Label:                  o.Name.ValueString(),
		Description:            o.Description.ValueString(),
		PolicyType:             apstra.DcRoutingPolicyTypeUser,
		ImportPolicy:           importPolicy,
		ExportPolicy:           *exportPolicy.request(),
		ExpectDefaultIpv4Route: o.ExpectV4Default.ValueBool(),
		ExpectDefaultIpv6Route: o.ExpectV6Default.ValueBool(),
		AggregatePrefixes:      aggregatePrefixes,
		ExtraImportRoutes:      extraImports,
		ExtraExportRoutes:      extraExports,
	}
}

func (o *DatacenterRoutingPolicy) LoadApiData(ctx context.Context, policyData *apstra.DcRoutingPolicyData, diags *diag.Diagnostics) {
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

	extraImports := make([]prefixFilter, len(policyData.ExtraImportRoutes))
	for i := range policyData.ExtraImportRoutes {
		extraImports[i].loadApiData(ctx, &policyData.ExtraImportRoutes[i], diags)
	}

	extraExports := make([]prefixFilter, len(policyData.ExtraExportRoutes))
	for i := range policyData.ExtraExportRoutes {
		extraExports[i].loadApiData(ctx, &policyData.ExtraExportRoutes[i], diags)
	}

	o.Name = types.StringValue(policyData.Label)
	o.Description = utils.StringValueOrNull(ctx, policyData.Description, diags)
	o.ImportPolicy = types.StringValue(policyData.ImportPolicy.String())
	o.ExportPolicy = exportPolicyObj
	o.ExpectV4Default = types.BoolValue(policyData.ExpectDefaultIpv4Route)
	o.ExpectV6Default = types.BoolValue(policyData.ExpectDefaultIpv6Route)
	o.AggregatePrefixes = utils.ListValueOrNull(ctx, types.StringType, aggregatePrefixStrings, diags)
	o.ExtraImports = utils.ListValueOrNull(ctx, types.ObjectType{AttrTypes: prefixFilter{}.attrTypes()}, extraImports, diags)
	o.ExtraExports = utils.ListValueOrNull(ctx, types.ObjectType{AttrTypes: prefixFilter{}.attrTypes()}, extraExports, diags)
}

// FilterMatch returns true if 'in' has values which match those in 'o'
func (o *DatacenterRoutingPolicy) FilterMatch(ctx context.Context, in *DatacenterRoutingPolicy, diags *diag.Diagnostics) bool {
	if !o.Id.IsNull() && !o.Id.Equal(in.Id) {
		return false
	}

	if !o.Name.IsNull() && !o.Name.Equal(in.Name) {
		return false
	}

	if !o.Description.IsNull() && !o.Description.Equal(in.Description) {
		return false
	}

	if !o.ImportPolicy.IsNull() && !o.ImportPolicy.Equal(in.ImportPolicy) {
		return false
	}

	if !o.ExportPolicy.IsNull() {
		var filterExportPolicy, candidateExportPolicy datacenterRoutingPolicyExport
		diags.Append(o.ExportPolicy.As(ctx, &filterExportPolicy, basetypes.ObjectAsOptions{})...)
		diags.Append(in.ExportPolicy.As(ctx, &candidateExportPolicy, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return false
		}

		if !filterExportPolicy.filterMatch(ctx, &candidateExportPolicy, diags) {
			return false
		}
	}

	if !o.ExpectV4Default.IsNull() && !o.ExpectV4Default.Equal(in.ExpectV4Default) {
		return false
	}

	if !o.ExpectV6Default.IsNull() && !o.ExpectV6Default.Equal(in.ExpectV6Default) {
		return false
	}

	if !o.AggregatePrefixes.IsNull() {
		var filterAggregatePrefixes, candidateAggregatePrefixes []string
		diags.Append(o.AggregatePrefixes.ElementsAs(ctx, &filterAggregatePrefixes, false)...)
		diags.Append(in.AggregatePrefixes.ElementsAs(ctx, &candidateAggregatePrefixes, false)...)
		if diags.HasError() {
			return false
		}
		for _, filterAggreatePrefix := range filterAggregatePrefixes {
			if !utils.SliceContains(filterAggreatePrefix, candidateAggregatePrefixes) {
				return false
			}
		}
	}

	if !o.ExtraImports.IsNull() {
		var filterExtraImports, candidateExtraImports []prefixFilter
		diags.Append(o.ExtraImports.ElementsAs(ctx, &filterExtraImports, false)...)
		diags.Append(in.ExtraImports.ElementsAs(ctx, &candidateExtraImports, false)...)

	importFilterLoop:
		for _, filter := range filterExtraImports {
			for _, candidate := range candidateExtraImports {
				if filter.filterMatch(ctx, &candidate, diags) {
					continue importFilterLoop
				}
			}
			// if we get here, then  no candidate matched the filter
			return false
		}
	}

	if !o.ExtraExports.IsNull() {
		var filterExtraExports, candidateExtraExports []prefixFilter
		diags.Append(o.ExtraExports.ElementsAs(ctx, &filterExtraExports, false)...)
		diags.Append(in.ExtraExports.ElementsAs(ctx, &candidateExtraExports, false)...)

	exportFilterLoop:
		for _, filter := range filterExtraExports {
			for _, candidate := range candidateExtraExports {
				if filter.filterMatch(ctx, &candidate, diags) {
					continue exportFilterLoop
				}
			}
			// if we get here, then  no candidate matched the filter
			return false
		}
	}

	return true
}

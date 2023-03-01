package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-apstra/apstra/utils"
)

type templateRackBased struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Spine                  types.Object `tfsdk:"spine"`
	AsnAllocation          types.String `tfsdk:"asn_allocation_scheme"`
	OverlayControlProtocol types.String `tfsdk:"overlay_control_protocol"`
	FabricAddressing       types.String `tfsdk:"fabric_link_addressing"`
	RackInfos              types.Map    `tfsdk:"rack_infos"`
}

func (o templateRackBased) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Template ID.  Required when the Template name is omitted.",
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
			MarkdownDescription: "Template name displayed in the Apstra web UI.  Required when Template ID is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
			},
		},
		"spine": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Spine layer details",
			Computed:            true,
			Attributes:          spine{}.dataSourceAttributes(),
		},
		"asn_allocation_scheme": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("%q is for 3-stage designs; %q is for 5-stage designs.",
				asnAllocationUnique, asnAllocationSingle),
			Computed: true,
		},
		"overlay_control_protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Defines the inter-rack virtual network overlay protocol in the fabric.",
			Computed:            true,
		},
		"fabric_link_addressing": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Fabric addressing scheme for spine/leaf links.",
			Computed:            true,
		},
		"rack_infos": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Rack Type info (count + details)",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: templateRackInfo{}.dataSourceAttributesNested(),
			},
		},
	}
}

func (o templateRackBased) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Rack Based Template.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra name of the Rack Based Template.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"spine": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Spine layer details",
			Required:            true,
			Attributes:          spine{}.resourceAttributes(),
		},
		"asn_allocation_scheme": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("%q is for 3-stage designs; %q is for 5-stage designs.",
				asnAllocationUnique, asnAllocationSingle),
			Validators: []validator.String{stringvalidator.OneOf(asnAllocationUnique, asnAllocationSingle)},
			Required:   true,
		},
		"overlay_control_protocol": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Defines the inter-rack virtual network overlay protocol in the fabric. [%q,%q]",
				overlayControlProtocolEvpn, overlayControlProtocolStatic),
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(overlayControlProtocolEvpn, overlayControlProtocolStatic),
				// todo make sure not ipv6 with evpn
			},
		},
		"fabric_link_addressing": resourceSchema.StringAttribute{
			MarkdownDescription: "Fabric addressing scheme for spine/leaf links. Required for " +
				"Apstra <= 4.1.0, not supported by Apstra >= 4.1.1.",
			Optional: true,
		},
		"rack_infos": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Rack Type info (count + details)",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: templateRackInfo{}.resourceAttributesNested(),
			},
		},
	}
}

func (o templateRackBased) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                       types.StringType,
		"name":                     types.StringType,
		"spine":                    types.ObjectType{AttrTypes: spine{}.attrTypes()},
		"asn_allocation_scheme":    types.StringType,
		"overlay_control_protocol": types.StringType,
		"fabric_link_addressing":   types.StringType,
		"rack_infos":               types.MapType{ElemType: types.ObjectType{AttrTypes: rackType{}.attrTypes()}},
	}
}

func (o *templateRackBased) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.CreateRackBasedTemplateRequest {
	var d diag.Diagnostics

	s := spine{}
	d = o.Spine.As(ctx, &s, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	rtMap := make(map[string]templateRackInfo, len(o.RackInfos.Elements()))
	d = o.RackInfos.ElementsAs(ctx, &rtMap, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	rackInfos := make(map[goapstra.ObjectId]goapstra.TemplateRackBasedRackInfo, len(rtMap))
	for k := range rtMap {
		rackInfos[goapstra.ObjectId(k)] = goapstra.TemplateRackBasedRackInfo{
			Count: int(rtMap[k].Count.ValueInt64()),
		}
	}

	var err error

	antiAffinityPolicy := &goapstra.AntiAffinityPolicy{
		Algorithm: goapstra.AlgorithmHeuristic,
	}

	var spineAsnScheme goapstra.AsnAllocationScheme
	err = spineAsnScheme.FromString(translateAsnAllocationSchemeFromWebUi(o.AsnAllocation.ValueString()))
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing ASN allocation scheme %q - %s",
				o.AsnAllocation.ValueString(), err.Error()))
	}
	asnAllocationPolicy := &goapstra.AsnAllocationPolicy{
		SpineAsnScheme: spineAsnScheme,
	}

	var fabricAddressingPolicy *goapstra.FabricAddressingPolicy
	if !o.FabricAddressing.IsNull() {
		var addressingScheme goapstra.AddressingScheme
		err = addressingScheme.FromString(o.FabricAddressing.ValueString())
		if err != nil {
			diags.AddError(errProviderBug,
				fmt.Sprintf("error parsing fabric addressing scheme %q - %s",
					o.FabricAddressing.ValueString(), err.Error()))
		}
		fabricAddressingPolicy = &goapstra.FabricAddressingPolicy{
			SpineSuperspineLinks: addressingScheme,
			SpineLeafLinks:       addressingScheme,
		}
	}

	var overlayControlProtocol goapstra.OverlayControlProtocol
	err = overlayControlProtocol.FromString(o.OverlayControlProtocol.ValueString())
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing overlay control protocol %q - %s",
				o.OverlayControlProtocol.ValueString(), err.Error()))
	}
	virtualNetworkPolicy := &goapstra.VirtualNetworkPolicy{
		OverlayControlProtocol: overlayControlProtocol,
	}

	return &goapstra.CreateRackBasedTemplateRequest{
		DisplayName:       o.Name.ValueString(),
		Capability:        goapstra.TemplateCapabilityNone,
		Spine:             s.request(ctx, diags),
		RackInfos:         rackInfos,
		DhcpServiceIntent: &goapstra.DhcpServiceIntent{Active: true},
		// todo: is this the right AntiAffinityPolicy?
		//  I'd have sent <nil>, but blocked by goapstra issue #2 (crash on nil pointer deref)
		AntiAffinityPolicy:     antiAffinityPolicy,
		AsnAllocationPolicy:    asnAllocationPolicy,
		FabricAddressingPolicy: fabricAddressingPolicy,
		VirtualNetworkPolicy:   virtualNetworkPolicy,
	}
}

func (o *templateRackBased) validate(ctx context.Context, diags *diag.Diagnostics) {
	rackInfoMap := make(map[string]templateRackInfo, len(o.RackInfos.Elements()))
	d := o.RackInfos.ElementsAs(ctx, &rackInfoMap, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	idMap := make(map[string]struct{}, len(rackInfoMap))
	for key := range rackInfoMap {
		if _, ok := idMap[key]; ok {
			diags.AddAttributeError(path.Root("rack_infos").AtMapKey(key), errInvalidConfig,
				fmt.Sprintf("rack type id %q used multiple times", key))
			return
		}
		idMap[key] = struct{}{}
	}
}

func (o *templateRackBased) loadApiData(ctx context.Context, in *goapstra.TemplateRackBasedData, diags *diag.Diagnostics) {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to load templateRackBased from nil source")
		return
	}

	fap := in.FabricAddressingPolicy
	if fap == nil {
		o.FabricAddressing = types.StringNull()
	} else {
		if fap.SpineLeafLinks != fap.SpineSuperspineLinks {
			diags.AddError(errProviderBug, "spine/leaf and spine/superspine addressing do not match - we cannot handle this situation")
		}
		o.FabricAddressing = types.StringValue(fap.SpineLeafLinks.String())
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Spine = newDesignTemplateSpineObject(ctx, &in.Spine, diags)
	o.AsnAllocation = types.StringValue(asnAllocationSchemeToString(in.AsnAllocationPolicy.SpineAsnScheme, diags))
	o.OverlayControlProtocol = types.StringValue(overlayControlProtocolToString(in.VirtualNetworkPolicy.OverlayControlProtocol, diags))
	o.RackInfos = newRackInfoMap(ctx, in, diags)
}

func (o *templateRackBased) minMaxApiVersions(_ context.Context, diags *diag.Diagnostics) (*version.Version, *version.Version) {
	var min, max *version.Version
	var err error
	if o.FabricAddressing.IsNull() {
		min, err = version.NewVersion("4.1.1")
	} else {
		max, err = version.NewVersion("4.1.0")
	}
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing min/max version - %s", err.Error()))
	}

	return min, max
}

func (o *templateRackBased) copyWriteOnlyElements(ctx context.Context, src *templateRackBased, diags *diag.Diagnostics) {
	var srcSpine, dstSpine *spine
	var d diag.Diagnostics

	// extract the source spine object from src
	d = src.Spine.As(ctx, &srcSpine, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	// extract the destination spine object from o
	d = o.Spine.As(ctx, &dstSpine, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	// clone missing spine bits
	dstSpine.copyWriteOnlyElements(ctx, srcSpine, diags)

	// repackage the destination spine in o
	o.Spine = utils.ObjectValueOrNull(ctx, spine{}.attrTypes(), dstSpine, diags)
}

// 	state.copyWriteOnlyElements(ctx, &plan, &resp.Diagnostics)

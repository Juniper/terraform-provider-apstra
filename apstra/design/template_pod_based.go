package design

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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
)

type TemplatePodBased struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	SuperSpine             types.Object `tfsdk:"super_spine"`
	OverlayControlProtocol types.String `tfsdk:"overlay_control_protocol"`
	FabricAddressing       types.String `tfsdk:"fabric_link_addressing"`
	PodInfos               types.Map    `tfsdk:"pod_infos"`
}

func (o TemplatePodBased) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                       types.StringType,
		"name":                     types.StringType,
		"spine":                    types.ObjectType{AttrTypes: Spine{}.AttrTypes()},
		"overlay_control_protocol": types.StringType,
		"fabric_link_addressing":   types.StringType,
		"pod_infos":                types.MapType{ElemType: types.ObjectType{AttrTypes: TemplatePodInfo{}.AttrTypes()}},
	}
}

func (o TemplatePodBased) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Template ID. Required when `id` is omitted.",
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
			MarkdownDescription: "Web UI name of the Template. Required when `id` is omitted.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
			},
		},
		"super_spine": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Super Spine layer details",
			Computed:            true,
			Attributes:          SuperSpine{}.DataSourceAttributes(),
		},
		"overlay_control_protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Defines the inter-pod virtual network overlay protocol in the fabric.",
			Computed:            true,
		},
		"fabric_link_addressing": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Fabric addressing scheme for Spine/Superspine links. Applies only to "+
				"Apstra %s.", apiversions.Apstra410),
			Computed: true,
		},
		"pod_infos": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Pod Type information (count + details)",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: TemplatePodInfo{}.DataSourceAttributes(),
			},
		},
	}
}

func (o TemplatePodBased) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Pod Based Template.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra name of the Pod Based Template.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"spine": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Spine layer details",
			Required:            true,
			Attributes:          Spine{}.ResourceAttributes(),
		},
		"overlay_control_protocol": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Defines the inter-pod virtual network overlay protocol in the fabric. [%q,%q]",
				OverlayControlProtocolEvpn, OverlayControlProtocolStatic),
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(OverlayControlProtocolEvpn, OverlayControlProtocolStatic),
				// todo make sure not ipv6 with evpn
			},
		},
		"fabric_link_addressing": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Fabric addressing scheme for Spine/Leaf links. Required for "+
				"Apstra <= %s, not supported by Apstra >= %s.", apiversions.Apstra410, apiversions.Apstra411),
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					apstra.AddressingSchemeIp4.String(),
					apstra.AddressingSchemeIp46.String(),
					apstra.AddressingSchemeIp6.String(),
				),
				apstravalidator.WhenValueIsString(
					types.StringValue(apstra.AddressingSchemeIp6.String()),
					apstravalidator.ValueAtMustBeString(
						path.MatchRelative().AtParent().AtName("overlay_control_protocol"),
						types.StringValue(OverlayControlProtocolStatic),
						false,
					),
				),
			},
		},
		"pod_infos": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Pod Type info (count + details)",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: TemplatePodInfo{}.ResourceAttributes(),
			},
		},
	}
}

func (o *TemplatePodBased) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.CreatePodBasedTemplateRequest {
	var d diag.Diagnostics

	s := Spine{}
	d = o.Spine.As(ctx, &s, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	rtMap := make(map[string]TemplatePodInfo, len(o.PodInfos.Elements()))
	d = o.PodInfos.ElementsAs(ctx, &rtMap, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	podInfos := make(map[apstra.ObjectId]apstra.TemplatePodBasedPodInfo, len(rtMap))
	for k := range rtMap {
		podInfos[apstra.ObjectId(k)] = apstra.TemplatePodBasedPodInfo{
			Count: int(rtMap[k].Count.ValueInt64()),
		}
	}

	var err error

	antiAffinityPolicy := &apstra.AntiAffinityPolicy{
		Algorithm: apstra.AlgorithmHeuristic,
	}

	var fabricAddressingPolicy *apstra.TemplateFabricAddressingPolicy410Only
	if utils.Known(o.FabricAddressing) {
		var addressingScheme apstra.AddressingScheme
		err = addressingScheme.FromString(o.FabricAddressing.ValueString())
		if err != nil {
			diags.AddError(errProviderBug,
				fmt.Sprintf("error parsing fabric addressing scheme %q - %s",
					o.FabricAddressing.ValueString(), err.Error()))
		}
		fabricAddressingPolicy = &apstra.TemplateFabricAddressingPolicy410Only{
			SpineSuperspineLinks: addressingScheme,
			SpineLeafLinks:       addressingScheme,
		}
	}

	var overlayControlProtocol apstra.OverlayControlProtocol
	err = utils.ApiStringerFromFriendlyString(&overlayControlProtocol, o.OverlayControlProtocol.ValueString())
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing overlay control protocol %q - %s",
				o.OverlayControlProtocol.ValueString(), err.Error()))
	}
	virtualNetworkPolicy := &apstra.VirtualNetworkPolicy{
		OverlayControlProtocol: overlayControlProtocol,
	}

	return &apstra.CreatePodBasedTemplateRequest{
		DisplayName:       o.Name.ValueString(),
		Spine:             s.Request(ctx, diags),
		PodInfos:          podInfos,
		DhcpServiceIntent: &apstra.DhcpServiceIntent{Active: true},
		// todo: is this the right AntiAffinityPolicy?
		//  I'd have sent <nil>, but blocked by sdk issue #2 (crash on nil pointer deref)
		AntiAffinityPolicy:     antiAffinityPolicy,
		FabricAddressingPolicy: fabricAddressingPolicy,
		VirtualNetworkPolicy:   virtualNetworkPolicy,
	}
}

func (o *TemplatePodBased) LoadApiData(ctx context.Context, in *apstra.TemplatePodBasedData, diags *diag.Diagnostics) {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to load TemplatePodBased from nil source")
		return
	}

	fabricAddressing := types.StringNull()
	if in.FabricAddressingPolicy != nil {
		if in.FabricAddressingPolicy.SpineLeafLinks != in.FabricAddressingPolicy.SpineSuperspineLinks {
			diags.AddError(errProviderBug,
				fmt.Sprintf("Spine/Leaf and Spine/Luperspine addressing do not match: %q vs. %q\n"+
					"We cannot handle this situation.",
					in.FabricAddressingPolicy.SpineLeafLinks.String(),
					in.FabricAddressingPolicy.SpineSuperspineLinks.String()),
			)
			return
		}
		fabricAddressing = types.StringValue(in.FabricAddressingPolicy.SpineLeafLinks.String())
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Spine = NewDesignTemplateSpineObject(ctx, &in.Spine, diags)
	o.OverlayControlProtocol = types.StringValue(utils.StringersToFriendlyString(in.VirtualNetworkPolicy.OverlayControlProtocol))
	o.PodInfos = NewPodInfoMap(ctx, in, diags)
	o.FabricAddressing = fabricAddressing
}

func (o *TemplatePodBased) CopyWriteOnlyElements(ctx context.Context, src *TemplatePodBased, diags *diag.Diagnostics) {
	var srcSpine, dstSpine *Spine

	// extract the source Spine object from src
	diags.Append(src.Spine.As(ctx, &srcSpine, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}

	// extract the destination Spine object from o
	diags.Append(o.Spine.As(ctx, &dstSpine, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}

	// clone missing Spine bits
	dstSpine.CopyWriteOnlyElements(ctx, srcSpine, diags)

	// repackage the destination Spine in o
	o.Spine = utils.ObjectValueOrNull(ctx, Spine{}.AttrTypes(), dstSpine, diags)
}

func (o TemplatePodBased) VersionConstraints() apiversions.Constraints {
	var response apiversions.Constraints

	if !o.FabricAddressing.IsNull() {
		response.AddAttributeConstraints(
			apiversions.AttributeConstraint{
				Path:        path.Root("fabric_link_addressing"),
				Constraints: version.MustConstraints(version.NewConstraint(apiversions.Apstra410)),
			},
		)
	}

	return response
}

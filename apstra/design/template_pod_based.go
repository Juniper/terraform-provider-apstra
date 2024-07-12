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
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	SuperSpine       types.Object `tfsdk:"super_spine"`
	FabricAddressing types.String `tfsdk:"fabric_link_addressing"`
	PodInfos         types.Map    `tfsdk:"pod_infos"`
}

func (o TemplatePodBased) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.StringType,
		"name":                   types.StringType,
		"super_spine":            types.ObjectType{AttrTypes: SuperSpine{}.AttrTypes()},
		"fabric_link_addressing": types.StringType,
		"pod_infos":              types.MapType{ElemType: types.ObjectType{AttrTypes: TemplatePodInfo{}.AttrTypes()}},
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
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"super_spine": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Super Spine layer details",
			Computed:            true,
			Attributes:          SuperSpine{}.DataSourceAttributes(),
		},
		"fabric_link_addressing": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Fabric addressing scheme for Spine/Superspine links. Applies only to "+
				"Apstra %s.", apiversions.Apstra410),
			Computed: true,
		},
		"pod_infos": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Pod Type info (count + details)",
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
		"super_spine": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "SuperSpine layer details",
			Required:            true,
			Attributes:          SuperSpine{}.ResourceAttributes(),
		},
		"fabric_link_addressing": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Fabric addressing scheme for Spine/SuperSpine links. Required for "+
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
			MarkdownDescription: "Map of Pod Type info (count + details) keyed by Pod Based Template ID.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: TemplatePodInfo{}.ResourceAttributes(),
			},
		},
	}
}

func (o *TemplatePodBased) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.CreatePodBasedTemplateRequest {
	ss := SuperSpine{}
	diags.Append(o.SuperSpine.As(ctx, &ss, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	piMap := make(map[string]TemplatePodInfo, len(o.PodInfos.Elements()))
	diags.Append(o.PodInfos.ElementsAs(ctx, &piMap, false)...)
	if diags.HasError() {
		return nil
	}

	podInfos := make(map[apstra.ObjectId]apstra.TemplatePodBasedInfo, len(piMap))
	for k, v := range piMap {
		podInfos[apstra.ObjectId(k)] = apstra.TemplatePodBasedInfo{
			Count: int(v.Count.ValueInt64()),
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

	return &apstra.CreatePodBasedTemplateRequest{
		DisplayName:            o.Name.ValueString(),
		Superspine:             ss.Request(ctx, diags),
		PodInfos:               podInfos,
		AntiAffinityPolicy:     antiAffinityPolicy,
		FabricAddressingPolicy: fabricAddressingPolicy,
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
	o.SuperSpine = NewDesignTemplateSuperSpineObject(ctx, &in.Superspine, diags)
	o.PodInfos = NewPodInfoMap(ctx, in, diags)
	o.FabricAddressing = fabricAddressing
}

func (o *TemplatePodBased) CopyWriteOnlyElements(ctx context.Context, src *TemplatePodBased, diags *diag.Diagnostics) {
	var srcSuperSpine, dstSuperSpine *SuperSpine

	// extract the source SuperSpine object from src
	diags.Append(src.SuperSpine.As(ctx, &srcSuperSpine, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}

	// extract the destination SuperSpine object from o
	diags.Append(o.SuperSpine.As(ctx, &dstSuperSpine, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}

	// clone missing SuperSpine bits
	dstSuperSpine.CopyWriteOnlyElements(ctx, srcSuperSpine, diags)

	// repackage the destination SuperSpine in o
	o.SuperSpine = utils.ObjectValueOrNull(ctx, SuperSpine{}.AttrTypes(), dstSuperSpine, diags)
	//
	//dstPodInfoMap := make(map[string]TemplatePodInfo)
	//diags.Append(o.PodInfos.ElementsAs(ctx, &dstPodInfoMap, false)...)
	//
	//srcPodInfoMap := make(map[string]TemplatePodInfo)
	//diags.Append(o.PodInfos.ElementsAs(ctx, &srcPodInfoMap, false)...)
	//
	//
}

func (o TemplatePodBased) VersionConstraints() apiversions.Constraints {
	var response apiversions.Constraints

	if o.FabricAddressing.IsUnknown() {
		return apiversions.Constraints{} // cannot validate
	}

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

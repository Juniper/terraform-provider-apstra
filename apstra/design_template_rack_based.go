package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
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
)

type templateRackBased struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Spine                  types.Object `tfsdk:"spine"`
	AsnAllocation          types.String `tfsdk:"asn_allocation_scheme"`
	OverlayControlProtocol types.String `tfsdk:"overlay_control_protocol"`
	FabricAddressing       types.String `tfsdk:"fabric_link_addressing"`
	RackInfos              types.Map    `tfsdk:"rack_infos"`
	//RackTypeIds            types.Map    `tfsdk:"rack_types_ids"`
	//RackTypes              types.Map    `tfsdk:"rack_types"`
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
		//"rack_types": dataSourceSchema.MapNestedAttribute{
		//	MarkdownDescription: "Details Rack Types included in the template",
		//	Computed:            true,
		//	NestedObject: dataSourceSchema.NestedAttributeObject{
		//		Attributes: templateRackInfo{}.dataSourceAttributesNested(),
		//	},
		//},
	}
}

func (o templateRackBased) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Agent Profile.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra name of the Agent Profile.",
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
			MarkdownDescription: "Fabric addressing scheme for spine/leaf links.",
			Required:            true,
		},
		//"rack_types": resourceSchema.MapNestedAttribute{
		//	MarkdownDescription: "Details Rack Types included in the template",
		//	Computed:            true,
		//	NestedObject: resourceSchema.NestedAttributeObject{
		//		Attributes: templateRackInfo{}.resourceAttributesNested(),
		//	},
		//},
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
		"rack_types":               types.MapType{ElemType: types.ObjectType{AttrTypes: rackType{}.attrTypes()}},
	}
}

//func (o *templateRackBased) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.CreateRackBasedTemplateRequest {
//	var d diag.Diagnostics
//
//	var s *spine
//	d = o.Spine.As(ctx, s, basetypes.ObjectAsOptions{})
//	diags.Append(d...)
//	if diags.HasError() {
//		return nil
//	}
//
//	rtMap := make(map[string]rackType, len(o.RackTypes.Elements()))
//	d = o.RackTypes.ElementsAs(ctx, &rtMap, false)
//	diags.Append(d...)
//	if diags.HasError() {
//		return nil
//	}
//
//	rackInfo := make([]goapstra.TemplateRackBasedRackInfo, len(rtMap))
//	for k, v := range rtMap {
//
//	}
//
//	return &goapstra.CreateRackBasedTemplateRequest{
//		DisplayName:            o.Name.ValueString(),
//		Capability:             goapstra.TemplateCapabilityNone,
//		Spine:                  s.request(ctx, diags),
//		RackInfo: ,
//		DhcpServiceIntent:      nil,
//		AntiAffinityPolicy:     nil,
//		AsnAllocationPolicy:    nil,
//		FabricAddressingPolicy: nil,
//		VirtualNetworkPolicy:   nil,
//	}
//}

func (o *templateRackBased) validate(ctx context.Context, diags *diag.Diagnostics) {
	rackInfoMap := make(map[string]templateRackInfo, len(o.RackInfos.Elements()))
	d := o.RackInfos.ElementsAs(ctx, &rackInfoMap, false)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	idMap := make(map[string]struct{}, len(rackInfoMap))
	for _, rackInfo := range rackInfoMap {
		id := rackInfo.RackTypeId.ValueString()
		if _, ok := idMap[id]; ok {
			diags.AddAttributeError(path.Root("rack_infos").AtMapKey(id), errInvalidConfig,
				fmt.Sprintf("rack type id %q used multiple times", id))
			return
		}
		idMap[rackInfo.RackTypeId.ValueString()] = struct{}{}
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

	riSlice := make([]templateRackInfo, len(in.RackInfo))
	for i := range in.RackInfo {
		riSlice[i].loadApiData(ctx, &in.RackInfo[i], diags)
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Spine = newDesignTemplateSpineObject(ctx, &in.Spine, diags)
	o.AsnAllocation = types.StringValue(asnAllocationSchemeToString(in.AsnAllocationPolicy.SpineAsnScheme, diags))
	o.OverlayControlProtocol = types.StringValue(overlayControlProtocolToString(in.VirtualNetworkPolicy.OverlayControlProtocol, diags))
	o.RackInfos = newRackInfoMap(ctx, in, diags)
}

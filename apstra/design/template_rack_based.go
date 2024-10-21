package design

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
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

type TemplateRackBased struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Spine                  types.Object `tfsdk:"spine"`
	AsnAllocation          types.String `tfsdk:"asn_allocation_scheme"`
	OverlayControlProtocol types.String `tfsdk:"overlay_control_protocol"`
	RackInfos              types.Map    `tfsdk:"rack_infos"`
}

func (o TemplateRackBased) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                       types.StringType,
		"name":                     types.StringType,
		"spine":                    types.ObjectType{AttrTypes: Spine{}.AttrTypes()},
		"asn_allocation_scheme":    types.StringType,
		"overlay_control_protocol": types.StringType,
		"rack_infos":               types.MapType{ElemType: types.ObjectType{AttrTypes: TemplateRackInfo{}.AttrTypes()}},
	}
}

func (o TemplateRackBased) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
		"spine": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Spine layer details",
			Computed:            true,
			Attributes:          Spine{}.DataSourceAttributes(),
		},
		"asn_allocation_scheme": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("%q is for 3-stage designs; %q is for 5-stage designs.",
				AsnAllocationUnique, AsnAllocationSingle),
			Computed: true,
		},
		"overlay_control_protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Defines the inter-rack virtual network overlay protocol in the fabric.",
			Computed:            true,
		},
		"rack_infos": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Rack Type info (count + details)",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: TemplateRackInfo{}.DataSourceAttributesNested(),
			},
		},
	}
}

func (o TemplateRackBased) DataSourceAttributesNested() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the pod inside the 5 stage template.",
			Computed:            true,
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name of the pod inside the 5 stage template.",
			Computed:            true,
		},
		"spine": dataSourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Spine layer details",
			Computed:            true,
			Attributes:          Spine{}.DataSourceAttributes(),
		},
		"asn_allocation_scheme": dataSourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("%q is for 3-stage designs; %q is for 5-stage designs.",
				AsnAllocationUnique, AsnAllocationSingle),
			Computed: true,
		},
		"overlay_control_protocol": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Defines the inter-rack virtual network overlay protocol in the fabric.",
			Computed:            true,
		},
		"rack_infos": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Rack Type info (count + details)",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: TemplateRackInfo{}.DataSourceAttributesNested(),
			},
		},
	}
}

func (o TemplateRackBased) ResourceAttributes() map[string]resourceSchema.Attribute {
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
			Attributes:          Spine{}.ResourceAttributes(),
		},
		"asn_allocation_scheme": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("%q is for 3-stage designs; %q is for 5-stage designs.",
				AsnAllocationUnique, AsnAllocationSingle),
			Validators: []validator.String{stringvalidator.OneOf(AsnAllocationUnique, AsnAllocationSingle)},
			Required:   true,
		},
		"overlay_control_protocol": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Defines the inter-rack virtual network overlay protocol in the fabric. [%q,%q]",
				OverlayControlProtocolEvpn, OverlayControlProtocolStatic),
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(OverlayControlProtocolEvpn, OverlayControlProtocolStatic),
				// todo make sure not ipv6 with evpn
			},
		},
		"rack_infos": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Rack Type info (count + details) keyed by Rack Type ID.",
			Required:            true,
			Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: TemplateRackInfo{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o TemplateRackBased) ResourceAttributesNested() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the pod inside the 5 stage template.",
			Computed:            true,
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name of the pod inside the 5 stage template.",
			Computed:            true,
		},
		"spine": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Spine layer details",
			Computed:            true,
			Attributes:          Spine{}.ResourceAttributes(),
		},
		"asn_allocation_scheme": resourceSchema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("%q is for 3-stage designs; %q is for 5-stage designs.",
				AsnAllocationUnique, AsnAllocationSingle),
			Computed: true,
		},
		"overlay_control_protocol": resourceSchema.StringAttribute{
			MarkdownDescription: "Defines the inter-rack virtual network overlay protocol in the fabric.",
			Computed:            true,
		},
		"rack_infos": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "Map of Rack Type info (count + details)",
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: TemplateRackInfo{}.ResourceAttributesNested(),
			},
		},
	}
}

func (o *TemplateRackBased) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.CreateRackBasedTemplateRequest {
	s := Spine{}
	diags.Append(o.Spine.As(ctx, &s, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	rtMap := make(map[string]TemplateRackInfo, len(o.RackInfos.Elements()))
	diags.Append(o.RackInfos.ElementsAs(ctx, &rtMap, false)...)
	if diags.HasError() {
		return nil
	}

	rackInfos := make(map[apstra.ObjectId]apstra.TemplateRackBasedRackInfo, len(rtMap))
	for k := range rtMap {
		rackInfos[apstra.ObjectId(k)] = apstra.TemplateRackBasedRackInfo{
			Count: int(rtMap[k].Count.ValueInt64()),
		}
	}

	var err error

	antiAffinityPolicy := &apstra.AntiAffinityPolicy{
		Algorithm: apstra.AlgorithmHeuristic,
	}

	var spineAsnScheme apstra.AsnAllocationScheme
	err = utils.ApiStringerFromFriendlyString(&spineAsnScheme, o.AsnAllocation.ValueString())
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing ASN allocation scheme %q - %s",
				o.AsnAllocation.ValueString(), err.Error()))
	}
	asnAllocationPolicy := &apstra.AsnAllocationPolicy{
		SpineAsnScheme: spineAsnScheme,
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

	return &apstra.CreateRackBasedTemplateRequest{
		DisplayName:       o.Name.ValueString(),
		Spine:             s.Request(ctx, diags),
		RackInfos:         rackInfos,
		DhcpServiceIntent: &apstra.DhcpServiceIntent{Active: true},
		// todo: is this the right AntiAffinityPolicy?
		//  I'd have sent <nil>, but blocked by sdk issue #2 (crash on nil pointer deref)
		AntiAffinityPolicy:   antiAffinityPolicy,
		AsnAllocationPolicy:  asnAllocationPolicy,
		VirtualNetworkPolicy: virtualNetworkPolicy,
	}
}

func (o *TemplateRackBased) LoadApiData(ctx context.Context, in *apstra.TemplateRackBasedData, diags *diag.Diagnostics) {
	if in == nil {
		diags.AddError(errProviderBug, "attempt to load TemplateRackBased from nil source")
		return
	}

	o.Name = types.StringValue(in.DisplayName)
	o.Spine = NewDesignTemplateSpineObject(ctx, &in.Spine, diags)
	o.AsnAllocation = types.StringValue(utils.StringersToFriendlyString(in.AsnAllocationPolicy.SpineAsnScheme))
	o.OverlayControlProtocol = types.StringValue(utils.StringersToFriendlyString(in.VirtualNetworkPolicy.OverlayControlProtocol))
	o.RackInfos = NewRackInfoMap(ctx, in, diags)
}

func (o *TemplateRackBased) CopyWriteOnlyElements(ctx context.Context, src *TemplateRackBased, diags *diag.Diagnostics) {
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

func NewTemplateRackBasedObject(ctx context.Context, in *apstra.TemplateRackBasedData, diags *diag.Diagnostics) types.Object {
	var trb TemplateRackBased
	trb.Id = types.StringNull()
	trb.LoadApiData(ctx, in, diags)
	if diags.HasError() {
		return types.ObjectNull(TemplateRackBased{}.AttrTypes())
	}

	trbObj, d := types.ObjectValueFrom(ctx, TemplateRackBased{}.AttrTypes(), &trb)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(TemplateRackBased{}.AttrTypes())
	}

	return trbObj
}

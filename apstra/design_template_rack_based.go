package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	RackTypes              types.Map    `tfsdk:"rack_types"`
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
		"rack_types": dataSourceSchema.MapNestedAttribute{
			MarkdownDescription: "Details Rack Types included in the template",
			Computed:            true,
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: templateRackInfo{}.dataSourceAttributes(),
			},
		},
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
	o.AsnAllocation = types.StringValue(asnAllocationSchemeToString(in.AsnAllocationPolicy.SpineAsnScheme, diags))
	o.Spine = newDesignTemplateSpineObject(ctx, &in.Spine, diags)
	o.OverlayControlProtocol = types.StringValue(overlayControlProtocolToString(in.VirtualNetworkPolicy.OverlayControlProtocol, diags))
	o.RackTypes = newRackTypeMap(ctx, in, diags)
}

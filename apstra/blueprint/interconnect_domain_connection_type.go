package blueprint

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	ierrors "github.com/Juniper/terraform-provider-apstra/internal/errors"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InterconnectDomainConnectionType struct {
	BlueprintID          types.String `tfsdk:"blueprint_id"`
	InterconnectDomainID types.String `tfsdk:"interconnect_domain_id"`
	VirtualNetworkID     types.String `tfsdk:"virtual_network_id"`
	Layer2Enabled        types.Bool   `tfsdk:"layer_2_enabled"`
	Layer3Enabled        types.Bool   `tfsdk:"layer_3_enabled"`
	TranslationVNI       types.Int64  `tfsdk:"translation_vni"`
}

func (ct InterconnectDomainConnectionType) DatasourceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"blueprint_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"interconnect_domain_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Interconnect Domain.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"virtual_network_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Virtual Network.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"layer_2_enabled": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Enables Layer 2 (EVPN Type 2) route exchange.",
			Computed:            true,
		},
		"layer_3_enabled": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Enables Layer 3 (EVPN Type 5) route exchange.",
			Computed:            true,
		},
		"translation_vni": datasourceSchema.Int64Attribute{
			MarkdownDescription: "The intermediate VNI to be used. It isn't required, but it needs to " +
				"match the remote VNI either by translation on the other side or by both data centers " +
				"using the same VNI for the virtual network that's being extended",
			Computed: true,
		},
	}
}

func (ct InterconnectDomainConnectionType) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"interconnect_domain_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Interconnect Domain.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"virtual_network_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Virtual Network.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"layer_2_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables Layer 2 (EVPN Type 2) route exchange.",
			Required:            true,
		},
		"layer_3_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Enables Layer 3 (EVPN Type 5) route exchange.",
			Required:            true,
		},
		"translation_vni": resourceSchema.Int64Attribute{
			MarkdownDescription: "The intermediate VNI to be used. It isn't required, but it needs to " +
				"match the remote VNI either by translation on the other side or by both data centers " +
				"using the same VNI for the virtual network that's being extended",
			Optional: true,
			Validators: []validator.Int64{
				int64validator.Between(1, 1<<24-1),
				apstravalidator.ForbiddenWhenValueIs(path.MatchRoot("layer_2_enabled"), types.BoolValue(false)),
			},
		},
	}
}

func (ct InterconnectDomainConnectionType) Request(_ context.Context, _ *diag.Diagnostics) apstra.EVPNInterconnectGroup {
	result := apstra.EVPNInterconnectGroup{
		InterconnectVirtualNetworks: map[string]apstra.InterconnectVirtualNetwork{
			ct.VirtualNetworkID.ValueString(): {
				L2Enabled:      ct.Layer2Enabled.ValueBool(),
				L3Enabled:      ct.Layer3Enabled.ValueBool(),
				TranslationVNI: pointer.To(uint32(ct.TranslationVNI.ValueInt64())),
			},
		},
	}
	_ = result.SetID(ct.InterconnectDomainID.ValueString())
	return result
}

// Read returns ResourceNotFound errors on 404 rather than diagnostics so that the caller can decide how to handle them.
func (ct *InterconnectDomainConnectionType) Read(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) error {
	g, err := bp.GetEVPNInterconnectGroup(ctx, ct.InterconnectDomainID.ValueString())
	if err != nil {
		if utils.IsApstra404(err) {
			return ierrors.ResourceNotFound(err.Error())
		}
		diags.AddError("Failed to fetch Interconnect Domain", err.Error())
		return nil
	}

	if ivn, ok := g.InterconnectVirtualNetworks[ct.VirtualNetworkID.ValueString()]; ok {
		ct.Layer2Enabled = types.BoolValue(ivn.L2Enabled)
		ct.Layer3Enabled = types.BoolValue(ivn.L3Enabled)
		ct.TranslationVNI = types.Int64PointerValue(pointer.ConvertInteger((*int64)(nil), ivn.TranslationVNI))
		return nil
	}

	return ierrors.ResourceNotFound(
		fmt.Sprintf(
			"Interconnect Domain %s of Blueprint %s does not have Interconnect Domain settings for Virtual Network %s",
			ct.InterconnectDomainID, ct.BlueprintID, ct.VirtualNetworkID,
		),
	)
}

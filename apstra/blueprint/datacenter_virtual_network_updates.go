package blueprint

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// This file contains the changes needed to add SVI IP support to the DatacenterVirtualNetwork resource

// Update DatacenterVirtualNetwork struct to include SviIps
type DatacenterVirtualNetworkWithSviIps struct {
	DatacenterVirtualNetwork
	SviIps types.Set `tfsdk:"svi_ips"`
}

// Update ResourceAttributes to include SviIps field
func (o DatacenterVirtualNetworkWithSviIps) ResourceAttributes() map[string]resourceSchema.Attribute {
	attrs := o.DatacenterVirtualNetwork.ResourceAttributes()
	attrs["svi_ips"] = resourceSchema.SetNestedAttribute{
		MarkdownDescription: "SVI IP assignments for switches in the virtual network",
		Optional:            true,
		NestedObject: resourceSchema.NestedAttributeObject{
			Attributes: SviIp{}.ResourceAttributes(),
		},
	}
	return attrs
}

// Update Request to include SviIps
func (o *DatacenterVirtualNetworkWithSviIps) RequestWithSviIps(ctx context.Context, diags *diag.Diagnostics) *apstra.VirtualNetworkData {
	// Get the base request
	request := o.DatacenterVirtualNetwork.Request(ctx, diags)
	if diags.HasError() {
		return nil
	}

	// Add SviIps to the request if provided
	if !o.SviIps.IsNull() {
		var sviIpsSlice []SviIp
		diags.Append(o.SviIps.ElementsAs(ctx, &sviIpsSlice, false)...)
		if diags.HasError() {
			return nil
		}

		apiSviIps := make([]apstra.SviIp, len(sviIpsSlice))
		for i, sviIp := range sviIpsSlice {
			apiSviIps[i] = *sviIp.Request(ctx, diags)
			if diags.HasError() {
				return nil
			}
		}
		request.SviIps = apiSviIps
	}

	return request
}

// Update LoadApiData to handle SviIps
func (o *DatacenterVirtualNetworkWithSviIps) LoadApiDataWithSviIps(ctx context.Context, in *apstra.VirtualNetworkData, diags *diag.Diagnostics) {
	// Load the base data
	o.DatacenterVirtualNetwork.LoadApiData(ctx, in, diags)
	
	// Load SviIps if present
	if len(in.SviIps) == 0 {
		o.SviIps = types.SetNull(types.ObjectType{AttrTypes: SviIp{}.AttrTypes()})
		return
	}

	// Convert API SviIps to Terraform objects
	tfSviIps := make([]attr.Value, len(in.SviIps))
	for i, apiSviIp := range in.SviIps {
		var tfSviIp SviIp
		tfSviIp.LoadApiData(ctx, apiSviIp, diags)
		if diags.HasError() {
			return
		}
		
		tfSviIps[i] = types.ObjectValueMust(
			SviIp{}.AttrTypes(),
			map[string]attr.Value{
				"system_id":    tfSviIp.SystemId,
				"ipv4_address": tfSviIp.IPv4Address,
				"ipv4_mode":    tfSviIp.IPv4Mode,
				"ipv6_address": tfSviIp.IPv6Address,
				"ipv6_mode":    tfSviIp.IPv6Mode,
			},
		)
	}

	o.SviIps = types.SetValueMust(types.ObjectType{AttrTypes: SviIp{}.AttrTypes()}, tfSviIps)
}

// SviIp struct definition and related functions would be defined in a separate file (already created)
type SviIp struct {
	SystemId    types.String `tfsdk:"system_id"`
	IPv4Address types.String `tfsdk:"ipv4_address"`
	IPv4Mode    types.String `tfsdk:"ipv4_mode"`
	IPv6Address types.String `tfsdk:"ipv6_address"`
	IPv6Mode    types.String `tfsdk:"ipv6_mode"`
}

// AttrTypes, ResourceAttributes, Request, and LoadApiData methods for SviIp would be defined in the separate file
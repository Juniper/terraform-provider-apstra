// Copyright (c) Juniper Networks, Inc., 2022. All rights reserved.

package blueprint

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DatacenterVirtualNetworkWithSviIps extends the DatacenterVirtualNetwork struct with SVI IP support
type DatacenterVirtualNetworkWithSviIps struct {
	DatacenterVirtualNetwork
	SviIps types.Set `tfsdk:"svi_ips"`
}

// ResourceAttributesWithSviIps adds SVI IPs to the DatacenterVirtualNetwork resource attributes
func (o DatacenterVirtualNetwork) ResourceAttributesWithSviIps() map[string]resourceSchema.Attribute {
	attributes := o.ResourceAttributes()
	
	// Add SviIps attribute
	attributes["svi_ips"] = resourceSchema.SetNestedAttribute{
		MarkdownDescription: "SVI IP assignments for switches in the virtual network. This allows explicit " +
			"control over the secondary virtual interface IPs assigned to switches, preventing overlaps " +
			"when identical virtual networks are created in multiple blueprints.",
		Optional: true,
		NestedObject: resourceSchema.NestedAttributeObject{
			Attributes: SviIp{}.ResourceAttributes(),
		},
	}
	
	return attributes
}

// RequestWithSviIps creates a VirtualNetworkData request that includes SVI IPs
func (o *DatacenterVirtualNetworkWithSviIps) RequestWithSviIps(ctx context.Context, diags *diag.Diagnostics) *apstra.VirtualNetworkData {
	// Get the base request from the standard DatacenterVirtualNetwork struct
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

// LoadApiDataWithSviIps loads data from the Apstra API response, including SVI IPs
func (o *DatacenterVirtualNetworkWithSviIps) LoadApiDataWithSviIps(ctx context.Context, in *apstra.VirtualNetworkData, diags *diag.Diagnostics) {
	// Load base data into the embedded DatacenterVirtualNetwork
	o.DatacenterVirtualNetwork.LoadApiData(ctx, in, diags)
	if diags.HasError() {
		return
	}
	
	// Load SVI IPs
	o.SviIps = loadApiSviIps(ctx, in.SviIps, diags)
}

// LoadApiSviIps is a public wrapper for the private loadApiSviIps function
func LoadApiSviIps(ctx context.Context, in []apstra.SviIp, diags *diag.Diagnostics) types.Set {
	return loadApiSviIps(ctx, in, diags)
}
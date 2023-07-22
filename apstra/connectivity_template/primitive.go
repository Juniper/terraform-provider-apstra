package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type JsonPrimitive interface {
	//Marshal(context.Context, *diag.Diagnostics) string
	//DataSourceAttributes() map[string]dataSourceSchema.Attribute
	attributes(context.Context, path.Path, *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes
	SdkPrimitive(context.Context, path.Path, *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive
}

// TfCfgPrimitive is a Terraform Config Primitive. it's the JSON structure used
// as a Connectivity Template Primitive (Single VN, IP Link, etc...) in a
// terraform config file (usually written by one of the primitive-specific data
// sources:
//   - apstra_datacenter_ct_virtual_network_single
//   - apstra_datacenter_ct_ip_link
//   - etc...
type TfCfgPrimitive struct { // todo make private
	PrimitiveType string          `json:"type"`
	Data          json.RawMessage `json:"data"`
}

// Rehydrate expands the TfCfgPrimitive (a type with raw json) into a type
// specific implementation of JsonPrimitive.
func (o TfCfgPrimitive) Rehydrate() (JsonPrimitive, error) { // todo make private
	var pType apstra.CtPrimitivePolicyTypeName
	err := pType.FromString(o.PrimitiveType)
	if err != nil {
		return nil, err
	}

	var jsonPrimitive JsonPrimitive
	switch pType {
	case apstra.CtPrimitivePolicyTypeNameAttachSingleVlan:
		jsonPrimitive = new(vnSinglePrototype)
		err = json.Unmarshal(o.Data, jsonPrimitive.(*vnSinglePrototype))
	//case apstra.CtPrimitivePolicyTypeNameAttachMultipleVLAN:
	//	primitive = new(VnMultiplePrototype)
	case apstra.CtPrimitivePolicyTypeNameAttachLogicalLink:
		jsonPrimitive = new(ipLinkPrototype)
		err = json.Unmarshal(o.Data, jsonPrimitive.(*ipLinkPrototype))
	case apstra.CtPrimitivePolicyTypeNameAttachStaticRoute:
		jsonPrimitive = new(staticRoutePrototype)
		err = json.Unmarshal(o.Data, jsonPrimitive.(*staticRoutePrototype))
	//case apstra.CtPrimitivePolicyTypeNameAttachCustomStaticRoute:
	//	primitive = new(CustomStaticRoute)
	//case apstra.CtPrimitivePolicyTypeNameAttachIpEndpointWithBgpNsxt:
	//	primitive = new(BgpPeeringIpEndpoint)
	//case apstra.CtPrimitivePolicyTypeNameAttachBgpOverSubinterfacesOrSvi:
	//	primitive = new(BgpPeeringGenericSystem)
	//case apstra.CtPrimitivePolicyTypeNameAttachBgpWithPrefixPeeringForSviOrSubinterface:
	//	primitive = new(DynamicBgpPeering)
	//case apstra.CtPrimitivePolicyTypeNameAttachExistingRoutingPolicy:
	//	primitive = new(RoutingPolicy)
	//case apstra.CtPrimitivePolicyTypeNameAttachRoutingZoneConstraint:
	//	primitive = new(RoutingZoneConstraint)
	default:
		return nil, fmt.Errorf("unhandled primitive type %q", pType.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rendered primitive data - %w", err)
	}

	return jsonPrimitive, nil
}

func ChildPrimitivesFromListOfJsonStrings(ctx context.Context, in []string, path path.Path, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	result := make([]*apstra.ConnectivityTemplatePrimitive, len(in))
	for i, s := range in {
		var rp TfCfgPrimitive // todo rename
		err := json.Unmarshal([]byte(s), &rp)
		if err != nil {
			diags.AddAttributeError(path.AtListIndex(i), "failed to marshal primitive", err.Error())
			return nil
		}

		primitive, err := rp.Rehydrate() // todo rename jsonPrimitive
		if err != nil {
			diags.AddAttributeError(path.AtListIndex(i), "failed parsing primitive", err.Error())
			return nil
		}

		sdkPrimitive := primitive.SdkPrimitive(ctx, path.AtListIndex(i), diags)
		if diags.HasError() {
			return nil
		}

		result[i] = sdkPrimitive
	}

	return result
}

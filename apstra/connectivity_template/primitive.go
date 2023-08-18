package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Primitive interface {
	Marshal(context.Context, *diag.Diagnostics) string
	DataSourceAttributes() map[string]dataSourceSchema.Attribute
	loadSdkPrimitive(context.Context, apstra.ConnectivityTemplatePrimitive, *diag.Diagnostics)
}

type JsonPrimitive interface {
	attributes(context.Context, path.Path, *diag.Diagnostics) apstra.ConnectivityTemplatePrimitiveAttributes
	ToSdkPrimitive(context.Context, path.Path, *diag.Diagnostics) *apstra.ConnectivityTemplatePrimitive
}

// tfCfgPrimitive is a Terraform Config Primitive. it's the JSON structure used
// as a Connectivity Template Primitive (Single VN, IP Link, etc...) in a
// terraform config file (usually written by one of the primitive-specific data
// sources:
//   - apstra_datacenter_ct_virtual_network_single
//   - apstra_datacenter_ct_ip_link
//   - etc...
type tfCfgPrimitive struct {
	PrimitiveType string          `json:"type"`
	Label         string          `json:"label"`
	Data          json.RawMessage `json:"data"`
}

// rehydrate expands the tfCfgPrimitive (a type with raw json) into a type
// specific implementation of JsonPrimitive.
func (o tfCfgPrimitive) rehydrate(_ context.Context, path path.Path, diags *diag.Diagnostics) JsonPrimitive {
	var pType apstra.CtPrimitivePolicyTypeName
	err := pType.FromString(o.PrimitiveType)
	if err != nil {
		diags.AddAttributeError(path, fmt.Sprintf("failed parsing primitive type string %q", o.PrimitiveType), err.Error())
		return nil
	}

	var jsonPrimitive JsonPrimitive
	switch pType {
	case apstra.CtPrimitivePolicyTypeNameAttachSingleVlan:
		jsonPrimitive = &vnSinglePrototype{}
	case apstra.CtPrimitivePolicyTypeNameAttachMultipleVlan:
		jsonPrimitive = &vnMultiplePrototype{}
	case apstra.CtPrimitivePolicyTypeNameAttachLogicalLink:
		jsonPrimitive = &ipLinkPrototype{}
	case apstra.CtPrimitivePolicyTypeNameAttachStaticRoute:
		jsonPrimitive = &staticRoutePrototype{Label: o.Label}
	case apstra.CtPrimitivePolicyTypeNameAttachCustomStaticRoute:
		jsonPrimitive = &customStaticRoutePrototype{}
	case apstra.CtPrimitivePolicyTypeNameAttachIpEndpointWithBgpNsxt:
		jsonPrimitive = &bgpPeeringIpEndpointPrototype{}
	case apstra.CtPrimitivePolicyTypeNameAttachBgpOverSubinterfacesOrSvi:
		jsonPrimitive = &bgpPeeringGenericSystemPrototype{}
	case apstra.CtPrimitivePolicyTypeNameAttachBgpWithPrefixPeeringForSviOrSubinterface:
		jsonPrimitive = &dynamicBgpPeeringPrototype{}
	case apstra.CtPrimitivePolicyTypeNameAttachExistingRoutingPolicy:
		jsonPrimitive = &routingPolicyPrototype{}
	case apstra.CtPrimitivePolicyTypeNameAttachRoutingZoneConstraint:
		jsonPrimitive = &routingZoneConstraintPrototype{}
	default:
		diags.AddAttributeError(path, "primitive rehydration failed", fmt.Sprintf("unhandled primitive type %q", pType.String()))
		return nil
	}
	err = json.Unmarshal(o.Data, jsonPrimitive)
	if err != nil {
		diags.AddAttributeError(path, "primitive rehydration failed", err.Error())
		return nil
	}

	return jsonPrimitive
}

func ChildPrimitivesFromListOfJsonStrings(ctx context.Context, in []string, path path.Path, diags *diag.Diagnostics) []*apstra.ConnectivityTemplatePrimitive {
	result := make([]*apstra.ConnectivityTemplatePrimitive, len(in))
	for i, s := range in {
		var rp tfCfgPrimitive // todo rename
		err := json.Unmarshal([]byte(s), &rp)
		if err != nil {
			diags.AddAttributeError(path.AtListIndex(i), "failed to marshal primitive", err.Error())
			return nil
		}

		jsonPrimitive := rp.rehydrate(ctx, path.AtListIndex(i), diags)
		if diags.HasError() {
			return nil
		}

		sdkPrimitive := jsonPrimitive.ToSdkPrimitive(ctx, path.AtListIndex(i), diags)
		if diags.HasError() {
			return nil
		}

		result[i] = sdkPrimitive
	}

	return result
}

func PrimitiveFromSdk(ctx context.Context, in *apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) Primitive {
	var primitive Primitive

	switch in.Attributes.PolicyTypeName() {
	case apstra.CtPrimitivePolicyTypeNameAttachSingleVlan:
		primitive = new(VnSingle)
	case apstra.CtPrimitivePolicyTypeNameAttachMultipleVlan:
		primitive = new(VnMultiple)
	case apstra.CtPrimitivePolicyTypeNameAttachLogicalLink:
		primitive = new(IpLink)
	case apstra.CtPrimitivePolicyTypeNameAttachStaticRoute:
		primitive = new(StaticRoute)
	case apstra.CtPrimitivePolicyTypeNameAttachCustomStaticRoute:
		primitive = new(CustomStaticRoute)
	case apstra.CtPrimitivePolicyTypeNameAttachIpEndpointWithBgpNsxt:
		primitive = new(BgpPeeringIpEndpoint)
	case apstra.CtPrimitivePolicyTypeNameAttachBgpOverSubinterfacesOrSvi:
		primitive = new(BgpPeeringGenericSystem)
	case apstra.CtPrimitivePolicyTypeNameAttachBgpWithPrefixPeeringForSviOrSubinterface:
		primitive = new(DynamicBgpPeering)
	case apstra.CtPrimitivePolicyTypeNameAttachExistingRoutingPolicy:
		primitive = new(RoutingPolicy)
	case apstra.CtPrimitivePolicyTypeNameAttachRoutingZoneConstraint:
		primitive = new(RoutingZoneConstraint)
	default:
		diags.AddError("parsing primitive from SDK failed", fmt.Sprintf("unhandled primitive type %q", in.Attributes.PolicyTypeName()))
		return nil
	}

	primitive.loadSdkPrimitive(ctx, *in, diags)
	if diags.HasError() {
		return nil
	}

	return primitive
}

func SdkPrimitivesToJsonStrings(ctx context.Context, in []*apstra.ConnectivityTemplatePrimitive, diags *diag.Diagnostics) []attr.Value {
	result := make([]attr.Value, len(in))
	for i := range in {
		p := PrimitiveFromSdk(ctx, in[i], diags)
		if diags.HasError() {
			return nil
		}
		result[i] = types.StringValue(p.Marshal(ctx, diags))
		if diags.HasError() {
			return nil
		}
	}
	return result
}

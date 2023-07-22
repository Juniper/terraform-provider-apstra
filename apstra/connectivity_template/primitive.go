package connectivitytemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type Primitive interface {
	Render(context.Context, *diag.Diagnostics) string
	DataSourceAttributes() map[string]dataSourceSchema.Attribute
	connectivityTemplateAttributes() (apstra.ConnectivityTemplateAttributes, error)
}

type RenderedPrimitive struct {
	PrimitiveType string          `json:"type"`
	Data          json.RawMessage `json:"data"`
}

func (o RenderedPrimitive) Parse() (apstra.ConnectivityTemplateAttributes, error) {
	var pType apstra.CtPrimitivePolicyTypeName
	err := pType.FromString(o.PrimitiveType)
	if err != nil {
		return nil, err
	}

	var primitive Primitive
	switch pType {
	case apstra.CtPrimitivePolicyTypeNameAttachSingleVlan:
		primitive = new(VnSingle)
	//case apstra.CtPrimitivePolicyTypeNameAttachMultipleVLAN:
	//	primitive = new(VnMultiple)
	case apstra.CtPrimitivePolicyTypeNameAttachLogicalLink:
		primitive = new(IpLink)
	case apstra.CtPrimitivePolicyTypeNameAttachStaticRoute:
		primitive = new(StaticRoute)
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
	err = json.Unmarshal(o.Data, primitive)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rendered primitive data - %w", err)
	}

	return primitive, nil
}

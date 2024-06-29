package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"strconv"
)

// GetCtIpLinkSubinterfaces returns a map of switch subinterfaces created by attaching a Connectivity
// Template with top-level "IP Link" (AttachLogialLInk) Primitivves keyed by VLAN number. Key 0
// represents the "no tag" condition. Inputs ctId and apId are both graph node IDs. They represent
// the Connectivity Template (top level ep_endpoint_policy object - the one used as "the CT ID" in
// the web UI) and the Application Point (switch port) ID respectively. Subinterfaces of apId which
// were not created by attaching ctId are ignored.
func GetCtIpLinkSubinterfaces(ctx context.Context, client *apstra.TwoStageL3ClosClient, ctId, apId apstra.ObjectId, diags *diag.Diagnostics) map[int64]apstra.ObjectId {
	ctNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_ct")}
	ctNodeIdAttr := apstra.QEEAttribute{Key: "id", Value: apstra.QEStringVal(ctId)}
	apNodeIdAttr := apstra.QEEAttribute{Key: "id", Value: apstra.QEStringVal(apId)}
	iplNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_ipl")}
	iplNodeTypeAttr := apstra.QEEAttribute{Key: "policy_type_name", Value: apstra.QEStringVal("AttachLogicalLink")}
	siNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_si")}
	siNodeTypeAttr := apstra.QEEAttribute{Key: "if_type", Value: apstra.QEStringVal("subinterface")}

	// query which identifies the Connectivity Template
	ctQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpEndpointPolicy.QEEAttribute(), ctNodeIdAttr, ctNodeNameAttr})

	// query which identifies IP Link CT primitives
	iplQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{ctNodeNameAttr}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeEpSubpolicy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpEndpointPolicy.QEEAttribute()}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeEpFirstSubpolicy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpEndpointPolicy.QEEAttribute(), iplNodeTypeAttr, iplNodeNameAttr})

	// query which identifies Subinterfaces
	siQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{ctNodeNameAttr}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeEpNested.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpApplicationInstance.QEEAttribute()}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeEpAffectedBy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpGroup.QEEAttribute()}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeEpMemberOf.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute(), apNodeIdAttr}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOf.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute(), siNodeTypeAttr, siNodeNameAttr})

	// query which ties together the previous queries via `match()`
	query := new(apstra.MatchQuery).
		SetBlueprintId(client.Id()).
		SetClient(client.Client()).
		Match(ctQuery).
		Match(iplQuery).
		Match(siQuery)

	qs := query.String()
	_ = qs

	// collect the query response here
	var queryResponse struct {
		Items []struct {
			IpL struct {
				Attributes json.RawMessage `json:"attributes"`
			} `json:"n_ipl"`
			Si struct {
				Id   apstra.ObjectId `json:"id"`
				Vlan *int64          `json:"vlan_id"`
			} `json:"n_si"`
		} `json:"items"`
	}

	// execute the query
	err := query.Do(ctx, &queryResponse)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to run graph query - %q", query.String()), err.Error())
		return nil
	}

	// the query result will include nested (and escaped) JSON. We'll unpack it here.
	var ipLinkAttributes struct {
		Vlan *int64 `json:"vlan_id"`
	}

	// prepare the result map
	result := make(map[int64]apstra.ObjectId)

	// iterate over the query response inspecting each item.
	// note that result will not necessarily be sized the same as queryResponse.Items
	// because the graph traversal can find multiple extraneous traversals:
	// - subinterfaces not related to the Connectivity Template
	// - mismatched combinations of IP Link primitive + subinterface
	for _, item := range queryResponse.Items {
		// un-quote the embedded JSON string
		attributes, err := strconv.Unquote(string(item.IpL.Attributes))
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to \"unquote\" IP Link attributes - '%s'", item.IpL.Attributes), err.Error())
			return nil
		}

		// unpack the embedded JSON string
		err = json.Unmarshal([]byte(attributes), &ipLinkAttributes)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to unmarshal IP Link attributes - '%s'", attributes), err.Error())
			return nil
		}

		// if the IP Link Primitive matches the Subinterface, collect the result
		switch {
		case ipLinkAttributes.Vlan == nil && item.Si.Vlan == nil:
			// found the matching untagged IP Link and Subinterface - save as VLAN 0
			result[0] = item.Si.Id
			continue
		case ipLinkAttributes.Vlan == nil || item.Si.Vlan == nil:
			// one item is untagged, but the other is not - not interesting
			continue
		case ipLinkAttributes.Vlan != nil && item.Si.Vlan != nil && *ipLinkAttributes.Vlan == *item.Si.Vlan:
			// IP link and subinterface are tagged - and they have matching values!
			result[*ipLinkAttributes.Vlan] = item.Si.Id
			continue
		}
	}

	return result
}

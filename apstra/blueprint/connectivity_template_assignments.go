package blueprint

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectivityTemplateAssignments struct {
	BlueprintId            types.String `tfsdk:"blueprint_id"`
	ConnectivityTemplateId types.String `tfsdk:"connectivity_template_id"`
	ApplicationPointIds    types.Set    `tfsdk:"application_point_ids"`
	FetchIpLinkIds         types.Bool   `tfsdk:"fetch_ip_link_ids"`
	IpLinkIds              types.Map    `tfsdk:"ip_links_ids"`
}

func (o ConnectivityTemplateAssignments) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"connectivity_template_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Connectivity Template ID which should be applied to the Application Points.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"application_point_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Apstra node IDs of the Interfaces or Systems where the Connectivity " +
				"Template should be applied.",
			Required:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"fetch_ip_link_ids": resourceSchema.BoolAttribute{
			MarkdownDescription: "When `true`, the read-only `ip_link_ids` attribute will be populated. Default " +
				"behavior skips retrieving `ip_link_ids` to improve performance in scenarios where this information " +
				"is not needed.",
			Optional: true,
		},
		"ip_links_ids": resourceSchema.MapAttribute{
			MarkdownDescription: "New Logical Links are created when Connectivity Templates containing *IP Link* " +
				"primitives are attached to a switch interface. These logical links may or may not be VLAN-tagged. " +
				"This attribute is a two-dimensional map. The outer map is keyed by Application Point ID. The inner " +
				"map is keyed by VLAN number. Untagged Logical Links are represented in the inner map by key `0`.\n" +
				"**Note:** requires `fetch_iplink_ids = true`",
			Computed:    true,
			ElementType: types.MapType{ElemType: types.StringType},
		},
	}
}

func (o *ConnectivityTemplateAssignments) Request(ctx context.Context, state *ConnectivityTemplateAssignments, diags *diag.Diagnostics) map[apstra.ObjectId]map[apstra.ObjectId]bool {
	var desired, current []apstra.ObjectId // Application Point IDs

	diags.Append(o.ApplicationPointIds.ElementsAs(ctx, &desired, false)...)
	if diags.HasError() {
		return nil
	}
	desiredMap := make(map[apstra.ObjectId]bool, len(desired))
	for _, apId := range desired {
		desiredMap[apId] = true
	}

	if state != nil {
		diags.Append(state.ApplicationPointIds.ElementsAs(ctx, &current, false)...)
		if diags.HasError() {
			return nil
		}
	}
	currentMap := make(map[apstra.ObjectId]bool, len(current))
	for _, apId := range current {
		currentMap[apId] = true
	}

	result := make(map[apstra.ObjectId]map[apstra.ObjectId]bool)
	ctId := apstra.ObjectId(o.ConnectivityTemplateId.ValueString())

	for _, ApplicationPointId := range desired {
		if _, ok := currentMap[ApplicationPointId]; !ok {
			// desired Application Point not found in currentMap -- need to add
			result[ApplicationPointId] = map[apstra.ObjectId]bool{ctId: true} // causes CT to be added
		}
	}

	for _, ApplicationPointId := range current {
		if _, ok := desiredMap[ApplicationPointId]; !ok {
			// current Application Point not found in desiredMap -- need to remove
			result[ApplicationPointId] = map[apstra.ObjectId]bool{ctId: false} // causes CT to be added
		}
	}

	return result
}

func (o *ConnectivityTemplateAssignments) GetIpLinkIds(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	o.IpLinkIds = types.MapNull(types.MapType{ElemType: types.StringType})

	if !o.FetchIpLinkIds.ValueBool() {
		return
	}

	var applicationPointIds []string
	diags.Append(o.ApplicationPointIds.ElementsAs(ctx, &applicationPointIds, false)...)
	if diags.HasError() {
		return
	}

	ctNodeIdAttr := apstra.QEEAttribute{Key: "id", Value: apstra.QEStringVal(o.ConnectivityTemplateId.ValueString())}
	ctNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_ct")}
	iplpNodeTypeAttr := apstra.QEEAttribute{Key: "policy_type_name", Value: apstra.QEStringVal("AttachLogicalLink")}
	iplpNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_iplp")}
	apNodeIdAttr := apstra.QEEAttribute{Key: "id", Value: apstra.QEStringValIsIn(applicationPointIds)}
	apNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_ap")}
	siNodeTypeAttr := apstra.QEEAttribute{Key: "if_type", Value: apstra.QEStringVal("subinterface")}
	siNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_si")}
	llNodeTypeAttr := apstra.QEEAttribute{Key: "link_type", Value: apstra.QEStringVal("logical_link")}
	llNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_ll")}

	// query which identifies the Connectivity Template
	ctQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpEndpointPolicy.QEEAttribute(), ctNodeIdAttr, ctNodeNameAttr})

	// query which identifies IP Link primitives within the CT
	iplpQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{ctNodeNameAttr}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeEpSubpolicy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpEndpointPolicy.QEEAttribute()}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeEpFirstSubpolicy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpEndpointPolicy.QEEAttribute(), iplpNodeTypeAttr, iplpNodeNameAttr})

	// query which identifies Subinterfaces
	llQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{ctNodeNameAttr}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeEpNested.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpApplicationInstance.QEEAttribute()}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeEpAffectedBy.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeEpGroup.QEEAttribute()}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeEpMemberOf.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute(), apNodeIdAttr, apNodeNameAttr}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeComposedOf.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute(), siNodeTypeAttr, siNodeNameAttr}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeLink.QEEAttribute(), llNodeTypeAttr, llNodeNameAttr})

	// query which ties together the previous queries via `match()`
	query := new(apstra.MatchQuery).
		SetBlueprintId(bp.Id()).
		SetClient(bp.Client()).
		Match(ctQuery).
		Match(iplpQuery).
		Match(llQuery)

	// collect the query response here
	var queryResponse struct {
		Items []struct {
			Iplp struct {
				Attributes json.RawMessage `json:"attributes"`
			} `json:"n_iplp"`
			Ap struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_ap"`
			Si struct {
				Vlan *int `json:"vlan_id"`
			} `json:"n_si"`
			Ll struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_ll"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResponse)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed to run graph query - %q", query.String()), err.Error())
		return
	}

	// the query result will include nested (and escaped) JSON. We'll unpack it here.
	var ipLinkAttributes struct {
		Vlan *int `json:"vlan_id"`
	}

	// prepare the result map
	result := make(map[apstra.ObjectId]map[string]apstra.ObjectId)

	addToResult := func(apId apstra.ObjectId, vlan int, linkId apstra.ObjectId) {
		innerMap, ok := result[apId]
		if !ok {
			innerMap = make(map[string]apstra.ObjectId)
		}
		innerMap[strconv.Itoa(vlan)] = linkId
		result[apId] = innerMap
	}
	_ = addToResult

	for _, item := range queryResponse.Items {
		// un-quote the embedded JSON string
		attributes, err := strconv.Unquote(string(item.Iplp.Attributes))
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to un-quote IP Link attributes - '%s'", item.Iplp.Attributes), err.Error())
			return
		}

		// unpack the embedded JSON string
		err = json.Unmarshal([]byte(attributes), &ipLinkAttributes)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to unmarshal IP Link attributes - '%s'", attributes), err.Error())
			return
		}

		// if the IP Link Primitive matches the Subinterface, collect the result
		switch {
		case ipLinkAttributes.Vlan == nil && item.Si.Vlan == nil:
			// found the matching untagged IP Link and Subinterface - save as VLAN 0
			addToResult(item.Ap.Id, 0, item.Ll.Id)
			continue
		case ipLinkAttributes.Vlan == nil || item.Si.Vlan == nil:
			// one item is untagged, but the other is not - not interesting
			continue
		case ipLinkAttributes.Vlan != nil && item.Si.Vlan != nil && *ipLinkAttributes.Vlan == *item.Si.Vlan:
			// IP link and subinterface are tagged - and they have matching values!
			addToResult(item.Ap.Id, *ipLinkAttributes.Vlan, item.Ll.Id)
			continue
		}
	}

	var d diag.Diagnostics
	o.IpLinkIds, d = types.MapValueFrom(ctx, types.MapType{ElemType: types.StringType}, result)
	diags.Append(d...)
}

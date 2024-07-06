package blueprint

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectivityTemplatesAssignment struct {
	BlueprintId             types.String `tfsdk:"blueprint_id"`
	ConnectivityTemplateIds types.Set    `tfsdk:"connectivity_template_ids"`
	ApplicationPointId      types.String `tfsdk:"application_point_id"`
	FetchIpLinkIds          types.Bool   `tfsdk:"fetch_ip_link_ids"`
	IpLinkIds               types.Map    `tfsdk:"ip_links_ids"`
}

func (o ConnectivityTemplatesAssignment) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra Blueprint ID.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"application_point_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra node ID of the Interface or System where the Connectivity Templates " +
				"should be applied.",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"connectivity_template_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of Connectivity Template IDs which should be applied to the Application Point.",
			Required:            true,
			ElementType:         types.StringType,
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
				"This attribute is a two-dimensional map. The outer map is keyed by Connectivity Template ID. The inner " +
				"map is keyed by VLAN number. Untagged Logical Links are represented in the inner map by key `0`.\n" +
				"**Note:** requires `fetch_iplink_ids = true`",
			Computed:    true,
			ElementType: types.MapType{ElemType: types.StringType},
		},
	}
}

func (o *ConnectivityTemplatesAssignment) AddDelRequest(ctx context.Context, state *ConnectivityTemplatesAssignment, diags *diag.Diagnostics) ([]apstra.ObjectId, []apstra.ObjectId) {
	var planIds, stateIds []apstra.ObjectId

	if o != nil { // o will be nil in Delete()
		diags.Append(o.ConnectivityTemplateIds.ElementsAs(ctx, &planIds, false)...)
		if diags.HasError() {
			return nil, nil
		}
	}

	if state != nil { // state will be nil in Create()
		diags.Append(state.ConnectivityTemplateIds.ElementsAs(ctx, &stateIds, false)...)
		if diags.HasError() {
			return nil, nil
		}
	}

	return utils.SliceComplementOfA(stateIds, planIds), utils.SliceComplementOfA(planIds, stateIds)
}

func (o *ConnectivityTemplatesAssignment) GetIpLinkIds(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
	o.IpLinkIds = types.MapNull(types.MapType{ElemType: types.StringType})

	if !o.FetchIpLinkIds.ValueBool() {
		return
	}

	var connectivityTemplateIds []string
	diags.Append(o.ConnectivityTemplateIds.ElementsAs(ctx, &connectivityTemplateIds, false)...)
	if diags.HasError() {
		return
	}

	ctNodeIdAttr := apstra.QEEAttribute{Key: "id", Value: apstra.QEStringValIsIn(connectivityTemplateIds)}
	ctNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_ct")}
	iplpNodeTypeAttr := apstra.QEEAttribute{Key: "policy_type_name", Value: apstra.QEStringVal("AttachLogicalLink")}
	iplpNodeNameAttr := apstra.QEEAttribute{Key: "name", Value: apstra.QEStringVal("n_iplp")}
	apNodeIdAttr := apstra.QEEAttribute{Key: "id", Value: apstra.QEStringVal(o.ApplicationPointId.ValueString())}
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
			Ct struct {
				Id apstra.ObjectId `json:"id"`
			} `json:"n_ct"`
			Iplp struct {
				Attributes json.RawMessage `json:"attributes"`
			} `json:"n_iplp"`
			//Ap struct {
			//	Id apstra.ObjectId `json:"id"`
			//} `json:"n_ap"`
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

	addToResult := func(ctId apstra.ObjectId, vlan int, linkId apstra.ObjectId) {
		innerMap, ok := result[ctId]
		if !ok {
			innerMap = make(map[string]apstra.ObjectId)
		}
		innerMap[strconv.Itoa(vlan)] = linkId
		result[ctId] = innerMap
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
			addToResult(item.Ct.Id, 0, item.Ll.Id)
			continue
		case ipLinkAttributes.Vlan == nil || item.Si.Vlan == nil:
			// one item is untagged, but the other is not - not interesting
			continue
		case ipLinkAttributes.Vlan != nil && item.Si.Vlan != nil && *ipLinkAttributes.Vlan == *item.Si.Vlan:
			// IP link and subinterface are tagged - and they have matching values!
			addToResult(item.Ct.Id, *ipLinkAttributes.Vlan, item.Ll.Id)
			continue
		}
	}

	var d diag.Diagnostics
	o.IpLinkIds, d = types.MapValueFrom(ctx, types.MapType{ElemType: types.StringType}, result)
	diags.Append(d...)
}

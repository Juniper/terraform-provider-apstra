package blueprint

import (
	"context"
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
	//IpLinkInfos            types.Map    `tfsdk:"ip_link_infos"`
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
		//"ip_link_infos": resourceSchema.MapAttribute{
		//	MarkdownDescription: "This is a Map of Application Point ID to Map of VLAN to IP Info. Outer map keys must " +
		//		"be values which appear in the `application_point_ids` attribute. Inner map keys are VLAN numbers in " +
		//		"the range 0 - 4904 which are specified by IP Link primitives in the Connectivity Template. VLAN \"0\" " +
		//		"stands in for \"Untagged\" IP Link primitives.",
		//	ElementType: types.MapType{
		//		ElemType: types.ObjectType{
		//			AttrTypes: IpLinkIps{}.AttrTypes(),
		//		},
		//	},
		//	Optional: true,
		//},
		//"ip_link_infos": resourceSchema.MapNestedAttribute{
		//	MarkdownDescription: "This is a Map of Application Point ID to Map of VLAN to IP Info. Outer map keys must " +
		//		"be values which appear in the `application_point_ids` attribute. Inner map keys are VLAN numbers in " +
		//		"the range 0 - 4904 which are specified by IP Link primitives in the Connectivity Template. VLAN \"0\" " +
		//		"stands in for \"Untagged\" IP Link primitives.",
		//	NestedObject: resourceSchema.MapNestedAttribute{
		//	},
		//	Optional:     true,
		//},
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

//func (o *ConnectivityTemplateAssignments) SetSubinterfaces(ctx context.Context, oState *ConnectivityTemplateAssignments, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	if oState != nil && o.ApplicationPointIds.Equal(oState.ApplicationPointIds) {
//		return // no changes required
//	}
//
//	// extract the config into a two-dimensional ip_link_infos map
//	var config map[string]map[string]IpLinkIps
//	diags.Append(o.IpLinkInfos.ElementsAs(ctx, &config, false)...)
//	if diags.HasError() {
//		return
//	}
//
//	//// extract the state into a two-dimensional ip_link_infos map
//	//var state map[string]map[string]IpLinkIps
//	//if state != nil {
//	//	diags.Append(oState.IpLinkInfos.ElementsAs(ctx, &state, false)...)
//	//	if diags.HasError() {
//	//		return
//	//	}
//	//}
//
//	// extract apIds - these are the valid keys for the outer map
//	var apIds []apstra.ObjectId
//	diags.Append(o.ApplicationPointIds.ElementsAs(ctx, &apIds, false)...)
//	if diags.HasError() {
//		return
//	}
//
//	for _, ap := range apIds {
//		vlanToSubinterfaces := utils.GetCtIpLinkSubinterfaces(ctx, bp, apstra.ObjectId(o.ConnectivityTemplateId.ValueString()), ap, diags)
//		if diags.HasError() {
//			return
//		}
//
//		update := make(map[apstra.ObjectId]apstra.TwoStageL3ClosSubinterface)
//		for vlan, siId := range vlanToSubinterfaces {
//			ipInfo, ok := config[ap.String()][strconv.Itoa(int(vlan))]
//			if !ok {
//				continue // no configuration specified for this subinterface
//			}
//
//			update[siId] = apstra.TwoStageL3ClosSubinterface{}
//
//		}
//
//		err := bp.UpdateSubinterfaces(ctx, update)
//
//	}
//
//	//extractSubinterfaceIdsByVlan := func(assignments *ConnectivityTemplateAssignments) map[int64]IpLinkIps {
//	//	if assignments == nil {
//	//		return nil
//	//	}
//	//
//	//	var result map[int64]IpLinkIps
//	//
//	//	return nil
//	//}
//	//
//	//vlanToSubinterfaces := utils.GetCtIpLinkSubinterfaces(ctx, bp, "", "", diags)
//	//if diags.HasError() {
//	//	return
//	//}
//
//}

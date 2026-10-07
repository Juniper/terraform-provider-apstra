package blueprint

import (
	"context"
	"encoding/json"

	"github.com/Juniper/apstra-go-sdk/apstra"
	importvalidation "github.com/Juniper/terraform-provider-apstra/internal/import_validation"
	"github.com/Juniper/terraform-provider-apstra/internal/value"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualNetworkAssignmentIdentity struct {
	BlueprintID types.String `tfsdk:"blueprint_id"`
	VNID        types.String `tfsdk:"virtual_network_id"`
	LeafID      types.String `tfsdk:"leaf_switch_id"`
}

func (vnai VirtualNetworkAssignmentIdentity) GetState(ctx context.Context, bp *apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) VirtualNetworkAssignment {
	state := VirtualNetworkAssignment{
		BlueprintID: vnai.BlueprintID,
		VNID:        vnai.VNID,
		LeafID:      vnai.LeafID,
	}
	ok := state.Read(ctx, bp, diags)
	if diags.HasError() {
		return state
	}
	if !ok {
		diags.AddError("failed reading virtual network assignment", "unable to read virtual network assignment from Apstra")
	}
	return state
}

func (vnai VirtualNetworkAssignmentIdentity) Attributes() map[string]identityschema.Attribute {
	return map[string]identityschema.Attribute{
		"blueprint_id": identityschema.StringAttribute{
			RequiredForImport: true,
			Description:       "Apstra Blueprint ID.",
		},
		"virtual_network_id": identityschema.StringAttribute{
			RequiredForImport: true,
			Description:       "Virtual Network graph node ID.",
		},
		"leaf_switch_id": identityschema.StringAttribute{
			RequiredForImport: true,
			Description:       "Leaf Switch system graph node ID.",
		},
	}
}

func (vnai *VirtualNetworkAssignmentIdentity) ParseLegacyID(ctx context.Context, id string, diags *diag.Diagnostics) {
	var model struct {
		BlueprintID string `json:"blueprint_id"`
		VNID        string `json:"virtual_network_id"`
		LeafID      string `json:"leaf_switch_id"`
	}
	err := json.Unmarshal([]byte(id), &model)
	if err != nil {
		diags.AddError("failed parsing import ID JSON string", err.Error())
		return
	}

	vnai.BlueprintID = value.StringOrNull(ctx, model.BlueprintID, diags)
	vnai.VNID = value.StringOrNull(ctx, model.VNID, diags)
	vnai.LeafID = value.StringOrNull(ctx, model.LeafID, diags)

	importvalidation.CheckRequiredFields(ctx, vnai, diags)
}

package utils

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func AllNodeDeployModes() []string {
	modes := apstra.AllNodeDeployModes()
	result := make([]string, len(modes))
	for i := range modes {
		result[i] = StringersToFriendlyString(modes[i])
	}

	return result
}

func GetNodeDeployMode(ctx context.Context, client *apstra.TwoStageL3ClosClient, nodeId string, diags *diag.Diagnostics) string {
	var node struct {
		Id         string `json:"id"`
		Type       string `json:"type"`
		DeployMode string `json:"deploy_mode"`
	}
	err := client.Client().GetNode(ctx, client.Id(), apstra.ObjectId(nodeId), &node)
	if err != nil {
		diags.AddError("failed to fetch blueprint node", err.Error())
		return ""
	}

	var deployMode apstra.NodeDeployMode
	err = deployMode.FromString(node.DeployMode)
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing deploy mode %q", node.DeployMode), err.Error())
		return ""
	}

	return StringersToFriendlyString(deployMode)
}

func SetNodeDeployMode(ctx context.Context, client *apstra.TwoStageL3ClosClient, nodeId string, modeString string, diags *diag.Diagnostics) {
	var modeIota apstra.NodeDeployMode
	err := ApiStringerFromFriendlyString(&modeIota, modeString)
	if err != nil {
		diags.AddError(fmt.Sprintf("error parsing deploy mode %q", modeString), err.Error())
		return
	}

	type patch struct {
		Id         string  `json:"id"`
		DeployMode *string `json:"deploy_mode"`
	}

	var stringPtr *string
	if modeIota == apstra.NodeDeployModeNone {
		stringPtr = nil
	} else {
		s := modeIota.String()
		stringPtr = &s
	}

	setDeployMode := patch{
		Id:         nodeId,
		DeployMode: stringPtr,
	}

	err = client.Client().PatchNode(ctx, client.Id(), apstra.ObjectId(nodeId), &setDeployMode, nil)
	if err != nil {
		diags.AddError("error setting deploy mode", err.Error())
		return
	}
}

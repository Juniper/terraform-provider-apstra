package utils

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
)

func AllNodeDeployModes() []string {
	members := enum.DeployModes.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = rosetta.StringersToFriendlyString(member)
	}

	return result
}

func GetNodeDeployMode(ctx context.Context, client *apstra.TwoStageL3ClosClient, nodeId string) (string, error) {
	var node struct {
		Id         string `json:"id"`
		Type       string `json:"type"`
		DeployMode string `json:"deploy_mode"`
	}
	err := client.Client().GetNode(ctx, client.Id(), apstra.ObjectId(nodeId), &node)
	if err != nil {
		return "", err
	}

	var deployMode enum.DeployMode
	err = deployMode.FromString(node.DeployMode)
	if err != nil {
		return "", err
	}

	return rosetta.StringersToFriendlyString(deployMode), nil
}

func SetNodeDeployMode(ctx context.Context, client *apstra.TwoStageL3ClosClient, nodeId string, modeString string) error {
	var modeIota enum.DeployMode
	err := rosetta.ApiStringerFromFriendlyString(&modeIota, modeString)
	if err != nil {
		return err
	}

	patch := struct {
		Id         string  `json:"id"`
		DeployMode *string `json:"deploy_mode"`
	}{
		Id: nodeId,
	}

	if modeIota != enum.DeployModeNone {
		s := modeIota.String()
		patch.DeployMode = &s
	}

	err = client.Client().PatchNode(ctx, client.Id(), apstra.ObjectId(nodeId), &patch, nil)
	if err != nil {
		return err
	}

	return nil
}

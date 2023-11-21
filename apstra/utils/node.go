package utils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
)

func AllNodeDeployModes() []string {
	modes := apstra.AllNodeDeployModes()
	result := make([]string, len(modes))
	for i := range modes {
		result[i] = StringersToFriendlyString(modes[i])
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

	var deployMode apstra.NodeDeployMode
	err = deployMode.FromString(node.DeployMode)
	if err != nil {
		return "", err
	}

	return StringersToFriendlyString(deployMode), nil
}

func SetNodeDeployMode(ctx context.Context, client *apstra.TwoStageL3ClosClient, nodeId string, modeString string) error {
	var modeIota apstra.NodeDeployMode
	err := ApiStringerFromFriendlyString(&modeIota, modeString)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

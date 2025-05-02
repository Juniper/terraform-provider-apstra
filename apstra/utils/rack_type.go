package utils

import (
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
)

func FcdModes() []string {
	members := enum.FabricConnectivityDesigns.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = member.String()
	}
	return result
}

package utils

import "github.com/Juniper/apstra-go-sdk/apstra"

func AllNodeDeployModes() []string {
	modes := apstra.AllNodeDeployModes()
	result := make([]string, len(modes))
	for i, mode := range modes {
		result[i] = StringersToFriendlyString(mode)
	}

	return result
}

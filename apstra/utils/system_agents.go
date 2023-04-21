package utils

import "github.com/Juniper/apstra-go-sdk/apstra"

func AllNodeDeployModes() []string {
	modes := apstra.AllNodeDeployModes()
	result := make([]string, len(modes))
	for i := range modes {
		result[i] = StringersToFriendlyString(modes[i])
	}

	return result
}

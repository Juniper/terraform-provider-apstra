package utils

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
)

func AllValidPrefixFilterActions() []string {
	actions := apstra.AllPrefixFilterActions()
	result := make([]string, len(actions))
	var i int
	for _, action := range actions {
		if action != apstra.PrefixFilterActionNone {
			result[i] = action.String()
			i++
		}
	}

	return result[:i]
}

func AllDcRoutingPolicyImportPolicy() []string {
	policies := apstra.AllDcRoutingPolicyImportPolicies()
	result := make([]string, len(policies))
	for i := range policies {
		result[i] = policies[i].String()
	}

	// remove empty string if present
	for i := len(result) - 1; i >= 0; i-- {
		if result[i] == "" {
			result[i] = result[len(result)-1]
			result = result[:len(result)-1]
		}
	}

	return result
}

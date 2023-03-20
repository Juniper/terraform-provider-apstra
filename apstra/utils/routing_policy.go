package utils

import "bitbucket.org/apstrktr/goapstra"

func AllPrefixFilterActions() []string {
	actions := goapstra.AllPrefixFilterActions()
	result := make([]string, len(actions))
	for i := range actions {
		result[i] = actions[i].String()
	}
	return result
}

func AllDcRoutingPolicyImportPolicy() []string {
	policies := goapstra.AllDcRoutingPolicyImportPolicies()
	result := make([]string, len(policies))
	for i := range policies {
		result[i] = policies[i].String()
	}
	return result
}

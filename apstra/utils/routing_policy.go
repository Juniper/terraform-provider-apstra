package utils

import "bitbucket.org/apstrktr/goapstra"

func AllValidPrefixFilterActions() []string {
	actions := goapstra.AllPrefixFilterActions()
	result := make([]string, len(actions))
	var i int
	for _, action := range actions {
		if action != goapstra.PrefixFilterActionNone {
			result[i] = action.String()
			i++
		}
	}

	return result[:i]
}

func AllDcRoutingPolicyImportPolicy() []string {
	policies := goapstra.AllDcRoutingPolicyImportPolicies()
	result := make([]string, len(policies))
	for i := range policies {
		result[i] = policies[i].String()
	}
	return result
}

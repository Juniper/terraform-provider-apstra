package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra"
)

func AllFFResourceTypes() []string {
	members := apstra.FFResourceTypes.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}

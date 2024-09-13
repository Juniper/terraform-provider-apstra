package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra/enum"
)

func AllFFResourceTypes() []string {
	members := enum.FFResourceTypes.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}

package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/enum"
)

func AllResourcePoolTypes() []string {
	members := enum.ResourcePoolTypes.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}

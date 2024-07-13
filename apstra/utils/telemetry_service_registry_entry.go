package utils

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"sort"
)

func AllStorageSchemaPaths() []string {
	members := apstra.StorageSchemaPaths.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}

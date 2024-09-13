package utils

import (
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"sort"
)

func AllStorageSchemaPaths() []string {
	members := enum.StorageSchemaPaths.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}

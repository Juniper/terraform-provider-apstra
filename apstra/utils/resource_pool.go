package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
)

func AllResourcePoolTypes() []string {
	members := enum.ResourcePoolTypes.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = rosetta.StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}
